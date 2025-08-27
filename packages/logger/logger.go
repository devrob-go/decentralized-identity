package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const (
	CorrelationIDKey  = "correlation_id"
	DefaultLevel      = "info"
	DefaultTimeFormat = time.RFC3339
)

// Config holds logger configuration
type Config struct {
	Level      string    // Log level: debug, info, warn, error, fatal, panic
	Output     io.Writer // Output destination (default: os.Stdout)
	JSONFormat bool      // Use JSON format if true, else console
	AddCaller  bool      // Include caller info (method name for errors)
	TimeFormat string    // Timestamp format
	Service    string    // Service name for structured logging
	Version    string    // Service version for structured logging
}

// Logger wraps zerolog.Logger with additional functionality
type Logger struct {
	*zerolog.Logger
	config Config
}

// singleton instance
var (
	instance *Logger
	once     sync.Once
)

// ctxKey is used for context-based correlation IDs
type ctxKey string

const correlationIDCtxKey ctxKey = "correlation_id"

// NewLogger initializes and returns the singleton logger
func NewLogger(config Config) *Logger {
	once.Do(func() {
		// Set defaults
		if config.Output == nil {
			config.Output = os.Stdout
		}
		if config.TimeFormat == "" {
			config.TimeFormat = DefaultTimeFormat
		}
		if config.Level == "" {
			config.Level = DefaultLevel
		}

		// Configure zerolog
		zerolog.TimeFieldFormat = config.TimeFormat

		var logger zerolog.Logger
		if config.JSONFormat {
			logger = zerolog.New(config.Output).With().Timestamp().Logger()
		} else {
			logger = zerolog.New(zerolog.ConsoleWriter{
				Out:        config.Output,
				TimeFormat: config.TimeFormat,
				FormatLevel: func(i any) string {
					if ll, ok := i.(string); ok {
						switch ll {
						case "debug":
							return "\x1b[36mDBG\x1b[0m"
						case "info":
							return "\x1b[32mINF\x1b[0m"
						case "warn":
							return "\x1b[33mWRN\x1b[0m"
						case "error":
							return "\x1b[31mERR\x1b[0m"
						case "fatal":
							return "\x1b[35mFTL\x1b[0m"
						case "panic":
							return "\x1b[35mPNC\x1b[0m"
						default:
							return ll
						}
					}
					return "???"
				},
			}).With().Timestamp().Logger()
		}

		// Set log level
		level, err := zerolog.ParseLevel(config.Level)
		if err != nil {
			level = zerolog.InfoLevel
		}
		logger = logger.Level(level)

		// Add service context if provided
		if config.Service != "" {
			logger = logger.With().Str("service", config.Service).Logger()
		}
		if config.Version != "" {
			logger = logger.With().Str("version", config.Version).Logger()
		}

		instance = &Logger{
			Logger: &logger,
			config: config,
		}
	})

	return instance
}

// WithCorrelationID adds a correlation ID to the context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	if correlationID == "" {
		correlationID = uuid.New().String()
	}
	return context.WithValue(ctx, correlationIDCtxKey, correlationID)
}

// getCorrelationID retrieves correlation ID from context
func getCorrelationID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if correlationID, ok := ctx.Value(correlationIDCtxKey).(string); ok {
		return correlationID
	}
	return ""
}

// Info logs an info message with optional fields
func (l *Logger) Info(ctx context.Context, message string, fields ...map[string]any) {
	event := l.Logger.Info()

	// Add correlation ID if available
	if correlationID := getCorrelationID(ctx); correlationID != "" {
		event = event.Str(CorrelationIDKey, correlationID)
	}

	// Add caller info if enabled
	if l.config.AddCaller {
		if _, file, line, ok := runtime.Caller(1); ok {
			event = event.Str("caller", fmt.Sprintf("%s:%d", filepath.Base(file), line))
		}
	}

	// Add additional fields
	if len(fields) > 0 {
		for key, value := range fields[0] {
			event = event.Interface(key, value)
		}
	}

	event.Msg(message)
}

