package db

import (
	"context"
	"testing"
	"time"
)

func TestRecord_Model(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)

	cat := &TestCat{Name: "Fluffy"}
	record := repo.Wrap(cat)

	if record.Model() != cat {
		t.Error("Model() should return the wrapped model")
	}
}

func TestRecord_Repository(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)

	record := repo.New()

	if record.Repository() != repo {
		t.Error("Repository() should return the associated repository")
	}
}

func TestRecord_ID(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	cat := &TestCat{Name: "Test"}
	repo.Create(ctx, cat)
	record := repo.Wrap(cat)

	if record.ID() != cat.ID {
		t.Errorf("ID() = %s, want %s", record.ID(), cat.ID)
	}
}

func TestRecord_IsNew(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	t.Run("new record", func(t *testing.T) {
		record := repo.New()
		if !record.IsNew() {
			t.Error("New record should be new")
		}
	})

	t.Run("persisted record", func(t *testing.T) {
		cat := &TestCat{Name: "Persisted"}
		repo.Create(ctx, cat)
		record := repo.Wrap(cat)

		if record.IsNew() {
			t.Error("Persisted record should not be new")
		}
	})
}

func TestRecord_IsPersisted(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	t.Run("new record", func(t *testing.T) {
		record := repo.New()
		if record.IsPersisted() {
			t.Error("New record should not be persisted")
		}
	})

	t.Run("persisted record", func(t *testing.T) {
		cat := &TestCat{Name: "Persisted"}
		repo.Create(ctx, cat)
		record := repo.Wrap(cat)

		if !record.IsPersisted() {
			t.Error("Saved record should be persisted")
		}
	})
}

func TestRecord_IsDeleted(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	cat := &TestCat{Name: "ToDelete"}
	repo.Create(ctx, cat)
	record := repo.Wrap(cat)

	if record.IsDeleted() {
		t.Error("Record should not be deleted initially")
	}

	record.Delete(ctx)

	if !record.IsDeleted() {
		t.Error("Record should be deleted after Delete()")
	}
}

func TestRecord_Save_New(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	cat := &TestCat{Name: "SaveNew"}
	record := repo.Wrap(cat)

	err := record.Save(ctx)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if cat.ID == "" {
		t.Error("Save should set ID on new record")
	}

	// Verify in database
	found, _ := repo.FindByID(ctx, cat.ID)
	if found.Model().Name != "SaveNew" {
		t.Error("Record should be saved in database")
	}
}

func TestRecord_Save_Existing(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	cat := &TestCat{Name: "Original"}
	repo.Create(ctx, cat)
	record := repo.Wrap(cat)

	originalID := cat.ID
	cat.Name = "Updated"

	err := record.Save(ctx)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if cat.ID != originalID {
		t.Error("Save should not change ID of existing record")
	}

	// Verify in database
	found, _ := repo.FindByID(ctx, cat.ID)
	if found.Model().Name != "Updated" {
		t.Error("Record should be updated in database")
	}
}

func TestRecord_Create(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	cat := &TestCat{Name: "CreateTest"}
	record := repo.Wrap(cat)

	err := record.Create(ctx)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if cat.ID == "" {
		t.Error("Create should set ID")
	}
}

func TestRecord_Update(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	cat := &TestCat{Name: "Original"}
	repo.Create(ctx, cat)
	record := repo.Wrap(cat)

	cat.Name = "Updated"
	err := record.Update(ctx)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, _ := repo.FindByID(ctx, cat.ID)
	if found.Model().Name != "Updated" {
		t.Error("Update should persist changes")
	}
}

func TestRecord_Delete(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	cat := &TestCat{Name: "ToDelete"}
	repo.Create(ctx, cat)
	record := repo.Wrap(cat)

	err := record.Delete(ctx)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Should not be findable
	_, err = repo.FindByID(ctx, cat.ID)
	if !IsNotFound(err) {
		t.Error("Deleted record should not be findable")
	}
}

func TestRecord_HardDelete(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	cat := &TestCat{Name: "ToHardDelete"}
	repo.Create(ctx, cat)
	id := cat.ID
	record := repo.Wrap(cat)

	err := record.HardDelete(ctx)
	if err != nil {
		t.Fatalf("HardDelete failed: %v", err)
	}

	// Should not exist even with WithDeleted
	records, _ := repo.FindAll(ctx, WithDeleted(), Where("id = ?", id))
	if len(records) != 0 {
		t.Error("HardDelete should permanently remove record")
	}
}

func TestRecord_Restore(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	cat := &TestCat{Name: "ToRestore"}
	repo.Create(ctx, cat)
	record := repo.Wrap(cat)
	record.Delete(ctx)

	err := record.Restore(ctx)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Should be findable again
	_, err = repo.FindByID(ctx, cat.ID)
	if err != nil {
		t.Error("Restored record should be findable")
	}
}

func TestRecord_Reload(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	cat := &TestCat{Name: "Original"}
	repo.Create(ctx, cat)
	record := repo.Wrap(cat)

	// Update directly in database
	repo.UpdateWhere(ctx, map[string]any{"name": "ChangedInDB"}, Where("id = ?", cat.ID))

	// Record still has old value
	if record.Model().Name != "Original" {
		t.Error("Record should still have old value before reload")
	}

	err := record.Reload(ctx)
	if err != nil {
		t.Fatalf("Reload failed: %v", err)
	}

	if record.Model().Name != "ChangedInDB" {
		t.Errorf("Name after reload = %s, want ChangedInDB", record.Model().Name)
	}
}

func TestRecord_Reload_NotPersisted(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)

	record := repo.New()

	err := record.Reload(context.Background())
	if err != ErrNotPersisted {
		t.Errorf("Expected ErrNotPersisted, got: %v", err)
	}
}

func TestRecord_Touch(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	cat := &TestCat{Name: "TouchTest"}
	repo.Create(ctx, cat)
	record := repo.Wrap(cat)

	originalUpdatedAt := cat.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	err := record.Touch(ctx)
	if err != nil {
		t.Fatalf("Touch failed: %v", err)
	}

	if !cat.UpdatedAt.After(originalUpdatedAt) {
		t.Error("Touch should update UpdatedAt")
	}
}

func TestRecord_Touch_NotPersisted(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)

	record := repo.New()

	err := record.Touch(context.Background())
	if err != ErrNotPersisted {
		t.Errorf("Expected ErrNotPersisted, got: %v", err)
	}
}

func TestRecord_Duplicate(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	cat := &TestCat{Name: "Original"}
	repo.Create(ctx, cat)
	record := repo.Wrap(cat)

	duplicate := record.Duplicate()

	if duplicate == nil {
		t.Fatal("Duplicate should return a new record")
	}
	if !duplicate.IsNew() {
		t.Error("Duplicate should be a new record")
	}
	if duplicate == record {
		t.Error("Duplicate should be a different record")
	}
}

// Test record chaining
func TestRecord_Chaining(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	// Create and immediately work with the record
	record := repo.New()
	record.Model().Name = "Chained"
	record.Model().Type = "Tabby"

	if err := record.Save(ctx); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify
	found, err := repo.FindByID(ctx, record.ID())
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if found.Model().Name != "Chained" {
		t.Errorf("Name = %s, want Chained", found.Model().Name)
	}
}
