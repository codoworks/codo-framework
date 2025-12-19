// Package redis provides a Redis client for the framework.
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/codoworks/codo-framework/core/clients"
	"github.com/redis/go-redis/v9"
)

const (
	// ClientName is the name of the Redis client.
	ClientName = "redis"
)

// Config holds Redis configuration.
type Config struct {
	Host         string        `json:"host" yaml:"host"`
	Port         int           `json:"port" yaml:"port"`
	Password     string        `json:"password" yaml:"password"`
	DB           int           `json:"db" yaml:"db"`
	PoolSize     int           `json:"pool_size" yaml:"pool_size"`
	MinIdleConns int           `json:"min_idle_conns" yaml:"min_idle_conns"`
	DialTimeout  time.Duration `json:"dial_timeout" yaml:"dial_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`
}

// DefaultConfig returns default Redis configuration.
func DefaultConfig() *Config {
	return &Config{
		Host:         "localhost",
		Port:         6379,
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 2,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}

// RedisClient is the interface for Redis operations.
type RedisClient interface {
	clients.Client

	// GetClient returns the underlying Redis client.
	GetClient() *redis.Client

	// Ping checks the connection to Redis.
	Ping(ctx context.Context) error

	// Set sets a key-value pair with expiration.
	Set(ctx context.Context, key string, value any, expiration time.Duration) error

	// Get retrieves a value by key.
	Get(ctx context.Context, key string) (string, error)

	// Del deletes keys.
	Del(ctx context.Context, keys ...string) error

	// Exists checks if keys exist.
	Exists(ctx context.Context, keys ...string) (int64, error)

	// Expire sets expiration on a key.
	Expire(ctx context.Context, key string, expiration time.Duration) error

	// TTL returns the time to live for a key.
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Keys returns keys matching the pattern.
	Keys(ctx context.Context, pattern string) ([]string, error)

	// Incr increments a key.
	Incr(ctx context.Context, key string) (int64, error)

	// Decr decrements a key.
	Decr(ctx context.Context, key string) (int64, error)

	// HSet sets hash fields.
	HSet(ctx context.Context, key string, values ...any) error

	// HGet gets a hash field.
	HGet(ctx context.Context, key, field string) (string, error)

	// HGetAll gets all hash fields.
	HGetAll(ctx context.Context, key string) (map[string]string, error)

	// HDel deletes hash fields.
	HDel(ctx context.Context, key string, fields ...string) error

	// LPush prepends values to a list.
	LPush(ctx context.Context, key string, values ...any) error

	// RPush appends values to a list.
	RPush(ctx context.Context, key string, values ...any) error

	// LPop removes and returns the first element.
	LPop(ctx context.Context, key string) (string, error)

	// RPop removes and returns the last element.
	RPop(ctx context.Context, key string) (string, error)

	// LRange returns a range of elements from a list.
	LRange(ctx context.Context, key string, start, stop int64) ([]string, error)

	// SAdd adds members to a set.
	SAdd(ctx context.Context, key string, members ...any) error

	// SMembers returns all members of a set.
	SMembers(ctx context.Context, key string) ([]string, error)

	// SRem removes members from a set.
	SRem(ctx context.Context, key string, members ...any) error

	// SIsMember checks if a value is a member of a set.
	SIsMember(ctx context.Context, key string, member any) (bool, error)
}

// Client is the Redis client implementation.
type Client struct {
	clients.BaseClient
	client *redis.Client
	config *Config
}

// New creates a new Redis client.
func New() *Client {
	return &Client{
		BaseClient: clients.NewBaseClient(ClientName),
	}
}

// Name returns the client name.
func (c *Client) Name() string {
	return ClientName
}

// Initialize sets up the Redis client with configuration.
func (c *Client) Initialize(cfg any) error {
	config := DefaultConfig()

	if cfg != nil {
		switch v := cfg.(type) {
		case *Config:
			config = v
		case Config:
			config = &v
		case map[string]any:
			if host, ok := v["host"].(string); ok {
				config.Host = host
			}
			if port, ok := v["port"].(int); ok {
				config.Port = port
			}
			if password, ok := v["password"].(string); ok {
				config.Password = password
			}
			if db, ok := v["db"].(int); ok {
				config.DB = db
			}
		default:
			return fmt.Errorf("invalid config type: %T", cfg)
		}
	}

	c.config = config

	c.client = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})

	return c.BaseClient.Initialize(cfg)
}

// Health checks if the Redis connection is healthy.
func (c *Client) Health() error {
	if c.client == nil {
		return fmt.Errorf("redis client not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.Ping(ctx)
}

// Shutdown closes the Redis connection.
func (c *Client) Shutdown() error {
	if c.client != nil {
		if err := c.client.Close(); err != nil {
			return err
		}
	}
	return c.BaseClient.Shutdown()
}

// GetClient returns the underlying Redis client.
func (c *Client) GetClient() *redis.Client {
	return c.client
}

// Ping checks the connection to Redis.
func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Set sets a key-value pair with expiration.
func (c *Client) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

// Get retrieves a value by key.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Del deletes keys.
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

// Exists checks if keys exist.
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.client.Exists(ctx, keys...).Result()
}

// Expire sets expiration on a key.
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.client.Expire(ctx, key, expiration).Err()
}

// TTL returns the time to live for a key.
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}

// Keys returns keys matching the pattern.
func (c *Client) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.client.Keys(ctx, pattern).Result()
}

// Incr increments a key.
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

// Decr decrements a key.
func (c *Client) Decr(ctx context.Context, key string) (int64, error) {
	return c.client.Decr(ctx, key).Result()
}

// HSet sets hash fields.
func (c *Client) HSet(ctx context.Context, key string, values ...any) error {
	return c.client.HSet(ctx, key, values...).Err()
}

// HGet gets a hash field.
func (c *Client) HGet(ctx context.Context, key, field string) (string, error) {
	return c.client.HGet(ctx, key, field).Result()
}

// HGetAll gets all hash fields.
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.client.HGetAll(ctx, key).Result()
}

// HDel deletes hash fields.
func (c *Client) HDel(ctx context.Context, key string, fields ...string) error {
	return c.client.HDel(ctx, key, fields...).Err()
}

// LPush prepends values to a list.
func (c *Client) LPush(ctx context.Context, key string, values ...any) error {
	return c.client.LPush(ctx, key, values...).Err()
}

// RPush appends values to a list.
func (c *Client) RPush(ctx context.Context, key string, values ...any) error {
	return c.client.RPush(ctx, key, values...).Err()
}

// LPop removes and returns the first element.
func (c *Client) LPop(ctx context.Context, key string) (string, error) {
	return c.client.LPop(ctx, key).Result()
}

// RPop removes and returns the last element.
func (c *Client) RPop(ctx context.Context, key string) (string, error) {
	return c.client.RPop(ctx, key).Result()
}

// LRange returns a range of elements from a list.
func (c *Client) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.client.LRange(ctx, key, start, stop).Result()
}

// SAdd adds members to a set.
func (c *Client) SAdd(ctx context.Context, key string, members ...any) error {
	return c.client.SAdd(ctx, key, members...).Err()
}

// SMembers returns all members of a set.
func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.client.SMembers(ctx, key).Result()
}

// SRem removes members from a set.
func (c *Client) SRem(ctx context.Context, key string, members ...any) error {
	return c.client.SRem(ctx, key, members...).Err()
}

// SIsMember checks if a value is a member of a set.
func (c *Client) SIsMember(ctx context.Context, key string, member any) (bool, error) {
	return c.client.SIsMember(ctx, key, member).Result()
}
