package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// ClientConfig holds database client configuration
type ClientConfig struct {
	Driver          string
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultClientConfig returns a ClientConfig with sensible defaults
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}
}

// Client is the database client
type Client struct {
	db     *sqlx.DB
	config *ClientConfig
}

// NewClient creates a new database client
func NewClient(cfg *ClientConfig) *Client {
	if cfg == nil {
		cfg = DefaultClientConfig()
	}
	return &Client{config: cfg}
}

// Name returns the client name
func (c *Client) Name() string {
	return "db"
}

// Initialize connects to the database
func (c *Client) Initialize(cfg any) error {
	if clientCfg, ok := cfg.(*ClientConfig); ok && clientCfg != nil {
		c.config = clientCfg
	}

	if c.config == nil {
		return fmt.Errorf("database configuration is required")
	}

	if c.config.Driver == "" {
		return fmt.Errorf("database driver is required")
	}

	if c.config.DSN == "" {
		return fmt.Errorf("database DSN is required")
	}

	// Normalize driver name: "sqlite" -> "sqlite3"
	driver := c.config.Driver
	if driver == "sqlite" {
		driver = "sqlite3"
	}

	db, err := sqlx.Connect(driver, c.config.DSN)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	if c.config.MaxOpenConns > 0 {
		db.SetMaxOpenConns(c.config.MaxOpenConns)
	}
	if c.config.MaxIdleConns > 0 {
		db.SetMaxIdleConns(c.config.MaxIdleConns)
	}
	if c.config.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(c.config.ConnMaxLifetime)
	}
	if c.config.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(c.config.ConnMaxIdleTime)
	}

	c.db = db
	return nil
}

// Health checks database connectivity
func (c *Client) Health() error {
	if c.db == nil {
		return fmt.Errorf("database not initialized")
	}
	return c.db.Ping()
}

// HealthContext checks database connectivity with context
func (c *Client) HealthContext(ctx context.Context) error {
	if c.db == nil {
		return fmt.Errorf("database not initialized")
	}
	return c.db.PingContext(ctx)
}

// Shutdown closes the database connection
func (c *Client) Shutdown() error {
	if c.db == nil {
		return nil
	}
	return c.db.Close()
}

// DB returns the underlying sqlx.DB
func (c *Client) DB() *sqlx.DB {
	return c.db
}

// Config returns the client configuration
func (c *Client) Config() *ClientConfig {
	return c.config
}

// BeginTx starts a transaction
func (c *Client) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return c.db.BeginTxx(ctx, nil)
}

// BeginTxWithOptions starts a transaction with options
func (c *Client) BeginTxWithOptions(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return c.db.BeginTxx(ctx, opts)
}

// ExecContext executes a query without returning rows
func (c *Client) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return c.db.ExecContext(ctx, query, args...)
}

// QueryContext executes a query that returns rows
func (c *Client) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return c.db.QueryContext(ctx, query, args...)
}

// QueryRowContext executes a query expected to return at most one row
func (c *Client) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return c.db.QueryRowContext(ctx, query, args...)
}

// GetContext queries a single row into dest
func (c *Client) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	if c.db == nil {
		return fmt.Errorf("database not initialized")
	}
	return c.db.GetContext(ctx, dest, query, args...)
}

// SelectContext queries multiple rows into dest slice
func (c *Client) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	if c.db == nil {
		return fmt.Errorf("database not initialized")
	}
	return c.db.SelectContext(ctx, dest, query, args...)
}

// NamedExecContext executes a named query
func (c *Client) NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return c.db.NamedExecContext(ctx, query, arg)
}

// Rebind transforms a query from ? placeholders to the driver-specific placeholder
func (c *Client) Rebind(query string) string {
	if c.db == nil {
		return query
	}
	return c.db.Rebind(query)
}

// IsInitialized returns true if the client is initialized
func (c *Client) IsInitialized() bool {
	return c.db != nil
}
