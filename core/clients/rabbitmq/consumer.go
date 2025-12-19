package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// subscribeOptions holds options for subscribing to a topic.
type subscribeOptions struct {
	queueName   string
	workers     int
	retryConfig *RetryConfig
	autoAck     bool
	exclusive   bool
}

// SubscribeOption is a function that configures subscribe options.
type SubscribeOption func(*subscribeOptions)

// WithQueueName sets a custom queue name.
// If not set, a queue name is auto-generated from the topic.
func WithQueueName(name string) SubscribeOption {
	return func(o *subscribeOptions) {
		o.queueName = name
	}
}

// WithWorkers sets the number of concurrent workers processing messages.
// Default is 1.
func WithWorkers(n int) SubscribeOption {
	return func(o *subscribeOptions) {
		if n < 1 {
			n = 1
		}
		o.workers = n
	}
}

// WithRetryPolicy overrides the default retry configuration for this consumer.
func WithRetryPolicy(cfg RetryConfig) SubscribeOption {
	return func(o *subscribeOptions) {
		o.retryConfig = &cfg
	}
}

// WithAutoAck enables automatic acknowledgement of messages.
// Use with caution - messages will be acked before processing completes.
func WithAutoAck() SubscribeOption {
	return func(o *subscribeOptions) {
		o.autoAck = true
	}
}

// WithExclusive makes the queue exclusive to this consumer.
func WithExclusive() SubscribeOption {
	return func(o *subscribeOptions) {
		o.exclusive = true
	}
}

// consumer represents a registered message consumer.
type consumer struct {
	id          string
	topic       string
	queueName   string
	handler     HandlerFunc
	workers     int
	retryConfig RetryConfig
	autoAck     bool
	exclusive   bool

	deliveries <-chan amqp.Delivery
	stopCh     chan struct{}
	wg         sync.WaitGroup
	running    bool
	mu         sync.Mutex
}

// Subscribe registers a handler for messages on the given topic pattern.
// The handler will be called for each message that matches the topic.
// Topic patterns support wildcards:
//   - * matches exactly one word (e.g., "payments.*.created")
//   - # matches zero or more words (e.g., "payments.#")
func (c *Client) Subscribe(topic string, handler HandlerFunc, opts ...SubscribeOption) error {
	options := &subscribeOptions{
		workers: 1,
	}
	for _, opt := range opts {
		opt(options)
	}

	// Generate queue name if not provided
	queueName := options.queueName
	if queueName == "" {
		queueName = c.generateQueueName(topic)
	}

	// Use default retry config if not overridden
	retryConfig := c.config.Retry
	if options.retryConfig != nil {
		retryConfig = *options.retryConfig
	}

	cons := &consumer{
		id:          uuid.New().String(),
		topic:       topic,
		queueName:   queueName,
		handler:     handler,
		workers:     options.workers,
		retryConfig: retryConfig,
		autoAck:     options.autoAck,
		exclusive:   options.exclusive,
		stopCh:      make(chan struct{}),
	}

	// Declare queue
	if err := c.declareQueue(cons); err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Declare DLQ for this queue
	if err := c.declareDLQForQueue(queueName, topic); err != nil {
		return fmt.Errorf("failed to declare DLQ: %w", err)
	}

	// Register consumer
	c.mu.Lock()
	c.consumers[cons.id] = cons
	c.mu.Unlock()

	return nil
}

// generateQueueName creates a queue name from a topic pattern.
// Replaces wildcards and special characters.
func (c *Client) generateQueueName(topic string) string {
	// Replace wildcards with readable names
	name := strings.ReplaceAll(topic, "*", "star")
	name = strings.ReplaceAll(name, "#", "hash")
	name = strings.ReplaceAll(name, ".", "-")

	// Prefix with service name if available
	return "q-" + name
}

