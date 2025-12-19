package db

import (
	"context"
	"testing"
	"time"
)

// Test model for repository tests
type TestCat struct {
	Model
	Name string `db:"name"`
	Type string `db:"type"`
	Age  int    `db:"age"`
}

func (c *TestCat) TableName() string {
	return "cats"
}

const testCatSchema = `
CREATE TABLE IF NOT EXISTS cats (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT,
    age INTEGER DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    deleted_at DATETIME
);
`

func setupTestDB(t *testing.T) *Client {
	t.Helper()

	client := newTestClient(t)

	_, err := client.ExecContext(context.Background(), testCatSchema)
	if err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	return client
}

func TestNewRepository(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)

	if repo == nil {
		t.Fatal("NewRepository returned nil")
	}
	if repo.TableName() != "cats" {
		t.Errorf("TableName() = %s, want cats", repo.TableName())
	}
	if repo.Client() != client {
		t.Error("Client() should return the client")
	}
}

func TestRepository_New(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)

	record := repo.New()

	if record == nil {
		t.Fatal("New() returned nil")
	}
	if record.Model() == nil {
		t.Error("New() should create a model")
	}
	if !record.IsNew() {
		t.Error("New record should be new")
	}
}

func TestRepository_Wrap(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)

	cat := &TestCat{Name: "Whiskers", Type: "Tabby"}
	record := repo.Wrap(cat)

	if record.Model() != cat {
		t.Error("Wrap should use the provided model")
	}
}

func TestRepository_Create(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)

	cat := &TestCat{Name: "Whiskers", Type: "Tabby", Age: 3}

	err := repo.Create(context.Background(), cat)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if cat.ID == "" {
		t.Error("Create should set ID")
	}
	if cat.CreatedAt.IsZero() {
		t.Error("Create should set CreatedAt")
	}
	if cat.UpdatedAt.IsZero() {
		t.Error("Create should set UpdatedAt")
	}
}

func TestRepository_Create_GeneratesID(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)

	cat := &TestCat{Name: "Fluffy"}

	err := repo.Create(context.Background(), cat)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if len(cat.ID) != 36 { // UUID length
		t.Errorf("ID should be a UUID, got length %d", len(cat.ID))
	}
}

func TestRepository_Create_SetsTimestamps(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)

	before := time.Now()
	cat := &TestCat{Name: "Mittens"}
	err := repo.Create(context.Background(), cat)
	after := time.Now()

	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if cat.CreatedAt.Before(before) || cat.CreatedAt.After(after) {
		t.Error("CreatedAt should be set to current time")
	}
	if cat.UpdatedAt.Before(before) || cat.UpdatedAt.After(after) {
		t.Error("UpdatedAt should be set to current time")
	}
}

func TestRepository_Update(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	cat := &TestCat{Name: "Original"}
	repo.Create(ctx, cat)

	originalUpdatedAt := cat.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	cat.Name = "Updated"
	err := repo.Update(ctx, cat)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if cat.UpdatedAt.Equal(originalUpdatedAt) {
		t.Error("Update should change UpdatedAt")
	}

	// Verify in database
	found, _ := repo.FindByID(ctx, cat.ID)
	if found.Model().Name != "Updated" {
		t.Errorf("Name = %s, want Updated", found.Model().Name)
	}
}

func TestRepository_Update_NotFound(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)

	cat := &TestCat{
		Model: Model{ID: "non-existent-id", CreatedAt: time.Now()},
		Name:  "Ghost",
	}

	err := repo.Update(context.Background(), cat)
	if err == nil {
		t.Error("Update should fail for non-existent record")
	}
	if !IsNotFound(err) {
		t.Errorf("Expected ErrNotFound, got: %v", err)
	}
}

func TestRepository_Update_NotPersisted(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)

	cat := &TestCat{Name: "New Cat"}

	err := repo.Update(context.Background(), cat)
	if err != ErrNotPersisted {
		t.Errorf("Expected ErrNotPersisted, got: %v", err)
	}
}

func TestRepository_Save(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	t.Run("creates new record", func(t *testing.T) {
		cat := &TestCat{Name: "NewCat"}
		err := repo.Save(ctx, cat)
		if err != nil {
			t.Fatalf("Save failed: %v", err)
		}
		if cat.ID == "" {
			t.Error("Save should create record with ID")
		}
	})

	t.Run("updates existing record", func(t *testing.T) {
		cat := &TestCat{Name: "Original"}
		repo.Create(ctx, cat)

		cat.Name = "Updated"
		err := repo.Save(ctx, cat)
		if err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		found, _ := repo.FindByID(ctx, cat.ID)
		if found.Model().Name != "Updated" {
			t.Error("Save should update existing record")
		}
	})
}

