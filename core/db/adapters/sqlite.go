package adapters

import (
	"fmt"
	"net/url"
	"strings"
)

// SQLiteAdapter implements Adapter for SQLite
type SQLiteAdapter struct{}

// DriverName returns the driver name for sqlx
func (a *SQLiteAdapter) DriverName() string {
	return "sqlite3"
}

// DSN builds the SQLite connection string
// For SQLite, this is typically just a file path or :memory:
func (a *SQLiteAdapter) DSN(host string, port int, user, password, dbname string, params map[string]string) string {
	// For SQLite, dbname is the file path
	dsn := dbname
	if dsn == "" {
		dsn = ":memory:"
	}

	// Add parameters if any
	if len(params) > 0 {
		values := url.Values{}
		for k, v := range params {
			values.Set(k, v)
		}
		dsn = fmt.Sprintf("%s?%s", dsn, values.Encode())
	}

	return dsn
}

// CreateDatabaseSQL returns SQL to create a database
// SQLite databases are created automatically when opened
func (a *SQLiteAdapter) CreateDatabaseSQL(dbname string) string {
	// SQLite creates the file automatically
	return ""
}

// DropDatabaseSQL returns SQL to drop a database
// For SQLite, this would mean deleting the file
func (a *SQLiteAdapter) DropDatabaseSQL(dbname string) string {
	// SQLite doesn't have DROP DATABASE - file must be deleted
	return ""
}

// Placeholder returns the SQLite placeholder (?)
func (a *SQLiteAdapter) Placeholder(index int) string {
	return "?"
}

// PlaceholderStyle returns the placeholder style name
func (a *SQLiteAdapter) PlaceholderStyle() string {
	return "question"
}

// QuoteIdentifier quotes an identifier using double quotes
func (a *SQLiteAdapter) QuoteIdentifier(name string) string {
	return fmt.Sprintf(`"%s"`, strings.ReplaceAll(name, `"`, `""`))
}

// SupportsReturning returns true as SQLite 3.35+ supports RETURNING
func (a *SQLiteAdapter) SupportsReturning() bool {
	return true
}

// SupportsLastInsertID returns true as SQLite supports LastInsertId
func (a *SQLiteAdapter) SupportsLastInsertID() bool {
	return true
}
