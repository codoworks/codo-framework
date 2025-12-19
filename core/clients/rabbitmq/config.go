// Package rabbitmq provides a RabbitMQ client for queue-based messaging.
package rabbitmq

import "time"

// Config holds RabbitMQ configuration.
type Config struct {
	// Connection - use URL or individual fields
	URL      string `json:"url" yaml:"url"`           // amqp://user:pass@host:5672/vhost
	Host     string `json:"host" yaml:"host"`         // Alternative to URL
	Port     int    `json:"port" yaml:"port"`         // Default: 5672
	User     string `json:"user" yaml:"user"`         // Default: guest
	Password string `json:"password" yaml:"password"` // Default: guest
	VHost    string `json:"vhost" yaml:"vhost"`       // Default: /

	// Exchange configuration
	Exchange     string `json:"exchange" yaml:"exchange"`           // e.g., "codo.events"
	ExchangeType string `json:"exchange_type" yaml:"exchange_type"` // Default: "topic"

	// Connection resilience
	ReconnectDelay    time.Duration `json:"reconnect_delay" yaml:"reconnect_delay"`         // Default: 1s
	MaxReconnectDelay time.Duration `json:"max_reconnect_delay" yaml:"max_reconnect_delay"` // Default: 30s

	// Consumer defaults
	PrefetchCount  int  `json:"prefetch_count" yaml:"prefetch_count"`   // Default: 10
	PrefetchGlobal bool `json:"prefetch_global" yaml:"prefetch_global"` // Default: false

	// Retry policy for failed messages
	Retry RetryConfig `json:"retry" yaml:"retry"`
}

// RetryConfig configures retry behavior for failed message processing.
type RetryConfig struct {
	MaxRetries     int           `json:"max_retries" yaml:"max_retries"`         // Default: 3
	InitialBackoff time.Duration `json:"initial_backoff" yaml:"initial_backoff"` // Default: 1s
	MaxBackoff     time.Duration `json:"max_backoff" yaml:"max_backoff"`         // Default: 30s
	Multiplier     float64       `json:"multiplier" yaml:"multiplier"`           // Default: 2.0
}

// DefaultConfig returns default RabbitMQ configuration.
func DefaultConfig() *Config {
	return &Config{
		Host:              "localhost",
		Port:              5672,
		User:              "guest",
		Password:          "guest",
		VHost:             "/",
		Exchange:          "codo.events",
		ExchangeType:      "topic",
		ReconnectDelay:    1 * time.Second,
		MaxReconnectDelay: 30 * time.Second,
		PrefetchCount:     10,
		PrefetchGlobal:    false,
		Retry:             DefaultRetryConfig(),
	}
}

// DefaultRetryConfig returns default retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		Multiplier:     2.0,
	}
}

// GetURL returns the AMQP connection URL.
// If URL is set, returns it directly. Otherwise builds from individual fields.
func (c *Config) GetURL() string {
	if c.URL != "" {
		return c.URL
	}
	// Build URL from components: amqp://user:pass@host:port/vhost
	vhost := c.VHost
	if vhost == "/" {
		vhost = ""
	}
	return "amqp://" + c.User + ":" + c.Password + "@" + c.Host + ":" + itoa(c.Port) + "/" + vhost
}

// itoa converts int to string without importing strconv.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b [20]byte
	pos := len(b)
	neg := i < 0
	if neg {
		i = -i
	}
	for i > 0 {
		pos--
		b[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		b[pos] = '-'
	}
	return string(b[pos:])
}
