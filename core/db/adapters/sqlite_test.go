package adapters

import (
	"strings"
	"testing"
)

func TestSQLiteAdapter_DriverName(t *testing.T) {
	a := &SQLiteAdapter{}
	if got := a.DriverName(); got != "sqlite3" {
		t.Errorf("DriverName() = %s, want sqlite3", got)
	}
}

func TestSQLiteAdapter_DSN(t *testing.T) {
	a := &SQLiteAdapter{}

	t.Run("file path", func(t *testing.T) {
		dsn := a.DSN("", 0, "", "", "/path/to/db.sqlite", nil)

		if dsn != "/path/to/db.sqlite" {
			t.Errorf("DSN = %s, want /path/to/db.sqlite", dsn)
		}
	})

	t.Run("memory database", func(t *testing.T) {
		dsn := a.DSN("", 0, "", "", ":memory:", nil)

		if dsn != ":memory:" {
			t.Errorf("DSN = %s, want :memory:", dsn)
		}
	})

	t.Run("empty defaults to memory", func(t *testing.T) {
		dsn := a.DSN("", 0, "", "", "", nil)

		if dsn != ":memory:" {
			t.Errorf("DSN = %s, want :memory:", dsn)
		}
	})

	t.Run("with params", func(t *testing.T) {
		dsn := a.DSN("", 0, "", "", "test.db", map[string]string{
			"cache": "shared",
			"mode":  "memory",
		})

		if !strings.HasPrefix(dsn, "test.db?") {
			t.Errorf("DSN should start with test.db?: %s", dsn)
		}
		if !strings.Contains(dsn, "cache=shared") {
			t.Error("DSN should contain cache param")
		}
		if !strings.Contains(dsn, "mode=memory") {
			t.Error("DSN should contain mode param")
		}
	})

	t.Run("ignores host/port/user", func(t *testing.T) {
		dsn := a.DSN("localhost", 5432, "user", "pass", "test.db", nil)

		if dsn != "test.db" {
			t.Errorf("DSN = %s, want test.db", dsn)
		}
	})
}

func TestSQLiteAdapter_CreateDatabaseSQL(t *testing.T) {
	a := &SQLiteAdapter{}
	sql := a.CreateDatabaseSQL("testdb")

	// SQLite creates databases automatically
	if sql != "" {
		t.Errorf("CreateDatabaseSQL should return empty string, got: %s", sql)
	}
}

func TestSQLiteAdapter_DropDatabaseSQL(t *testing.T) {
	a := &SQLiteAdapter{}
	sql := a.DropDatabaseSQL("testdb")

	// SQLite databases are files
	if sql != "" {
		t.Errorf("DropDatabaseSQL should return empty string, got: %s", sql)
	}
}

func TestSQLiteAdapter_Placeholder(t *testing.T) {
	a := &SQLiteAdapter{}

	// SQLite always uses ?
	for i := 1; i <= 10; i++ {
		if got := a.Placeholder(i); got != "?" {
			t.Errorf("Placeholder(%d) = %s, want ?", i, got)
		}
	}
}

func TestSQLiteAdapter_PlaceholderStyle(t *testing.T) {
	a := &SQLiteAdapter{}
	if got := a.PlaceholderStyle(); got != "question" {
		t.Errorf("PlaceholderStyle() = %s, want question", got)
	}
}

func TestSQLiteAdapter_QuoteIdentifier(t *testing.T) {
	a := &SQLiteAdapter{}

	tests := []struct {
		name string
		want string
	}{
		{"column", `"column"`},
		{"table_name", `"table_name"`},
		{`with"quote`, `"with""quote"`},
	}

	for _, tt := range tests {
		if got := a.QuoteIdentifier(tt.name); got != tt.want {
			t.Errorf("QuoteIdentifier(%s) = %s, want %s", tt.name, got, tt.want)
		}
	}
}

func TestSQLiteAdapter_SupportsReturning(t *testing.T) {
	a := &SQLiteAdapter{}
	if !a.SupportsReturning() {
		t.Error("SQLite 3.35+ should support RETURNING")
	}
}

func TestSQLiteAdapter_SupportsLastInsertID(t *testing.T) {
	a := &SQLiteAdapter{}
	if !a.SupportsLastInsertID() {
		t.Error("SQLite should support LastInsertID")
	}
}
