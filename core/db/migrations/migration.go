package migrations

import (
	"time"
)

// Direction indicates the migration direction
type Direction string

const (
	// Up runs the migration forward
	Up Direction = "up"
	// Down reverts the migration
	Down Direction = "down"
)

// Migration represents a database migration
type Migration struct {
	// Version is a unique identifier for the migration (typically timestamp)
	Version string

	// Name is a human-readable description
	Name string

	// UpSQL is the SQL to apply the migration
	UpSQL string

	// DownSQL is the SQL to revert the migration
	DownSQL string

	// UpFunc is an optional function to run instead of UpSQL
	UpFunc func(tx Executor) error

	// DownFunc is an optional function to run instead of DownSQL
	DownFunc func(tx Executor) error
}

// Executor is the interface for executing SQL
type Executor interface {
	Exec(query string, args ...any) error
}

// MigrationRecord represents a migration entry in the database
type MigrationRecord struct {
	Version   string    `db:"version"`
	Name      string    `db:"name"`
	AppliedAt time.Time `db:"applied_at"`
}

// NewMigration creates a new migration with the given version and name
func NewMigration(version, name string) *Migration {
	return &Migration{
		Version: version,
		Name:    name,
	}
}

// WithUpSQL sets the up SQL
func (m *Migration) WithUpSQL(sql string) *Migration {
	m.UpSQL = sql
	return m
}

// WithDownSQL sets the down SQL
func (m *Migration) WithDownSQL(sql string) *Migration {
	m.DownSQL = sql
	return m
}

// WithUpFunc sets the up function
func (m *Migration) WithUpFunc(fn func(tx Executor) error) *Migration {
	m.UpFunc = fn
	return m
}

// WithDownFunc sets the down function
func (m *Migration) WithDownFunc(fn func(tx Executor) error) *Migration {
	m.DownFunc = fn
	return m
}

// HasUp returns true if the migration has an up operation
func (m *Migration) HasUp() bool {
	return m.UpSQL != "" || m.UpFunc != nil
}

// HasDown returns true if the migration has a down operation
func (m *Migration) HasDown() bool {
	return m.DownSQL != "" || m.DownFunc != nil
}

// FullName returns the version and name combined
func (m *Migration) FullName() string {
	if m.Name != "" {
		return m.Version + "_" + m.Name
	}
	return m.Version
}

// GenerateVersion generates a version string based on current time
func GenerateVersion() string {
	return time.Now().Format("20060102150405")
}

// MigrationList is a sortable list of migrations
type MigrationList []*Migration

func (l MigrationList) Len() int           { return len(l) }
func (l MigrationList) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l MigrationList) Less(i, j int) bool { return l[i].Version < l[j].Version }
