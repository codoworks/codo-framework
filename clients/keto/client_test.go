package keto

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	cfg := &ClientConfig{
		ReadURL:  "http://localhost:4466",
		WriteURL: "http://localhost:4467",
		Timeout:  5 * time.Second,
	}

	client := NewClient(cfg)

	assert.NotNil(t, client)
	assert.Equal(t, cfg, client.config)
	assert.NotNil(t, client.httpClient)
}

func TestNewClient_DefaultTimeout(t *testing.T) {
	cfg := &ClientConfig{
		ReadURL: "http://localhost:4466",
		Timeout: 0, // should default to 10s
	}

	client := NewClient(cfg)

	assert.NotNil(t, client.httpClient)
	assert.Equal(t, 10*time.Second, client.httpClient.Timeout)
}

func TestClient_Name(t *testing.T) {
	client := NewClient(&ClientConfig{})
	assert.Equal(t, "keto", client.Name())
}

func TestClient_Initialize(t *testing.T) {
	client := &Client{}
	cfg := &ClientConfig{
		ReadURL:  "http://localhost:4466",
		WriteURL: "http://localhost:4467",
		Timeout:  5 * time.Second,
	}

	err := client.Initialize(cfg)

	assert.NoError(t, err)
	assert.Equal(t, cfg, client.config)
	assert.NotNil(t, client.httpClient)
}

func TestClient_Initialize_WrongType(t *testing.T) {
	client := &Client{
		config: &ClientConfig{ReadURL: "original"},
	}

	err := client.Initialize("not a config")

	assert.NoError(t, err) // doesn't error, just ignores
	assert.Equal(t, "original", client.config.ReadURL)
}

func TestClient_Initialize_DefaultTimeout(t *testing.T) {
	client := &Client{}
	cfg := &ClientConfig{
		ReadURL: "http://localhost:4466",
		Timeout: 0,
	}

	err := client.Initialize(cfg)

	assert.NoError(t, err)
	assert.Equal(t, 10*time.Second, client.httpClient.Timeout)
}

func TestClient_Health(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/health/alive", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(&ClientConfig{
		ReadURL: server.URL,
		Timeout: 5 * time.Second,
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
		ReadURL: server.URL,
		Timeout: 5 * time.Second,
	})

	err := client.Health()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "keto unhealthy")
}

func TestClient_Health_ConnectionError(t *testing.T) {
	client := NewClient(&ClientConfig{
		ReadURL: "http://localhost:59999", // unlikely to be running
		Timeout: 100 * time.Millisecond,
	})

	err := client.Health()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "health check failed")
}

func TestClient_Health_InvalidURL(t *testing.T) {
	client := NewClient(&ClientConfig{
		ReadURL: "http://invalid\x00url", // control character makes URL invalid
		Timeout: 5 * time.Second,
	})

	err := client.Health()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create health request")
}

func TestClient_CheckPermission(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/relation-tuples/check", r.URL.Path)
		assert.Equal(t, "user-123", r.URL.Query().Get("subject_id"))
		assert.Equal(t, "viewer", r.URL.Query().Get("relation"))
		assert.Equal(t, "documents", r.URL.Query().Get("namespace"))
		assert.Equal(t, "doc-456", r.URL.Query().Get("object"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"allowed": true})
	}))
	defer server.Close()

	client := NewClient(&ClientConfig{
		ReadURL: server.URL,
		Timeout: 5 * time.Second,
	})

	allowed, err := client.CheckPermission(context.Background(), "user-123", "viewer", "documents", "doc-456")

	assert.NoError(t, err)
	assert.True(t, allowed)
}

func TestClient_CheckPermission_Denied(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"allowed": false})
	}))
	defer server.Close()

	client := NewClient(&ClientConfig{
		ReadURL: server.URL,
		Timeout: 5 * time.Second,
	})

	allowed, err := client.CheckPermission(context.Background(), "user-123", "admin", "documents", "doc-456")

	assert.NoError(t, err)
	assert.False(t, allowed)
}

func TestClient_CheckPermission_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := NewClient(&ClientConfig{
		ReadURL: server.URL,
		Timeout: 5 * time.Second,
	})

	allowed, err := client.CheckPermission(context.Background(), "user-123", "admin", "documents", "doc-456")

	assert.NoError(t, err)
	assert.False(t, allowed)
}

func TestClient_CheckPermission_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(&ClientConfig{
		ReadURL: server.URL,
		Timeout: 5 * time.Second,
	})

	allowed, err := client.CheckPermission(context.Background(), "user-123", "admin", "documents", "doc-456")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status")
	assert.False(t, allowed)
}

func TestClient_CheckPermission_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{invalid json}`))
	}))
	defer server.Close()

	client := NewClient(&ClientConfig{
		ReadURL: server.URL,
		Timeout: 5 * time.Second,
	})

	allowed, err := client.CheckPermission(context.Background(), "user-123", "admin", "documents", "doc-456")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode response")
	assert.False(t, allowed)
}

func TestClient_CheckPermission_ConnectionError(t *testing.T) {
	client := NewClient(&ClientConfig{
		ReadURL: "http://localhost:59999",
		Timeout: 100 * time.Millisecond,
	})

	allowed, err := client.CheckPermission(context.Background(), "user-123", "admin", "documents", "doc-456")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission check failed")
	assert.False(t, allowed)
}

func TestClient_CheckPermission_InvalidURL(t *testing.T) {
	client := NewClient(&ClientConfig{
		ReadURL: "http://invalid\x00url", // control character makes URL invalid
		Timeout: 5 * time.Second,
	})

	allowed, err := client.CheckPermission(context.Background(), "user-123", "admin", "documents", "doc-456")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create request")
	assert.False(t, allowed)
}

func TestClient_Shutdown(t *testing.T) {
	client := NewClient(&ClientConfig{
		ReadURL: "http://localhost:4466",
		Timeout: 5 * time.Second,
	})

	err := client.Shutdown()
	assert.NoError(t, err)
}
