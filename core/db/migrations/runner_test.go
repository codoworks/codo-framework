package migrations

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func newTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestNewRunner(t *testing.T) {
	db := newTestDB(t)
	runner := NewRunner(db)

	if runner == nil {
		t.Fatal("NewRunner returned nil")
	}
	if runner.tableName != "schema_migrations" {
		t.Errorf("tableName = %s, want schema_migrations", runner.tableName)
	}
}

func TestRunner_WithTableName(t *testing.T) {
	db := newTestDB(t)
	runner := NewRunner(db).WithTableName("custom_migrations")

	if runner.tableName != "custom_migrations" {
		t.Errorf("tableName = %s, want custom_migrations", runner.tableName)
	}
}

func TestRunner_Add(t *testing.T) {
	db := newTestDB(t)
	runner := NewRunner(db)

	m1 := NewMigration("1", "first")
	m2 := NewMigration("2", "second")

	runner.Add(m1, m2)

	if len(runner.Migrations()) != 2 {
		t.Errorf("len(Migrations()) = %d, want 2", len(runner.Migrations()))
	}
}

func TestRunner_Initialize(t *testing.T) {
	db := newTestDB(t)
	runner := NewRunner(db)
	ctx := context.Background()

	err := runner.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Verify table exists
	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM schema_migrations")
	if err != nil {
		t.Errorf("Table should exist: %v", err)
	}
}

func TestRunner_Up(t *testing.T) {
	db := newTestDB(t)
	runner := NewRunner(db)
	ctx := context.Background()

	runner.Add(
		NewMigration("001", "create_users").WithUpSQL("CREATE TABLE users (id INTEGER PRIMARY KEY)"),
		NewMigration("002", "create_posts").WithUpSQL("CREATE TABLE posts (id INTEGER PRIMARY KEY)"),
	)

	runner.Initialize(ctx)

	count, err := runner.Up(ctx)
	if err != nil {
		t.Fatalf("Up failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Up count = %d, want 2", count)
	}

	// Verify tables exist
	var tableCount int
	db.Get(&tableCount, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('users', 'posts')")
	if tableCount != 2 {
		t.Errorf("Tables not created: %d", tableCount)
	}
}

func TestRunner_UpOne(t *testing.T) {
	db := newTestDB(t)
	runner := NewRunner(db)
	ctx := context.Background()

	runner.Add(
		NewMigration("001", "first").WithUpSQL("CREATE TABLE t1 (id INTEGER)"),
		NewMigration("002", "second").WithUpSQL("CREATE TABLE t2 (id INTEGER)"),
	)

	runner.Initialize(ctx)

	err := runner.UpOne(ctx)
	if err != nil {
		t.Fatalf("UpOne failed: %v", err)
	}

	applied, _ := runner.Applied(ctx)
	if len(applied) != 1 {
		t.Errorf("len(applied) = %d, want 1", len(applied))
	}
	if applied[0].Version != "001" {
		t.Errorf("applied version = %s, want 001", applied[0].Version)
	}
}

func TestRunner_UpTo(t *testing.T) {
	db := newTestDB(t)
	runner := NewRunner(db)
	ctx := context.Background()

	runner.Add(
		NewMigration("001", "first").WithUpSQL("CREATE TABLE t1 (id INTEGER)"),
		NewMigration("002", "second").WithUpSQL("CREATE TABLE t2 (id INTEGER)"),
		NewMigration("003", "third").WithUpSQL("CREATE TABLE t3 (id INTEGER)"),
	)

	runner.Initialize(ctx)

	count, err := runner.UpTo(ctx, "002")
	if err != nil {
		t.Fatalf("UpTo failed: %v", err)
	}

	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}

	pending, _ := runner.Pending(ctx)
	if len(pending) != 1 {
		t.Errorf("pending count = %d, want 1", len(pending))
	}
}

func TestRunner_Down(t *testing.T) {
	db := newTestDB(t)
	runner := NewRunner(db)
	ctx := context.Background()

	runner.Add(
		NewMigration("001", "first").
			WithUpSQL("CREATE TABLE t1 (id INTEGER)").
			WithDownSQL("DROP TABLE t1"),
	)

	runner.Initialize(ctx)
	runner.Up(ctx)

	err := runner.Down(ctx)
	if err != nil {
		t.Fatalf("Down failed: %v", err)
	}

	applied, _ := runner.Applied(ctx)
	if len(applied) != 0 {
		t.Errorf("len(applied) = %d, want 0", len(applied))
	}
}

