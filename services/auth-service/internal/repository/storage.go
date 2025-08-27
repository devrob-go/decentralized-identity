package repository

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"auth-service/config"

	zlog "packages/logger"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pressly/goose"
)

const (
	APP_LAYER = "storage"
)

// Postgres error codes for specific constraint violations
var (
	ErrUniqueViolation     = fmt.Errorf("unique constraint violation")
	ErrForeignKeyViolation = fmt.Errorf("foreign key violation")
	ErrNotNullViolation    = fmt.Errorf("not-null violation")
	ErrCheckViolation      = fmt.Errorf("check constraint violation")
	ErrExclusionViolation  = fmt.Errorf("exclusion constraint violation")
)

// DB wraps sqlx.DB with additional functionality
type DB struct {
	*sqlx.DB
	logger *zlog.Logger
}

// Config holds database configuration
type Config struct {
	ConnStr         string
	MigrationsDir   string
	ConnTimeout     time.Duration
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type NamedPreparer interface {
	PrepareNamedContext(ctx context.Context, query string) (*sqlx.NamedStmt, error)
}

// NewDB initializes a new database connection with the provided configuration
func NewDB(ctx context.Context, cfg *Config, logger *zlog.Logger) (*DB, error) {
	if cfg == nil || logger == nil {
		return nil, fmt.Errorf("config and logger must not be nil")
	}

	dbx, err := sqlx.Open("postgres", cfg.ConnStr)
	if err != nil {
		logger.Error(ctx, err, "Failed to open database connection", http.StatusInternalServerError)
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	dbx.SetMaxOpenConns(cfg.MaxOpenConns)
	dbx.SetMaxIdleConns(cfg.MaxIdleConns)
	dbx.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	ctxTimeout, cancel := context.WithTimeout(ctx, cfg.ConnTimeout)
	defer cancel()

	// Ping the database to verify connection
	if err := dbx.PingContext(ctxTimeout); err != nil {
		logger.Error(ctx, err, "Failed to ping database", http.StatusInternalServerError)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations
	if err := runMigrations(ctx, dbx, cfg.MigrationsDir, logger); err != nil {
		logger.Error(ctx, err, "Database migrations failed", http.StatusInternalServerError)
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info(ctx, "Database connection established and migrations applied successfully")
	return &DB{DB: dbx, logger: func() *zlog.Logger {
		return logger.WithFields(map[string]any{
			"layer": APP_LAYER,
		})
	}()}, nil
}

// InitDB initializes the database using the application config
func InitDB(ctx context.Context, appCfg *config.Config, logger *zlog.Logger) (*DB, error) {
	cfg := FromConfig(appCfg)
	return NewDB(ctx, cfg, logger)
}

const (
	DefaultMigrationsDir   = "./storage/migrations"
	DefaultConnTimeout     = 5 * time.Second
	DefaultMaxOpenConns    = 25
	DefaultMaxIdleConns    = 10
	DefaultConnMaxLifetime = 30 * time.Minute
)

// FromConfig creates a storage Config from the application config
func FromConfig(appCfg *config.Config) *Config {
	return &Config{
		ConnStr: fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			appCfg.PostgresUser,
			appCfg.PostgresPassword,
			appCfg.PostgresHost,
			appCfg.PostgresPort,
			appCfg.PostgresDB,
		),
		MigrationsDir:   DefaultMigrationsDir,
		ConnTimeout:     DefaultConnTimeout,
		MaxOpenConns:    DefaultMaxOpenConns,
		MaxIdleConns:    DefaultMaxIdleConns,
		ConnMaxLifetime: DefaultConnMaxLifetime,
	}
}

// runMigrations applies database migrations from the specified directory
func runMigrations(ctx context.Context, db *sqlx.DB, migrationsDir string, logger *zlog.Logger) error {
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		logger.Error(ctx, err, "Migrations directory not found", http.StatusInternalServerError, map[string]any{"path": migrationsDir})
		return fmt.Errorf("migrations directory does not exist: %w", err)
	}

	logger.Info(ctx, "Applying migrations", map[string]any{"migrations_dir": migrationsDir})

	if err := goose.SetDialect("postgres"); err != nil {
		logger.Error(ctx, err, "Failed to set goose dialect", http.StatusInternalServerError)
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db.DB, migrationsDir); err != nil {
		logger.Error(ctx, err, "Failed to apply migrations", http.StatusInternalServerError)
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	logger.Info(ctx, "Migrations applied successfully")
	return nil
}

// defaultUniqueMessage generates a user-friendly message for unique constraint violations
func defaultUniqueMessage(constraint string) string {
	parts := strings.Split(constraint, "_")
	if len(parts) >= 3 {
		return fmt.Sprintf("%s must be unique", parts[len(parts)-2])
	}
	return "value must be unique"
}

// HandlePgError maps PostgreSQL errors to HTTP status codes and custom errors
func HandlePgError(err error) (int, error) {
	var pgErr *pq.Error
	if !errors.As(err, &pgErr) {
		return http.StatusInternalServerError, fmt.Errorf("database error: %w", err)
	}

	// Use a map for better performance and maintainability
	errorCodeMap := map[string]struct {
		status int
		err    error
	}{
		"unique_violation":     {http.StatusConflict, ErrUniqueViolation},
		"foreign_key_violation": {http.StatusBadRequest, ErrForeignKeyViolation},
		"not_null_violation":   {http.StatusBadRequest, ErrNotNullViolation},
		"check_violation":      {http.StatusBadRequest, ErrCheckViolation},
		"exclusion_violation":  {http.StatusBadRequest, ErrExclusionViolation},
	}

	if errorInfo, exists := errorCodeMap[pgErr.Code.Name()]; exists {
		if pgErr.Code.Name() == "unique_violation" {
			return errorInfo.status, fmt.Errorf("%w: %s", errorInfo.err, defaultUniqueMessage(pgErr.Constraint))
		}
		return errorInfo.status, errorInfo.err
	}

	return http.StatusInternalServerError, fmt.Errorf("unhandled database error: %w", err)
}

// Close gracefully closes the database connection
func (db *DB) Close(ctx context.Context) error {
	if err := db.DB.Close(); err != nil {
		db.logger.Error(ctx, err, "Failed to close database connection", http.StatusInternalServerError)
		return fmt.Errorf("failed to close database: %w", err)
	}
	db.logger.Info(ctx, "Database connection closed successfully")
	return nil
}
