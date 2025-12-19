// Package errors provides error types with HTTP status mapping for the framework.
package errors

// Error codes for categorizing framework errors.
const (
	// CodeInternal represents internal server errors.
	CodeInternal = "INTERNAL_ERROR"

	// CodeNotFound represents resource not found errors.
	CodeNotFound = "NOT_FOUND"

	// CodeBadRequest represents invalid request errors.
	CodeBadRequest = "BAD_REQUEST"

	// CodeUnauthorized represents authentication errors.
	CodeUnauthorized = "UNAUTHORIZED"

	// CodeForbidden represents authorization errors.
	CodeForbidden = "FORBIDDEN"

	// CodeConflict represents resource conflict errors.
	CodeConflict = "CONFLICT"

	// CodeValidation represents validation errors.
	CodeValidation = "VALIDATION_ERROR"

	// CodeTimeout represents timeout errors.
	CodeTimeout = "TIMEOUT"

	// CodeUnavailable represents service unavailable errors.
	CodeUnavailable = "SERVICE_UNAVAILABLE"
)

// AllCodes returns all defined error codes.
func AllCodes() []string {
	return []string{
		CodeInternal,
		CodeNotFound,
		CodeBadRequest,
		CodeUnauthorized,
		CodeForbidden,
		CodeConflict,
		CodeValidation,
		CodeTimeout,
		CodeUnavailable,
	}
}

// IsValidCode checks if a code is a valid error code.
func IsValidCode(code string) bool {
	for _, c := range AllCodes() {
		if c == code {
			return true
		}
	}
	return false
}
