package adapters

import (
	"strings"
	"testing"
)

func TestPostgresAdapter_DriverName(t *testing.T) {
	a := &PostgresAdapter{}
	if got := a.DriverName(); got != "postgres" {
		t.Errorf("DriverName() = %s, want postgres", got)
	}
}

func TestPostgresAdapter_DSN(t *testing.T) {
	a := &PostgresAdapter{}

	t.Run("full DSN", func(t *testing.T) {
		dsn := a.DSN("localhost", 5432, "user", "pass", "mydb", map[string]string{
			"sslmode": "disable",
		})

		if !strings.Contains(dsn, "host=localhost") {
			t.Error("DSN should contain host")
		}
		if !strings.Contains(dsn, "port=5432") {
			t.Error("DSN should contain port")
		}
		if !strings.Contains(dsn, "user=user") {
			t.Error("DSN should contain user")
		}
		if !strings.Contains(dsn, "password=pass") {
			t.Error("DSN should contain password")
		}
		if !strings.Contains(dsn, "dbname=mydb") {
			t.Error("DSN should contain dbname")
		}
		if !strings.Contains(dsn, "sslmode=disable") {
			t.Error("DSN should contain params")
		}
	})

	t.Run("minimal DSN", func(t *testing.T) {
		dsn := a.DSN("", 0, "", "", "", nil)
		if dsn != "" {
			t.Errorf("Minimal DSN should be empty, got: %s", dsn)
		}
	})

	t.Run("without password", func(t *testing.T) {
		dsn := a.DSN("localhost", 5432, "user", "", "mydb", nil)
		if strings.Contains(dsn, "password") {
			t.Error("DSN should not contain password when empty")
		}
	})
}

func TestPostgresAdapter_CreateDatabaseSQL(t *testing.T) {
	a := &PostgresAdapter{}
	sql := a.CreateDatabaseSQL("testdb")

	if !strings.Contains(sql, "CREATE DATABASE") {
		t.Error("CreateDatabaseSQL should contain CREATE DATABASE")
	}
	if !strings.Contains(sql, `"testdb"`) {
		t.Error("CreateDatabaseSQL should quote the database name")
	}
}

func TestPostgresAdapter_DropDatabaseSQL(t *testing.T) {
	a := &PostgresAdapter{}
	sql := a.DropDatabaseSQL("testdb")

	if !strings.Contains(sql, "DROP DATABASE IF EXISTS") {
		t.Error("DropDatabaseSQL should contain DROP DATABASE IF EXISTS")
	}
}

func TestPostgresAdapter_Placeholder(t *testing.T) {
	a := &PostgresAdapter{}

	tests := []struct {
		index int
		want  string
	}{
		{1, "$1"},
		{2, "$2"},
		{10, "$10"},
		{100, "$100"},
	}

	for _, tt := range tests {
		if got := a.Placeholder(tt.index); got != tt.want {
			t.Errorf("Placeholder(%d) = %s, want %s", tt.index, got, tt.want)
		}
	}
}

func TestPostgresAdapter_PlaceholderStyle(t *testing.T) {
	a := &PostgresAdapter{}
	if got := a.PlaceholderStyle(); got != "dollar" {
		t.Errorf("PlaceholderStyle() = %s, want dollar", got)
	}
}

func TestPostgresAdapter_QuoteIdentifier(t *testing.T) {
	a := &PostgresAdapter{}

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

func TestPostgresAdapter_SupportsReturning(t *testing.T) {
	a := &PostgresAdapter{}
	if !a.SupportsReturning() {
		t.Error("PostgreSQL should support RETURNING")
	}
}

func TestPostgresAdapter_SupportsLastInsertID(t *testing.T) {
	a := &PostgresAdapter{}
	if a.SupportsLastInsertID() {
		t.Error("PostgreSQL should not support LastInsertID")
	}
}
