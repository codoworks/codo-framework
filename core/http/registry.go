package http

import (
	"sync"
)

var (
	handlers   []Handler
	handlersMu sync.RWMutex
)

// RegisterHandler adds a handler to the registry
func RegisterHandler(h Handler) {
	handlersMu.Lock()
	defer handlersMu.Unlock()
	handlers = append(handlers, h)
}

// GetHandlers returns handlers for a specific scope
func GetHandlers(scope RouterScope) []Handler {
	handlersMu.RLock()
	defer handlersMu.RUnlock()

	var result []Handler
	for _, h := range handlers {
		if h.Scope() == scope {
			result = append(result, h)
		}
	}
	return result
}

// AllHandlers returns all registered handlers
func AllHandlers() []Handler {
	handlersMu.RLock()
	defer handlersMu.RUnlock()

	result := make([]Handler, len(handlers))
	copy(result, handlers)
	return result
}

// ClearHandlers removes all handlers (for testing)
func ClearHandlers() {
	handlersMu.Lock()
	defer handlersMu.Unlock()
	handlers = nil
}

// HandlerCount returns the number of registered handlers
func HandlerCount() int {
	handlersMu.RLock()
	defer handlersMu.RUnlock()
	return len(handlers)
}