func TestRepository_Delete(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	cat := &TestCat{Name: "ToDelete"}
	repo.Create(ctx, cat)

	err := repo.Delete(ctx, cat)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if cat.DeletedAt == nil {
		t.Error("Delete should set DeletedAt")
	}

	// Should not be findable
	_, err = repo.FindByID(ctx, cat.ID)
	if !IsNotFound(err) {
		t.Error("Deleted record should not be found")
	}
}

func TestRepository_Delete_NotFound(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)

	cat := &TestCat{
		Model: Model{ID: "non-existent"},
		Name:  "Ghost",
	}

	err := repo.Delete(context.Background(), cat)
	if !IsNotFound(err) {
		t.Errorf("Expected ErrNotFound, got: %v", err)
	}
}

func TestRepository_Delete_NotPersisted(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)

	cat := &TestCat{Name: "New"}

	err := repo.Delete(context.Background(), cat)
	if err != ErrNotPersisted {
		t.Errorf("Expected ErrNotPersisted, got: %v", err)
	}
}

func TestRepository_HardDelete(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	cat := &TestCat{Name: "ToHardDelete"}
	repo.Create(ctx, cat)
	id := cat.ID

	err := repo.HardDelete(ctx, cat)
	if err != nil {
		t.Fatalf("HardDelete failed: %v", err)
	}

	// Should not exist even with WithDeleted
	records, _ := repo.FindAll(ctx, WithDeleted(), Where("id = ?", id))
	if len(records) != 0 {
		t.Error("HardDelete should permanently remove record")
	}
}

func TestRepository_Restore(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	cat := &TestCat{Name: "ToRestore"}
	repo.Create(ctx, cat)
	repo.Delete(ctx, cat)

	err := repo.Restore(ctx, cat)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	if cat.DeletedAt != nil {
		t.Error("Restore should clear DeletedAt")
	}

	// Should be findable again
	_, err = repo.FindByID(ctx, cat.ID)
	if err != nil {
		t.Error("Restored record should be findable")
	}
}

func TestRepository_FindByID(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	cat := &TestCat{Name: "Findable", Type: "Persian", Age: 5}
	repo.Create(ctx, cat)

	record, err := repo.FindByID(ctx, cat.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if record.Model().Name != "Findable" {
		t.Errorf("Name = %s, want Findable", record.Model().Name)
	}
	if record.Model().Type != "Persian" {
		t.Errorf("Type = %s, want Persian", record.Model().Type)
	}
	if record.Model().Age != 5 {
		t.Errorf("Age = %d, want 5", record.Model().Age)
	}
}

func TestRepository_FindByID_NotFound(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)

	_, err := repo.FindByID(context.Background(), "non-existent")
	if !IsNotFound(err) {
		t.Errorf("Expected ErrNotFound, got: %v", err)
	}
}

func TestRepository_FindByID_ExcludesDeleted(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	cat := &TestCat{Name: "Deleted"}
	repo.Create(ctx, cat)
	repo.Delete(ctx, cat)

	_, err := repo.FindByID(ctx, cat.ID)
	if !IsNotFound(err) {
		t.Error("FindByID should not return deleted records")
	}
}

func TestRepository_FindAll(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	// Create test data
	repo.Create(ctx, &TestCat{Name: "Cat1", Type: "Tabby", Age: 1})
	repo.Create(ctx, &TestCat{Name: "Cat2", Type: "Tabby", Age: 2})
	repo.Create(ctx, &TestCat{Name: "Cat3", Type: "Persian", Age: 3})

	records, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}

	if len(records) != 3 {
		t.Errorf("len(records) = %d, want 3", len(records))
	}
}

func TestRepository_FindAll_Empty(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)

	records, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}

	if len(records) != 0 {
		t.Errorf("len(records) = %d, want 0", len(records))
	}
}

func TestRepository_FindAll_Where(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	repo.Create(ctx, &TestCat{Name: "Cat1", Type: "Tabby"})
	repo.Create(ctx, &TestCat{Name: "Cat2", Type: "Tabby"})
	repo.Create(ctx, &TestCat{Name: "Cat3", Type: "Persian"})

	records, err := repo.FindAll(ctx, Where("type = ?", "Tabby"))
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("len(records) = %d, want 2", len(records))
	}
}

