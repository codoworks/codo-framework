package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// publishOptions holds options for publishing messages.
type publishOptions struct {
	correlationID string
	headers       map[string]string
	priority      uint8
	expiration    string
	contentType   string
	persistent    bool
}

// PublishOption is a function that configures publish options.
type PublishOption func(*publishOptions)

// WithCorrelationID sets the correlation ID for the message.
func WithCorrelationID(id string) PublishOption {
	return func(o *publishOptions) {
		o.correlationID = id
	}
}

// WithHeaders sets custom headers for the message.
func WithHeaders(h map[string]string) PublishOption {
	return func(o *publishOptions) {
		o.headers = h
	}
}

// WithPriority sets the message priority (0-9).
func WithPriority(p uint8) PublishOption {
	return func(o *publishOptions) {
		if p > 9 {
			p = 9
		}
		o.priority = p
	}
}

// WithExpiration sets the message TTL.
func WithExpiration(ttl time.Duration) PublishOption {
	return func(o *publishOptions) {
		o.expiration = fmt.Sprintf("%d", ttl.Milliseconds())
	}
}

// WithPersistent marks the message as persistent (survives broker restart).
func WithPersistent() PublishOption {
	return func(o *publishOptions) {
		o.persistent = true
	}
}

// Publish sends a message to the specified topic.
func (c *Client) Publish(ctx context.Context, topic string, payload any, opts ...PublishOption) error {
	if err := c.Health(); err != nil {
		return fmt.Errorf("cannot publish: %w", err)
	}

	// Create message
	msg, err := NewMessage(topic, payload)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	// Apply options
	options := &publishOptions{
		contentType: "application/json",
		persistent:  true, // Default to persistent
	}
	for _, opt := range opts {
		opt(options)
	}

	// Set correlation ID if provided
	if options.correlationID != "" {
		msg.CorrelationID = options.correlationID
	}

	// Merge headers
	if options.headers != nil {
		for k, v := range options.headers {
			msg.SetHeader(k, v)
		}
	}

	return c.publishMessage(ctx, msg, options)
}

// PublishMessage publishes a pre-constructed message.
func (c *Client) PublishMessage(ctx context.Context, msg *Message, opts ...PublishOption) error {
	if err := c.Health(); err != nil {
		return fmt.Errorf("cannot publish: %w", err)
	}

	options := &publishOptions{
		contentType: "application/json",
		persistent:  true,
	}
	for _, opt := range opts {
		opt(options)
	}

	return c.publishMessage(ctx, msg, options)
}

// publishMessage is the internal publish implementation.
func (c *Client) publishMessage(ctx context.Context, msg *Message, opts *publishOptions) error {
	// Serialize the full message (includes metadata)
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Convert headers to AMQP table
	headers := make(amqp.Table)
	for k, v := range msg.Headers {
		headers[k] = v
	}
	headers["x-message-id"] = msg.ID
	headers["x-retry-count"] = msg.RetryCount

	// Build AMQP publishing
	publishing := amqp.Publishing{
		Headers:       headers,
		ContentType:   opts.contentType,
		Body:          body,
		Timestamp:     msg.Timestamp,
		MessageId:     msg.ID,
		CorrelationId: msg.CorrelationID,
		Priority:      opts.priority,
		Expiration:    opts.expiration,
	}

	if opts.persistent {
		publishing.DeliveryMode = amqp.Persistent
	} else {
		publishing.DeliveryMode = amqp.Transient
	}

	// Publish to exchange with topic as routing key
	return c.channel.PublishWithContext(
		ctx,
		c.config.Exchange, // exchange
		msg.Topic,         // routing key (topic)
		false,             // mandatory
		false,             // immediate
		publishing,
	)
}

// PublishDelayed publishes a message with a delay before delivery.
// Note: Requires the rabbitmq_delayed_message_exchange plugin.
func (c *Client) PublishDelayed(ctx context.Context, topic string, payload any, delay time.Duration, opts ...PublishOption) error {
	// Add delay header
	opts = append(opts, func(o *publishOptions) {
		if o.headers == nil {
			o.headers = make(map[string]string)
		}
		o.headers["x-delay"] = fmt.Sprintf("%d", delay.Milliseconds())
	})

	return c.Publish(ctx, topic, payload, opts...)
}
