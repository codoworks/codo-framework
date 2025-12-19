package config

import (
	"fmt"
	"time"
)

// RabbitMQConfig holds RabbitMQ configuration
type RabbitMQConfig struct {
	// Connection - use URL or individual fields
	URL      string `yaml:"url"`      // amqp://user:pass@host:5672/vhost
	Host     string `yaml:"host"`     // Alternative to URL
	Port     int    `yaml:"port"`     // Default: 5672
	User     string `yaml:"user"`     // Default: guest
	Password string `yaml:"password"` // Default: guest
	VHost    string `yaml:"vhost"`    // Default: /

	// Exchange configuration
	Exchange     string `yaml:"exchange"`      // e.g., "codo.events"
	ExchangeType string `yaml:"exchange_type"` // Default: "topic"

	// Connection resilience
	ReconnectDelay    Duration `yaml:"reconnect_delay"`     // e.g., "1s", "5s"
	MaxReconnectDelay Duration `yaml:"max_reconnect_delay"` // e.g., "30s", "1m"

	// Consumer defaults
	PrefetchCount  int  `yaml:"prefetch_count"`  // Default: 10
	PrefetchGlobal bool `yaml:"prefetch_global"` // Default: false

	// Retry policy
	Retry RabbitMQRetryConfig `yaml:"retry"`
}

// RabbitMQRetryConfig configures retry behavior for failed messages
type RabbitMQRetryConfig struct {
	MaxRetries     int      `yaml:"max_retries"`     // Default: 3
	InitialBackoff Duration `yaml:"initial_backoff"` // e.g., "1s", "5s"
	MaxBackoff     Duration `yaml:"max_backoff"`     // e.g., "30s", "1m"
	Multiplier     float64  `yaml:"multiplier"`      // Default: 2.0
}

// DefaultRabbitMQConfig returns default RabbitMQ configuration
func DefaultRabbitMQConfig() RabbitMQConfig {
	return RabbitMQConfig{
		Host:              "localhost",
		Port:              5672,
		User:              "guest",
		Password:          "guest",
		VHost:             "/",
		Exchange:          "codo.events",
		ExchangeType:      "topic",
		ReconnectDelay:    Duration(1 * time.Second),
		MaxReconnectDelay: Duration(30 * time.Second),
		PrefetchCount:     10,
		PrefetchGlobal:    false,
		Retry:             DefaultRabbitMQRetryConfig(),
	}
}

// DefaultRabbitMQRetryConfig returns default retry configuration
func DefaultRabbitMQRetryConfig() RabbitMQRetryConfig {
	return RabbitMQRetryConfig{
		MaxRetries:     3,
		InitialBackoff: Duration(1 * time.Second),
		MaxBackoff:     Duration(30 * time.Second),
		Multiplier:     2.0,
	}
}

// Validate validates RabbitMQ configuration
func (c *RabbitMQConfig) Validate() error {
	// URL takes precedence, so only validate components if URL is not set
	if c.URL == "" {
		if c.Host == "" {
			return fmt.Errorf("rabbitmq.host is required when url is not set")
		}
		if c.Port < 1 || c.Port > 65535 {
			return fmt.Errorf("rabbitmq.port must be between 1 and 65535")
		}
	}

	if c.Exchange == "" {
		return fmt.Errorf("rabbitmq.exchange is required")
	}

	validExchangeTypes := map[string]bool{"topic": true, "direct": true, "fanout": true, "headers": true}
	if c.ExchangeType != "" && !validExchangeTypes[c.ExchangeType] {
		return fmt.Errorf("rabbitmq.exchange_type must be one of: topic, direct, fanout, headers")
	}

	if c.PrefetchCount < 0 {
		return fmt.Errorf("rabbitmq.prefetch_count must be non-negative")
	}

	if c.Retry.MaxRetries < 0 {
		return fmt.Errorf("rabbitmq.retry.max_retries must be non-negative")
	}

	if c.Retry.Multiplier <= 0 {
		return fmt.Errorf("rabbitmq.retry.multiplier must be positive")
	}

	return nil
}

// GetURL returns the AMQP connection URL.
// If URL is set, returns it directly. Otherwise builds from individual fields.
func (c *RabbitMQConfig) GetURL() string {
	if c.URL != "" {
		return c.URL
	}
	// Build URL from components: amqp://user:pass@host:port/vhost
	vhost := c.VHost
	if vhost == "/" {
		vhost = ""
	}
	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s", c.User, c.Password, c.Host, c.Port, vhost)
}

// IsEnabled returns true if RabbitMQ is configured (has URL or host)
func (c *RabbitMQConfig) IsEnabled() bool {
	return c.URL != "" || c.Host != ""
}