func TestRunner_DownTo(t *testing.T) {
	db := newTestDB(t)
	runner := NewRunner(db)
	ctx := context.Background()

	runner.Add(
		NewMigration("001", "first").
			WithUpSQL("CREATE TABLE t1 (id INTEGER)").
			WithDownSQL("DROP TABLE t1"),
		NewMigration("002", "second").
			WithUpSQL("CREATE TABLE t2 (id INTEGER)").
			WithDownSQL("DROP TABLE t2"),
		NewMigration("003", "third").
			WithUpSQL("CREATE TABLE t3 (id INTEGER)").
			WithDownSQL("DROP TABLE t3"),
	)

	runner.Initialize(ctx)
	runner.Up(ctx)

	count, err := runner.DownTo(ctx, "001")
	if err != nil {
		t.Fatalf("DownTo failed: %v", err)
	}

	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}

	applied, _ := runner.Applied(ctx)
	if len(applied) != 1 {
		t.Errorf("len(applied) = %d, want 1", len(applied))
	}
}

func TestRunner_Reset(t *testing.T) {
	db := newTestDB(t)
	runner := NewRunner(db)
	ctx := context.Background()

	runner.Add(
		NewMigration("001", "first").
			WithUpSQL("CREATE TABLE t1 (id INTEGER)").
			WithDownSQL("DROP TABLE t1"),
		NewMigration("002", "second").
			WithUpSQL("CREATE TABLE t2 (id INTEGER)").
			WithDownSQL("DROP TABLE t2"),
	)

	runner.Initialize(ctx)
	runner.Up(ctx)

	count, err := runner.Reset(ctx)
	if err != nil {
		t.Fatalf("Reset failed: %v", err)
	}

	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}

	applied, _ := runner.Applied(ctx)
	if len(applied) != 0 {
		t.Errorf("len(applied) = %d, want 0", len(applied))
	}
}

func TestRunner_Refresh(t *testing.T) {
	db := newTestDB(t)
	runner := NewRunner(db)
	ctx := context.Background()

	runner.Add(
		NewMigration("001", "first").
			WithUpSQL("CREATE TABLE t1 (id INTEGER)").
			WithDownSQL("DROP TABLE t1"),
	)

	runner.Initialize(ctx)
	runner.Up(ctx)

	err := runner.Refresh(ctx)
	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}

	applied, _ := runner.Applied(ctx)
	if len(applied) != 1 {
		t.Errorf("len(applied) = %d, want 1", len(applied))
	}
}

func TestRunner_Pending(t *testing.T) {
	db := newTestDB(t)
	runner := NewRunner(db)
	ctx := context.Background()

	runner.Add(
		NewMigration("001", "first").WithUpSQL("CREATE TABLE t1 (id INTEGER)"),
		NewMigration("002", "second").WithUpSQL("CREATE TABLE t2 (id INTEGER)"),
	)

	runner.Initialize(ctx)

	pending, err := runner.Pending(ctx)
	if err != nil {
		t.Fatalf("Pending failed: %v", err)
	}

	if len(pending) != 2 {
		t.Errorf("len(pending) = %d, want 2", len(pending))
	}

	// Run one migration
	runner.UpOne(ctx)

	pending, _ = runner.Pending(ctx)
	if len(pending) != 1 {
		t.Errorf("len(pending) = %d, want 1", len(pending))
	}
}

func TestRunner_Applied(t *testing.T) {
	db := newTestDB(t)
	runner := NewRunner(db)
	ctx := context.Background()

	runner.Add(
		NewMigration("001", "first").WithUpSQL("CREATE TABLE t1 (id INTEGER)"),
	)

	runner.Initialize(ctx)

	applied, _ := runner.Applied(ctx)
	if len(applied) != 0 {
		t.Errorf("len(applied) = %d, want 0", len(applied))
	}

	runner.Up(ctx)

	applied, _ = runner.Applied(ctx)
	if len(applied) != 1 {
		t.Errorf("len(applied) = %d, want 1", len(applied))
	}
	if applied[0].Version != "001" {
		t.Errorf("version = %s, want 001", applied[0].Version)
	}
}

