package adapters

import (
	"os"

	"github.com/codoworks/codo-framework/core/errors"
)

// Adapter defines the interface for database adapters
type Adapter interface {
	// DriverName returns the driver name for sqlx
	DriverName() string

	// DSN builds the connection string
	DSN(host string, port int, user, password, dbname string, params map[string]string) string

	// CreateDatabaseSQL returns SQL to create a database
	CreateDatabaseSQL(dbname string) string

	// DropDatabaseSQL returns SQL to drop a database
	DropDatabaseSQL(dbname string) string

	// Placeholder returns the placeholder style ($1, ?, etc.)
	Placeholder(index int) string

	// PlaceholderStyle returns the placeholder style name
	PlaceholderStyle() string

	// QuoteIdentifier quotes an identifier (table/column name)
	QuoteIdentifier(name string) string

	// SupportsReturning returns true if the adapter supports RETURNING clause
	SupportsReturning() bool

	// SupportsLastInsertID returns true if the adapter supports LastInsertId
	SupportsLastInsertID() bool
}

// GetAdapter returns the adapter for a driver name
func GetAdapter(driver string) Adapter {
	switch driver {
	case "postgres", "postgresql":
		return &PostgresAdapter{}
	case "mysql":
		return &MySQLAdapter{}
	case "sqlite", "sqlite3":
		return &SQLiteAdapter{}
	default:
		return nil
	}
}

// MustGetAdapter returns the adapter or exits with error if not found
func MustGetAdapter(driver string) Adapter {
	adapter := GetAdapter(driver)
	if adapter == nil {
		frameworkErr := errors.BadRequest("Unsupported database driver").
			WithPhase(errors.PhaseBootstrap).
			WithDetail("driver", driver).
			WithDetail("supported_drivers", SupportedDrivers())
		errors.RenderCLI(frameworkErr)
		os.Exit(1)
	}
	return adapter
}

// SupportedDrivers returns a list of supported driver names
func SupportedDrivers() []string {
	return []string{"postgres", "mysql", "sqlite3"}
}

// IsSupported returns true if the driver is supported
func IsSupported(driver string) bool {
	return GetAdapter(driver) != nil
}
