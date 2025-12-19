package rabbitmq

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Message represents a RabbitMQ message with metadata.
type Message struct {
	ID            string            `json:"id"`
	Topic         string            `json:"topic"`                       // Routing key, e.g., "payments.order.created"
	Payload       json.RawMessage   `json:"payload"`                     // Raw JSON payload
	Headers       map[string]string `json:"headers,omitempty"`           // Custom headers
	Timestamp     time.Time         `json:"timestamp"`                   // When the message was created
	RetryCount    int               `json:"retry_count"`                 // Number of retry attempts
	CorrelationID string            `json:"correlation_id,omitempty"`    // For request/response patterns
	Error         string            `json:"error,omitempty"`             // Last error (for DLQ messages)
	OriginalTopic string            `json:"original_topic,omitempty"`    // Original topic (for DLQ messages)
}

// NewMessage creates a new message with the given topic and payload.
// Generates a UUID for the message ID and sets the timestamp to now.
func NewMessage(topic string, payload any) (*Message, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &Message{
		ID:        uuid.New().String(),
		Topic:     topic,
		Payload:   data,
		Headers:   make(map[string]string),
		Timestamp: time.Now().UTC(),
	}, nil
}

// Bind unmarshals the message payload into the provided value.
func (m *Message) Bind(v any) error {
	return json.Unmarshal(m.Payload, v)
}

// SetPayload marshals the given value and sets it as the message payload.
func (m *Message) SetPayload(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	m.Payload = data
	return nil
}

// SetHeader sets a header value.
func (m *Message) SetHeader(key, value string) {
	if m.Headers == nil {
		m.Headers = make(map[string]string)
	}
	m.Headers[key] = value
}

// GetHeader returns a header value.
func (m *Message) GetHeader(key string) string {
	if m.Headers == nil {
		return ""
	}
	return m.Headers[key]
}

// Clone creates a deep copy of the message.
func (m *Message) Clone() *Message {
	clone := &Message{
		ID:            m.ID,
		Topic:         m.Topic,
		Payload:       make(json.RawMessage, len(m.Payload)),
		Timestamp:     m.Timestamp,
		RetryCount:    m.RetryCount,
		CorrelationID: m.CorrelationID,
		Error:         m.Error,
		OriginalTopic: m.OriginalTopic,
	}
	copy(clone.Payload, m.Payload)

	if m.Headers != nil {
		clone.Headers = make(map[string]string, len(m.Headers))
		for k, v := range m.Headers {
			clone.Headers[k] = v
		}
	}

	return clone
}

// IsDLQ returns true if this message is from a dead letter queue.
func (m *Message) IsDLQ() bool {
	return m.OriginalTopic != ""
}