func TestRunner_Status(t *testing.T) {
	db := newTestDB(t)
	runner := NewRunner(db)
	ctx := context.Background()

	runner.Add(
		NewMigration("001", "first").WithUpSQL("CREATE TABLE t1 (id INTEGER)"),
		NewMigration("002", "second").WithUpSQL("CREATE TABLE t2 (id INTEGER)"),
	)

	runner.Initialize(ctx)
	runner.UpOne(ctx)

	statuses, err := runner.Status(ctx)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	if len(statuses) != 2 {
		t.Errorf("len(statuses) = %d, want 2", len(statuses))
	}

	// First should be applied
	if !statuses[0].Applied {
		t.Error("First migration should be applied")
	}
	if !statuses[0].Registered {
		t.Error("First migration should be registered")
	}

	// Second should not be applied
	if statuses[1].Applied {
		t.Error("Second migration should not be applied")
	}
	if !statuses[1].Registered {
		t.Error("Second migration should be registered")
	}
}

func TestRunner_Version(t *testing.T) {
	db := newTestDB(t)
	runner := NewRunner(db)
	ctx := context.Background()

	runner.Add(
		NewMigration("001", "first").WithUpSQL("CREATE TABLE t1 (id INTEGER)"),
		NewMigration("002", "second").WithUpSQL("CREATE TABLE t2 (id INTEGER)"),
	)

	runner.Initialize(ctx)

	// No migrations applied
	version, err := runner.Version(ctx)
	if err != nil {
		t.Fatalf("Version failed: %v", err)
	}
	if version != "" {
		t.Errorf("version = %s, want empty", version)
	}

	// Apply one
	runner.UpOne(ctx)
	version, _ = runner.Version(ctx)
	if version != "001" {
		t.Errorf("version = %s, want 001", version)
	}

	// Apply all
	runner.Up(ctx)
	version, _ = runner.Version(ctx)
	if version != "002" {
		t.Errorf("version = %s, want 002", version)
	}
}

func TestRunner_WithFunc(t *testing.T) {
	db := newTestDB(t)
	runner := NewRunner(db)
	ctx := context.Background()

	upCalled := false
	downCalled := false

	runner.Add(
		NewMigration("001", "with_func").
			WithUpFunc(func(tx Executor) error {
				upCalled = true
				return tx.Exec("CREATE TABLE func_table (id INTEGER)")
			}).
			WithDownFunc(func(tx Executor) error {
				downCalled = true
				return tx.Exec("DROP TABLE func_table")
			}),
	)

	runner.Initialize(ctx)

	// Test up
	runner.Up(ctx)
	if !upCalled {
		t.Error("UpFunc should be called")
	}

	// Test down
	runner.Down(ctx)
	if !downCalled {
		t.Error("DownFunc should be called")
	}
}

func TestRunner_MultipleStatements(t *testing.T) {
	db := newTestDB(t)
	runner := NewRunner(db)
	ctx := context.Background()

	runner.Add(
		NewMigration("001", "multiple").
			WithUpSQL(`
				CREATE TABLE t1 (id INTEGER);
				CREATE TABLE t2 (id INTEGER);
				CREATE TABLE t3 (id INTEGER);
			`),
	)

	runner.Initialize(ctx)

	_, err := runner.Up(ctx)
	if err != nil {
		t.Fatalf("Up failed: %v", err)
	}

	// Verify all tables exist
	var count int
	db.Get(&count, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('t1', 't2', 't3')")
	if count != 3 {
		t.Errorf("table count = %d, want 3", count)
	}
}

func TestRunner_UpOne_Empty(t *testing.T) {
	db := newTestDB(t)
	runner := NewRunner(db)
	ctx := context.Background()

	runner.Initialize(ctx)

	// No migrations to run
	err := runner.UpOne(ctx)
	if err != nil {
		t.Errorf("UpOne should not fail with no migrations: %v", err)
	}
}

func TestRunner_Down_Empty(t *testing.T) {
	db := newTestDB(t)
	runner := NewRunner(db)
	ctx := context.Background()

	runner.Initialize(ctx)

	// No migrations to revert
	err := runner.Down(ctx)
	if err != nil {
		t.Errorf("Down should not fail with no applied migrations: %v", err)
	}
}

func TestRunner_Down_NotFound(t *testing.T) {
	db := newTestDB(t)
	runner := NewRunner(db)
	ctx := context.Background()

	runner.Add(
		NewMigration("001", "first").WithUpSQL("CREATE TABLE t1 (id INTEGER)"),
	)

	runner.Initialize(ctx)
	runner.Up(ctx)

	// Remove migration from registry
	runner.migrations = MigrationList{}

	err := runner.Down(ctx)
	if err == nil {
		t.Error("Down should fail when migration not found")
	}
}
