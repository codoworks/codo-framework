package errors

import (
	"testing"
)

func TestCodes(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"internal", CodeInternal, "INTERNAL_ERROR"},
		{"not found", CodeNotFound, "NOT_FOUND"},
		{"bad request", CodeBadRequest, "BAD_REQUEST"},
		{"unauthorized", CodeUnauthorized, "UNAUTHORIZED"},
		{"forbidden", CodeForbidden, "FORBIDDEN"},
		{"method not allowed", CodeMethodNotAllowed, "METHOD_NOT_ALLOWED"},
		{"conflict", CodeConflict, "CONFLICT"},
		{"gone", CodeGone, "GONE"},
		{"precondition failed", CodePreconditionFailed, "PRECONDITION_FAILED"},
		{"unsupported media", CodeUnsupportedMedia, "UNSUPPORTED_MEDIA_TYPE"},
		{"validation", CodeValidation, "VALIDATION_ERROR"},
		{"too many requests", CodeTooManyRequests, "TOO_MANY_REQUESTS"},
		{"timeout", CodeTimeout, "TIMEOUT"},
		{"unavailable", CodeUnavailable, "SERVICE_UNAVAILABLE"},
		{"bad gateway", CodeBadGateway, "BAD_GATEWAY"},
		{"gateway timeout", CodeGatewayTimeout, "GATEWAY_TIMEOUT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.code)
			}
		})
	}
}

func TestAllCodes(t *testing.T) {
	codes := AllCodes()

	t.Run("returns all codes", func(t *testing.T) {
		expectedCodes := []string{
			"INTERNAL_ERROR",
			"NOT_FOUND",
			"BAD_REQUEST",
			"UNAUTHORIZED",
			"FORBIDDEN",
			"METHOD_NOT_ALLOWED",
			"CONFLICT",
			"GONE",
			"PRECONDITION_FAILED",
			"UNSUPPORTED_MEDIA_TYPE",
			"VALIDATION_ERROR",
			"TOO_MANY_REQUESTS",
			"TIMEOUT",
			"SERVICE_UNAVAILABLE",
			"BAD_GATEWAY",
			"GATEWAY_TIMEOUT",
		}

		if len(codes) != len(expectedCodes) {
			t.Errorf("expected %d codes, got %d", len(expectedCodes), len(codes))
		}

		for i, code := range expectedCodes {
			if codes[i] != code {
				t.Errorf("expected code at index %d to be %s, got %s", i, code, codes[i])
			}
		}
	})

	t.Run("returns a copy", func(t *testing.T) {
		codes1 := AllCodes()
		codes2 := AllCodes()

		// Modify the first slice
		codes1[0] = "MODIFIED"

		// Second slice should not be affected
		if codes2[0] == "MODIFIED" {
			t.Error("AllCodes should return a new slice each time")
		}
	})
}

func TestIsValidCode(t *testing.T) {
	tests := []struct {
		name  string
		code  string
		valid bool
	}{
		{"internal valid", CodeInternal, true},
		{"not found valid", CodeNotFound, true},
		{"bad request valid", CodeBadRequest, true},
		{"unauthorized valid", CodeUnauthorized, true},
		{"forbidden valid", CodeForbidden, true},
		{"method not allowed valid", CodeMethodNotAllowed, true},
		{"conflict valid", CodeConflict, true},
		{"gone valid", CodeGone, true},
		{"precondition failed valid", CodePreconditionFailed, true},
		{"unsupported media valid", CodeUnsupportedMedia, true},
		{"validation valid", CodeValidation, true},
		{"too many requests valid", CodeTooManyRequests, true},
		{"timeout valid", CodeTimeout, true},
		{"unavailable valid", CodeUnavailable, true},
		{"bad gateway valid", CodeBadGateway, true},
		{"gateway timeout valid", CodeGatewayTimeout, true},
		{"invalid code", "INVALID_CODE", false},
		{"empty code", "", false},
		{"lowercase version", "internal_error", false},
		{"partial match", "INTERNAL", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidCode(tt.code)
			if got != tt.valid {
				t.Errorf("IsValidCode(%q) = %v, want %v", tt.code, got, tt.valid)
			}
		})
	}
}
