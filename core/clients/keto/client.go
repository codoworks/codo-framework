package keto

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// ClientConfig holds Keto client configuration
type ClientConfig struct {
	ReadURL  string
	WriteURL string
	Timeout  time.Duration
}

// Client is the Ory Keto client
type Client struct {
	config     *ClientConfig
	httpClient *http.Client
}

// NewClient creates a new Keto client
func NewClient(cfg *ClientConfig) *Client {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Name returns the client name
func (c *Client) Name() string {
	return "keto"
}

// Initialize sets up the client
func (c *Client) Initialize(cfg any) error {
	if clientCfg, ok := cfg.(*ClientConfig); ok {
		c.config = clientCfg
		timeout := clientCfg.Timeout
		if timeout <= 0 {
			timeout = 10 * time.Second
		}
		c.httpClient = &http.Client{
			Timeout: timeout,
		}
	}
	return nil
}

// Health checks if Keto is reachable
func (c *Client) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", c.config.ReadURL+"/health/alive", nil)
	if err != nil {
		return fmt.Errorf("failed to create health request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("keto unhealthy: status %d", resp.StatusCode)
	}

	return nil
}

// Shutdown closes the client
func (c *Client) Shutdown() error {
	c.httpClient.CloseIdleConnections()
	return nil
}

// CheckPermission checks if a subject has a relation on an object
func (c *Client) CheckPermission(ctx context.Context, subject, relation, namespace, object string) (bool, error) {
	params := url.Values{}
	params.Set("subject_id", subject)
	params.Set("relation", relation)
	params.Set("namespace", namespace)
	params.Set("object", object)

	reqURL := fmt.Sprintf("%s/relation-tuples/check?%s", c.config.ReadURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("permission check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return false, nil
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result struct {
		Allowed bool `json:"allowed"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Allowed, nil
}
