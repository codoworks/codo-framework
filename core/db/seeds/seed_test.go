package seeds

import (
	"context"
	"errors"
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

func TestNewSeed(t *testing.T) {
	seed := NewSeed("test_seed", func(ctx context.Context, db *sqlx.DB) error {
		return nil
	})

	if seed.Name != "test_seed" {
		t.Errorf("Name = %s, want test_seed", seed.Name)
	}
	if seed.Run == nil {
		t.Error("Run should not be nil")
	}
}

func TestNewSeeder(t *testing.T) {
	db := newTestDB(t)
	seeder := NewSeeder(db)

	if seeder == nil {
		t.Fatal("NewSeeder returned nil")
	}
	if seeder.Count() != 0 {
		t.Errorf("Count() = %d, want 0", seeder.Count())
	}
}

func TestSeeder_Add(t *testing.T) {
	db := newTestDB(t)
	seeder := NewSeeder(db)

	seed1 := NewSeed("seed1", func(ctx context.Context, db *sqlx.DB) error { return nil })
	seed2 := NewSeed("seed2", func(ctx context.Context, db *sqlx.DB) error { return nil })

	seeder.Add(seed1, seed2)

	if seeder.Count() != 2 {
		t.Errorf("Count() = %d, want 2", seeder.Count())
	}
}

func TestSeeder_Seeds(t *testing.T) {
	db := newTestDB(t)
	seeder := NewSeeder(db)

	seed := NewSeed("test", func(ctx context.Context, db *sqlx.DB) error { return nil })
	seeder.Add(seed)

	seeds := seeder.Seeds()
	if len(seeds) != 1 {
		t.Errorf("len(Seeds()) = %d, want 1", len(seeds))
	}
	if seeds[0].Name != "test" {
		t.Errorf("Name = %s, want test", seeds[0].Name)
	}
}

func TestSeeder_Run(t *testing.T) {
	db := newTestDB(t)
	seeder := NewSeeder(db)
	ctx := context.Background()

	// Create test table
	db.Exec("CREATE TABLE test_seeds (id INTEGER PRIMARY KEY, value TEXT)")

	ran := []string{}

	seeder.Add(
		NewSeed("seed1", func(ctx context.Context, db *sqlx.DB) error {
			ran = append(ran, "seed1")
			_, err := db.Exec("INSERT INTO test_seeds (value) VALUES ('one')")
			return err
		}),
		NewSeed("seed2", func(ctx context.Context, db *sqlx.DB) error {
			ran = append(ran, "seed2")
			_, err := db.Exec("INSERT INTO test_seeds (value) VALUES ('two')")
			return err
		}),
	)

	err := seeder.Run(ctx)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if len(ran) != 2 {
		t.Errorf("len(ran) = %d, want 2", len(ran))
	}

	// Verify data
	var count int
	db.Get(&count, "SELECT COUNT(*) FROM test_seeds")
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func TestSeeder_Run_Error(t *testing.T) {
	db := newTestDB(t)
	seeder := NewSeeder(db)
	ctx := context.Background()

	seeder.Add(
		NewSeed("failing", func(ctx context.Context, db *sqlx.DB) error {
			return errors.New("seed error")
		}),
	)

	err := seeder.Run(ctx)
	if err == nil {
		t.Error("Run should fail when seed fails")
	}
}

func TestSeeder_RunOne(t *testing.T) {
	db := newTestDB(t)
	seeder := NewSeeder(db)
	ctx := context.Background()

	db.Exec("CREATE TABLE test_run_one (id INTEGER PRIMARY KEY, value TEXT)")

	seeder.Add(
		NewSeed("first", func(ctx context.Context, db *sqlx.DB) error {
			_, err := db.Exec("INSERT INTO test_run_one (value) VALUES ('first')")
			return err
		}),
		NewSeed("second", func(ctx context.Context, db *sqlx.DB) error {
			_, err := db.Exec("INSERT INTO test_run_one (value) VALUES ('second')")
			return err
		}),
	)

	err := seeder.RunOne(ctx, "first")
	if err != nil {
		t.Fatalf("RunOne failed: %v", err)
	}

	var count int
	db.Get(&count, "SELECT COUNT(*) FROM test_run_one")
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}

func TestSeeder_RunOne_NotFound(t *testing.T) {
	db := newTestDB(t)
	seeder := NewSeeder(db)

	err := seeder.RunOne(context.Background(), "nonexistent")
	if err == nil {
		t.Error("RunOne should fail for nonexistent seed")
	}
}

func TestSeeder_RunByNames(t *testing.T) {
	db := newTestDB(t)
	seeder := NewSeeder(db)
	ctx := context.Background()

	db.Exec("CREATE TABLE test_by_names (id INTEGER PRIMARY KEY, value TEXT)")

	seeder.Add(
		NewSeed("first", func(ctx context.Context, db *sqlx.DB) error {
			_, err := db.Exec("INSERT INTO test_by_names (value) VALUES ('first')")
			return err
		}),
		NewSeed("second", func(ctx context.Context, db *sqlx.DB) error {
			_, err := db.Exec("INSERT INTO test_by_names (value) VALUES ('second')")
			return err
		}),
		NewSeed("third", func(ctx context.Context, db *sqlx.DB) error {
			_, err := db.Exec("INSERT INTO test_by_names (value) VALUES ('third')")
			return err
		}),
	)

	err := seeder.RunByNames(ctx, "first", "third")
	if err != nil {
		t.Fatalf("RunByNames failed: %v", err)
	}

	var count int
	db.Get(&count, "SELECT COUNT(*) FROM test_by_names")
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func TestSeeder_Clear(t *testing.T) {
	db := newTestDB(t)
	seeder := NewSeeder(db)
	ctx := context.Background()

	db.Exec("CREATE TABLE t1 (id INTEGER PRIMARY KEY)")
	db.Exec("CREATE TABLE t2 (id INTEGER PRIMARY KEY)")
	db.Exec("INSERT INTO t1 (id) VALUES (1), (2)")
	db.Exec("INSERT INTO t2 (id) VALUES (1), (2), (3)")

	err := seeder.Clear(ctx, "t1", "t2")
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	var count1, count2 int
	db.Get(&count1, "SELECT COUNT(*) FROM t1")
	db.Get(&count2, "SELECT COUNT(*) FROM t2")

	if count1 != 0 || count2 != 0 {
		t.Errorf("Tables not cleared: t1=%d, t2=%d", count1, count2)
	}
}

func TestSeeder_Refresh(t *testing.T) {
	db := newTestDB(t)
	seeder := NewSeeder(db)
	ctx := context.Background()

	db.Exec("CREATE TABLE refresh_test (id INTEGER PRIMARY KEY, value TEXT)")
	db.Exec("INSERT INTO refresh_test (value) VALUES ('old1'), ('old2')")

	seeder.Add(
		NewSeed("new_data", func(ctx context.Context, db *sqlx.DB) error {
			_, err := db.Exec("INSERT INTO refresh_test (value) VALUES ('new')")
			return err
		}),
	)

	err := seeder.Refresh(ctx, "refresh_test")
	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}

	var count int
	db.Get(&count, "SELECT COUNT(*) FROM refresh_test")
	if count != 1 {
		t.Errorf("count = %d, want 1 (old data cleared, new seed ran)", count)
	}
}

func TestSeeder_HasSeed(t *testing.T) {
	db := newTestDB(t)
	seeder := NewSeeder(db)

	seeder.Add(
		NewSeed("exists", func(ctx context.Context, db *sqlx.DB) error { return nil }),
	)

	if !seeder.HasSeed("exists") {
		t.Error("HasSeed should return true for existing seed")
	}
	if seeder.HasSeed("notexists") {
		t.Error("HasSeed should return false for non-existing seed")
	}
}

func TestSeeder_Chaining(t *testing.T) {
	db := newTestDB(t)

	seed1 := NewSeed("s1", func(ctx context.Context, db *sqlx.DB) error { return nil })
	seed2 := NewSeed("s2", func(ctx context.Context, db *sqlx.DB) error { return nil })

	seeder := NewSeeder(db).Add(seed1).Add(seed2)

	if seeder.Count() != 2 {
		t.Errorf("Chaining should work: Count() = %d", seeder.Count())
	}
}

func TestSeeder_RunByNames_Error(t *testing.T) {
	db := newTestDB(t)
	seeder := NewSeeder(db)
	ctx := context.Background()

	seeder.Add(
		NewSeed("good", func(ctx context.Context, db *sqlx.DB) error {
			return nil
		}),
		NewSeed("bad", func(ctx context.Context, db *sqlx.DB) error {
			return errors.New("seed failed")
		}),
	)

	err := seeder.RunByNames(ctx, "good", "bad")
	if err == nil {
		t.Error("RunByNames should fail when a seed fails")
	}
	if err != nil && !errors.Is(err, errors.New("seed failed")) {
		// Check error message contains expected info
		if !contains(err.Error(), "seed bad failed") {
			t.Errorf("Error should mention failed seed: %v", err)
		}
	}
}

func TestSeeder_Clear_Error(t *testing.T) {
	db := newTestDB(t)
	seeder := NewSeeder(db)
	ctx := context.Background()

	// Try to clear a non-existent table
	err := seeder.Clear(ctx, "nonexistent_table")
	if err == nil {
		t.Error("Clear should fail for non-existent table")
	}
}

func TestSeeder_Refresh_ClearError(t *testing.T) {
	db := newTestDB(t)
	seeder := NewSeeder(db)
	ctx := context.Background()

	seeder.Add(
		NewSeed("test", func(ctx context.Context, db *sqlx.DB) error { return nil }),
	)

	// Try to refresh with a non-existent table (Clear should fail)
	err := seeder.Refresh(ctx, "nonexistent_table")
	if err == nil {
		t.Error("Refresh should fail when Clear fails")
	}
}

func TestSeeder_Refresh_RunError(t *testing.T) {
	db := newTestDB(t)
	seeder := NewSeeder(db)
	ctx := context.Background()

	// Create a table that can be cleared
	db.Exec("CREATE TABLE refresh_err_test (id INTEGER PRIMARY KEY)")

	seeder.Add(
		NewSeed("failing", func(ctx context.Context, db *sqlx.DB) error {
			return errors.New("seed run failed")
		}),
	)

	// Clear succeeds but Run fails
	err := seeder.Refresh(ctx, "refresh_err_test")
	if err == nil {
		t.Error("Refresh should fail when Run fails")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
