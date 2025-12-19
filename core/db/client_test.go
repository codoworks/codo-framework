package db

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestDefaultClientConfig(t *testing.T) {
	cfg := DefaultClientConfig()

	if cfg.MaxOpenConns != 25 {
		t.Errorf("MaxOpenConns = %d, want 25", cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns != 5 {
		t.Errorf("MaxIdleConns = %d, want 5", cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime != 5*time.Minute {
		t.Errorf("ConnMaxLifetime = %v, want 5m", cfg.ConnMaxLifetime)
	}
	if cfg.ConnMaxIdleTime != 5*time.Minute {
		t.Errorf("ConnMaxIdleTime = %v, want 5m", cfg.ConnMaxIdleTime)
	}
}

func TestNewClient(t *testing.T) {
	t.Run("with config", func(t *testing.T) {
		cfg := &ClientConfig{Driver: "sqlite3", DSN: ":memory:"}
		client := NewClient(cfg)

		if client == nil {
			t.Fatal("NewClient returned nil")
		}
		if client.config != cfg {
			t.Error("NewClient did not use provided config")
		}
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		client := NewClient(nil)

		if client == nil {
			t.Fatal("NewClient returned nil")
		}
		if client.config == nil {
			t.Error("NewClient should set default config")
		}
	})
}

func TestClient_Name(t *testing.T) {
	client := NewClient(nil)
	if name := client.Name(); name != "db" {
		t.Errorf("Name() = %s, want db", name)
	}
}

func TestClient_Initialize(t *testing.T) {
	t.Run("successful connection", func(t *testing.T) {
		cfg := &ClientConfig{
			Driver:       "sqlite3",
			DSN:          ":memory:",
			MaxOpenConns: 1,
			MaxIdleConns: 1,
		}
		client := NewClient(cfg)

		err := client.Initialize(nil)
		if err != nil {
			t.Fatalf("Initialize failed: %v", err)
		}
		defer client.Shutdown()

		if client.db == nil {
			t.Error("Initialize should set db")
		}
	})

	t.Run("with config parameter", func(t *testing.T) {
		client := NewClient(nil)

		cfg := &ClientConfig{
			Driver: "sqlite3",
			DSN:    ":memory:",
		}
		err := client.Initialize(cfg)
		if err != nil {
			t.Fatalf("Initialize failed: %v", err)
		}
		defer client.Shutdown()

		if client.config != cfg {
			t.Error("Initialize should use config from parameter")
		}
	})

	t.Run("with connection pool settings", func(t *testing.T) {
		cfg := &ClientConfig{
			Driver:          "sqlite3",
			DSN:             ":memory:",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 10 * time.Minute,
			ConnMaxIdleTime: 5 * time.Minute,
		}
		client := NewClient(cfg)

		err := client.Initialize(nil)
		if err != nil {
			t.Fatalf("Initialize failed: %v", err)
		}
		defer client.Shutdown()
	})
}

func TestClient_Initialize_Errors(t *testing.T) {
	t.Run("nil config", func(t *testing.T) {
		client := &Client{}

		err := client.Initialize(nil)
		if err == nil {
			t.Error("Initialize should fail with nil config")
		}
	})

	t.Run("missing driver", func(t *testing.T) {
		cfg := &ClientConfig{
			DSN: ":memory:",
		}
		client := NewClient(cfg)

		err := client.Initialize(nil)
		if err == nil {
			t.Error("Initialize should fail without driver")
		}
	})

	t.Run("missing DSN", func(t *testing.T) {
		cfg := &ClientConfig{
			Driver: "sqlite3",
		}
		client := NewClient(cfg)

		err := client.Initialize(nil)
		if err == nil {
			t.Error("Initialize should fail without DSN")
		}
	})

	t.Run("invalid DSN", func(t *testing.T) {
		cfg := &ClientConfig{
			Driver: "postgres",
			DSN:    "invalid://invalid:invalid@invalid:9999/invalid?sslmode=disable&connect_timeout=1",
		}
		client := NewClient(cfg)

		err := client.Initialize(nil)
		if err == nil {
			t.Error("Initialize should fail with invalid DSN")
		}
	})

	t.Run("unknown driver", func(t *testing.T) {
		cfg := &ClientConfig{
			Driver: "unknown_driver",
			DSN:    ":memory:",
		}
		client := NewClient(cfg)

		err := client.Initialize(nil)
		if err == nil {
			t.Error("Initialize should fail with unknown driver")
		}
	})
}

func TestClient_Health(t *testing.T) {
	t.Run("healthy connection", func(t *testing.T) {
		client := newTestClient(t)

		err := client.Health()
		if err != nil {
			t.Errorf("Health() returned error: %v", err)
		}
	})

	t.Run("not initialized", func(t *testing.T) {
		client := NewClient(nil)

		err := client.Health()
		if err == nil {
			t.Error("Health() should fail when not initialized")
		}
	})
}

func TestClient_HealthContext(t *testing.T) {
	t.Run("healthy connection", func(t *testing.T) {
		client := newTestClient(t)

		err := client.HealthContext(context.Background())
		if err != nil {
			t.Errorf("HealthContext() returned error: %v", err)
		}
	})

	t.Run("not initialized", func(t *testing.T) {
		client := NewClient(nil)

		err := client.HealthContext(context.Background())
		if err == nil {
			t.Error("HealthContext() should fail when not initialized")
		}
	})

	t.Run("cancelled context", func(t *testing.T) {
		client := newTestClient(t)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := client.HealthContext(ctx)
		if err == nil {
			t.Error("HealthContext() should fail with cancelled context")
		}
	})
}

func TestClient_Shutdown(t *testing.T) {
	t.Run("clean shutdown", func(t *testing.T) {
		cfg := &ClientConfig{
			Driver: "sqlite3",
			DSN:    ":memory:",
		}
		client := NewClient(cfg)
		if err := client.Initialize(nil); err != nil {
			t.Fatal(err)
		}

		err := client.Shutdown()
		if err != nil {
			t.Errorf("Shutdown() returned error: %v", err)
		}

		// Health check should now fail
		if err := client.Health(); err == nil {
			t.Error("Health() should fail after shutdown")
		}
	})

	t.Run("not initialized", func(t *testing.T) {
		client := NewClient(nil)

		err := client.Shutdown()
		if err != nil {
			t.Errorf("Shutdown() should not fail when not initialized: %v", err)
		}
	})

	t.Run("double shutdown", func(t *testing.T) {
		client := newTestClient(t)
		// First shutdown is in cleanup

		// This should not panic
		client.Shutdown()
	})
}

func TestClient_DB(t *testing.T) {
	t.Run("returns sqlx.DB", func(t *testing.T) {
		client := newTestClient(t)

		db := client.DB()
		if db == nil {
			t.Error("DB() should return sqlx.DB")
		}
	})

	t.Run("not initialized", func(t *testing.T) {
		client := NewClient(nil)

		db := client.DB()
		if db != nil {
			t.Error("DB() should return nil when not initialized")
		}
	})
}

func TestClient_Config(t *testing.T) {
	cfg := &ClientConfig{
		Driver: "sqlite3",
		DSN:    ":memory:",
	}
	client := NewClient(cfg)

	if client.Config() != cfg {
		t.Error("Config() should return the client config")
	}
}

func TestClient_BeginTx(t *testing.T) {
	t.Run("successful transaction", func(t *testing.T) {
		client := newTestClient(t)

		tx, err := client.BeginTx(context.Background())
		if err != nil {
			t.Fatalf("BeginTx() failed: %v", err)
		}
		defer tx.Rollback()

		if tx == nil {
			t.Error("BeginTx() should return transaction")
		}
	})

	t.Run("not initialized", func(t *testing.T) {
		client := NewClient(nil)

		_, err := client.BeginTx(context.Background())
		if err == nil {
			t.Error("BeginTx() should fail when not initialized")
		}
	})
}

func TestClient_BeginTxWithOptions(t *testing.T) {
	t.Run("with options", func(t *testing.T) {
		client := newTestClient(t)

		opts := &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
			ReadOnly:  true,
		}
		tx, err := client.BeginTxWithOptions(context.Background(), opts)
		if err != nil {
			t.Fatalf("BeginTxWithOptions() failed: %v", err)
		}
		defer tx.Rollback()
	})

	t.Run("not initialized", func(t *testing.T) {
		client := NewClient(nil)

		_, err := client.BeginTxWithOptions(context.Background(), nil)
		if err == nil {
			t.Error("BeginTxWithOptions() should fail when not initialized")
		}
	})
}

func TestClient_ExecContext(t *testing.T) {
	t.Run("successful exec", func(t *testing.T) {
		client := newTestClient(t)

		_, err := client.ExecContext(context.Background(),
			"CREATE TABLE test_exec (id INTEGER PRIMARY KEY)")
		if err != nil {
			t.Fatalf("ExecContext() failed: %v", err)
		}

		result, err := client.ExecContext(context.Background(),
			"INSERT INTO test_exec (id) VALUES (?)", 1)
		if err != nil {
			t.Fatalf("ExecContext() insert failed: %v", err)
		}

		affected, _ := result.RowsAffected()
		if affected != 1 {
			t.Errorf("RowsAffected = %d, want 1", affected)
		}
	})

	t.Run("not initialized", func(t *testing.T) {
		client := NewClient(nil)

		_, err := client.ExecContext(context.Background(), "SELECT 1")
		if err == nil {
			t.Error("ExecContext() should fail when not initialized")
		}
	})
}

func TestClient_QueryContext(t *testing.T) {
	t.Run("successful query", func(t *testing.T) {
		client := newTestClient(t)

		rows, err := client.QueryContext(context.Background(), "SELECT 1 as num")
		if err != nil {
			t.Fatalf("QueryContext() failed: %v", err)
		}
		defer rows.Close()

		if !rows.Next() {
			t.Error("QueryContext() should return rows")
		}
	})

	t.Run("not initialized", func(t *testing.T) {
		client := NewClient(nil)

		_, err := client.QueryContext(context.Background(), "SELECT 1")
		if err == nil {
			t.Error("QueryContext() should fail when not initialized")
		}
	})
}

func TestClient_QueryRowContext(t *testing.T) {
	client := newTestClient(t)

	var result int
	err := client.QueryRowContext(context.Background(), "SELECT 42").Scan(&result)
	if err != nil {
		t.Fatalf("QueryRowContext() failed: %v", err)
	}

	if result != 42 {
		t.Errorf("result = %d, want 42", result)
	}
}

func TestClient_GetContext(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		client := newTestClient(t)

		var result int
		err := client.GetContext(context.Background(), &result, "SELECT 42 as num")
		if err != nil {
			t.Fatalf("GetContext() failed: %v", err)
		}

		if result != 42 {
			t.Errorf("result = %d, want 42", result)
		}
	})

	t.Run("not initialized", func(t *testing.T) {
		client := NewClient(nil)

		var result int
		err := client.GetContext(context.Background(), &result, "SELECT 1")
		if err == nil {
			t.Error("GetContext() should fail when not initialized")
		}
	})
}

func TestClient_SelectContext(t *testing.T) {
	t.Run("successful select", func(t *testing.T) {
		client := newTestClient(t)

		// Create and populate test table
		_, err := client.ExecContext(context.Background(),
			"CREATE TABLE test_select (id INTEGER PRIMARY KEY, name TEXT)")
		if err != nil {
			t.Fatal(err)
		}
		_, _ = client.ExecContext(context.Background(),
			"INSERT INTO test_select (id, name) VALUES (1, 'one'), (2, 'two')")

		type row struct {
			ID   int    `db:"id"`
			Name string `db:"name"`
		}

		var results []row
		err = client.SelectContext(context.Background(), &results,
			"SELECT id, name FROM test_select ORDER BY id")
		if err != nil {
			t.Fatalf("SelectContext() failed: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("len(results) = %d, want 2", len(results))
		}
	})

	t.Run("not initialized", func(t *testing.T) {
		client := NewClient(nil)

		var results []int
		err := client.SelectContext(context.Background(), &results, "SELECT 1")
		if err == nil {
			t.Error("SelectContext() should fail when not initialized")
		}
	})
}

func TestClient_NamedExecContext(t *testing.T) {
	t.Run("successful named exec", func(t *testing.T) {
		client := newTestClient(t)

		_, err := client.ExecContext(context.Background(),
			"CREATE TABLE test_named (id INTEGER PRIMARY KEY, name TEXT)")
		if err != nil {
			t.Fatal(err)
		}

		type insertData struct {
			ID   int    `db:"id"`
			Name string `db:"name"`
		}

		result, err := client.NamedExecContext(context.Background(),
			"INSERT INTO test_named (id, name) VALUES (:id, :name)",
			insertData{ID: 1, Name: "test"})
		if err != nil {
			t.Fatalf("NamedExecContext() failed: %v", err)
		}

		affected, _ := result.RowsAffected()
		if affected != 1 {
			t.Errorf("RowsAffected = %d, want 1", affected)
		}
	})

	t.Run("not initialized", func(t *testing.T) {
		client := NewClient(nil)

		_, err := client.NamedExecContext(context.Background(), "SELECT 1", nil)
		if err == nil {
			t.Error("NamedExecContext() should fail when not initialized")
		}
	})
}

func TestClient_Rebind(t *testing.T) {
	t.Run("with initialized client", func(t *testing.T) {
		client := newTestClient(t)

		query := "SELECT * FROM users WHERE id = ? AND name = ?"
		rebound := client.Rebind(query)

		// SQLite uses ? so should be unchanged
		if rebound != query {
			t.Logf("Rebind result: %s", rebound)
		}
	})

	t.Run("not initialized returns original", func(t *testing.T) {
		client := NewClient(nil)

		query := "SELECT * FROM users WHERE id = ?"
		result := client.Rebind(query)

		if result != query {
			t.Error("Rebind() should return original when not initialized")
		}
	})
}

func TestClient_IsInitialized(t *testing.T) {
	t.Run("before initialization", func(t *testing.T) {
		client := NewClient(nil)

		if client.IsInitialized() {
			t.Error("IsInitialized() should return false before Initialize()")
		}
	})

	t.Run("after initialization", func(t *testing.T) {
		client := newTestClient(t)

		if !client.IsInitialized() {
			t.Error("IsInitialized() should return true after Initialize()")
		}
	})
}

// Test helper to create a test client
func newTestClient(t *testing.T) *Client {
	t.Helper()

	cfg := &ClientConfig{
		Driver:       "sqlite3",
		DSN:          ":memory:",
		MaxOpenConns: 1,
		MaxIdleConns: 1,
	}
	client := NewClient(cfg)

	if err := client.Initialize(nil); err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	t.Cleanup(func() {
		client.Shutdown()
	})

	return client
}

// Test transaction commit and rollback
func TestClient_Transaction(t *testing.T) {
	t.Run("commit", func(t *testing.T) {
		client := newTestClient(t)

		_, err := client.ExecContext(context.Background(),
			"CREATE TABLE tx_test (id INTEGER PRIMARY KEY)")
		if err != nil {
			t.Fatal(err)
		}

		tx, _ := client.BeginTx(context.Background())

		_, err = tx.Exec("INSERT INTO tx_test (id) VALUES (1)")
		if err != nil {
			t.Fatal(err)
		}

		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}

		var count int
		client.GetContext(context.Background(), &count, "SELECT COUNT(*) FROM tx_test")
		if count != 1 {
			t.Errorf("count = %d after commit, want 1", count)
		}
	})

	t.Run("rollback", func(t *testing.T) {
		client := newTestClient(t)

		_, err := client.ExecContext(context.Background(),
			"CREATE TABLE tx_rollback (id INTEGER PRIMARY KEY)")
		if err != nil {
			t.Fatal(err)
		}

		tx, _ := client.BeginTx(context.Background())

		_, err = tx.Exec("INSERT INTO tx_rollback (id) VALUES (1)")
		if err != nil {
			t.Fatal(err)
		}

		if err := tx.Rollback(); err != nil {
			t.Fatal(err)
		}

		var count int
		client.GetContext(context.Background(), &count, "SELECT COUNT(*) FROM tx_rollback")
		if count != 0 {
			t.Errorf("count = %d after rollback, want 0", count)
		}
	})
}
