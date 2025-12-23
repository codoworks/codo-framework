package errors

import (
	"fmt"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// CaptureConfig holds settings for error capture behavior
// Set via SetCaptureConfig during initialization
type CaptureConfig struct {
	StackTraceOn5xx bool // Whether to capture stack traces for 5xx errors
	StackTraceDepth int  // How many stack frames to capture
	AutoDetectPhase bool // Whether to auto-detect lifecycle phase from package name
}

// defaultCaptureConfig returns safe defaults
func defaultCaptureConfig() CaptureConfig {
	return CaptureConfig{
		StackTraceOn5xx: true,
		StackTraceDepth: 32,
		AutoDetectPhase: true,
	}
}

var (
	captureConfig   = defaultCaptureConfig()
	captureConfigMu sync.RWMutex
)

// SetCaptureConfig sets the global capture configuration
// Should be called during bootstrap before creating errors
func SetCaptureConfig(cfg CaptureConfig) {
	captureConfigMu.Lock()
	defer captureConfigMu.Unlock()

	// Validate and apply defaults for invalid values
	if cfg.StackTraceDepth <= 0 {
		cfg.StackTraceDepth = 32
	}
	captureConfig = cfg
}

// GetCaptureConfig returns the current capture configuration
func GetCaptureConfig() CaptureConfig {
	captureConfigMu.RLock()
	defer captureConfigMu.RUnlock()
	return captureConfig
}

// Phase identifies where in the application lifecycle the error occurred
type Phase string

const (
	PhaseBootstrap   Phase = "bootstrap"   // App initialization
	PhaseConfig      Phase = "config"      // Configuration loading
	PhaseClient      Phase = "client"      // Client initialization
	PhaseMigration   Phase = "migration"   // Database migration
	PhaseMiddleware  Phase = "middleware"  // Middleware execution
	PhaseHandler     Phase = "handler"     // HTTP handler
	PhaseService     Phase = "service"     // Business logic
	PhaseRepository  Phase = "repository"  // Data access
	PhaseWorker      Phase = "worker"      // Background worker
	PhaseShutdown    Phase = "shutdown"    // Graceful shutdown
)

// CallerInfo captures where the error was created
type CallerInfo struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Function string `json:"function"`
	Package  string `json:"package"`
}

// StackFrame represents one frame in a stack trace
type StackFrame struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Function string `json:"function"`
}

// RequestContext captures HTTP request details
type RequestContext struct {
	RequestID  string            `json:"requestId,omitempty"`
	Method     string            `json:"method,omitempty"`
	Path       string            `json:"path,omitempty"`
	Query      map[string]string `json:"query,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
	RemoteAddr string            `json:"remoteAddr,omitempty"`
	UserID     string            `json:"userId,omitempty"`
	SessionID  string            `json:"sessionId,omitempty"`
}

// Error represents a framework error with HTTP status mapping.
type Error struct {
	// Core fields
	Code       string         `json:"code"`
	Message    string         `json:"message"`
	HTTPStatus int            `json:"-"`
	Phase      Phase          `json:"phase,omitempty"`
	Timestamp  time.Time      `json:"timestamp"`

	// Error chain
	Cause   error          `json:"-"`
	Details map[string]any `json:"details,omitempty"`

	// Auto-captured context
	Caller     *CallerInfo     `json:"caller,omitempty"`
	StackTrace []StackFrame    `json:"stackTrace,omitempty"`
	RequestCtx *RequestContext `json:"requestContext,omitempty"`

	// Rendering hints
	UserMessage string `json:"userMessage,omitempty"`
	Internal    bool   `json:"-"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause for errors.Unwrap support.
func (e *Error) Unwrap() error {
	return e.Cause
}

// Is implements errors.Is for error comparison by code.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// WithCause adds a cause to the error and returns the error for chaining.
func (e *Error) WithCause(err error) *Error {
	e.Cause = err
	return e
}

// WithDetails adds details to the error and returns the error for chaining.
func (e *Error) WithDetails(details map[string]any) *Error {
	e.Details = details
	return e
}

// WithDetail adds a single detail to the error and returns the error for chaining.
func (e *Error) WithDetail(key string, value any) *Error {
	if e.Details == nil {
		e.Details = make(map[string]any)
	}
	e.Details[key] = value
	return e
}

// WithPhase sets the phase and returns the error for chaining.
func (e *Error) WithPhase(phase Phase) *Error {
	e.Phase = phase
	return e
}

// WithUserMessage sets a user-friendly message and returns the error for chaining.
func (e *Error) WithUserMessage(msg string) *Error {
	e.UserMessage = msg
	return e
}

// MarkInternal marks this error as internal (won't expose details in HTTP responses).
func (e *Error) MarkInternal() *Error {
	e.Internal = true
	return e
}

// captureCallerInfo captures information about where the error was created
func captureCallerInfo(skip int) *CallerInfo {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return nil
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return nil
	}

	fnName := fn.Name()
	pkg := extractPackage(fnName)

	return &CallerInfo{
		File:     filepath.Base(file),
		Line:     line,
		Function: extractFunctionName(fnName),
		Package:  pkg,
	}
}

