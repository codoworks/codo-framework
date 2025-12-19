package rabbitmq

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/codoworks/codo-framework/core/clients"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	// ClientName is the name of the RabbitMQ client.
	ClientName = "rabbitmq"
)

// HandlerFunc is the function signature for message handlers.
type HandlerFunc func(ctx context.Context, msg *Message) error

// RabbitMQClient is the interface for RabbitMQ operations.
type RabbitMQClient interface {
	clients.Client

	// Publish sends a message to the specified topic.
	Publish(ctx context.Context, topic string, payload any, opts ...PublishOption) error

	// Subscribe registers a handler for messages on the given topic pattern.
	Subscribe(topic string, handler HandlerFunc, opts ...SubscribeOption) error

	// GetChannel returns the underlying AMQP channel.
	GetChannel() *amqp.Channel

	// GetConnection returns the underlying AMQP connection.
	GetConnection() *amqp.Connection
}

// Client is the RabbitMQ client implementation.
type Client struct {
	clients.BaseClient

	config    *Config
	conn      *amqp.Connection
	channel   *amqp.Channel
	consumers map[string]*consumer
	mu        sync.RWMutex

	// For reconnection handling
	connClose   chan *amqp.Error
	chanClose   chan *amqp.Error
	stopReconnect chan struct{}
	wg          sync.WaitGroup
}

// New creates a new RabbitMQ client.
func New() *Client {
	return &Client{
		BaseClient:    clients.NewBaseClient(ClientName),
		consumers:     make(map[string]*consumer),
		stopReconnect: make(chan struct{}),
	}
}

// Name returns the client name.
func (c *Client) Name() string {
	return ClientName
}

// Initialize sets up the RabbitMQ client with configuration.
func (c *Client) Initialize(cfg any) error {
	config := DefaultConfig()

	if cfg != nil {
		switch v := cfg.(type) {
		case *Config:
			config = v
		case Config:
			config = &v
		case map[string]any:
			if err := c.parseMapConfig(config, v); err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid config type: %T", cfg)
		}
	}

	c.config = config

	// Connect to RabbitMQ
	if err := c.connect(); err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Declare the exchange
	if err := c.declareExchange(); err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Start reconnection handler
	c.wg.Add(1)
	go c.handleReconnect()

	return c.BaseClient.Initialize(cfg)
}

// parseMapConfig parses a map[string]any into the Config struct.
func (c *Client) parseMapConfig(config *Config, m map[string]any) error {
	if url, ok := m["url"].(string); ok {
		config.URL = url
	}
	if host, ok := m["host"].(string); ok {
		config.Host = host
	}
	if port, ok := m["port"].(int); ok {
		config.Port = port
	}
	if user, ok := m["user"].(string); ok {
		config.User = user
	}
	if password, ok := m["password"].(string); ok {
		config.Password = password
	}
	if vhost, ok := m["vhost"].(string); ok {
		config.VHost = vhost
	}
	if exchange, ok := m["exchange"].(string); ok {
		config.Exchange = exchange
	}
	if exchangeType, ok := m["exchange_type"].(string); ok {
		config.ExchangeType = exchangeType
	}
	if prefetch, ok := m["prefetch_count"].(int); ok {
		config.PrefetchCount = prefetch
	}
	if retry, ok := m["retry"].(map[string]any); ok {
		if maxRetries, ok := retry["max_retries"].(int); ok {
			config.Retry.MaxRetries = maxRetries
		}
	}
	return nil
}

// connect establishes a connection to RabbitMQ.
func (c *Client) connect() error {
	conn, err := amqp.Dial(c.config.GetURL())
	if err != nil {
		return err
	}
	c.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return err
	}
	c.channel = ch

	// Set QoS
	if err := ch.Qos(c.config.PrefetchCount, 0, c.config.PrefetchGlobal); err != nil {
		ch.Close()
		conn.Close()
		return err
	}

	// Set up close notifications
	c.connClose = make(chan *amqp.Error, 1)
	c.chanClose = make(chan *amqp.Error, 1)
	c.conn.NotifyClose(c.connClose)
	c.channel.NotifyClose(c.chanClose)

	return nil
}

// declareExchange declares the configured exchange.
func (c *Client) declareExchange() error {
	return c.channel.ExchangeDeclare(
		c.config.Exchange,     // name
		c.config.ExchangeType, // type (topic)
		true,                  // durable
		false,                 // auto-deleted
		false,                 // internal
		false,                 // no-wait
		nil,                   // arguments
	)
}

// handleReconnect handles connection loss and attempts to reconnect.
func (c *Client) handleReconnect() {
	defer c.wg.Done()

	for {
		select {
		case <-c.stopReconnect:
			return
		case err := <-c.connClose:
			if err != nil {
				c.reconnect()
			}
		case err := <-c.chanClose:
			if err != nil {
				c.reconnect()
			}
		}
	}
}

// reconnect attempts to reconnect with exponential backoff.
func (c *Client) reconnect() {
	delay := c.config.ReconnectDelay

	for {
		select {
		case <-c.stopReconnect:
			return
		default:
		}

		if err := c.connect(); err == nil {
			if err := c.declareExchange(); err == nil {
				// Restart all consumers
				c.restartConsumers()
				return
			}
		}

		time.Sleep(delay)
		delay *= 2
		if delay > c.config.MaxReconnectDelay {
			delay = c.config.MaxReconnectDelay
		}
	}
}

// restartConsumers restarts all registered consumers after reconnection.
func (c *Client) restartConsumers() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, consumer := range c.consumers {
		go consumer.start(c)
	}
}

// Health checks if the RabbitMQ connection is healthy.
func (c *Client) Health() error {
	if c.conn == nil || c.conn.IsClosed() {
		return fmt.Errorf("rabbitmq connection is closed")
	}
	if c.channel == nil || c.channel.IsClosed() {
		return fmt.Errorf("rabbitmq channel is closed")
	}
	return nil
}

// Shutdown gracefully shuts down the RabbitMQ client.
func (c *Client) Shutdown() error {
	// Stop reconnection handler
	close(c.stopReconnect)

	// Stop all consumers
	c.mu.Lock()
	for _, consumer := range c.consumers {
		consumer.stop()
	}
	c.mu.Unlock()

	// Wait for goroutines to finish
	c.wg.Wait()

	// Close channel and connection
	var errs []error
	if c.channel != nil && !c.channel.IsClosed() {
		if err := c.channel.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if c.conn != nil && !c.conn.IsClosed() {
		if err := c.conn.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	return c.BaseClient.Shutdown()
}

// Start begins consuming messages for all registered consumers.
// Implements the clients.Startable interface.
func (c *Client) Start() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, cons := range c.consumers {
		c.wg.Add(1)
		go func(cn *consumer) {
			defer c.wg.Done()
			cn.start(c)
		}(cons)
	}

	return nil
}

// Stop stops all consumers.
// Implements the clients.Stoppable interface.
func (c *Client) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, consumer := range c.consumers {
		consumer.stop()
	}

	return nil
}

// GetChannel returns the underlying AMQP channel.
func (c *Client) GetChannel() *amqp.Channel {
	return c.channel
}

// GetConnection returns the underlying AMQP connection.
func (c *Client) GetConnection() *amqp.Connection {
	return c.conn
}

// GetConfig returns the client configuration.
func (c *Client) GetConfig() *Config {
	return c.config
}
