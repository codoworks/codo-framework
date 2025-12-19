// Package registry provides a thread-safe generic registry for storing and retrieving items by name.
package registry

import (
	"fmt"
	"sync"
)

// Registry is a thread-safe generic registry for storing items by name.
type Registry[T any] struct {
	mu    sync.RWMutex
	items map[string]T
}

// New creates a new empty registry.
func New[T any]() *Registry[T] {
	return &Registry[T]{
		items: make(map[string]T),
	}
}

// Register adds an item to the registry.
// Returns an error if an item with the same name already exists.
func (r *Registry[T]) Register(name string, item T) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.items[name]; exists {
		return fmt.Errorf("item %q already registered", name)
	}
	r.items[name] = item
	return nil
}

// MustRegister adds an item to the registry and panics if it fails.
func (r *Registry[T]) MustRegister(name string, item T) {
	if err := r.Register(name, item); err != nil {
		panic(err)
	}
}

// Get retrieves an item from the registry by name.
// Returns an error if the item doesn't exist.
func (r *Registry[T]) Get(name string) (T, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, exists := r.items[name]
	if !exists {
		var zero T
		return zero, fmt.Errorf("item %q not found", name)
	}
	return item, nil
}

// MustGet retrieves an item from the registry by name and panics if it doesn't exist.
func (r *Registry[T]) MustGet(name string) T {
	item, err := r.Get(name)
	if err != nil {
		panic(err)
	}
	return item
}

// Has checks if an item with the given name exists in the registry.
func (r *Registry[T]) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.items[name]
	return exists
}

// All returns a copy of all items in the registry.
func (r *Registry[T]) All() map[string]T {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]T, len(r.items))
	for k, v := range r.items {
		result[k] = v
	}
	return result
}

// Keys returns all keys in the registry.
func (r *Registry[T]) Keys() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	keys := make([]string, 0, len(r.items))
	for k := range r.items {
		keys = append(keys, k)
	}
	return keys
}

// Count returns the number of items in the registry.
func (r *Registry[T]) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.items)
}

// Remove removes an item from the registry.
// Returns true if the item was removed, false if it didn't exist.
func (r *Registry[T]) Remove(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.items[name]; !exists {
		return false
	}
	delete(r.items, name)
	return true
}

// Clear removes all items from the registry.
func (r *Registry[T]) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = make(map[string]T)
}

// Replace replaces an existing item or adds a new one.
// Returns true if the item was replaced, false if it was added.
func (r *Registry[T]) Replace(name string, item T) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.items[name]
	r.items[name] = item
	return exists
}

// Range calls fn for each item in the registry.
// If fn returns false, iteration stops.
func (r *Registry[T]) Range(fn func(name string, item T) bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for name, item := range r.items {
		if !fn(name, item) {
			break
		}
	}
}
