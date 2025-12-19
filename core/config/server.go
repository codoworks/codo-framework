package config

import (
	"fmt"
	"time"
)

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	PublicPort    int      `yaml:"public_port"`
	ProtectedPort int      `yaml:"protected_port"`
	HiddenPort    int      `yaml:"hidden_port"`
	ReadTimeout   Duration `yaml:"read_timeout"`   // e.g., "30s", "1m"
	WriteTimeout  Duration `yaml:"write_timeout"`  // e.g., "30s", "1m"
	IdleTimeout   Duration `yaml:"idle_timeout"`   // e.g., "60s", "1m"
	ShutdownGrace Duration `yaml:"shutdown_grace"` // e.g., "20s", "30s"
}

// DefaultServerConfig returns default server configuration
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		PublicPort:    8081,
		ProtectedPort: 8080,
		HiddenPort:    8079,
		ReadTimeout:   Duration(30 * time.Second),
		WriteTimeout:  Duration(30 * time.Second),
		IdleTimeout:   Duration(60 * time.Second),
		ShutdownGrace: Duration(20 * time.Second),
	}
}

// Validate validates server configuration
func (c *ServerConfig) Validate() error {
	if c.PublicPort < 1 || c.PublicPort > 65535 {
		return fmt.Errorf("server.public_port must be between 1 and 65535")
	}
	if c.ProtectedPort < 1 || c.ProtectedPort > 65535 {
		return fmt.Errorf("server.protected_port must be between 1 and 65535")
	}
	if c.HiddenPort < 1 || c.HiddenPort > 65535 {
		return fmt.Errorf("server.hidden_port must be between 1 and 65535")
	}
	// Check for port conflicts
	if c.PublicPort == c.ProtectedPort {
		return fmt.Errorf("server ports must be unique: public_port and protected_port are both %d", c.PublicPort)
	}
	if c.PublicPort == c.HiddenPort {
		return fmt.Errorf("server ports must be unique: public_port and hidden_port are both %d", c.PublicPort)
	}
	if c.ProtectedPort == c.HiddenPort {
		return fmt.Errorf("server ports must be unique: protected_port and hidden_port are both %d", c.ProtectedPort)
	}
	if c.ReadTimeout.Duration() <= 0 {
		return fmt.Errorf("server.read_timeout must be positive")
	}
	if c.WriteTimeout.Duration() <= 0 {
		return fmt.Errorf("server.write_timeout must be positive")
	}
	if c.IdleTimeout.Duration() <= 0 {
		return fmt.Errorf("server.idle_timeout must be positive")
	}
	if c.ShutdownGrace.Duration() <= 0 {
		return fmt.Errorf("server.shutdown_grace must be positive")
	}
	return nil
}

// PublicAddr returns the public API address
func (c *ServerConfig) PublicAddr() string {
	return fmt.Sprintf(":%d", c.PublicPort)
}

// ProtectedAddr returns the protected API address
func (c *ServerConfig) ProtectedAddr() string {
	return fmt.Sprintf(":%d", c.ProtectedPort)
}

// HiddenAddr returns the hidden API address
func (c *ServerConfig) HiddenAddr() string {
	return fmt.Sprintf(":%d", c.HiddenPort)
}
