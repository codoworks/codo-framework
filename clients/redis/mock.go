package redis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/codoworks/codo-framework/core/clients"
	"github.com/redis/go-redis/v9"
)

// MockClient is a mock Redis client for testing.
type MockClient struct {
	clients.BaseClient
	mu      sync.RWMutex
	data    map[string]mockValue
	hashes  map[string]map[string]string
	lists   map[string][]string
	sets    map[string]map[string]struct{}
	healthy bool
	closed  bool
}

type mockValue struct {
	value      string
	expiration time.Time
}

// NewMock creates a new mock Redis client.
func NewMock() *MockClient {
	return &MockClient{
		BaseClient: clients.NewBaseClient(ClientName),
		data:       make(map[string]mockValue),
		hashes:     make(map[string]map[string]string),
		lists:      make(map[string][]string),
		sets:       make(map[string]map[string]struct{}),
		healthy:    true,
	}
}

// Name returns the client name.
func (m *MockClient) Name() string {
	return ClientName
}

// Initialize initializes the mock client.
func (m *MockClient) Initialize(cfg any) error {
	return m.BaseClient.Initialize(cfg)
}

// Health checks if the mock client is healthy.
func (m *MockClient) Health() error {
	if !m.healthy {
		return fmt.Errorf("mock redis is unhealthy")
	}
	if m.closed {
		return fmt.Errorf("mock redis is closed")
	}
	return nil
}

// Shutdown closes the mock client.
func (m *MockClient) Shutdown() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return m.BaseClient.Shutdown()
}

// GetClient returns nil for mock.
func (m *MockClient) GetClient() *redis.Client {
	return nil
}

// SetHealthy sets the health status.
func (m *MockClient) SetHealthy(healthy bool) {
	m.healthy = healthy
}

// Reset clears all data.
func (m *MockClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[string]mockValue)
	m.hashes = make(map[string]map[string]string)
	m.lists = make(map[string][]string)
	m.sets = make(map[string]map[string]struct{})
	m.closed = false
}

// Ping checks the connection.
func (m *MockClient) Ping(ctx context.Context) error {
	if m.closed {
		return fmt.Errorf("connection closed")
	}
	return nil
}

// Set sets a key-value pair.
func (m *MockClient) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var exp time.Time
	if expiration > 0 {
		exp = time.Now().Add(expiration)
	}

	m.data[key] = mockValue{
		value:      fmt.Sprintf("%v", value),
		expiration: exp,
	}
	return nil
}

// Get retrieves a value.
func (m *MockClient) Get(ctx context.Context, key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	mv, exists := m.data[key]
	if !exists {
		return "", redis.Nil
	}

	if !mv.expiration.IsZero() && time.Now().After(mv.expiration) {
		return "", redis.Nil
	}

	return mv.value, nil
}

// Del deletes keys.
func (m *MockClient) Del(ctx context.Context, keys ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, key := range keys {
		delete(m.data, key)
		delete(m.hashes, key)
		delete(m.lists, key)
		delete(m.sets, key)
	}
	return nil
}

// Exists checks if keys exist.
func (m *MockClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var count int64
	for _, key := range keys {
		if _, exists := m.data[key]; exists {
			count++
		} else if _, exists := m.hashes[key]; exists {
			count++
		} else if _, exists := m.lists[key]; exists {
			count++
		} else if _, exists := m.sets[key]; exists {
			count++
		}
	}
	return count, nil
}

// Expire sets expiration on a key.
func (m *MockClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if mv, exists := m.data[key]; exists {
		mv.expiration = time.Now().Add(expiration)
		m.data[key] = mv
	}
	return nil
}

// TTL returns the time to live.
func (m *MockClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	mv, exists := m.data[key]
	if !exists {
		return -2 * time.Second, nil // Key doesn't exist
	}

	if mv.expiration.IsZero() {
		return -1 * time.Second, nil // No expiration
	}

	ttl := time.Until(mv.expiration)
	if ttl < 0 {
		return -2 * time.Second, nil // Expired
	}
	return ttl, nil
}

// Keys returns keys matching pattern.
func (m *MockClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var keys []string
	for key := range m.data {
		keys = append(keys, key)
	}
	return keys, nil
}

// Incr increments a key.
func (m *MockClient) Incr(ctx context.Context, key string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	mv, exists := m.data[key]
	var val int64 = 0
	if exists {
		fmt.Sscanf(mv.value, "%d", &val)
	}
	val++
	m.data[key] = mockValue{value: fmt.Sprintf("%d", val)}
	return val, nil
}

