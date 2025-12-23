package kratos

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	cfg := &ClientConfig{
		PublicURL:  "http://localhost:4433",
		AdminURL:   "http://localhost:4434",
		CookieName: "ory_kratos_session",
		Timeout:    5 * time.Second,
	}

	client := NewClient(cfg)

	assert.NotNil(t, client)
	assert.Equal(t, cfg, client.config)
	assert.NotNil(t, client.httpClient)
}

func TestNewClient_DefaultTimeout(t *testing.T) {
	cfg := &ClientConfig{
		PublicURL:  "http://localhost:4433",
		CookieName: "ory_kratos_session",
		Timeout:    0, // should default to 10s
	}

	client := NewClient(cfg)

	assert.NotNil(t, client.httpClient)
	assert.Equal(t, 10*time.Second, client.httpClient.Timeout)
}

func TestClient_Name(t *testing.T) {
	client := NewClient(&ClientConfig{})
	assert.Equal(t, "kratos", client.Name())
}

func TestClient_Initialize(t *testing.T) {
	client := &Client{}
	cfg := &ClientConfig{
		PublicURL:  "http://localhost:4433",
		CookieName: "test_cookie",
		Timeout:    5 * time.Second,
	}

	err := client.Initialize(cfg)

	assert.NoError(t, err)
	assert.Equal(t, cfg, client.config)
	assert.NotNil(t, client.httpClient)
}

func TestClient_Initialize_WrongType(t *testing.T) {
	client := &Client{
		config: &ClientConfig{PublicURL: "original"},
	}

	err := client.Initialize("not a config")

	assert.NoError(t, err) // doesn't error, just ignores
	assert.Equal(t, "original", client.config.PublicURL)
}

func TestClient_Initialize_DefaultTimeout(t *testing.T) {
	client := &Client{}
	cfg := &ClientConfig{
		PublicURL: "http://localhost:4433",
		Timeout:   0,
	}

	err := client.Initialize(cfg)

	assert.NoError(t, err)
	assert.Equal(t, 10*time.Second, client.httpClient.Timeout)
}

func TestClient_Health(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/health/live", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(&ClientConfig{
		PublicURL: server.URL,
		Timeout:   5 * time.Second,
	})

	err := client.Health()
	assert.NoError(t, err)
}

func TestClient_Health_Unhealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := NewClient(&ClientConfig{
		PublicURL: server.URL,
		Timeout:   5 * time.Second,
	})

	err := client.Health()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kratos unhealthy")
}

func TestClient_Health_ConnectionError(t *testing.T) {
	client := NewClient(&ClientConfig{
		PublicURL: "http://localhost:59999", // unlikely to be running
		Timeout:   100 * time.Millisecond,
	})

	err := client.Health()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "health check failed")
}

func TestClient_Health_InvalidURL(t *testing.T) {
	client := NewClient(&ClientConfig{
		PublicURL: "http://invalid\x00url", // control character makes URL invalid
		Timeout:   5 * time.Second,
	})

	err := client.Health()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create health request")
}

func TestClient_ValidateSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/sessions/whoami", r.URL.Path)
		assert.Contains(t, r.Header.Get("Cookie"), "ory_kratos_session=valid-session")

		session := Session{
			ID:     "session-123",
			Active: true,
		}
		session.Identity.ID = "user-456"
		session.Identity.Traits = map[string]any{
			"email": "test@example.com",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(session)
	}))
	defer server.Close()

	client := NewClient(&ClientConfig{
		PublicURL:  server.URL,
		CookieName: "ory_kratos_session",
		Timeout:    5 * time.Second,
	})

	identity, err := client.ValidateSession(context.Background(), "valid-session")

	require.NoError(t, err)
	assert.Equal(t, "user-456", identity.ID)
	assert.Equal(t, "session-123", identity.SessionID)
	assert.Equal(t, "test@example.com", identity.GetTraitString("email"))
}

func TestClient_ValidateSession_NoSession(t *testing.T) {
	client := NewClient(&ClientConfig{
		PublicURL:  "http://localhost:4433",
		CookieName: "ory_kratos_session",
	})

	identity, err := client.ValidateSession(context.Background(), "")

	assert.ErrorIs(t, err, ErrNoSession)
	assert.Nil(t, identity)
}

func TestClient_ValidateSession_Invalid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClient(&ClientConfig{
		PublicURL:  server.URL,
		CookieName: "ory_kratos_session",
		Timeout:    5 * time.Second,
	})

	identity, err := client.ValidateSession(context.Background(), "invalid-session")

	assert.ErrorIs(t, err, ErrInvalidSession)
	assert.Nil(t, identity)
}

func TestClient_ValidateSession_Expired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := Session{
			ID:        "session-123",
			Active:    false, // inactive = expired
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}
		session.Identity.ID = "user-456"

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(session)
	}))
	defer server.Close()

	client := NewClient(&ClientConfig{
		PublicURL:  server.URL,
		CookieName: "ory_kratos_session",
		Timeout:    5 * time.Second,
	})

	identity, err := client.ValidateSession(context.Background(), "expired-session")

	assert.ErrorIs(t, err, ErrSessionExpired)
	assert.Nil(t, identity)
}

func TestClient_ValidateSession_UnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(&ClientConfig{
		PublicURL:  server.URL,
		CookieName: "ory_kratos_session",
		Timeout:    5 * time.Second,
	})

	identity, err := client.ValidateSession(context.Background(), "some-session")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status")
	assert.Nil(t, identity)
}

func TestClient_ValidateSession_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{invalid json}`))
	}))
	defer server.Close()

	client := NewClient(&ClientConfig{
		PublicURL:  server.URL,
		CookieName: "ory_kratos_session",
		Timeout:    5 * time.Second,
	})

	identity, err := client.ValidateSession(context.Background(), "some-session")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode session")
	assert.Nil(t, identity)
}

func TestClient_ValidateSession_ConnectionError(t *testing.T) {
	client := NewClient(&ClientConfig{
		PublicURL:  "http://localhost:59999",
		CookieName: "ory_kratos_session",
		Timeout:    100 * time.Millisecond,
	})

	identity, err := client.ValidateSession(context.Background(), "some-session")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session validation failed")
	assert.Nil(t, identity)
}

func TestClient_ValidateSession_InvalidURL(t *testing.T) {
	client := NewClient(&ClientConfig{
		PublicURL:  "http://invalid\x00url", // control character makes URL invalid
		CookieName: "ory_kratos_session",
		Timeout:    5 * time.Second,
	})

	identity, err := client.ValidateSession(context.Background(), "some-session")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create request")
	assert.Nil(t, identity)
}

func TestClient_GetCookieName(t *testing.T) {
	client := NewClient(&ClientConfig{
		CookieName: "custom_session",
	})

	assert.Equal(t, "custom_session", client.GetCookieName())
}

func TestClient_Shutdown(t *testing.T) {
	client := NewClient(&ClientConfig{
		PublicURL: "http://localhost:4433",
		Timeout:   5 * time.Second,
	})

	err := client.Shutdown()
	assert.NoError(t, err)
}

func TestErrors(t *testing.T) {
	assert.Equal(t, "no session cookie", ErrNoSession.Error())
	assert.Equal(t, "invalid session", ErrInvalidSession.Error())
	assert.Equal(t, "session expired", ErrSessionExpired.Error())
}
