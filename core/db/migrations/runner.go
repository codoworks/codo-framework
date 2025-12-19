package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// Runner executes migrations
type Runner struct {
	db         *sqlx.DB
	tableName  string
	migrations MigrationList
}

// NewRunner creates a new migration runner
func NewRunner(db *sqlx.DB) *Runner {
	return &Runner{
		db:         db,
		tableName:  "schema_migrations",
		migrations: make(MigrationList, 0),
	}
}

// WithTableName sets the migration tracking table name
func (r *Runner) WithTableName(name string) *Runner {
	r.tableName = name
	return r
}

// Add adds migrations to the runner
func (r *Runner) Add(migrations ...*Migration) *Runner {
	r.migrations = append(r.migrations, migrations...)
	return r
}

// Migrations returns the registered migrations
func (r *Runner) Migrations() MigrationList {
	return r.migrations
}

// Initialize creates the migrations table if it doesn't exist
func (r *Runner) Initialize(ctx context.Context) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			version VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`, r.tableName)

	_, err := r.db.ExecContext(ctx, query)
	return err
}

// Applied returns the list of applied migration versions
func (r *Runner) Applied(ctx context.Context) ([]MigrationRecord, error) {
	query := fmt.Sprintf("SELECT version, name, applied_at FROM %s ORDER BY version ASC", r.tableName)

	var records []MigrationRecord
	err := r.db.SelectContext(ctx, &records, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	return records, nil
}

// Pending returns migrations that haven't been applied yet
func (r *Runner) Pending(ctx context.Context) (MigrationList, error) {
	applied, err := r.Applied(ctx)
	if err != nil {
		return nil, err
	}

	appliedMap := make(map[string]bool)
	for _, record := range applied {
		appliedMap[record.Version] = true
	}

	pending := make(MigrationList, 0)
	for _, m := range r.migrations {
		if !appliedMap[m.Version] {
			pending = append(pending, m)
		}
	}

	sort.Sort(pending)
	return pending, nil
}

// Up runs all pending migrations
func (r *Runner) Up(ctx context.Context) (int, error) {
	pending, err := r.Pending(ctx)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, m := range pending {
		if err := r.runMigration(ctx, m, Up); err != nil {
			return count, fmt.Errorf("migration %s failed: %w", m.FullName(), err)
		}
		count++
	}

	return count, nil
}

// UpTo runs migrations up to and including the specified version
func (r *Runner) UpTo(ctx context.Context, version string) (int, error) {
	pending, err := r.Pending(ctx)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, m := range pending {
		if m.Version > version {
			break
		}
		if err := r.runMigration(ctx, m, Up); err != nil {
			return count, fmt.Errorf("migration %s failed: %w", m.FullName(), err)
		}
		count++
	}

	return count, nil
}

// UpOne runs the next pending migration
func (r *Runner) UpOne(ctx context.Context) error {
	pending, err := r.Pending(ctx)
	if err != nil {
		return err
	}

	if len(pending) == 0 {
		return nil
	}

	return r.runMigration(ctx, pending[0], Up)
}

// Down reverts the last applied migration
func (r *Runner) Down(ctx context.Context) error {
	applied, err := r.Applied(ctx)
	if err != nil {
		return err
	}

	if len(applied) == 0 {
		return nil
	}

	lastVersion := applied[len(applied)-1].Version

	// Find the migration
	var migration *Migration
	for _, m := range r.migrations {
		if m.Version == lastVersion {
			migration = m
			break
		}
	}

	if migration == nil {
		return fmt.Errorf("migration %s not found", lastVersion)
	}

	return r.runMigration(ctx, migration, Down)
}

// DownTo reverts migrations down to (but not including) the specified version
func (r *Runner) DownTo(ctx context.Context, version string) (int, error) {
	applied, err := r.Applied(ctx)
	if err != nil {
		return 0, err
	}

	// Sort applied in reverse order
	sort.Slice(applied, func(i, j int) bool {
		return applied[i].Version > applied[j].Version
	})

	count := 0
	for _, record := range applied {
		if record.Version <= version {
			break
		}

		// Find the migration
		var migration *Migration
		for _, m := range r.migrations {
			if m.Version == record.Version {
				migration = m
				break
			}
		}

		if migration == nil {
			return count, fmt.Errorf("migration %s not found", record.Version)
		}

		if err := r.runMigration(ctx, migration, Down); err != nil {
			return count, fmt.Errorf("migration %s failed: %w", migration.FullName(), err)
		}
		count++
	}

	return count, nil
}

// Reset reverts all applied migrations
func (r *Runner) Reset(ctx context.Context) (int, error) {
	applied, err := r.Applied(ctx)
	if err != nil {
		return 0, err
	}

	// Sort applied in reverse order
	sort.Slice(applied, func(i, j int) bool {
		return applied[i].Version > applied[j].Version
	})

	count := 0
	for _, record := range applied {
		// Find the migration
		var migration *Migration
		for _, m := range r.migrations {
			if m.Version == record.Version {
				migration = m
				break
			}
		}

		if migration == nil {
			return count, fmt.Errorf("migration %s not found", record.Version)
		}

		if err := r.runMigration(ctx, migration, Down); err != nil {
			return count, fmt.Errorf("migration %s failed: %w", migration.FullName(), err)
		}
		count++
	}

	return count, nil
}

// Refresh resets and re-runs all migrations
func (r *Runner) Refresh(ctx context.Context) error {
	if _, err := r.Reset(ctx); err != nil {
		return fmt.Errorf("reset failed: %w", err)
	}

	if _, err := r.Up(ctx); err != nil {
		return fmt.Errorf("up failed: %w", err)
	}

	return nil
}

// Status returns the status of all migrations
func (r *Runner) Status(ctx context.Context) ([]MigrationStatus, error) {
	applied, err := r.Applied(ctx)
	if err != nil {
		return nil, err
	}

	appliedMap := make(map[string]*MigrationRecord)
	for i := range applied {
		appliedMap[applied[i].Version] = &applied[i]
	}

	// Combine registered migrations with applied
	allVersions := make(map[string]bool)
	for _, m := range r.migrations {
		allVersions[m.Version] = true
	}
	for _, record := range applied {
		allVersions[record.Version] = true
	}

	statuses := make([]MigrationStatus, 0, len(allVersions))
	for version := range allVersions {
		status := MigrationStatus{Version: version}

		// Find registered migration
		for _, m := range r.migrations {
			if m.Version == version {
				status.Name = m.Name
				status.Registered = true
				break
			}
		}

		// Check if applied
		if record, ok := appliedMap[version]; ok {
			status.Applied = true
			status.AppliedAt = &record.AppliedAt
		}

		statuses = append(statuses, status)
	}

	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Version < statuses[j].Version
	})

	return statuses, nil
}

// runMigration executes a single migration
func (r *Runner) runMigration(ctx context.Context, m *Migration, direction Direction) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}

	// Create an executor wrapper
	executor := &txExecutor{tx: tx}

	if direction == Up {
		if err := r.runUp(ctx, tx, executor, m); err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if err := r.runDown(ctx, tx, executor, m); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *Runner) runUp(ctx context.Context, tx *sqlx.Tx, executor Executor, m *Migration) error {
	// Run migration
	if m.UpFunc != nil {
		if err := m.UpFunc(executor); err != nil {
			return err
		}
	} else if m.UpSQL != "" {
		// Split and execute statements
		for _, stmt := range splitStatements(m.UpSQL) {
			if _, err := tx.ExecContext(ctx, stmt); err != nil {
				return fmt.Errorf("SQL failed: %w", err)
			}
		}
	}

	// Record migration
	query := fmt.Sprintf("INSERT INTO %s (version, name, applied_at) VALUES (?, ?, ?)", r.tableName)
	query = tx.Rebind(query)
	_, err := tx.ExecContext(ctx, query, m.Version, m.Name, time.Now())
	return err
}

func (r *Runner) runDown(ctx context.Context, tx *sqlx.Tx, executor Executor, m *Migration) error {
	// Run migration
	if m.DownFunc != nil {
		if err := m.DownFunc(executor); err != nil {
			return err
		}
	} else if m.DownSQL != "" {
		for _, stmt := range splitStatements(m.DownSQL) {
			if _, err := tx.ExecContext(ctx, stmt); err != nil {
				return fmt.Errorf("SQL failed: %w", err)
			}
		}
	}

	// Remove record
	query := fmt.Sprintf("DELETE FROM %s WHERE version = ?", r.tableName)
	query = tx.Rebind(query)
	_, err := tx.ExecContext(ctx, query, m.Version)
	return err
}

// MigrationStatus represents the status of a migration
type MigrationStatus struct {
	Version    string
	Name       string
	Applied    bool
	AppliedAt  *time.Time
	Registered bool
}

// txExecutor wraps a transaction for the Executor interface
type txExecutor struct {
	tx *sqlx.Tx
}

func (e *txExecutor) Exec(query string, args ...any) error {
	_, err := e.tx.Exec(query, args...)
	return err
}

// splitStatements splits SQL into individual statements
func splitStatements(sql string) []string {
	statements := strings.Split(sql, ";")
	result := make([]string, 0, len(statements))

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt != "" {
			result = append(result, stmt)
		}
	}

	return result
}

// Version returns the current schema version (last applied migration)
func (r *Runner) Version(ctx context.Context) (string, error) {
	query := fmt.Sprintf("SELECT version FROM %s ORDER BY version DESC LIMIT 1", r.tableName)

	var version string
	err := r.db.GetContext(ctx, &version, query)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return version, err
}
