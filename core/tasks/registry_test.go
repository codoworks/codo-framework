package tasks

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	Clear()
	defer Clear()

	task := Task{
		Name:        "test-task",
		Description: "A test task",
		Run: func(ctx context.Context, args []string) error {
			return nil
		},
	}

	Register(task)

	got, ok := Get("test-task")
	assert.True(t, ok)
	assert.Equal(t, "test-task", got.Name)
	assert.Equal(t, "A test task", got.Description)
}

func TestRegister_Overwrite(t *testing.T) {
	Clear()
	defer Clear()

	task1 := Task{Name: "task", Description: "first"}
	task2 := Task{Name: "task", Description: "second"}

	Register(task1)
	Register(task2)

	got, ok := Get("task")
	assert.True(t, ok)
	assert.Equal(t, "second", got.Description)
}

func TestGet(t *testing.T) {
	Clear()
	defer Clear()

	t.Run("existing task", func(t *testing.T) {
		Register(Task{Name: "exists", Description: "test"})

		got, ok := Get("exists")

		assert.True(t, ok)
		assert.Equal(t, "exists", got.Name)
	})

	t.Run("non-existing task", func(t *testing.T) {
		got, ok := Get("does-not-exist")

		assert.False(t, ok)
		assert.Equal(t, "", got.Name)
	})
}

func TestAll(t *testing.T) {
	Clear()
	defer Clear()

	t.Run("empty registry", func(t *testing.T) {
		all := All()
		assert.Empty(t, all)
	})

	t.Run("with tasks", func(t *testing.T) {
		Register(Task{Name: "task1", Description: "first"})
		Register(Task{Name: "task2", Description: "second"})
		Register(Task{Name: "task3", Description: "third"})

		all := All()

		assert.Len(t, all, 3)

		names := make(map[string]bool)
		for _, task := range all {
			names[task.Name] = true
		}
		assert.True(t, names["task1"])
		assert.True(t, names["task2"])
		assert.True(t, names["task3"])
	})
}

func TestClear(t *testing.T) {
	Register(Task{Name: "to-clear", Description: "will be cleared"})

	Clear()

	_, ok := Get("to-clear")
	assert.False(t, ok)
	assert.Empty(t, All())
}

func TestExecute(t *testing.T) {
	Clear()
	defer Clear()

	t.Run("successful execution", func(t *testing.T) {
		var executed bool
		var receivedArgs []string

		Register(Task{
			Name: "success-task",
			Run: func(ctx context.Context, args []string) error {
				executed = true
				receivedArgs = args
				return nil
			},
		})

		err := Execute(context.Background(), "success-task", []string{"arg1", "arg2"})

		assert.NoError(t, err)
		assert.True(t, executed)
		assert.Equal(t, []string{"arg1", "arg2"}, receivedArgs)
	})

	t.Run("task returns error", func(t *testing.T) {
		expectedErr := errors.New("task failed")
		Register(Task{
			Name: "error-task",
			Run: func(ctx context.Context, args []string) error {
				return expectedErr
			},
		})

		err := Execute(context.Background(), "error-task", nil)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("task not found", func(t *testing.T) {
		err := Execute(context.Background(), "nonexistent", nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task not found")
		assert.Contains(t, err.Error(), "nonexistent")
	})

	t.Run("task with nil run function", func(t *testing.T) {
		Register(Task{
			Name:        "nil-run",
			Description: "has no run function",
			Run:         nil,
		})

		err := Execute(context.Background(), "nil-run", nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "has no run function")
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		Register(Task{
			Name: "ctx-task",
			Run: func(ctx context.Context, args []string) error {
				return ctx.Err()
			},
		})

		err := Execute(ctx, "ctx-task", nil)

		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("with empty args", func(t *testing.T) {
		var receivedArgs []string
		Register(Task{
			Name: "empty-args-task",
			Run: func(ctx context.Context, args []string) error {
				receivedArgs = args
				return nil
			},
		})

		err := Execute(context.Background(), "empty-args-task", []string{})

		assert.NoError(t, err)
		assert.Empty(t, receivedArgs)
	})
}

func TestConcurrentAccess(t *testing.T) {
	Clear()
	defer Clear()

	done := make(chan bool)

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(n int) {
			Register(Task{
				Name:        string(rune('a' + n)),
				Description: "concurrent task",
			})
			done <- true
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			_ = All()
			_, _ = Get("a")
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Verify some tasks were registered
	all := All()
	assert.NotEmpty(t, all)
}
