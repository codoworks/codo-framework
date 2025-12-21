package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	// DLQSuffix is appended to queue names for dead letter queues.
	DLQSuffix = ".dlq"
)

// DLQName returns the dead letter queue name for a given queue.
func DLQName(queue string) string {
	return queue + DLQSuffix
}

// sendToDLQ sends a failed message to the dead letter queue.
func (c *Client) sendToDLQ(ctx context.Context, msg *Message, originalQueue string, lastError error) error {
	dlqMsg := msg.Clone()
	dlqMsg.OriginalTopic = msg.Topic
	dlqMsg.Topic = msg.Topic + DLQSuffix
	dlqMsg.Error = lastError.Error()
	dlqMsg.SetHeader("x-dlq-reason", lastError.Error())
	dlqMsg.SetHeader("x-dlq-timestamp", time.Now().UTC().Format(time.RFC3339))
	dlqMsg.SetHeader("x-original-queue", originalQueue)
	dlqMsg.SetHeader("x-retry-count", itoa(msg.RetryCount))

	// Ensure DLQ exists
	dlqName := DLQName(originalQueue)
	if err := c.declareDLQ(dlqName, msg.Topic+DLQSuffix); err != nil {
		return fmt.Errorf("failed to declare DLQ: %w", err)
	}

	return c.PublishMessage(ctx, dlqMsg)
}

// declareDLQ declares a dead letter queue and binds it to the exchange.
func (c *Client) declareDLQ(queueName, routingKey string) error {
	// Declare the DLQ
	_, err := c.channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	// Bind to exchange
	return c.channel.QueueBind(
		queueName,
		routingKey,
		c.config.Exchange,
		false,
		nil,
	)
}

// ReplayDLQ replays messages from a dead letter queue back to their original topic.
// Returns the number of messages replayed.
func (c *Client) ReplayDLQ(ctx context.Context, queue string, count int) (int, error) {
	if err := c.Health(); err != nil {
		return 0, fmt.Errorf("cannot replay DLQ: %w", err)
	}

	dlqName := DLQName(queue)
	replayed := 0

	for i := 0; i < count; i++ {
		// Get a message from DLQ (no auto-ack)
		delivery, ok, err := c.channel.Get(dlqName, false)
		if err != nil {
			return replayed, fmt.Errorf("failed to get message from DLQ: %w", err)
		}
		if !ok {
			// No more messages
			break
		}

		// Parse the message
		var msg Message
		if err := json.Unmarshal(delivery.Body, &msg); err != nil {
			// Reject malformed message
			delivery.Reject(false)
			continue
		}

		// Reset retry count and restore original topic
		msg.RetryCount = 0
		if msg.OriginalTopic != "" {
			msg.Topic = msg.OriginalTopic
		}
		msg.OriginalTopic = ""
		msg.Error = ""

		// Re-publish to original topic
		if err := c.PublishMessage(ctx, &msg); err != nil {
			// Requeue the message on failure
			delivery.Reject(true)
			return replayed, fmt.Errorf("failed to republish message: %w", err)
		}

		// Acknowledge successful replay
		if err := delivery.Ack(false); err != nil {
			return replayed, fmt.Errorf("failed to ack message: %w", err)
		}

		replayed++
	}

	return replayed, nil
}

// ReplayAll replays all messages from a dead letter queue.
func (c *Client) ReplayAll(ctx context.Context, queue string) (int, error) {
	dlqName := DLQName(queue)

	// Get queue info to find message count
	q, err := c.channel.QueueInspect(dlqName)
	if err != nil {
		return 0, fmt.Errorf("failed to inspect DLQ: %w", err)
	}

	if q.Messages == 0 {
		return 0, nil
	}

	return c.ReplayDLQ(ctx, queue, q.Messages)
}

// PurgeDLQ removes all messages from a dead letter queue.
func (c *Client) PurgeDLQ(queue string) (int, error) {
	dlqName := DLQName(queue)

	count, err := c.channel.QueuePurge(dlqName, false)
	if err != nil {
		return 0, fmt.Errorf("failed to purge DLQ: %w", err)
	}

	return count, nil
}

// InspectDLQ returns information about a dead letter queue.
func (c *Client) InspectDLQ(queue string) (*DLQInfo, error) {
	dlqName := DLQName(queue)

	q, err := c.channel.QueueInspect(dlqName)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect DLQ: %w", err)
	}

	return &DLQInfo{
		Name:          dlqName,
		MessageCount:  q.Messages,
		ConsumerCount: q.Consumers,
	}, nil
}

// DLQInfo contains information about a dead letter queue.
type DLQInfo struct {
	Name          string `json:"name"`
	MessageCount  int    `json:"message_count"`
	ConsumerCount int    `json:"consumer_count"`
}

// GetDLQMessage retrieves a single message from the DLQ without removing it.
func (c *Client) GetDLQMessage(queue string) (*Message, error) {
	dlqName := DLQName(queue)

	delivery, ok, err := c.channel.Get(dlqName, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get message from DLQ: %w", err)
	}
	if !ok {
		return nil, nil
	}

	// Requeue the message (we're just peeking)
	if err := delivery.Reject(true); err != nil {
		return nil, fmt.Errorf("failed to requeue message: %w", err)
	}

	var msg Message
	if err := json.Unmarshal(delivery.Body, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return &msg, nil
}

// SubscribeDLQ registers a handler for dead letter messages.
// This allows custom processing of failed messages.
func (c *Client) SubscribeDLQ(queue string, handler HandlerFunc, opts ...SubscribeOption) error {
	dlqTopic := queue + DLQSuffix
	dlqQueueName := DLQName(queue)

	// Override queue name option
	opts = append([]SubscribeOption{WithQueueName(dlqQueueName)}, opts...)

	return c.Subscribe(dlqTopic, handler, opts...)
}

// declareDLQForQueue ensures a DLQ exists for a given queue.
// Called automatically when setting up consumers.
func (c *Client) declareDLQForQueue(queueName string, topic string) error {
	dlqName := DLQName(queueName)
	dlqTopic := topic + DLQSuffix

	// Declare DLQ
	_, err := c.channel.QueueDeclare(
		dlqName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		amqp.Table{
			"x-queue-type": "classic",
		},
	)
	if err != nil {
		return err
	}

	// Bind to exchange with DLQ topic pattern
	return c.channel.QueueBind(
		dlqName,
		dlqTopic,
		c.config.Exchange,
		false,
		nil,
	)
}
