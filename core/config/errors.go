package config

// ErrorsConfig holds error handling configuration
type ErrorsConfig struct {
	Handler ErrorHandlerConfig `yaml:"handler"`
	Capture ErrorCaptureConfig `yaml:"capture"`
}

// ErrorHandlerConfig controls HTTP error response behavior
type ErrorHandlerConfig struct {
	// ExposeDetails controls whether error details are included in responses
	// Should be true in dev mode, false in production
	ExposeDetails bool `yaml:"expose_details"`

	// ExposeStackTraces controls whether stack traces are included in responses
	// Should NEVER be true in production (security risk)
	ExposeStackTraces bool `yaml:"expose_stack_traces"`

	// ResponseFormat controls the response format
	// "standard" - full response with all fields
	// "minimal" - only code and message
	ResponseFormat string `yaml:"response_format"`
}

// ErrorCaptureConfig controls error capture behavior
type ErrorCaptureConfig struct {
	// StackTraceOn5xx controls whether stack traces are captured for 5xx errors
	// Captured traces are logged and available for debugging
	StackTraceOn5xx bool `yaml:"stack_trace_on_5xx"`

	// StackTraceDepth controls how many stack frames to capture
	StackTraceDepth int `yaml:"stack_trace_depth"`

	// AutoDetectPhase controls whether to auto-detect lifecycle phase from package name
	AutoDetectPhase bool `yaml:"auto_detect_phase"`
}

// DefaultErrorsConfig returns default error configuration
func DefaultErrorsConfig() ErrorsConfig {
	return ErrorsConfig{
		Handler: ErrorHandlerConfig{
			ExposeDetails:     false,      // Don't expose details by default
			ExposeStackTraces: false,      // NEVER expose stack traces by default
			ResponseFormat:    "standard", // Use standard format
		},
		Capture: ErrorCaptureConfig{
			StackTraceOn5xx: true, // Capture stack traces for 5xx errors
			StackTraceDepth: 32,   // Capture up to 32 frames
			AutoDetectPhase: true, // Auto-detect phase from package name
		},
	}
}
