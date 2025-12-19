// Package tasks provides a task registry and execution system.
// This is a minimal stub for CLI layer - will be replaced by WS-1 implementation.
package tasks

import (
	"context"
	"fmt"
	"sync"
)

// Task represents a runnable task.
type Task struct {
	Name        string
	Description string
	Run         func(ctx context.Context, args []string) error
}

var (
	tasks   map[string]Task
	tasksMu sync.RWMutex
)

func init() {
	tasks = make(map[string]Task)
}

// Register registers a task.
func Register(t Task) {
	tasksMu.Lock()
	defer tasksMu.Unlock()
	tasks[t.Name] = t
}

// Get returns a task by name.
func Get(name string) (Task, bool) {
	tasksMu.RLock()
	defer tasksMu.RUnlock()
	t, ok := tasks[name]
	return t, ok
}

// All returns all registered tasks.
func All() []Task {
	tasksMu.RLock()
	defer tasksMu.RUnlock()
	result := make([]Task, 0, len(tasks))
	for _, t := range tasks {
		result = append(result, t)
	}
	return result
}

// Clear clears all registered tasks (for testing).
func Clear() {
	tasksMu.Lock()
	defer tasksMu.Unlock()
	tasks = make(map[string]Task)
}

// Execute runs a task by name with the given arguments.
func Execute(ctx context.Context, name string, args []string) error {
	t, ok := Get(name)
	if !ok {
		return fmt.Errorf("task not found: %s", name)
	}
	if t.Run == nil {
		return fmt.Errorf("task %s has no run function", name)
	}
	return t.Run(ctx, args)
}