func TestRepository_FindAll_OrderBy(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	repo.Create(ctx, &TestCat{Name: "Zoe"})
	repo.Create(ctx, &TestCat{Name: "Alice"})
	repo.Create(ctx, &TestCat{Name: "Bob"})

	records, err := repo.FindAll(ctx, OrderByAsc("name"))
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}

	if records[0].Model().Name != "Alice" {
		t.Errorf("First name = %s, want Alice", records[0].Model().Name)
	}
	if records[2].Model().Name != "Zoe" {
		t.Errorf("Last name = %s, want Zoe", records[2].Model().Name)
	}
}

func TestRepository_FindAll_Limit(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		repo.Create(ctx, &TestCat{Name: "Cat"})
	}

	records, err := repo.FindAll(ctx, Limit(3))
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}

	if len(records) != 3 {
		t.Errorf("len(records) = %d, want 3", len(records))
	}
}

func TestRepository_FindAll_Offset(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	repo.Create(ctx, &TestCat{Name: "Cat1"})
	repo.Create(ctx, &TestCat{Name: "Cat2"})
	repo.Create(ctx, &TestCat{Name: "Cat3"})

	records, err := repo.FindAll(ctx, OrderByAsc("name"), Offset(1))
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("len(records) = %d, want 2", len(records))
	}
	if records[0].Model().Name != "Cat2" {
		t.Errorf("First name = %s, want Cat2", records[0].Model().Name)
	}
}

func TestRepository_FindAll_WithDeleted(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	cat1 := &TestCat{Name: "Active"}
	cat2 := &TestCat{Name: "Deleted"}
	repo.Create(ctx, cat1)
	repo.Create(ctx, cat2)
	repo.Delete(ctx, cat2)

	// Without WithDeleted
	records, _ := repo.FindAll(ctx)
	if len(records) != 1 {
		t.Errorf("Without WithDeleted: len = %d, want 1", len(records))
	}

	// With WithDeleted
	records, _ = repo.FindAll(ctx, WithDeleted())
	if len(records) != 2 {
		t.Errorf("With WithDeleted: len = %d, want 2", len(records))
	}
}

func TestRepository_FindOne(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	repo.Create(ctx, &TestCat{Name: "Cat1", Type: "Tabby"})
	repo.Create(ctx, &TestCat{Name: "Cat2", Type: "Tabby"})

	record, err := repo.FindOne(ctx, Where("type = ?", "Tabby"))
	if err != nil {
		t.Fatalf("FindOne failed: %v", err)
	}

	if record.Model().Type != "Tabby" {
		t.Error("FindOne should return matching record")
	}
}

func TestRepository_First(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	repo.Create(ctx, &TestCat{Name: "First"})
	time.Sleep(10 * time.Millisecond)
	repo.Create(ctx, &TestCat{Name: "Second"})

	record, err := repo.First(ctx)
	if err != nil {
		t.Fatalf("First failed: %v", err)
	}

	if record.Model().Name != "First" {
		t.Errorf("Name = %s, want First", record.Model().Name)
	}
}

func TestRepository_Last(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	repo.Create(ctx, &TestCat{Name: "First"})
	time.Sleep(10 * time.Millisecond)
	repo.Create(ctx, &TestCat{Name: "Last"})

	record, err := repo.Last(ctx)
	if err != nil {
		t.Fatalf("Last failed: %v", err)
	}

	if record.Model().Name != "Last" {
		t.Errorf("Name = %s, want Last", record.Model().Name)
	}
}

func TestRepository_Count(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		repo.Create(ctx, &TestCat{Name: "Cat", Type: "Tabby"})
	}
	repo.Create(ctx, &TestCat{Name: "Cat", Type: "Persian"})

	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}

	if count != 6 {
		t.Errorf("Count = %d, want 6", count)
	}
}

func TestRepository_Count_Where(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	repo.Create(ctx, &TestCat{Name: "Cat1", Type: "Tabby"})
	repo.Create(ctx, &TestCat{Name: "Cat2", Type: "Tabby"})
	repo.Create(ctx, &TestCat{Name: "Cat3", Type: "Persian"})

	count, err := repo.Count(ctx, Where("type = ?", "Tabby"))
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Count = %d, want 2", count)
	}
}

func TestRepository_Exists(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	cat := &TestCat{Name: "Existing"}
	repo.Create(ctx, cat)

	exists, err := repo.Exists(ctx, cat.ID)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}

	if !exists {
		t.Error("Exists should return true for existing record")
	}
}

