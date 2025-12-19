package migrations

import (
	"sort"
	"testing"
	"time"
)

func TestNewMigration(t *testing.T) {
	m := NewMigration("20240101120000", "create_users")

	if m.Version != "20240101120000" {
		t.Errorf("Version = %s, want 20240101120000", m.Version)
	}
	if m.Name != "create_users" {
		t.Errorf("Name = %s, want create_users", m.Name)
	}
}

func TestMigration_WithUpSQL(t *testing.T) {
	m := NewMigration("1", "test").WithUpSQL("CREATE TABLE test (id INT)")

	if m.UpSQL != "CREATE TABLE test (id INT)" {
		t.Error("WithUpSQL should set UpSQL")
	}
}

func TestMigration_WithDownSQL(t *testing.T) {
	m := NewMigration("1", "test").WithDownSQL("DROP TABLE test")

	if m.DownSQL != "DROP TABLE test" {
		t.Error("WithDownSQL should set DownSQL")
	}
}

func TestMigration_WithUpFunc(t *testing.T) {
	fn := func(tx Executor) error { return nil }
	m := NewMigration("1", "test").WithUpFunc(fn)

	if m.UpFunc == nil {
		t.Error("WithUpFunc should set UpFunc")
	}
}

func TestMigration_WithDownFunc(t *testing.T) {
	fn := func(tx Executor) error { return nil }
	m := NewMigration("1", "test").WithDownFunc(fn)

	if m.DownFunc == nil {
		t.Error("WithDownFunc should set DownFunc")
	}
}

func TestMigration_Chaining(t *testing.T) {
	m := NewMigration("1", "test").
		WithUpSQL("CREATE TABLE").
		WithDownSQL("DROP TABLE")

	if m.UpSQL == "" || m.DownSQL == "" {
		t.Error("Chaining should work")
	}
}

func TestMigration_HasUp(t *testing.T) {
	t.Run("with SQL", func(t *testing.T) {
		m := NewMigration("1", "test").WithUpSQL("CREATE TABLE")
		if !m.HasUp() {
			t.Error("HasUp should return true with UpSQL")
		}
	})

	t.Run("with func", func(t *testing.T) {
		m := NewMigration("1", "test").WithUpFunc(func(tx Executor) error { return nil })
		if !m.HasUp() {
			t.Error("HasUp should return true with UpFunc")
		}
	})

	t.Run("without", func(t *testing.T) {
		m := NewMigration("1", "test")
		if m.HasUp() {
			t.Error("HasUp should return false without UpSQL or UpFunc")
		}
	})
}

func TestMigration_HasDown(t *testing.T) {
	t.Run("with SQL", func(t *testing.T) {
		m := NewMigration("1", "test").WithDownSQL("DROP TABLE")
		if !m.HasDown() {
			t.Error("HasDown should return true with DownSQL")
		}
	})

	t.Run("with func", func(t *testing.T) {
		m := NewMigration("1", "test").WithDownFunc(func(tx Executor) error { return nil })
		if !m.HasDown() {
			t.Error("HasDown should return true with DownFunc")
		}
	})

	t.Run("without", func(t *testing.T) {
		m := NewMigration("1", "test")
		if m.HasDown() {
			t.Error("HasDown should return false without DownSQL or DownFunc")
		}
	})
}

func TestMigration_FullName(t *testing.T) {
	t.Run("with name", func(t *testing.T) {
		m := NewMigration("20240101", "create_users")
		if m.FullName() != "20240101_create_users" {
			t.Errorf("FullName() = %s, want 20240101_create_users", m.FullName())
		}
	})

	t.Run("without name", func(t *testing.T) {
		m := NewMigration("20240101", "")
		if m.FullName() != "20240101" {
			t.Errorf("FullName() = %s, want 20240101", m.FullName())
		}
	})
}

func TestGenerateVersion(t *testing.T) {
	before := time.Now().Format("20060102150405")
	version := GenerateVersion()
	after := time.Now().Format("20060102150405")

	if version < before || version > after {
		t.Errorf("GenerateVersion() = %s, should be between %s and %s", version, before, after)
	}

	if len(version) != 14 {
		t.Errorf("GenerateVersion() length = %d, want 14", len(version))
	}
}

func TestMigrationList_Sort(t *testing.T) {
	list := MigrationList{
		NewMigration("20240103", "third"),
		NewMigration("20240101", "first"),
		NewMigration("20240102", "second"),
	}

	sort.Sort(list)

	if list[0].Version != "20240101" {
		t.Errorf("First version = %s, want 20240101", list[0].Version)
	}
	if list[1].Version != "20240102" {
		t.Errorf("Second version = %s, want 20240102", list[1].Version)
	}
	if list[2].Version != "20240103" {
		t.Errorf("Third version = %s, want 20240103", list[2].Version)
	}
}

func TestMigrationList_Len(t *testing.T) {
	list := MigrationList{
		NewMigration("1", ""),
		NewMigration("2", ""),
	}

	if list.Len() != 2 {
		t.Errorf("Len() = %d, want 2", list.Len())
	}
}

func TestMigrationList_Swap(t *testing.T) {
	list := MigrationList{
		NewMigration("1", "first"),
		NewMigration("2", "second"),
	}

	list.Swap(0, 1)

	if list[0].Version != "2" || list[1].Version != "1" {
		t.Error("Swap should swap elements")
	}
}

func TestDirection(t *testing.T) {
	if Up != "up" {
		t.Errorf("Up = %s, want up", Up)
	}
	if Down != "down" {
		t.Errorf("Down = %s, want down", Down)
	}
}
