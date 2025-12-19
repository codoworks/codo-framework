package registry

import (
	"sort"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegistry_New(t *testing.T) {
	t.Run("creates empty registry", func(t *testing.T) {
		r := New[string]()

		assert.NotNil(t, r)
		assert.Equal(t, 0, r.Count())
	})

	t.Run("creates registry with different types", func(t *testing.T) {
		stringReg := New[string]()
		intReg := New[int]()
		structReg := New[struct{ Name string }]()

		assert.NotNil(t, stringReg)
		assert.NotNil(t, intReg)
		assert.NotNil(t, structReg)
	})
}

func TestRegistry_Register(t *testing.T) {
	t.Run("registers item successfully", func(t *testing.T) {
		r := New[string]()

		err := r.Register("key1", "value1")

		assert.NoError(t, err)
		assert.True(t, r.Has("key1"))
	})

	t.Run("registers multiple items", func(t *testing.T) {
		r := New[int]()

		err1 := r.Register("one", 1)
		err2 := r.Register("two", 2)
		err3 := r.Register("three", 3)

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NoError(t, err3)
		assert.Equal(t, 3, r.Count())
	})
}

func TestRegistry_Register_Duplicate(t *testing.T) {
	r := New[string]()

	err1 := r.Register("key", "value1")
	err2 := r.Register("key", "value2")

	assert.NoError(t, err1)
	assert.Error(t, err2)
	assert.Contains(t, err2.Error(), "already registered")

	// Original value should be unchanged
	val, _ := r.Get("key")
	assert.Equal(t, "value1", val)
}

func TestRegistry_MustRegister(t *testing.T) {
	t.Run("registers successfully", func(t *testing.T) {
		r := New[string]()

		assert.NotPanics(t, func() {
			r.MustRegister("key", "value")
		})
		assert.True(t, r.Has("key"))
	})

	t.Run("panics on duplicate", func(t *testing.T) {
		r := New[string]()
		r.MustRegister("key", "value1")

		assert.Panics(t, func() {
			r.MustRegister("key", "value2")
		})
	})
}

func TestRegistry_Get(t *testing.T) {
	r := New[string]()
	r.Register("key", "value")

	t.Run("retrieves existing item", func(t *testing.T) {
		val, err := r.Get("key")

		assert.NoError(t, err)
		assert.Equal(t, "value", val)
	})
}

func TestRegistry_Get_NotFound(t *testing.T) {
	r := New[string]()

	val, err := r.Get("nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Equal(t, "", val) // zero value for string
}

func TestRegistry_MustGet(t *testing.T) {
	r := New[string]()
	r.Register("key", "value")

	t.Run("returns item", func(t *testing.T) {
		var val string
		assert.NotPanics(t, func() {
			val = r.MustGet("key")
		})
		assert.Equal(t, "value", val)
	})
}

func TestRegistry_MustGet_Panics(t *testing.T) {
	r := New[string]()

	assert.Panics(t, func() {
		r.MustGet("nonexistent")
	})
}

func TestRegistry_Has(t *testing.T) {
	r := New[string]()
	r.Register("exists", "value")

	t.Run("returns true for existing", func(t *testing.T) {
		assert.True(t, r.Has("exists"))
	})
}

func TestRegistry_Has_Missing(t *testing.T) {
	r := New[string]()

	assert.False(t, r.Has("missing"))
}

func TestRegistry_All(t *testing.T) {
	r := New[int]()
	r.Register("one", 1)
	r.Register("two", 2)
	r.Register("three", 3)

	t.Run("returns all items", func(t *testing.T) {
		all := r.All()

		assert.Len(t, all, 3)
		assert.Equal(t, 1, all["one"])
		assert.Equal(t, 2, all["two"])
		assert.Equal(t, 3, all["three"])
	})

	t.Run("returns a copy", func(t *testing.T) {
		all := r.All()
		all["four"] = 4

		// Original registry should be unchanged
		assert.False(t, r.Has("four"))
		assert.Equal(t, 3, r.Count())
	})
}

func TestRegistry_All_Empty(t *testing.T) {
	r := New[string]()

	all := r.All()

	assert.NotNil(t, all)
	assert.Len(t, all, 0)
}

func TestRegistry_Keys(t *testing.T) {
	r := New[string]()
	r.Register("alpha", "a")
	r.Register("beta", "b")
	r.Register("gamma", "g")

	t.Run("returns all keys", func(t *testing.T) {
		keys := r.Keys()

		assert.Len(t, keys, 3)
		sort.Strings(keys)
		assert.Equal(t, []string{"alpha", "beta", "gamma"}, keys)
	})

	t.Run("returns empty slice for empty registry", func(t *testing.T) {
		empty := New[string]()
		keys := empty.Keys()

		assert.NotNil(t, keys)
		assert.Len(t, keys, 0)
	})
}

func TestRegistry_Count(t *testing.T) {
	r := New[string]()

	assert.Equal(t, 0, r.Count())

	r.Register("one", "1")
	assert.Equal(t, 1, r.Count())

	r.Register("two", "2")
	assert.Equal(t, 2, r.Count())
}

func TestRegistry_Remove(t *testing.T) {
	r := New[string]()
	r.Register("key", "value")

	t.Run("removes existing item", func(t *testing.T) {
		removed := r.Remove("key")

		assert.True(t, removed)
		assert.False(t, r.Has("key"))
	})

	t.Run("returns false for non-existent item", func(t *testing.T) {
		removed := r.Remove("nonexistent")

		assert.False(t, removed)
	})
}

func TestRegistry_Clear(t *testing.T) {
	r := New[string]()
	r.Register("one", "1")
	r.Register("two", "2")
	r.Register("three", "3")

	r.Clear()

	assert.Equal(t, 0, r.Count())
	assert.False(t, r.Has("one"))
	assert.False(t, r.Has("two"))
	assert.False(t, r.Has("three"))
}

func TestRegistry_Replace(t *testing.T) {
	r := New[string]()

	t.Run("adds new item", func(t *testing.T) {
		replaced := r.Replace("new", "value")

		assert.False(t, replaced)
		val, _ := r.Get("new")
		assert.Equal(t, "value", val)
	})

	t.Run("replaces existing item", func(t *testing.T) {
		r.Register("key", "original")

		replaced := r.Replace("key", "updated")

		assert.True(t, replaced)
		val, _ := r.Get("key")
		assert.Equal(t, "updated", val)
	})
}

func TestRegistry_Range(t *testing.T) {
	r := New[int]()
	r.Register("one", 1)
	r.Register("two", 2)
	r.Register("three", 3)

	t.Run("iterates all items", func(t *testing.T) {
		sum := 0
		r.Range(func(name string, item int) bool {
			sum += item
			return true
		})

		assert.Equal(t, 6, sum)
	})

	t.Run("stops when fn returns false", func(t *testing.T) {
		count := 0
		r.Range(func(name string, item int) bool {
			count++
			return count < 2
		})

		assert.Equal(t, 2, count)
	})

	t.Run("works with empty registry", func(t *testing.T) {
		empty := New[string]()
		called := false
		empty.Range(func(name string, item string) bool {
			called = true
			return true
		})

		assert.False(t, called)
	})
}

func TestRegistry_Concurrent(t *testing.T) {
	r := New[int]()
	var wg sync.WaitGroup
	iterations := 100

	// Concurrent writes
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := string(rune('a' + (i % 26)))
			r.Replace(key, i)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := string(rune('a' + (i % 26)))
			r.Get(key)
			r.Has(key)
			r.All()
			r.Keys()
			r.Count()
		}(i)
	}

	wg.Wait()

	// Should not panic or have race conditions
	assert.True(t, r.Count() > 0)
}