// captureStackTrace captures the current stack trace
func captureStackTrace(skip int) []StackFrame {
	cfg := GetCaptureConfig()
	maxFrames := cfg.StackTraceDepth
	if maxFrames <= 0 {
		maxFrames = 32 // Fallback default
	}

	pcs := make([]uintptr, maxFrames)
	n := runtime.Callers(skip, pcs)

	frames := make([]StackFrame, 0, n)
	for i := 0; i < n; i++ {
		fn := runtime.FuncForPC(pcs[i])
		if fn == nil {
			continue
		}

		file, line := fn.FileLine(pcs[i])
		frames = append(frames, StackFrame{
			File:     filepath.Base(file),
			Line:     line,
			Function: extractFunctionName(fn.Name()),
		})
	}

	return frames
}

// detectPhaseFromPackage attempts to detect the phase from package name
func detectPhaseFromPackage(pkg string) Phase {
	switch {
	case strings.Contains(pkg, "/core/app"):
		return PhaseBootstrap
	case strings.Contains(pkg, "/core/config"):
		return PhaseConfig
	case strings.Contains(pkg, "/core/clients"):
		return PhaseClient
	case strings.Contains(pkg, "/core/db"):
		return PhaseRepository
	case strings.Contains(pkg, "/core/middleware"):
		return PhaseMiddleware
	case strings.Contains(pkg, "/handlers"):
		return PhaseHandler
	case strings.Contains(pkg, "/services"):
		return PhaseService
	case strings.Contains(pkg, "/repositories"):
		return PhaseRepository
	default:
		return ""
	}
}

// extractPackage extracts the package path from a function name
func extractPackage(fnName string) string {
	// Function name format: github.com/user/repo/pkg.Type.Method
	lastSlash := strings.LastIndex(fnName, "/")
	if lastSlash == -1 {
		return ""
	}

	lastDot := strings.Index(fnName[lastSlash:], ".")
	if lastDot == -1 {
		return fnName
	}

	return fnName[:lastSlash+lastDot]
}

// extractFunctionName extracts just the function name from full path
func extractFunctionName(fnName string) string {
	// Get everything after last slash
	parts := strings.Split(fnName, "/")
	return parts[len(parts)-1]
}

// New creates a new error with custom code, message, and HTTP status.
// Automatically captures caller info, phase (if enabled), and stack trace (for 5xx if enabled).
func New(code, message string, httpStatus int) *Error {
	cfg := GetCaptureConfig()

	err := &Error{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Timestamp:  time.Now(),
	}

	// Auto-capture caller (skip 2 frames: New + constructor)
	err.Caller = captureCallerInfo(2)

	// Auto-detect phase from package (if enabled)
	if cfg.AutoDetectPhase && err.Caller != nil {
		err.Phase = detectPhaseFromPackage(err.Caller.Package)
	}

	// Capture stack trace for 5xx errors (if enabled)
	if cfg.StackTraceOn5xx && httpStatus >= 500 {
		err.StackTrace = captureStackTrace(2)
	}

	return err
}

// Internal creates an internal server error.
func Internal(msg string) *Error {
	return New(CodeInternal, msg, http.StatusInternalServerError)
}

// NotFound creates a not found error.
func NotFound(msg string) *Error {
	return New(CodeNotFound, msg, http.StatusNotFound)
}

// BadRequest creates a bad request error.
func BadRequest(msg string) *Error {
	return New(CodeBadRequest, msg, http.StatusBadRequest)
}

// Unauthorized creates an unauthorized error.
func Unauthorized(msg string) *Error {
	return New(CodeUnauthorized, msg, http.StatusUnauthorized)
}

