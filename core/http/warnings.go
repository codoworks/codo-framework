package http

// Warning represents a non-fatal issue in the response
// Warnings allow APIs to return partial success responses, such as:
// - Batch operations with some failures
// - Non-critical background tasks that failed
// - Deprecated API usage notices
type Warning struct {
	Code    string         `json:"code"`              // Warning code (e.g., "SYNC_FAILED", "DEPRECATED")
	Message string         `json:"message"`           // Human-readable warning message
	Details map[string]any `json:"details,omitempty"` // Optional additional context
}

// NewWarning creates a new warning with code and message
func NewWarning(code, message string) Warning {
	return Warning{
		Code:    code,
		Message: message,
		Details: make(map[string]any),
	}
}

// WithDetail adds a detail field to the warning
func (w Warning) WithDetail(key string, value any) Warning {
	if w.Details == nil {
		w.Details = make(map[string]any)
	}
	w.Details[key] = value
	return w
}
