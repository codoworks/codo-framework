package adapters

import (
	"fmt"
	"net/url"
	"strings"
)

// MySQLAdapter implements Adapter for MySQL
type MySQLAdapter struct{}

// DriverName returns the driver name for sqlx
func (a *MySQLAdapter) DriverName() string {
	return "mysql"
}

// DSN builds the MySQL connection string
// Format: user:password@tcp(host:port)/dbname?param=value
func (a *MySQLAdapter) DSN(host string, port int, user, password, dbname string, params map[string]string) string {
	var dsn strings.Builder

	// user:password@
	if user != "" {
		dsn.WriteString(user)
		if password != "" {
			dsn.WriteString(":")
			dsn.WriteString(password)
		}
		dsn.WriteString("@")
	}

	// tcp(host:port)
	dsn.WriteString("tcp(")
	if host != "" {
		dsn.WriteString(host)
	} else {
		dsn.WriteString("localhost")
	}
	if port > 0 {
		dsn.WriteString(fmt.Sprintf(":%d", port))
	} else {
		dsn.WriteString(":3306")
	}
	dsn.WriteString(")")

	// /dbname
	dsn.WriteString("/")
	if dbname != "" {
		dsn.WriteString(dbname)
	}

	// ?params
	if len(params) > 0 {
		dsn.WriteString("?")
		values := url.Values{}
		for k, v := range params {
			values.Set(k, v)
		}
		dsn.WriteString(values.Encode())
	}

	return dsn.String()
}

// CreateDatabaseSQL returns SQL to create a database
func (a *MySQLAdapter) CreateDatabaseSQL(dbname string) string {
	return fmt.Sprintf("CREATE DATABASE %s", a.QuoteIdentifier(dbname))
}

// DropDatabaseSQL returns SQL to drop a database
func (a *MySQLAdapter) DropDatabaseSQL(dbname string) string {
	return fmt.Sprintf("DROP DATABASE IF EXISTS %s", a.QuoteIdentifier(dbname))
}

// Placeholder returns the MySQL placeholder (?)
func (a *MySQLAdapter) Placeholder(index int) string {
	return "?"
}

// PlaceholderStyle returns the placeholder style name
func (a *MySQLAdapter) PlaceholderStyle() string {
	return "question"
}

// QuoteIdentifier quotes an identifier using backticks
func (a *MySQLAdapter) QuoteIdentifier(name string) string {
	return fmt.Sprintf("`%s`", strings.ReplaceAll(name, "`", "``"))
}

// SupportsReturning returns false as MySQL doesn't support RETURNING
func (a *MySQLAdapter) SupportsReturning() bool {
	return false
}

// SupportsLastInsertID returns true as MySQL supports LastInsertId
func (a *MySQLAdapter) SupportsLastInsertID() bool {
	return true
}
