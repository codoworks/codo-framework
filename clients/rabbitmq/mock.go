package rabbitmq

import (
	"context"
	"sync"

	"github.com/codoworks/codo-framework/core/clients"
	amqp "github.com/rabbitmq/amqp091-go"
)

// MockClient is a mock implementation of RabbitMQClient for testing.
type MockClient struct {
	clients.BaseClient

	config       *Config
	published    []*Message
	handlers     map[string]HandlerFunc
	dlqMessages  map[string][]*Message
	mu           sync.RWMutex

	// Hooks for testing
	OnPublish    func(ctx context.Context, msg *Message) error
	OnSubscribe  func(topic string, handler HandlerFunc) error
}

// NewMock creates a new mock RabbitMQ client.
func NewMock() *MockClient {
	return &MockClient{
		BaseClient:  clients.NewBaseClient(ClientName),
		config:      DefaultConfig(),
		published:   make([]*Message, 0),
		handlers:    make(map[string]HandlerFunc),
		dlqMessages: make(map[string][]*Message),
	}
}

// Name returns the client name.
func (m *MockClient) Name() string {
	return ClientName
}

// Initialize sets up the mock client.
func (m *MockClient) Initialize(cfg any) error {
	if c, ok := cfg.(*Config); ok {
		m.config = c
	}
	return m.BaseClient.Initialize(cfg)
}

// Health always returns nil (healthy).
func (m *MockClient) Health() error {
	return nil
}

// Shutdown cleans up the mock client.
func (m *MockClient) Shutdown() error {
	m.mu.Lock()
	m.published = nil
	m.handlers = make(map[string]HandlerFunc)
	m.mu.Unlock()
	return m.BaseClient.Shutdown()
}

// Start is a no-op for mock.
func (m *MockClient) Start() error {
	return nil
}

// Stop is a no-op for mock.
func (m *MockClient) Stop() error {
	return nil
}

// Publish records the message for later inspection.
func (m *MockClient) Publish(ctx context.Context, topic string, payload any, opts ...PublishOption) error {
	msg, err := NewMessage(topic, payload)
	if err != nil {
		return err
	}

	if m.OnPublish != nil {
		if err := m.OnPublish(ctx, msg); err != nil {
			return err
		}
	}

	m.mu.Lock()
	m.published = append(m.published, msg)
	m.mu.Unlock()

	return nil
}

// PublishMessage records the message for later inspection.
func (m *MockClient) PublishMessage(ctx context.Context, msg *Message, opts ...PublishOption) error {
	if m.OnPublish != nil {
		if err := m.OnPublish(ctx, msg); err != nil {
			return err
		}
	}

	m.mu.Lock()
	m.published = append(m.published, msg)
	m.mu.Unlock()

	return nil
}

// Subscribe records the handler for testing.
func (m *MockClient) Subscribe(topic string, handler HandlerFunc, opts ...SubscribeOption) error {
	if m.OnSubscribe != nil {
		if err := m.OnSubscribe(topic, handler); err != nil {
			return err
		}
	}

	m.mu.Lock()
	m.handlers[topic] = handler
	m.mu.Unlock()

	return nil
}

// GetChannel returns nil for mock.
func (m *MockClient) GetChannel() *amqp.Channel {
	return nil
}

// GetConnection returns nil for mock.
func (m *MockClient) GetConnection() *amqp.Connection {
	return nil
}

// --- Test Helpers ---

// Published returns all published messages.
func (m *MockClient) Published() []*Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Message, len(m.published))
	copy(result, m.published)
	return result
}

// PublishedCount returns the number of published messages.
func (m *MockClient) PublishedCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.published)
}

// PublishedByTopic returns messages published to a specific topic.
func (m *MockClient) PublishedByTopic(topic string) []*Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Message
	for _, msg := range m.published {
		if msg.Topic == topic {
			result = append(result, msg)
		}
	}
	return result
}

// LastPublished returns the most recently published message.
func (m *MockClient) LastPublished() *Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.published) == 0 {
		return nil
	}
	return m.published[len(m.published)-1]
}

// ClearPublished clears all recorded published messages.
func (m *MockClient) ClearPublished() {
	m.mu.Lock()
	m.published = make([]*Message, 0)
	m.mu.Unlock()
}

// HasHandler checks if a handler is registered for a topic.
func (m *MockClient) HasHandler(topic string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.handlers[topic]
	return ok
}

// GetHandler returns the registered handler for a topic.
func (m *MockClient) GetHandler(topic string) HandlerFunc {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.handlers[topic]
}

// SimulateMessage simulates receiving a message on a topic.
// Calls the registered handler with the message.
func (m *MockClient) SimulateMessage(ctx context.Context, topic string, payload any) error {
	m.mu.RLock()
	handler, ok := m.handlers[topic]
	m.mu.RUnlock()

	if !ok {
		return nil // No handler registered
	}

	msg, err := NewMessage(topic, payload)
	if err != nil {
		return err
	}

	return handler(ctx, msg)
}

// SimulateMessageWithMsg simulates receiving a pre-constructed message.
func (m *MockClient) SimulateMessageWithMsg(ctx context.Context, msg *Message) error {
	m.mu.RLock()
	handler, ok := m.handlers[msg.Topic]
	m.mu.RUnlock()

	if !ok {
		return nil
	}

	return handler(ctx, msg)
}

// AddToDLQ adds a message to the mock DLQ.
func (m *MockClient) AddToDLQ(queue string, msg *Message) {
	m.mu.Lock()
	m.dlqMessages[queue] = append(m.dlqMessages[queue], msg)
	m.mu.Unlock()
}

// GetDLQMessages returns messages in the mock DLQ.
func (m *MockClient) GetDLQMessages(queue string) []*Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Message, len(m.dlqMessages[queue]))
	copy(result, m.dlqMessages[queue])
	return result
}

// DLQCount returns the number of messages in a mock DLQ.
func (m *MockClient) DLQCount(queue string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.dlqMessages[queue])
}

// Ensure MockClient implements RabbitMQClient interface.
var _ RabbitMQClient = (*MockClient)(nil)
