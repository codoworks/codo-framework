package adapters

import (
	"fmt"
	"strings"
)

// PostgresAdapter implements Adapter for PostgreSQL
type PostgresAdapter struct{}

// DriverName returns the driver name for sqlx
func (a *PostgresAdapter) DriverName() string {
	return "postgres"
}

// DSN builds the PostgreSQL connection string
func (a *PostgresAdapter) DSN(host string, port int, user, password, dbname string, params map[string]string) string {
	var parts []string

	if host != "" {
		parts = append(parts, fmt.Sprintf("host=%s", host))
	}
	if port > 0 {
		parts = append(parts, fmt.Sprintf("port=%d", port))
	}
	if user != "" {
		parts = append(parts, fmt.Sprintf("user=%s", user))
	}
	if password != "" {
		parts = append(parts, fmt.Sprintf("password=%s", password))
	}
	if dbname != "" {
		parts = append(parts, fmt.Sprintf("dbname=%s", dbname))
	}

	for k, v := range params {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(parts, " ")
}

// CreateDatabaseSQL returns SQL to create a database
func (a *PostgresAdapter) CreateDatabaseSQL(dbname string) string {
	return fmt.Sprintf("CREATE DATABASE %s", a.QuoteIdentifier(dbname))
}

// DropDatabaseSQL returns SQL to drop a database
func (a *PostgresAdapter) DropDatabaseSQL(dbname string) string {
	return fmt.Sprintf("DROP DATABASE IF EXISTS %s", a.QuoteIdentifier(dbname))
}

// Placeholder returns the PostgreSQL placeholder ($1, $2, etc.)
func (a *PostgresAdapter) Placeholder(index int) string {
	return fmt.Sprintf("$%d", index)
}

// PlaceholderStyle returns the placeholder style name
func (a *PostgresAdapter) PlaceholderStyle() string {
	return "dollar"
}

// QuoteIdentifier quotes an identifier using double quotes
func (a *PostgresAdapter) QuoteIdentifier(name string) string {
	return fmt.Sprintf(`"%s"`, strings.ReplaceAll(name, `"`, `""`))
}

// SupportsReturning returns true as PostgreSQL supports RETURNING
func (a *PostgresAdapter) SupportsReturning() bool {
	return true
}

// SupportsLastInsertID returns false as PostgreSQL uses RETURNING
func (a *PostgresAdapter) SupportsLastInsertID() bool {
	return false
}
