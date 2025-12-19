package db

import (
	"reflect"
	"strings"
	"testing"
)

func TestNewQueryBuilder(t *testing.T) {
	qb := NewQueryBuilder("users")
	if qb.tableName != "users" {
		t.Errorf("tableName = %s, want users", qb.tableName)
	}
}

func TestQueryBuilder_TableName(t *testing.T) {
	qb := NewQueryBuilder("products")
	if qb.TableName() != "products" {
		t.Errorf("TableName() = %s, want products", qb.TableName())
	}
}

func TestWhere(t *testing.T) {
	qb := NewQueryBuilder("users")
	Where("name = ?", "John")(qb)

	query, args := qb.Build()

	if !strings.Contains(query, "name = ?") {
		t.Errorf("query should contain 'name = ?', got: %s", query)
	}
	if len(args) != 1 || args[0] != "John" {
		t.Errorf("args = %v, want [John]", args)
	}
}

func TestWhere_Multiple(t *testing.T) {
	qb := NewQueryBuilder("users")
	Where("name = ?", "John")(qb)
	Where("age > ?", 18)(qb)

	query, args := qb.Build()

	if !strings.Contains(query, "name = ?") || !strings.Contains(query, "age > ?") {
		t.Errorf("query should contain both conditions, got: %s", query)
	}
	if !strings.Contains(query, " AND ") {
		t.Errorf("query should use AND for multiple conditions, got: %s", query)
	}
	if len(args) != 2 {
		t.Errorf("args length = %d, want 2", len(args))
	}
}

func TestWhere_WithArgs(t *testing.T) {
	qb := NewQueryBuilder("users")
	Where("created_at BETWEEN ? AND ?", "2024-01-01", "2024-12-31")(qb)

	_, args := qb.Build()

	if len(args) != 2 {
		t.Errorf("args length = %d, want 2", len(args))
	}
	if args[0] != "2024-01-01" || args[1] != "2024-12-31" {
		t.Errorf("args = %v, want [2024-01-01, 2024-12-31]", args)
	}
}

func TestWhereEq(t *testing.T) {
	qb := NewQueryBuilder("users")
	WhereEq("status", "active")(qb)

	query, args := qb.Build()

	if !strings.Contains(query, "status = ?") {
		t.Errorf("query should contain 'status = ?', got: %s", query)
	}
	if len(args) != 1 || args[0] != "active" {
		t.Errorf("args = %v, want [active]", args)
	}
}

func TestWhereIn(t *testing.T) {
	t.Run("with values", func(t *testing.T) {
		qb := NewQueryBuilder("users")
		WhereIn("status", "active", "pending", "approved")(qb)

		query, args := qb.Build()

		if !strings.Contains(query, "status IN (?, ?, ?)") {
			t.Errorf("query should contain 'status IN (?, ?, ?)', got: %s", query)
		}
		if len(args) != 3 {
			t.Errorf("args length = %d, want 3", len(args))
		}
	})

	t.Run("empty values", func(t *testing.T) {
		qb := NewQueryBuilder("users")
		WhereIn("status")(qb)

		query, _ := qb.Build()

		if strings.Contains(query, "IN") {
			t.Errorf("query should not contain IN clause with empty values, got: %s", query)
		}
	})
}

func TestWhereNotNull(t *testing.T) {
	qb := NewQueryBuilder("users")
	WhereNotNull("email")(qb)

	query, _ := qb.Build()

	if !strings.Contains(query, "email IS NOT NULL") {
		t.Errorf("query should contain 'email IS NOT NULL', got: %s", query)
	}
}

func TestWhereNull(t *testing.T) {
	qb := NewQueryBuilder("users")
	WhereNull("deleted_at")(qb)

	query, _ := qb.Build()

	// Note: deleted_at IS NULL is added by default, so we have two
	if !strings.Contains(query, "deleted_at IS NULL") {
		t.Errorf("query should contain 'deleted_at IS NULL', got: %s", query)
	}
}

