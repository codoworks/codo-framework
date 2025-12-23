// Package errors provides error types with HTTP status mapping for the framework.
package errors

// Error codes for categorizing framework errors.
const (
	// CodeInternal represents internal server errors (500).
	CodeInternal = "INTERNAL_ERROR"

	// CodeNotFound represents resource not found errors (404).
	CodeNotFound = "NOT_FOUND"

	// CodeBadRequest represents invalid request errors (400).
	CodeBadRequest = "BAD_REQUEST"

	// CodeUnauthorized represents authentication errors (401).
	CodeUnauthorized = "UNAUTHORIZED"

	// CodeForbidden represents authorization errors (403).
	CodeForbidden = "FORBIDDEN"

	// CodeMethodNotAllowed represents method not allowed errors (405).
	CodeMethodNotAllowed = "METHOD_NOT_ALLOWED"

	// CodeConflict represents resource conflict errors (409).
	CodeConflict = "CONFLICT"

	// CodeGone represents permanently deleted resources (410).
	CodeGone = "GONE"

	// CodePreconditionFailed represents precondition failed errors (412).
	CodePreconditionFailed = "PRECONDITION_FAILED"

	// CodeUnsupportedMedia represents unsupported media type errors (415).
	CodeUnsupportedMedia = "UNSUPPORTED_MEDIA_TYPE"

	// CodeValidation represents validation errors (422).
	CodeValidation = "VALIDATION_ERROR"

	// CodeTooManyRequests represents rate limiting errors (429).
	CodeTooManyRequests = "TOO_MANY_REQUESTS"

	// CodeTimeout represents timeout errors (408).
	CodeTimeout = "TIMEOUT"

	// CodeUnavailable represents service unavailable errors (503).
	CodeUnavailable = "SERVICE_UNAVAILABLE"

	// CodeBadGateway represents bad gateway errors (502).
	CodeBadGateway = "BAD_GATEWAY"

	// CodeGatewayTimeout represents gateway timeout errors (504).
	CodeGatewayTimeout = "GATEWAY_TIMEOUT"
)

// AllCodes returns all defined error codes.
func AllCodes() []string {
	return []string{
		CodeInternal,
		CodeNotFound,
		CodeBadRequest,
		CodeUnauthorized,
		CodeForbidden,
		CodeMethodNotAllowed,
		CodeConflict,
		CodeGone,
		CodePreconditionFailed,
		CodeUnsupportedMedia,
		CodeValidation,
		CodeTooManyRequests,
		CodeTimeout,
		CodeUnavailable,
		CodeBadGateway,
		CodeGatewayTimeout,
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
