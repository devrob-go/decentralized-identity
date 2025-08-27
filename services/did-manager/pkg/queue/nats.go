package queue

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

// NATSQueue handles message queuing using NATS
type NATSQueue struct {
	conn *nats.Conn
	js   nats.JetStreamContext
}

// NewNATSQueue creates a new NATS queue instance
func NewNATSQueue(natsURL string) (*NATSQueue, error) {
	// Connect to NATS
	conn, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Create JetStream context
	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	// Create stream for blockchain jobs
	stream, err := js.AddStream(&nats.StreamConfig{
		Name:      "BLOCKCHAIN_JOBS",
		Subjects:  []string{"blockchain.jobs.*"},
		Storage:   nats.FileStorage,
		Retention: nats.LimitsPolicy,
		MaxAge:    24 * time.Hour, // Keep messages for 24 hours
		MaxMsgs:   10000,          // Max 10k messages
	})

	if err != nil && err.Error() != "stream name already in use" {
		log.Printf("Warning: failed to create stream: %v", err)
	} else if err == nil {
		log.Printf("Created stream: %s", stream.Config.Name)
	}

	// Create consumer for processing jobs
	_, err = js.AddConsumer("BLOCKCHAIN_JOBS", &nats.ConsumerConfig{
		Durable:       "did-manager-worker",
		FilterSubject: "blockchain.jobs.register_did",
		AckPolicy:     nats.AckExplicitPolicy,
		MaxAckPending: 100,
		MaxDeliver:    3, // Retry failed jobs up to 3 times
	})

	if err != nil && err.Error() != "consumer name already in use" {
		log.Printf("Warning: failed to create consumer: %v", err)
	}

	return &NATSQueue{
		conn: conn,
		js:   js,
	}, nil
}

// BlockchainJob represents a job to be processed on the blockchain
type BlockchainJob struct {
	ID        string    `json:"id"`
	JobType   string    `json:"job_type"`
	DIDID     string    `json:"did_id"`
	UserHash  string    `json:"user_hash"`
	DID       string    `json:"did"`
	CreatedAt time.Time `json:"created_at"`
}

// PublishJob publishes a blockchain job to the queue
func (n *NATSQueue) PublishJob(job *BlockchainJob) error {
	subject := fmt.Sprintf("blockchain.jobs.%s", job.JobType)

	// Publish with JetStream for persistence
	ack, err := n.js.Publish(subject, job.toJSON())
	if err != nil {
		return fmt.Errorf("failed to publish job: %w", err)
	}

	log.Printf("Published job %s to subject %s, stream sequence: %d",
		job.ID, subject, ack.Sequence)

	return nil
}

// SubscribeToJobs subscribes to blockchain jobs for processing
func (n *NATSQueue) SubscribeToJobs(jobType string, handler func(*BlockchainJob) error) error {
	subject := fmt.Sprintf("blockchain.jobs.%s", jobType)

	// Subscribe with JetStream for reliable delivery
	_, err := n.js.Subscribe(subject, func(msg *nats.Msg) {
		var job BlockchainJob
		if err := json.Unmarshal(msg.Data, &job); err != nil {
			log.Printf("Failed to unmarshal job: %v", err)
			msg.Nak() // Negative acknowledgment - will retry
			return
		}

		log.Printf("Processing job %s of type %s", job.ID, job.JobType)

		// Process the job
		if err := handler(&job); err != nil {
			log.Printf("Failed to process job %s: %v", job.ID, err)
			msg.Nak() // Negative acknowledgment - will retry
			return
		}

		// Acknowledge successful processing
		msg.Ack()
		log.Printf("Successfully processed job %s", job.ID)
	}, nats.Durable("did-manager-worker"), nats.AckExplicit())

	if err != nil {
		return fmt.Errorf("failed to subscribe to jobs: %w", err)
	}

	log.Printf("Subscribed to %s with durable consumer", subject)

	return nil
}

// Close closes the NATS connection
func (n *NATSQueue) Close() {
	if n.conn != nil {
		n.conn.Close()
	}
}

// toJSON converts a BlockchainJob to JSON bytes
func (j *BlockchainJob) toJSON() []byte {
	data, err := json.Marshal(j)
	if err != nil {
		log.Printf("Failed to marshal job: %v", err)
		return nil
	}
	return data
}

// GetPendingJobs retrieves pending jobs count from the stream
func (n *NATSQueue) GetPendingJobs(jobType string, limit int) (int, error) {
	// Get stream info to check pending messages
	streamInfo, err := n.js.StreamInfo("BLOCKCHAIN_JOBS")
	if err != nil {
		return 0, fmt.Errorf("failed to get stream info: %w", err)
	}

	// For simplicity, return the number of messages in the stream
	pending := int(streamInfo.State.Msgs)
	if pending > limit {
		pending = limit
	}

	return pending, nil
}