// Error logs an error message with optional fields and status code
func (l *Logger) Error(ctx context.Context, err error, message string, statusCode int, fields ...map[string]any) {
	event := l.Logger.Error().Err(err)

	// Add correlation ID if available
	if correlationID := getCorrelationID(ctx); correlationID != "" {
		event = event.Str(CorrelationIDKey, correlationID)
	}

	// Add status code if provided
	if statusCode > 0 {
		event = event.Int("status_code", statusCode)
	}

	// Add caller info if enabled
	if l.config.AddCaller {
		if _, file, line, ok := runtime.Caller(1); ok {
			event = event.Str("caller", fmt.Sprintf("%s:%d", filepath.Base(file), line))
		}
	}

	// Add additional fields
	if len(fields) > 0 {
		for key, value := range fields[0] {
			event = event.Interface(key, value)
		}
	}

	event.Msg(message)
}

// Debug logs a debug message with optional fields
func (l *Logger) Debug(ctx context.Context, message string, fields ...map[string]any) {
	event := l.Logger.Debug()

	// Add correlation ID if available
	if correlationID := getCorrelationID(ctx); correlationID != "" {
		event = event.Str(CorrelationIDKey, correlationID)
	}

	// Add caller info if enabled
	if l.config.AddCaller {
		if _, file, line, ok := runtime.Caller(1); ok {
			event = event.Str("caller", fmt.Sprintf("%s:%d", filepath.Base(file), line))
		}
	}

	// Add additional fields
	if len(fields) > 0 {
		for key, value := range fields[0] {
			event = event.Interface(key, value)
		}
	}

	event.Msg(message)
}

// Warn logs a warning message with optional fields
func (l *Logger) Warn(ctx context.Context, message string, fields ...map[string]any) {
	event := l.Logger.Warn()

	// Add correlation ID if available
	if correlationID := getCorrelationID(ctx); correlationID != "" {
		event = event.Str(CorrelationIDKey, correlationID)
	}

	// Add caller info if enabled
	if l.config.AddCaller {
		if _, file, line, ok := runtime.Caller(1); ok {
			event = event.Str("caller", fmt.Sprintf("%s:%d", filepath.Base(file), line))
		}
	}

	// Add additional fields
	if len(fields) > 0 {
		for key, value := range fields[0] {
			event = event.Interface(key, value)
		}
	}

	event.Msg(message)
}

// Fatal logs a fatal message and exits the program
func (l *Logger) Fatal(ctx context.Context, err error, message string, fields ...map[string]any) {
	event := l.Logger.Fatal().Err(err)

	// Add correlation ID if available
	if correlationID := getCorrelationID(ctx); correlationID != "" {
		event = event.Str(CorrelationIDKey, correlationID)
	}

	// Add caller info if enabled
	if l.config.AddCaller {
		if _, file, line, ok := runtime.Caller(1); ok {
			event = event.Str("caller", fmt.Sprintf("%s:%d", filepath.Base(file), line))
		}
	}

	// Add additional fields
	if len(fields) > 0 {
		for key, value := range fields[0] {
			event = event.Interface(key, value)
		}
	}

	event.Msg(message)
}

// WithFields creates a new logger with additional fields
func (l *Logger) WithFields(fields map[string]any) *Logger {
	newLogger := l.Logger.With()
	for key, value := range fields {
		newLogger = newLogger.Interface(key, value)
	}

	logger := newLogger.Logger()
	return &Logger{
		Logger: &logger,
		config: l.config,
	}
}

// SetLevel changes the log level dynamically
func (l *Logger) SetLevel(level string) {
	if parsedLevel, err := zerolog.ParseLevel(level); err == nil {
		newLogger := l.Logger.Level(parsedLevel)
		l.Logger = &newLogger
	}
}

// GetLevel returns the current log level
func (l *Logger) GetLevel() string {
	return l.Logger.GetLevel().String()
}
