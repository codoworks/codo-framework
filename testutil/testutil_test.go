package testutil

import (
	"bytes"
	"errors"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestAssertNoError(t *testing.T) {
	t.Run("nil error passes", func(t *testing.T) {
		mockT := &testing.T{}
		AssertNoError(mockT, nil)
		if mockT.Failed() {
			t.Error("AssertNoError should not fail with nil error")
		}
	})

	t.Run("non-nil error fails", func(t *testing.T) {
		mockT := &testing.T{}
		AssertNoError(mockT, errors.New("test error"))
		if !mockT.Failed() {
			t.Error("AssertNoError should fail with non-nil error")
		}
	})
}

func TestAssertError(t *testing.T) {
	targetErr := errors.New("target error")

	t.Run("matching error passes", func(t *testing.T) {
		mockT := &testing.T{}
		AssertError(mockT, targetErr, targetErr)
		if mockT.Failed() {
			t.Error("AssertError should not fail with matching error")
		}
	})

	t.Run("nil error fails", func(t *testing.T) {
		mockT := &testing.T{}
		AssertError(mockT, nil, targetErr)
		if !mockT.Failed() {
			t.Error("AssertError should fail with nil error")
		}
	})

	t.Run("different error fails", func(t *testing.T) {
		mockT := &testing.T{}
		AssertError(mockT, errors.New("different"), targetErr)
		if !mockT.Failed() {
			t.Error("AssertError should fail with different error")
		}
	})
}

func TestAssertErrorContains(t *testing.T) {
	t.Run("error containing substring passes", func(t *testing.T) {
		mockT := &testing.T{}
		AssertErrorContains(mockT, errors.New("test error message"), "error")
		if mockT.Failed() {
			t.Error("AssertErrorContains should not fail when error contains substring")
		}
	})

	t.Run("nil error fails", func(t *testing.T) {
		mockT := &testing.T{}
		AssertErrorContains(mockT, nil, "error")
		if !mockT.Failed() {
			t.Error("AssertErrorContains should fail with nil error")
		}
	})

	t.Run("error not containing substring fails", func(t *testing.T) {
		mockT := &testing.T{}
		AssertErrorContains(mockT, errors.New("test message"), "xyz")
		if !mockT.Failed() {
			t.Error("AssertErrorContains should fail when error doesn't contain substring")
		}
	})

	t.Run("empty substring passes", func(t *testing.T) {
		mockT := &testing.T{}
		AssertErrorContains(mockT, errors.New("any error"), "")
		if mockT.Failed() {
			t.Error("AssertErrorContains should pass with empty substring")
		}
	})
}

func TestContainsSubstring(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		substr string
		want   bool
	}{
		{"empty both", "", "", true},
		{"empty substring", "hello", "", true},
		{"empty string", "", "hello", false},
		{"contains at start", "hello world", "hello", true},
		{"contains at end", "hello world", "world", true},
		{"contains in middle", "hello world", "lo wo", true},
		{"not contains", "hello world", "xyz", false},
		{"exact match", "hello", "hello", true},
		{"substring longer", "hi", "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsSubstring(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("containsSubstring(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}

func TestNewTestContext(t *testing.T) {
	t.Run("creates context with correct method and path", func(t *testing.T) {
		ctx, rec := NewTestContext(http.MethodPost, "/test/path", nil)

		if ctx.Request().Method != http.MethodPost {
			t.Errorf("expected method %s, got %s", http.MethodPost, ctx.Request().Method)
		}
		if ctx.Request().URL.Path != "/test/path" {
			t.Errorf("expected path /test/path, got %s", ctx.Request().URL.Path)
		}
		if ctx.Request().Header.Get(echo.HeaderContentType) != echo.MIMEApplicationJSON {
			t.Error("expected Content-Type application/json")
		}
		if rec == nil {
			t.Error("expected non-nil recorder")
		}
	})

	t.Run("creates context with body", func(t *testing.T) {
		body := bytes.NewReader([]byte(`{"key":"value"}`))
		ctx, _ := NewTestContext(http.MethodPost, "/", body)

		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(ctx.Request().Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}
		if buf.String() != `{"key":"value"}` {
			t.Errorf("expected body {\"key\":\"value\"}, got %s", buf.String())
		}
	})
}

func TestNewTestContextWithHeaders(t *testing.T) {
	t.Run("creates context with custom headers", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "Bearer token123",
			"X-Custom":      "custom-value",
		}
		ctx, rec := NewTestContextWithHeaders(http.MethodGet, "/test", nil, headers)

		if ctx.Request().Header.Get("Authorization") != "Bearer token123" {
			t.Error("expected Authorization header")
		}
		if ctx.Request().Header.Get("X-Custom") != "custom-value" {
			t.Error("expected X-Custom header")
		}
		if rec == nil {
			t.Error("expected non-nil recorder")
		}
	})

	t.Run("works with empty headers", func(t *testing.T) {
		ctx, _ := NewTestContextWithHeaders(http.MethodGet, "/test", nil, nil)
		if ctx == nil {
			t.Error("expected non-nil context")
		}
	})
}

func TestAssertEqual(t *testing.T) {
	t.Run("equal values pass", func(t *testing.T) {
		mockT := &testing.T{}
		AssertEqual(mockT, 42, 42)
		if mockT.Failed() {
			t.Error("AssertEqual should not fail with equal values")
		}
	})

	t.Run("different values fail", func(t *testing.T) {
		mockT := &testing.T{}
		AssertEqual(mockT, 42, 43)
		if !mockT.Failed() {
			t.Error("AssertEqual should fail with different values")
		}
	})

	t.Run("equal strings pass", func(t *testing.T) {
		mockT := &testing.T{}
		AssertEqual(mockT, "hello", "hello")
		if mockT.Failed() {
			t.Error("AssertEqual should not fail with equal strings")
		}
	})

	t.Run("different strings fail", func(t *testing.T) {
		mockT := &testing.T{}
		AssertEqual(mockT, "hello", "world")
		if !mockT.Failed() {
			t.Error("AssertEqual should fail with different strings")
		}
	})
}

func TestAssertTrue(t *testing.T) {
	t.Run("true passes", func(t *testing.T) {
		mockT := &testing.T{}
		AssertTrue(mockT, true, "should be true")
		if mockT.Failed() {
			t.Error("AssertTrue should not fail with true")
		}
	})

	t.Run("false fails", func(t *testing.T) {
		mockT := &testing.T{}
		AssertTrue(mockT, false, "should be true")
		if !mockT.Failed() {
			t.Error("AssertTrue should fail with false")
		}
	})
}

func TestAssertFalse(t *testing.T) {
	t.Run("false passes", func(t *testing.T) {
		mockT := &testing.T{}
		AssertFalse(mockT, false, "should be false")
		if mockT.Failed() {
			t.Error("AssertFalse should not fail with false")
		}
	})

	t.Run("true fails", func(t *testing.T) {
		mockT := &testing.T{}
		AssertFalse(mockT, true, "should be false")
		if !mockT.Failed() {
			t.Error("AssertFalse should fail with true")
		}
	})
}

func TestAssertNotNil(t *testing.T) {
	t.Run("non-nil passes", func(t *testing.T) {
		mockT := &testing.T{}
		AssertNotNil(mockT, "something")
		if mockT.Failed() {
			t.Error("AssertNotNil should not fail with non-nil value")
		}
	})

	t.Run("nil fails", func(t *testing.T) {
		mockT := &testing.T{}
		AssertNotNil(mockT, nil)
		if !mockT.Failed() {
			t.Error("AssertNotNil should fail with nil value")
		}
	})
}

func TestAssertNil(t *testing.T) {
	t.Run("nil passes", func(t *testing.T) {
		mockT := &testing.T{}
		AssertNil(mockT, nil)
		if mockT.Failed() {
			t.Error("AssertNil should not fail with nil value")
		}
	})

	t.Run("non-nil fails", func(t *testing.T) {
		mockT := &testing.T{}
		AssertNil(mockT, "something")
		if !mockT.Failed() {
			t.Error("AssertNil should fail with non-nil value")
		}
	})
}

func TestAssertPanics(t *testing.T) {
	t.Run("panic passes", func(t *testing.T) {
		mockT := &testing.T{}
		AssertPanics(mockT, func() {
			panic("expected panic")
		})
		if mockT.Failed() {
			t.Error("AssertPanics should not fail when function panics")
		}
	})

	t.Run("no panic fails", func(t *testing.T) {
		mockT := &testing.T{}
		AssertPanics(mockT, func() {
			// no panic
		})
		if !mockT.Failed() {
			t.Error("AssertPanics should fail when function doesn't panic")
		}
	})
}

func TestAssertNotPanics(t *testing.T) {
	t.Run("no panic passes", func(t *testing.T) {
		mockT := &testing.T{}
		AssertNotPanics(mockT, func() {
			// no panic
		})
		if mockT.Failed() {
			t.Error("AssertNotPanics should not fail when function doesn't panic")
		}
	})

	t.Run("panic fails", func(t *testing.T) {
		mockT := &testing.T{}
		AssertNotPanics(mockT, func() {
			panic("unexpected panic")
		})
		if !mockT.Failed() {
			t.Error("AssertNotPanics should fail when function panics")
		}
	})
}
