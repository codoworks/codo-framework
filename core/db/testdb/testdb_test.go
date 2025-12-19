package testdb

import (
	"context"
	"testing"

	"github.com/codoworks/codo-framework/core/db"
	"github.com/jmoiron/sqlx"
)

func TestNew(t *testing.T) {
	client := New(t)
	if client == nil {
		t.Fatal("New should return a client")
	}
	if !client.IsInitialized() {
		t.Error("Client should be initialized")
	}
}

func TestNewWithSchema(t *testing.T) {
	schema := "CREATE TABLE test_schema (id INTEGER PRIMARY KEY)"
	client := NewWithSchema(t, schema)

	// Verify table exists
	var count int
	err := client.GetContext(context.Background(), &count, "SELECT COUNT(*) FROM test_schema")
	if err != nil {
		t.Fatalf("Table should exist: %v", err)
	}
}

func TestNewWithSchemas(t *testing.T) {
	schemas := []string{
		"CREATE TABLE t1 (id INTEGER PRIMARY KEY)",
		"CREATE TABLE t2 (id INTEGER PRIMARY KEY)",
	}
	client := NewWithSchemas(t, schemas...)

	// Verify tables exist
	var count int
	client.GetContext(context.Background(), &count, "SELECT COUNT(*) FROM t1")
	client.GetContext(context.Background(), &count, "SELECT COUNT(*) FROM t2")
}

func TestTruncate(t *testing.T) {
	client := NewWithSchema(t, "CREATE TABLE truncate_test (id INTEGER PRIMARY KEY)")

	// Insert data
	client.ExecContext(context.Background(), "INSERT INTO truncate_test (id) VALUES (1), (2), (3)")

	// Truncate
	Truncate(t, client, "truncate_test")

	// Verify empty
	count := Count(t, client, "truncate_test")
	if count != 0 {
		t.Errorf("Count after truncate = %d, want 0", count)
	}
}

func TestExec(t *testing.T) {
	client := NewWithSchema(t, "CREATE TABLE exec_test (id INTEGER PRIMARY KEY, val TEXT)")

	Exec(t, client, "INSERT INTO exec_test (id, val) VALUES (?, ?)", 1, "test")

	count := Count(t, client, "exec_test")
	if count != 1 {
		t.Errorf("Count = %d, want 1", count)
	}
}

func TestCount(t *testing.T) {
	client := NewWithSchema(t, "CREATE TABLE count_test (id INTEGER PRIMARY KEY)")

	// Empty table
	count := Count(t, client, "count_test")
	if count != 0 {
		t.Errorf("Empty count = %d, want 0", count)
	}

	// With data
	Exec(t, client, "INSERT INTO count_test (id) VALUES (1), (2)")
	count = Count(t, client, "count_test")
	if count != 2 {
		t.Errorf("Count = %d, want 2", count)
	}
}

func TestMustCreate(t *testing.T) {
	cfg := &db.ClientConfig{
		Driver:       "sqlite3",
		DSN:          ":memory:",
		MaxOpenConns: 1,
	}

	client := MustCreate(t, cfg)
	if client == nil {
		t.Fatal("MustCreate should return a client")
	}
}

func TestTransaction(t *testing.T) {
	client := NewWithSchema(t, "CREATE TABLE tx_test (id INTEGER PRIMARY KEY)")

	Transaction(t, client, func(tx *sqlx.Tx) {
		// Execute something in the transaction
		tx.Exec("INSERT INTO tx_test (id) VALUES (1)")
	})

	// Data should be rolled back
	count := Count(t, client, "tx_test")
	if count != 0 {
		t.Errorf("Transaction should rollback, count = %d", count)
	}
}

func TestTxExec(t *testing.T) {
	client := NewWithSchema(t, "CREATE TABLE tx_exec_test (id INTEGER PRIMARY KEY)")

	tx, err := client.BeginTx(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	result := TxExec(t, tx, "INSERT INTO tx_exec_test (id) VALUES (?)", 1)
	affected, _ := result.RowsAffected()
	if affected != 1 {
		t.Errorf("RowsAffected = %d, want 1", affected)
	}
}
