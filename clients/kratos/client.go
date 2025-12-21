package kratos

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/codoworks/codo-framework/core/auth"
)

// ClientConfig holds Kratos client configuration
type ClientConfig struct {
	PublicURL  string
	AdminURL   string
	CookieName string
	Timeout    time.Duration
}

// Client is the Ory Kratos client
type Client struct {
	config     *ClientConfig
	httpClient *http.Client
}

// NewClient creates a new Kratos client
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
	return "kratos"
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

// Health checks if Kratos is reachable
func (c *Client) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", c.config.PublicURL+"/health/alive", nil)
	if err != nil {
		return fmt.Errorf("failed to create health request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("kratos unhealthy: status %d", resp.StatusCode)
	}

	return nil
}

// Shutdown closes the client
func (c *Client) Shutdown() error {
	c.httpClient.CloseIdleConnections()
	return nil
}

// ValidateSession validates a session cookie and returns the identity
func (c *Client) ValidateSession(ctx context.Context, cookie string) (*auth.Identity, error) {
	if cookie == "" {
		return nil, ErrNoSession
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.config.PublicURL+"/sessions/whoami", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Cookie", fmt.Sprintf("%s=%s", c.config.CookieName, cookie))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("session validation failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrInvalidSession
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var session Session
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("failed to decode session: %w", err)
	}

	if !session.Active {
		return nil, ErrSessionExpired
	}

	return &auth.Identity{
		ID:     session.Identity.ID,
		Traits: session.Identity.Traits,
	}, nil
}

// GetCookieName returns the session cookie name
func (c *Client) GetCookieName() string {
	return c.config.CookieName
}

// Errors
var (
	ErrNoSession      = fmt.Errorf("no session cookie")
	ErrInvalidSession = fmt.Errorf("invalid session")
	ErrSessionExpired = fmt.Errorf("session expired")
)