func TestWhereLike(t *testing.T) {
	qb := NewQueryBuilder("users")
	WhereLike("name", "%John%")(qb)

	query, args := qb.Build()

	if !strings.Contains(query, "name LIKE ?") {
		t.Errorf("query should contain 'name LIKE ?', got: %s", query)
	}
	if len(args) != 1 || args[0] != "%John%" {
		t.Errorf("args = %v, want [%%John%%]", args)
	}
}

func TestWhereBetween(t *testing.T) {
	qb := NewQueryBuilder("products")
	WhereBetween("price", 10.0, 100.0)(qb)

	query, args := qb.Build()

	if !strings.Contains(query, "price BETWEEN ? AND ?") {
		t.Errorf("query should contain 'price BETWEEN ? AND ?', got: %s", query)
	}
	if len(args) != 2 {
		t.Errorf("args length = %d, want 2", len(args))
	}
}

func TestOrderBy(t *testing.T) {
	t.Run("ascending", func(t *testing.T) {
		qb := NewQueryBuilder("users")
		OrderBy("name", "ASC")(qb)

		query, _ := qb.Build()

		if !strings.Contains(query, "ORDER BY name ASC") {
			t.Errorf("query should contain 'ORDER BY name ASC', got: %s", query)
		}
	})

	t.Run("descending", func(t *testing.T) {
		qb := NewQueryBuilder("users")
		OrderBy("created_at", "DESC")(qb)

		query, _ := qb.Build()

		if !strings.Contains(query, "ORDER BY created_at DESC") {
			t.Errorf("query should contain 'ORDER BY created_at DESC', got: %s", query)
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		qb := NewQueryBuilder("users")
		OrderBy("name", "desc")(qb)

		query, _ := qb.Build()

		if !strings.Contains(query, "ORDER BY name DESC") {
			t.Errorf("query should normalize to 'ORDER BY name DESC', got: %s", query)
		}
	})
}

func TestOrderBy_Multiple(t *testing.T) {
	qb := NewQueryBuilder("users")
	OrderBy("last_name", "ASC")(qb)
	OrderBy("first_name", "ASC")(qb)

	query, _ := qb.Build()

	if !strings.Contains(query, "ORDER BY last_name ASC, first_name ASC") {
		t.Errorf("query should contain multiple order clauses, got: %s", query)
	}
}

func TestOrderBy_InvalidDirection(t *testing.T) {
	qb := NewQueryBuilder("users")
	OrderBy("name", "INVALID")(qb)

	query, _ := qb.Build()

	if !strings.Contains(query, "ORDER BY name ASC") {
		t.Errorf("invalid direction should default to ASC, got: %s", query)
	}
}

func TestOrderByAsc(t *testing.T) {
	qb := NewQueryBuilder("users")
	OrderByAsc("name")(qb)

	query, _ := qb.Build()

	if !strings.Contains(query, "ORDER BY name ASC") {
		t.Errorf("query should contain 'ORDER BY name ASC', got: %s", query)
	}
}

func TestOrderByDesc(t *testing.T) {
	qb := NewQueryBuilder("users")
	OrderByDesc("created_at")(qb)

	query, _ := qb.Build()

	if !strings.Contains(query, "ORDER BY created_at DESC") {
		t.Errorf("query should contain 'ORDER BY created_at DESC', got: %s", query)
	}
}

func TestGroupBy(t *testing.T) {
	qb := NewQueryBuilder("orders")
	GroupBy("customer_id", "status")(qb)

	query, _ := qb.Build()

	if !strings.Contains(query, "GROUP BY customer_id, status") {
		t.Errorf("query should contain 'GROUP BY customer_id, status', got: %s", query)
	}
}

func TestHaving(t *testing.T) {
	qb := NewQueryBuilder("orders")
	GroupBy("customer_id")(qb)
	Having("COUNT(*) > ?", 5)(qb)

	query, args := qb.Build()

	if !strings.Contains(query, "HAVING COUNT(*) > ?") {
		t.Errorf("query should contain HAVING clause, got: %s", query)
	}
	// Check args include having args
	if len(args) != 1 || args[0] != 5 {
		t.Errorf("args should include HAVING args, got: %v", args)
	}
}

func TestLimit(t *testing.T) {
	t.Run("positive limit", func(t *testing.T) {
		qb := NewQueryBuilder("users")
		Limit(10)(qb)

		query, _ := qb.Build()

		if !strings.Contains(query, "LIMIT 10") {
			t.Errorf("query should contain 'LIMIT 10', got: %s", query)
		}
	})

	t.Run("zero is ignored", func(t *testing.T) {
		qb := NewQueryBuilder("users")
		Limit(0)(qb)

		query, _ := qb.Build()

		if strings.Contains(query, "LIMIT") {
			t.Errorf("zero limit should be ignored, got: %s", query)
		}
	})

	t.Run("negative is ignored", func(t *testing.T) {
		qb := NewQueryBuilder("users")
		Limit(-5)(qb)

		query, _ := qb.Build()

		if strings.Contains(query, "LIMIT") {
			t.Errorf("negative limit should be ignored, got: %s", query)
		}
	})
}

func TestOffset(t *testing.T) {
	t.Run("positive offset", func(t *testing.T) {
		qb := NewQueryBuilder("users")
		Offset(20)(qb)

		query, _ := qb.Build()

		if !strings.Contains(query, "OFFSET 20") {
			t.Errorf("query should contain 'OFFSET 20', got: %s", query)
		}
	})

	t.Run("zero is valid", func(t *testing.T) {
		qb := NewQueryBuilder("users")
		qb.offset = -1 // Set to invalid first
		Offset(0)(qb)

		if qb.offset != 0 {
			t.Errorf("zero offset should be valid, got: %d", qb.offset)
		}
	})

	t.Run("negative is ignored", func(t *testing.T) {
		qb := NewQueryBuilder("users")
		qb.offset = 10
		Offset(-5)(qb)

		if qb.offset != 10 {
			t.Errorf("negative offset should be ignored, offset = %d", qb.offset)
		}
	})
}

func TestPaginate(t *testing.T) {
	t.Run("valid pagination", func(t *testing.T) {
		qb := NewQueryBuilder("users")
		Paginate(2, 25)(qb)

		query, _ := qb.Build()

		if !strings.Contains(query, "LIMIT 25") {
			t.Errorf("query should contain 'LIMIT 25', got: %s", query)
		}
		if !strings.Contains(query, "OFFSET 25") {
			t.Errorf("query should contain 'OFFSET 25' for page 2, got: %s", query)
		}
	})

	t.Run("page 1", func(t *testing.T) {
		qb := NewQueryBuilder("users")
		Paginate(1, 10)(qb)

		if qb.offset != 0 {
			t.Errorf("page 1 should have offset 0, got: %d", qb.offset)
		}
	})

	t.Run("invalid page defaults to 1", func(t *testing.T) {
		qb := NewQueryBuilder("users")
		Paginate(0, 10)(qb)

		if qb.offset != 0 {
			t.Errorf("invalid page should default to 1 (offset 0), got: %d", qb.offset)
		}
	})

	t.Run("invalid perPage defaults to 10", func(t *testing.T) {
		qb := NewQueryBuilder("users")
		Paginate(1, 0)(qb)

		if qb.limit != 10 {
			t.Errorf("invalid perPage should default to 10, got: %d", qb.limit)
		}
	})
}

func TestWithDeleted(t *testing.T) {
	t.Run("includes deleted records", func(t *testing.T) {
		qb := NewQueryBuilder("users")
		WithDeleted()(qb)

		query, _ := qb.Build()

		if strings.Contains(query, "deleted_at IS NULL") {
			t.Errorf("withDeleted should not add deleted_at filter, got: %s", query)
		}
	})

	t.Run("default excludes deleted", func(t *testing.T) {
		qb := NewQueryBuilder("users")

		query, _ := qb.Build()

		if !strings.Contains(query, "deleted_at IS NULL") {
			t.Errorf("default should exclude deleted records, got: %s", query)
		}
	})
}

func TestOnlyDeleted(t *testing.T) {
	qb := NewQueryBuilder("users")
	OnlyDeleted()(qb)

	query, _ := qb.Build()

	if !strings.Contains(query, "deleted_at IS NOT NULL") {
		t.Errorf("onlyDeleted should filter for deleted records, got: %s", query)
	}
	// Should not have the default "deleted_at IS NULL" filter
	if strings.Contains(query, "deleted_at IS NULL") {
		t.Errorf("onlyDeleted should not have deleted_at IS NULL filter, got: %s", query)
	}
}

func TestSelect(t *testing.T) {
	qb := NewQueryBuilder("users")
	Select("id", "name", "email")(qb)

	query, _ := qb.Build()

	if !strings.Contains(query, "SELECT id, name, email FROM") {
		t.Errorf("query should select specific columns, got: %s", query)
	}
}

func TestJoin(t *testing.T) {
	qb := NewQueryBuilder("users")
	Join("LEFT JOIN orders ON users.id = orders.user_id")(qb)

	query, _ := qb.Build()

	if !strings.Contains(query, "LEFT JOIN orders ON users.id = orders.user_id") {
		t.Errorf("query should contain JOIN clause, got: %s", query)
	}
}

func TestQueryBuilder_Apply(t *testing.T) {
	qb := NewQueryBuilder("users")
	qb.Apply(
		Where("status = ?", "active"),
		OrderBy("name", "ASC"),
		Limit(10),
	)

	query, args := qb.Build()

	if !strings.Contains(query, "status = ?") {
		t.Error("Apply should add WHERE condition")
	}
	if !strings.Contains(query, "ORDER BY name ASC") {
		t.Error("Apply should add ORDER BY")
	}
	if !strings.Contains(query, "LIMIT 10") {
		t.Error("Apply should add LIMIT")
	}
	if len(args) != 1 {
		t.Errorf("args length = %d, want 1", len(args))
	}
}

func TestQueryBuilder_Build(t *testing.T) {
	qb := NewQueryBuilder("users")
	qb.Apply(
		Select("id", "name"),
		Where("status = ?", "active"),
		Where("age >= ?", 18),
		OrderBy("name", "ASC"),
		Limit(10),
		Offset(20),
	)

	query, args := qb.Build()

	expected := "SELECT id, name FROM users WHERE deleted_at IS NULL AND status = ? AND age >= ? ORDER BY name ASC LIMIT 10 OFFSET 20"
	if query != expected {
		t.Errorf("Build() =\n%s\nwant:\n%s", query, expected)
	}
	if len(args) != 2 {
		t.Errorf("args length = %d, want 2", len(args))
	}
}

func TestQueryBuilder_Build_Count(t *testing.T) {
	qb := NewQueryBuilder("users")
	qb.Apply(
		Where("status = ?", "active"),
		OrderBy("name", "ASC"), // Should be ignored in count
		Limit(10),              // Should be ignored in count
	)

	query, args := qb.BuildCount()

	if !strings.HasPrefix(query, "SELECT COUNT(*) FROM") {
		t.Errorf("count query should start with SELECT COUNT(*), got: %s", query)
	}
	if strings.Contains(query, "ORDER BY") {
		t.Errorf("count query should not have ORDER BY, got: %s", query)
	}
	if strings.Contains(query, "LIMIT") {
		t.Errorf("count query should not have LIMIT, got: %s", query)
	}
	if len(args) != 1 {
		t.Errorf("args length = %d, want 1", len(args))
	}
}

func TestQueryBuilder_Clone(t *testing.T) {
	original := NewQueryBuilder("users")
	original.Apply(
		Where("status = ?", "active"),
		OrderBy("name", "ASC"),
		Limit(10),
	)

	clone := original.Clone()

	// Modify clone
	clone.Apply(Where("age > ?", 18))

	// Original should be unchanged
	origQuery, origArgs := original.Build()
	cloneQuery, cloneArgs := clone.Build()

	if origQuery == cloneQuery {
		t.Error("Clone should be independent from original")
	}
	if len(origArgs) != 1 {
		t.Errorf("original args should be unchanged, got %d", len(origArgs))
	}
	if len(cloneArgs) != 2 {
		t.Errorf("clone args should have additional arg, got %d", len(cloneArgs))
	}
}

func TestQueryBuilder_Reset(t *testing.T) {
	qb := NewQueryBuilder("users")
	qb.Apply(
		Where("status = ?", "active"),
		OrderBy("name", "ASC"),
		Limit(10),
		Offset(20),
		WithDeleted(),
	)

	qb.Reset()

	query, args := qb.Build()

	// After reset, should only have default deleted_at filter
	expected := "SELECT * FROM users WHERE deleted_at IS NULL"
	if query != expected {
		t.Errorf("Reset() should clear all options, got: %s", query)
	}
	if len(args) != 0 {
		t.Errorf("Reset() should clear args, got: %v", args)
	}
}

func TestQueryBuilder_ComplexQuery(t *testing.T) {
	qb := NewQueryBuilder("orders")
	qb.Apply(
		Select("customer_id", "COUNT(*) as order_count", "SUM(total) as total_amount"),
		Join("INNER JOIN customers ON orders.customer_id = customers.id"),
		Where("orders.status = ?", "completed"),
		Where("orders.created_at > ?", "2024-01-01"),
		GroupBy("customer_id"),
		Having("COUNT(*) > ?", 5),
		OrderByDesc("total_amount"),
		Limit(100),
	)

	query, args := qb.Build()

	// Verify all parts are present
	if !strings.Contains(query, "SELECT customer_id, COUNT(*) as order_count, SUM(total) as total_amount") {
		t.Error("Missing SELECT columns")
	}
	if !strings.Contains(query, "INNER JOIN customers") {
		t.Error("Missing JOIN")
	}
	if !strings.Contains(query, "orders.status = ?") {
		t.Error("Missing WHERE condition")
	}
	if !strings.Contains(query, "GROUP BY customer_id") {
		t.Error("Missing GROUP BY")
	}
	if !strings.Contains(query, "HAVING COUNT(*) > ?") {
		t.Error("Missing HAVING")
	}
	if !strings.Contains(query, "ORDER BY total_amount DESC") {
		t.Error("Missing ORDER BY")
	}
	if !strings.Contains(query, "LIMIT 100") {
		t.Error("Missing LIMIT")
	}
	if len(args) != 3 {
		t.Errorf("Expected 3 args (2 WHERE + 1 HAVING), got %d", len(args))
	}
}

func TestQueryOption_ChainReturnsSameBuilder(t *testing.T) {
	qb := NewQueryBuilder("users")
	result := qb.Apply(Where("a = ?", 1))

	if result != qb {
		t.Error("Apply should return same QueryBuilder for chaining")
	}

	result = qb.Reset()
	if result != qb {
		t.Error("Reset should return same QueryBuilder for chaining")
	}
}

func TestQueryBuilder_EmptyBuild(t *testing.T) {
	qb := NewQueryBuilder("users")
	query, args := qb.Build()

	expected := "SELECT * FROM users WHERE deleted_at IS NULL"
	if query != expected {
		t.Errorf("Empty build should return basic query, got: %s", query)
	}
	if len(args) != 0 {
		t.Errorf("Empty build should have no args, got: %v", args)
	}
}

func TestQueryBuilder_PreservesArgOrder(t *testing.T) {
	qb := NewQueryBuilder("users")
	qb.Apply(
		Where("a = ?", "first"),
		Where("b = ?", "second"),
		Where("c = ?", "third"),
	)

	_, args := qb.Build()

	expected := []any{"first", "second", "third"}
	if !reflect.DeepEqual(args, expected) {
		t.Errorf("Args should preserve order, got: %v, want: %v", args, expected)
	}
}
