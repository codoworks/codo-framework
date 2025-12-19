package testdb

import (
	"context"
	"database/sql"
	"testing"

	"github.com/codoworks/codo-framework/core/db"
	"github.com/jmoiron/sqlx"
)

// New creates an in-memory SQLite database for testing
func New(t *testing.T) *db.Client {
	t.Helper()

	client := db.NewClient(&db.ClientConfig{
		Driver:       "sqlite3",
		DSN:          ":memory:",
		MaxOpenConns: 1,
		MaxIdleConns: 1,
	})

	if err := client.Initialize(nil); err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	t.Cleanup(func() {
		client.Shutdown()
	})

	return client
}

// NewWithSchema creates a test database and runs schema SQL
func NewWithSchema(t *testing.T, schema string) *db.Client {
	t.Helper()

	client := New(t)

	if _, err := client.DB().Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return client
}

// NewWithSchemas creates a test database and runs multiple schema SQL statements
func NewWithSchemas(t *testing.T, schemas ...string) *db.Client {
	t.Helper()

	client := New(t)

	for _, schema := range schemas {
		if _, err := client.DB().Exec(schema); err != nil {
			t.Fatalf("failed to create schema: %v", err)
		}
	}

	return client
}

// Truncate clears all data from the given tables
func Truncate(t *testing.T, client *db.Client, tables ...string) {
	t.Helper()

	ctx := context.Background()
	for _, table := range tables {
		if _, err := client.ExecContext(ctx, "DELETE FROM "+table); err != nil {
			t.Fatalf("failed to truncate table %s: %v", table, err)
		}
	}
}

// Exec executes SQL on the test database
func Exec(t *testing.T, client *db.Client, query string, args ...any) {
	t.Helper()

	if _, err := client.ExecContext(context.Background(), query, args...); err != nil {
		t.Fatalf("failed to execute query: %v", err)
	}
}

// Count returns the count of rows in a table
func Count(t *testing.T, client *db.Client, table string) int64 {
	t.Helper()

	var count int64
	err := client.GetContext(context.Background(), &count, "SELECT COUNT(*) FROM "+table)
	if err != nil {
		t.Fatalf("failed to count rows in %s: %v", table, err)
	}
	return count
}

// MustCreate creates a client or fails the test
func MustCreate(t *testing.T, cfg *db.ClientConfig) *db.Client {
	t.Helper()

	client := db.NewClient(cfg)
	if err := client.Initialize(nil); err != nil {
		t.Fatalf("failed to create database client: %v", err)
	}

	t.Cleanup(func() {
		client.Shutdown()
	})

	return client
}

// Transaction runs a function in a transaction that is always rolled back
// Useful for testing database operations without persisting changes
func Transaction(t *testing.T, client *db.Client, fn func(tx *sqlx.Tx)) {
	t.Helper()

	tx, err := client.BeginTx(context.Background())
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	defer func() {
		if err := tx.Rollback(); err != nil {
			t.Logf("rollback warning: %v", err)
		}
	}()

	fn(tx)
}

// TxExec executes a query in a transaction, failing the test on error
func TxExec(t *testing.T, tx *sqlx.Tx, query string, args ...any) sql.Result {
	t.Helper()
	result, err := tx.Exec(query, args...)
	if err != nil {
		t.Fatalf("tx exec failed: %v", err)
	}
	return result
}
