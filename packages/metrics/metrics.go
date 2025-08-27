package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTPRequestsTotal counts total HTTP requests by method and status
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "status", "service"},
	)

	// HTTPRequestDuration tracks HTTP request duration
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "status", "service"},
	)

	// gRPCRequestsTotal counts total gRPC requests by method and status
	GRPCRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "status", "service"},
	)

	// gRPCRequestDuration tracks gRPC request duration
	GRPCRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "status", "service"},
	)

	// DatabaseConnections tracks active database connections
	DatabaseConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "database_connections_active",
			Help: "Number of active database connections",
		},
		[]string{"service"},
	)

	// DatabaseQueryDuration tracks database query duration
	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "operation"},
	)
)

// Metrics provides methods to record various metrics
type Metrics struct {
	serviceName string
}

// NewMetrics creates a new metrics instance
func NewMetrics(serviceName string) *Metrics {
	return &Metrics{
		serviceName: serviceName,
	}
}

// RecordHTTPRequest records an HTTP request
func (m *Metrics) RecordHTTPRequest(method, status string, duration float64) {
	HTTPRequestsTotal.WithLabelValues(method, status, m.serviceName).Inc()
	HTTPRequestDuration.WithLabelValues(method, status, m.serviceName).Observe(duration)
}

// RecordGRPCRequest records a gRPC request
func (m *Metrics) RecordGRPCRequest(method, status string, duration float64) {
	GRPCRequestsTotal.WithLabelValues(method, status, m.serviceName).Inc()
	GRPCRequestDuration.WithLabelValues(method, status, m.serviceName).Observe(duration)
}

// SetDatabaseConnections sets the number of active database connections
func (m *Metrics) SetDatabaseConnections(count float64) {
	DatabaseConnections.WithLabelValues(m.serviceName).Set(count)
}

// RecordDatabaseQuery records a database query
func (m *Metrics) RecordDatabaseQuery(operation string, duration float64) {
	DatabaseQueryDuration.WithLabelValues(m.serviceName, operation).Observe(duration)
}
