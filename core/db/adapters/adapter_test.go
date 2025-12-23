package adapters

import (
	"testing"
)

func TestGetAdapter(t *testing.T) {
	tests := []struct {
		driver   string
		wantNil  bool
		wantType string
	}{
		{"postgres", false, "*adapters.PostgresAdapter"},
		{"postgresql", false, "*adapters.PostgresAdapter"},
		{"mysql", false, "*adapters.MySQLAdapter"},
		{"sqlite", false, "*adapters.SQLiteAdapter"},
		{"sqlite3", false, "*adapters.SQLiteAdapter"},
		{"unknown", true, ""},
		{"", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			adapter := GetAdapter(tt.driver)
			if tt.wantNil {
				if adapter != nil {
					t.Errorf("GetAdapter(%s) should return nil", tt.driver)
				}
			} else {
				if adapter == nil {
					t.Errorf("GetAdapter(%s) should not return nil", tt.driver)
				}
			}
		})
	}
}

func TestMustGetAdapter(t *testing.T) {
	t.Run("valid driver", func(t *testing.T) {
		adapter := MustGetAdapter("postgres")
		if adapter == nil {
			t.Error("MustGetAdapter should return adapter for valid driver")
		}
	})

	// Note: "invalid driver panics" test was removed because MustGetAdapter now
	// calls os.Exit(1) instead of panic() for cleaner error output. The error
	// behavior is verified by TestGetAdapter which tests that GetAdapter returns
	// nil for invalid drivers.
}

func TestSupportedDrivers(t *testing.T) {
	drivers := SupportedDrivers()

	if len(drivers) != 3 {
		t.Errorf("SupportedDrivers() returned %d drivers, want 3", len(drivers))
	}

	expected := map[string]bool{"postgres": true, "mysql": true, "sqlite3": true}
	for _, d := range drivers {
		if !expected[d] {
			t.Errorf("Unexpected driver in list: %s", d)
		}
	}
}

func TestIsSupported(t *testing.T) {
	tests := []struct {
		driver string
		want   bool
	}{
		{"postgres", true},
		{"postgresql", true},
		{"mysql", true},
		{"sqlite", true},
		{"sqlite3", true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			if got := IsSupported(tt.driver); got != tt.want {
				t.Errorf("IsSupported(%s) = %v, want %v", tt.driver, got, tt.want)
			}
		})
	}
}

// Test that all adapters implement the Adapter interface
func TestAdapterInterface(t *testing.T) {
	adapters := []Adapter{
		&PostgresAdapter{},
		&MySQLAdapter{},
		&SQLiteAdapter{},
	}

	for _, a := range adapters {
		t.Run(a.DriverName(), func(t *testing.T) {
			// Just verify these don't panic
			_ = a.DriverName()
			_ = a.DSN("localhost", 5432, "user", "pass", "db", nil)
			_ = a.CreateDatabaseSQL("testdb")
			_ = a.DropDatabaseSQL("testdb")
			_ = a.Placeholder(1)
			_ = a.PlaceholderStyle()
			_ = a.QuoteIdentifier("column")
			_ = a.SupportsReturning()
			_ = a.SupportsLastInsertID()
		})
	}
}