func TestRegistry_ConcurrentRegister(t *testing.T) {
	r := New[int]()
	var wg sync.WaitGroup
	iterations := 100

	// Try to register the same key from multiple goroutines
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			r.Register("same-key", i)
		}(i)
	}

	wg.Wait()

	// Only one registration should succeed
	assert.Equal(t, 1, r.Count())
}

func TestRegistry_ConcurrentRemove(t *testing.T) {
	r := New[int]()

	// Add items
	for i := 0; i < 100; i++ {
		r.Register(string(rune('a'+i)), i)
	}

	var wg sync.WaitGroup

	// Concurrent removes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			r.Remove(string(rune('a' + i)))
		}(i)
	}

	wg.Wait()

	assert.Equal(t, 0, r.Count())
}

func TestRegistry_TypeSafety(t *testing.T) {
	t.Run("struct type", func(t *testing.T) {
		type User struct {
			ID   int
			Name string
		}

		r := New[User]()
		r.Register("user1", User{ID: 1, Name: "Alice"})

		user, err := r.Get("user1")
		assert.NoError(t, err)
		assert.Equal(t, 1, user.ID)
		assert.Equal(t, "Alice", user.Name)
	})

	t.Run("pointer type", func(t *testing.T) {
		type Config struct {
			Value string
		}

		r := New[*Config]()
		cfg := &Config{Value: "test"}
		r.Register("config", cfg)

		retrieved, err := r.Get("config")
		assert.NoError(t, err)
		assert.Equal(t, cfg, retrieved)
	})

	t.Run("interface type", func(t *testing.T) {
		r := New[any]()
		r.Register("string", "hello")
		r.Register("int", 42)
		r.Register("bool", true)

		s, _ := r.Get("string")
		i, _ := r.Get("int")
		b, _ := r.Get("bool")

		assert.Equal(t, "hello", s)
		assert.Equal(t, 42, i)
		assert.Equal(t, true, b)
	})
}