// Forbidden creates a forbidden error.
func Forbidden(msg string) *Error {
	return New(CodeForbidden, msg, http.StatusForbidden)
}

// Conflict creates a conflict error.
func Conflict(msg string) *Error {
	return New(CodeConflict, msg, http.StatusConflict)
}

// Validation creates a validation error with a list of validation errors.
func Validation(msg string, errs []string) *Error {
	err := New(CodeValidation, msg, http.StatusUnprocessableEntity)
	err.Details = map[string]any{"errors": errs}
	return err
}

// Timeout creates a timeout error.
func Timeout(msg string) *Error {
	return New(CodeTimeout, msg, http.StatusRequestTimeout)
}

// Unavailable creates a service unavailable error.
func Unavailable(msg string) *Error {
	return New(CodeUnavailable, msg, http.StatusServiceUnavailable)
}

// MethodNotAllowed creates a method not allowed error.
func MethodNotAllowed(msg string) *Error {
	return New(CodeMethodNotAllowed, msg, http.StatusMethodNotAllowed)
}

// Gone creates a gone error for permanently deleted resources.
func Gone(msg string) *Error {
	return New(CodeGone, msg, http.StatusGone)
}

// PreconditionFailed creates a precondition failed error.
func PreconditionFailed(msg string) *Error {
	return New(CodePreconditionFailed, msg, http.StatusPreconditionFailed)
}

// UnsupportedMediaType creates an unsupported media type error.
func UnsupportedMediaType(msg string) *Error {
	return New(CodeUnsupportedMedia, msg, http.StatusUnsupportedMediaType)
}

// TooManyRequests creates a rate limiting error.
func TooManyRequests(msg string) *Error {
	return New(CodeTooManyRequests, msg, http.StatusTooManyRequests)
}

// BadGateway creates a bad gateway error.
func BadGateway(msg string) *Error {
	return New(CodeBadGateway, msg, http.StatusBadGateway)
}

// GatewayTimeout creates a gateway timeout error.
func GatewayTimeout(msg string) *Error {
	return New(CodeGatewayTimeout, msg, http.StatusGatewayTimeout)
}

// Wrap wraps an existing error with a framework error.
func Wrap(err error, code, message string, httpStatus int) *Error {
	e := New(code, message, httpStatus)
	e.Cause = err
	return e
}

// WrapInternal wraps an error as an internal error.
func WrapInternal(err error, msg string) *Error {
	return Internal(msg).WithCause(err)
}

// WrapNotFound wraps an error as a not found error.
func WrapNotFound(err error, msg string) *Error {
	return NotFound(msg).WithCause(err)
}

// WrapBadRequest wraps an error as a bad request error.
func WrapBadRequest(err error, msg string) *Error {
	return BadRequest(msg).WithCause(err)
}

// IsError checks if an error is a framework Error with a specific code.
func IsError(err error, code string) bool {
	if err == nil {
		return false
	}
	e, ok := err.(*Error)
	if !ok {
		return false
	}
	return e.Code == code
}

// IsInternal checks if an error is an internal error.
func IsInternal(err error) bool {
	return IsError(err, CodeInternal)
}

// IsNotFound checks if an error is a not found error.
func IsNotFound(err error) bool {
	return IsError(err, CodeNotFound)
}

// IsBadRequest checks if an error is a bad request error.
func IsBadRequest(err error) bool {
	return IsError(err, CodeBadRequest)
}

// IsUnauthorized checks if an error is an unauthorized error.
func IsUnauthorized(err error) bool {
	return IsError(err, CodeUnauthorized)
}

// IsForbidden checks if an error is a forbidden error.
func IsForbidden(err error) bool {
	return IsError(err, CodeForbidden)
}

// IsConflict checks if an error is a conflict error.
func IsConflict(err error) bool {
	return IsError(err, CodeConflict)
}

// IsValidation checks if an error is a validation error.
func IsValidation(err error) bool {
	return IsError(err, CodeValidation)
}

// GetHTTPStatus returns the HTTP status code for an error.
// Returns 500 for non-framework errors.
func GetHTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	e, ok := err.(*Error)
	if !ok {
		return http.StatusInternalServerError
	}
	return e.HTTPStatus
}

// GetCode returns the error code for an error.
// Returns CodeInternal for non-framework errors.
func GetCode(err error) string {
	if err == nil {
		return ""
	}
	e, ok := err.(*Error)
	if !ok {
		return CodeInternal
	}
	return e.Code
}