// Decr decrements a key.
func (m *MockClient) Decr(ctx context.Context, key string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	mv, exists := m.data[key]
	var val int64 = 0
	if exists {
		fmt.Sscanf(mv.value, "%d", &val)
	}
	val--
	m.data[key] = mockValue{value: fmt.Sprintf("%d", val)}
	return val, nil
}

// HSet sets hash fields.
func (m *MockClient) HSet(ctx context.Context, key string, values ...any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.hashes[key] == nil {
		m.hashes[key] = make(map[string]string)
	}

	for i := 0; i < len(values)-1; i += 2 {
		field := fmt.Sprintf("%v", values[i])
		value := fmt.Sprintf("%v", values[i+1])
		m.hashes[key][field] = value
	}
	return nil
}

// HGet gets a hash field.
func (m *MockClient) HGet(ctx context.Context, key, field string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if hash, exists := m.hashes[key]; exists {
		if val, ok := hash[field]; ok {
			return val, nil
		}
	}
	return "", redis.Nil
}

// HGetAll gets all hash fields.
func (m *MockClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if hash, exists := m.hashes[key]; exists {
		result := make(map[string]string)
		for k, v := range hash {
			result[k] = v
		}
		return result, nil
	}
	return make(map[string]string), nil
}

// HDel deletes hash fields.
func (m *MockClient) HDel(ctx context.Context, key string, fields ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if hash, exists := m.hashes[key]; exists {
		for _, field := range fields {
			delete(hash, field)
		}
	}
	return nil
}

// LPush prepends values to a list.
// Values are inserted one after another to the head of the list, from the leftmost to the rightmost.
func (m *MockClient) LPush(ctx context.Context, key string, values ...any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := 0; i < len(values); i++ {
		m.lists[key] = append([]string{fmt.Sprintf("%v", values[i])}, m.lists[key]...)
	}
	return nil
}

// RPush appends values to a list.
func (m *MockClient) RPush(ctx context.Context, key string, values ...any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, v := range values {
		m.lists[key] = append(m.lists[key], fmt.Sprintf("%v", v))
	}
	return nil
}

// LPop removes and returns the first element.
func (m *MockClient) LPop(ctx context.Context, key string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if list, exists := m.lists[key]; exists && len(list) > 0 {
		val := list[0]
		m.lists[key] = list[1:]
		return val, nil
	}
	return "", redis.Nil
}

// RPop removes and returns the last element.
func (m *MockClient) RPop(ctx context.Context, key string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if list, exists := m.lists[key]; exists && len(list) > 0 {
		val := list[len(list)-1]
		m.lists[key] = list[:len(list)-1]
		return val, nil
	}
	return "", redis.Nil
}

// LRange returns a range of elements.
func (m *MockClient) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list, exists := m.lists[key]
	if !exists {
		return []string{}, nil
	}

	length := int64(len(list))
	if start < 0 {
		start = length + start
	}
	if stop < 0 {
		stop = length + stop
	}

	if start < 0 {
		start = 0
	}
	if stop >= length {
		stop = length - 1
	}

	if start > stop || start >= length {
		return []string{}, nil
	}

	return list[start : stop+1], nil
}

// SAdd adds members to a set.
func (m *MockClient) SAdd(ctx context.Context, key string, members ...any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.sets[key] == nil {
		m.sets[key] = make(map[string]struct{})
	}

	for _, member := range members {
		m.sets[key][fmt.Sprintf("%v", member)] = struct{}{}
	}
	return nil
}

// SMembers returns all members of a set.
func (m *MockClient) SMembers(ctx context.Context, key string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var members []string
	if set, exists := m.sets[key]; exists {
		for member := range set {
			members = append(members, member)
		}
	}
	return members, nil
}

// SRem removes members from a set.
func (m *MockClient) SRem(ctx context.Context, key string, members ...any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if set, exists := m.sets[key]; exists {
		for _, member := range members {
			delete(set, fmt.Sprintf("%v", member))
		}
	}
	return nil
}

// SIsMember checks if a value is a member of a set.
func (m *MockClient) SIsMember(ctx context.Context, key string, member any) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if set, exists := m.sets[key]; exists {
		_, isMember := set[fmt.Sprintf("%v", member)]
		return isMember, nil
	}
	return false, nil
}

// Ensure MockClient implements RedisClient interface.
var _ RedisClient = (*MockClient)(nil)