// declareQueue declares a queue and binds it to the exchange.
func (c *Client) declareQueue(cons *consumer) error {
	// Declare the queue
	_, err := c.channel.QueueDeclare(
		cons.queueName,
		true,           // durable
		false,          // auto-delete
		cons.exclusive, // exclusive
		false,          // no-wait
		amqp.Table{
			"x-queue-type": "classic",
		},
	)
	if err != nil {
		return err
	}

	// Bind queue to exchange with topic pattern
	return c.channel.QueueBind(
		cons.queueName,
		cons.topic,
		c.config.Exchange,
		false,
		nil,
	)
}

// start begins consuming messages for this consumer.
func (cons *consumer) start(c *Client) {
	cons.mu.Lock()
	if cons.running {
		cons.mu.Unlock()
		return
	}
	cons.running = true
	cons.stopCh = make(chan struct{})
	cons.mu.Unlock()

	// Start consuming
	deliveries, err := c.channel.Consume(
		cons.queueName,
		cons.id,       // consumer tag
		cons.autoAck,  // auto-ack
		cons.exclusive, // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)
	if err != nil {
		cons.mu.Lock()
		cons.running = false
		cons.mu.Unlock()
		return
	}

	cons.deliveries = deliveries

	// Start worker goroutines
	for i := 0; i < cons.workers; i++ {
		cons.wg.Add(1)
		go cons.worker(c, i)
	}
}

// worker processes messages from the delivery channel.
func (cons *consumer) worker(c *Client, id int) {
	defer cons.wg.Done()

	retryHandler := &retryHandler{
		client:  c,
		handler: cons.handler,
		config:  cons.retryConfig,
		queue:   cons.queueName,
	}

	for {
		select {
		case <-cons.stopCh:
			return
		case delivery, ok := <-cons.deliveries:
			if !ok {
				return
			}
			cons.processDelivery(c, delivery, retryHandler)
		}
	}
}

// processDelivery handles a single message delivery.
func (cons *consumer) processDelivery(c *Client, delivery amqp.Delivery, retryHandler *retryHandler) {
	ctx := context.Background()

	// Parse message
	var msg Message
	if err := json.Unmarshal(delivery.Body, &msg); err != nil {
		// Malformed message - reject without requeue
		delivery.Reject(false)
		return
	}

	// Process with retry handler
	if err := retryHandler.handle(ctx, &msg); err != nil {
		// Handler failed after all retries - message is in DLQ
		if !cons.autoAck {
			delivery.Ack(false)
		}
		return
	}

	// Success - acknowledge
	if !cons.autoAck {
		delivery.Ack(false)
	}
}

// stop signals the consumer to stop processing.
func (cons *consumer) stop() {
	cons.mu.Lock()
	if !cons.running {
		cons.mu.Unlock()
		return
	}
	cons.running = false
	close(cons.stopCh)
	cons.mu.Unlock()

	cons.wg.Wait()
}

// Unsubscribe removes a consumer by its topic.
func (c *Client) Unsubscribe(topic string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for id, cons := range c.consumers {
		if cons.topic == topic {
			cons.stop()
			delete(c.consumers, id)
			return nil
		}
	}

	return fmt.Errorf("no consumer found for topic: %s", topic)
}

// ConsumerCount returns the number of registered consumers.
func (c *Client) ConsumerCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.consumers)
}

// ConsumerInfo contains information about a consumer.
type ConsumerInfo struct {
	ID        string `json:"id"`
	Topic     string `json:"topic"`
	QueueName string `json:"queue_name"`
	Workers   int    `json:"workers"`
	Running   bool   `json:"running"`
}

// ListConsumers returns information about all registered consumers.
func (c *Client) ListConsumers() []ConsumerInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	infos := make([]ConsumerInfo, 0, len(c.consumers))
	for _, cons := range c.consumers {
		cons.mu.Lock()
		infos = append(infos, ConsumerInfo{
			ID:        cons.id,
			Topic:     cons.topic,
			QueueName: cons.queueName,
			Workers:   cons.workers,
			Running:   cons.running,
		})
		cons.mu.Unlock()
	}

	return infos
}