func TestRepository_Exists_NotFound(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)

	exists, err := repo.Exists(context.Background(), "non-existent")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}

	if exists {
		t.Error("Exists should return false for non-existent record")
	}
}

func TestRepository_ExistsWhere(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	repo.Create(ctx, &TestCat{Name: "Cat", Type: "Tabby"})

	exists, _ := repo.ExistsWhere(ctx, Where("type = ?", "Tabby"))
	if !exists {
		t.Error("ExistsWhere should return true")
	}

	exists, _ = repo.ExistsWhere(ctx, Where("type = ?", "NonExistent"))
	if exists {
		t.Error("ExistsWhere should return false")
	}
}

func TestRepository_DeleteWhere(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	repo.Create(ctx, &TestCat{Name: "Cat1", Type: "Tabby"})
	repo.Create(ctx, &TestCat{Name: "Cat2", Type: "Tabby"})
	repo.Create(ctx, &TestCat{Name: "Cat3", Type: "Persian"})

	affected, err := repo.DeleteWhere(ctx, Where("type = ?", "Tabby"))
	if err != nil {
		t.Fatalf("DeleteWhere failed: %v", err)
	}

	if affected != 2 {
		t.Errorf("affected = %d, want 2", affected)
	}

	count, _ := repo.Count(ctx)
	if count != 1 {
		t.Errorf("remaining count = %d, want 1", count)
	}
}

func TestRepository_UpdateWhere(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	repo.Create(ctx, &TestCat{Name: "Cat1", Type: "Tabby", Age: 1})
	repo.Create(ctx, &TestCat{Name: "Cat2", Type: "Tabby", Age: 2})
	repo.Create(ctx, &TestCat{Name: "Cat3", Type: "Persian", Age: 3})

	affected, err := repo.UpdateWhere(ctx,
		map[string]any{"age": 10},
		Where("type = ?", "Tabby"))

	if err != nil {
		t.Fatalf("UpdateWhere failed: %v", err)
	}

	if affected != 2 {
		t.Errorf("affected = %d, want 2", affected)
	}
}

func TestRepository_Transaction(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	t.Run("successful transaction", func(t *testing.T) {
		err := repo.Transaction(ctx, func(txRepo *TxRepository[*TestCat]) error {
			cat1 := &TestCat{Name: "TxCat1"}
			if err := txRepo.Create(ctx, cat1); err != nil {
				return err
			}

			cat2 := &TestCat{Name: "TxCat2"}
			return txRepo.Create(ctx, cat2)
		})

		if err != nil {
			t.Fatalf("Transaction failed: %v", err)
		}

		count, _ := repo.Count(ctx, Where("name LIKE ?", "TxCat%"))
		if count != 2 {
			t.Errorf("count = %d, want 2", count)
		}
	})

	t.Run("rolled back transaction", func(t *testing.T) {
		initialCount, _ := repo.Count(ctx)

		err := repo.Transaction(ctx, func(txRepo *TxRepository[*TestCat]) error {
			cat := &TestCat{Name: "Rollback"}
			if err := txRepo.Create(ctx, cat); err != nil {
				return err
			}
			return ErrNotFound // Force rollback
		})

		if err == nil {
			t.Error("Transaction should return error")
		}

		finalCount, _ := repo.Count(ctx)
		if finalCount != initialCount {
			t.Error("Rolled back transaction should not change count")
		}
	})
}

func TestTxRepository_FindByID(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	cat := &TestCat{Name: "TxFind"}
	repo.Create(ctx, cat)

	err := repo.Transaction(ctx, func(txRepo *TxRepository[*TestCat]) error {
		record, err := txRepo.FindByID(ctx, cat.ID)
		if err != nil {
			return err
		}
		if record.Model().Name != "TxFind" {
			t.Errorf("Name = %s, want TxFind", record.Model().Name)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}
}

func TestTxRepository_Update(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRepository[*TestCat](client)
	ctx := context.Background()

	cat := &TestCat{Name: "Original"}
	repo.Create(ctx, cat)

	err := repo.Transaction(ctx, func(txRepo *TxRepository[*TestCat]) error {
		cat.Name = "UpdatedInTx"
		return txRepo.Update(ctx, cat)
	})

	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}

	record, _ := repo.FindByID(ctx, cat.ID)
	if record.Model().Name != "UpdatedInTx" {
		t.Errorf("Name = %s, want UpdatedInTx", record.Model().Name)
	}
}
