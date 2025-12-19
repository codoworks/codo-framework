package adapters

import (
	"strings"
	"testing"
)

func TestMySQLAdapter_DriverName(t *testing.T) {
	a := &MySQLAdapter{}
	if got := a.DriverName(); got != "mysql" {
		t.Errorf("DriverName() = %s, want mysql", got)
	}
}

func TestMySQLAdapter_DSN(t *testing.T) {
	a := &MySQLAdapter{}

	t.Run("full DSN", func(t *testing.T) {
		dsn := a.DSN("localhost", 3306, "user", "pass", "mydb", map[string]string{
			"parseTime": "true",
		})

		if !strings.HasPrefix(dsn, "user:pass@tcp(localhost:3306)/mydb") {
			t.Errorf("DSN format incorrect: %s", dsn)
		}
		if !strings.Contains(dsn, "parseTime=true") {
			t.Error("DSN should contain params")
		}
	})

	t.Run("default host and port", func(t *testing.T) {
		dsn := a.DSN("", 0, "user", "pass", "mydb", nil)

		if !strings.Contains(dsn, "tcp(localhost:3306)") {
			t.Errorf("DSN should use default host and port: %s", dsn)
		}
	})

	t.Run("without password", func(t *testing.T) {
		dsn := a.DSN("localhost", 3306, "user", "", "mydb", nil)

		if !strings.HasPrefix(dsn, "user@tcp") {
			t.Errorf("DSN should not have colon without password: %s", dsn)
		}
	})

	t.Run("without user", func(t *testing.T) {
		dsn := a.DSN("localhost", 3306, "", "", "mydb", nil)

		if strings.HasPrefix(dsn, "@") {
			t.Errorf("DSN should not start with @: %s", dsn)
		}
	})

	t.Run("without dbname", func(t *testing.T) {
		dsn := a.DSN("localhost", 3306, "user", "pass", "", nil)

		if !strings.HasSuffix(dsn, ")/") && !strings.Contains(dsn, ")/?") {
			t.Errorf("DSN should have trailing slash: %s", dsn)
		}
	})
}

func TestMySQLAdapter_CreateDatabaseSQL(t *testing.T) {
	a := &MySQLAdapter{}
	sql := a.CreateDatabaseSQL("testdb")

	if !strings.Contains(sql, "CREATE DATABASE") {
		t.Error("CreateDatabaseSQL should contain CREATE DATABASE")
	}
	if !strings.Contains(sql, "`testdb`") {
		t.Error("CreateDatabaseSQL should quote the database name with backticks")
	}
}

func TestMySQLAdapter_DropDatabaseSQL(t *testing.T) {
	a := &MySQLAdapter{}
	sql := a.DropDatabaseSQL("testdb")

	if !strings.Contains(sql, "DROP DATABASE IF EXISTS") {
		t.Error("DropDatabaseSQL should contain DROP DATABASE IF EXISTS")
	}
}

func TestMySQLAdapter_Placeholder(t *testing.T) {
	a := &MySQLAdapter{}

	// MySQL always uses ?
	for i := 1; i <= 10; i++ {
		if got := a.Placeholder(i); got != "?" {
			t.Errorf("Placeholder(%d) = %s, want ?", i, got)
		}
	}
}

func TestMySQLAdapter_PlaceholderStyle(t *testing.T) {
	a := &MySQLAdapter{}
	if got := a.PlaceholderStyle(); got != "question" {
		t.Errorf("PlaceholderStyle() = %s, want question", got)
	}
}

func TestMySQLAdapter_QuoteIdentifier(t *testing.T) {
	a := &MySQLAdapter{}

	tests := []struct {
		name string
		want string
	}{
		{"column", "`column`"},
		{"table_name", "`table_name`"},
		{"with`tick", "`with``tick`"},
	}

	for _, tt := range tests {
		if got := a.QuoteIdentifier(tt.name); got != tt.want {
			t.Errorf("QuoteIdentifier(%s) = %s, want %s", tt.name, got, tt.want)
		}
	}
}

func TestMySQLAdapter_SupportsReturning(t *testing.T) {
	a := &MySQLAdapter{}
	if a.SupportsReturning() {
		t.Error("MySQL should not support RETURNING")
	}
}

func TestMySQLAdapter_SupportsLastInsertID(t *testing.T) {
	a := &MySQLAdapter{}
	if !a.SupportsLastInsertID() {
		t.Error("MySQL should support LastInsertID")
	}
}
