package redis

import (
	"context"
	"testing"
	"time"

	"github.com/codoworks/codo-framework/core/clients"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestClient_ImplementsClient(t *testing.T) {
	var _ clients.Client = (*Client)(nil)
}

func TestClient_ImplementsRedisClient(t *testing.T) {
	var _ RedisClient = (*Client)(nil)
}

func TestNew(t *testing.T) {
	c := New()

	assert.NotNil(t, c)
	assert.Equal(t, ClientName, c.Name())
}

func TestClient_Name(t *testing.T) {
	c := New()
	assert.Equal(t, "redis", c.Name())
}

func TestClient_Initialize(t *testing.T) {
	t.Run("with nil config uses defaults", func(t *testing.T) {
		c := New()

		err := c.Initialize(nil)

		assert.NoError(t, err)
		assert.True(t, c.IsInitialized())
		assert.NotNil(t, c.client)
	})

	t.Run("with Config pointer", func(t *testing.T) {
		c := New()
		cfg := &Config{
			Host: "redis.example.com",
			Port: 6380,
		}

		err := c.Initialize(cfg)

		assert.NoError(t, err)
		assert.Equal(t, cfg, c.config)
	})

	t.Run("with Config value", func(t *testing.T) {
		c := New()
		cfg := Config{
			Host: "redis.example.com",
			Port: 6380,
		}

		err := c.Initialize(cfg)

		assert.NoError(t, err)
	})

	t.Run("with map config", func(t *testing.T) {
		c := New()
		cfg := map[string]any{
			"host":     "redis.local",
			"port":     6381,
			"password": "secret",
			"db":       1,
		}

		err := c.Initialize(cfg)

		assert.NoError(t, err)
		assert.Equal(t, "redis.local", c.config.Host)
		assert.Equal(t, 6381, c.config.Port)
		assert.Equal(t, "secret", c.config.Password)
		assert.Equal(t, 1, c.config.DB)
	})

	t.Run("with invalid config type", func(t *testing.T) {
		c := New()

		err := c.Initialize("invalid")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid config type")
	})
}

func TestClient_GetClient(t *testing.T) {
	c := New()
	c.Initialize(nil)

	client := c.GetClient()

	assert.NotNil(t, client)
	assert.IsType(t, &redis.Client{}, client)
}

func TestClient_Health_NotInitialized(t *testing.T) {
	c := &Client{
		BaseClient: clients.NewBaseClient(ClientName),
		client:     nil,
	}

	err := c.Health()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestClient_Shutdown(t *testing.T) {
	c := New()
	c.Initialize(nil)

	err := c.Shutdown()

	assert.NoError(t, err)
	assert.False(t, c.IsInitialized())
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 6379, cfg.Port)
	assert.Equal(t, "", cfg.Password)
	assert.Equal(t, 0, cfg.DB)
	assert.Equal(t, 10, cfg.PoolSize)
	assert.Equal(t, 2, cfg.MinIdleConns)
	assert.Equal(t, 5*time.Second, cfg.DialTimeout)
	assert.Equal(t, 3*time.Second, cfg.ReadTimeout)
	assert.Equal(t, 3*time.Second, cfg.WriteTimeout)
}

func TestClientName(t *testing.T) {
	assert.Equal(t, "redis", ClientName)
}

// Tests using MockClient

func TestMockClient_ImplementsRedisClient(t *testing.T) {
	var _ RedisClient = (*MockClient)(nil)
}

func TestMockClient_New(t *testing.T) {
	m := NewMock()

	assert.NotNil(t, m)
	assert.Equal(t, ClientName, m.Name())
}

func TestMockClient_Initialize(t *testing.T) {
	m := NewMock()

	err := m.Initialize(nil)

	assert.NoError(t, err)
	assert.True(t, m.IsInitialized())
}

func TestMockClient_Health(t *testing.T) {
	t.Run("healthy", func(t *testing.T) {
		m := NewMock()

		err := m.Health()

		assert.NoError(t, err)
	})

	t.Run("unhealthy", func(t *testing.T) {
		m := NewMock()
		m.SetHealthy(false)

		err := m.Health()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unhealthy")
	})

	t.Run("closed", func(t *testing.T) {
		m := NewMock()
		m.Shutdown()

		err := m.Health()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "closed")
	})
}

func TestMockClient_Shutdown(t *testing.T) {
	m := NewMock()
	m.Initialize(nil)

	err := m.Shutdown()

	assert.NoError(t, err)
	assert.False(t, m.IsInitialized())
}

func TestMockClient_GetClient(t *testing.T) {
	m := NewMock()

	client := m.GetClient()

	assert.Nil(t, client)
}

func TestMockClient_Reset(t *testing.T) {
	m := NewMock()
	ctx := context.Background()

	m.Set(ctx, "key", "value", 0)
	m.Shutdown()

	m.Reset()

	_, err := m.Get(ctx, "key")
	assert.Error(t, err) // Should be redis.Nil

	pingErr := m.Ping(ctx)
	assert.NoError(t, pingErr)
}

func TestMockClient_Ping(t *testing.T) {
	t.Run("open connection", func(t *testing.T) {
		m := NewMock()
		ctx := context.Background()

		err := m.Ping(ctx)

		assert.NoError(t, err)
	})

	t.Run("closed connection", func(t *testing.T) {
		m := NewMock()
		m.Shutdown()
		ctx := context.Background()

		err := m.Ping(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "closed")
	})
}

func TestMockClient_SetGet(t *testing.T) {
	m := NewMock()
	ctx := context.Background()

	t.Run("set and get", func(t *testing.T) {
		err := m.Set(ctx, "key1", "value1", 0)
		assert.NoError(t, err)

		val, err := m.Get(ctx, "key1")
		assert.NoError(t, err)
		assert.Equal(t, "value1", val)
	})

	t.Run("get nonexistent key", func(t *testing.T) {
		_, err := m.Get(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Equal(t, redis.Nil, err)
	})

	t.Run("set with expiration", func(t *testing.T) {
		err := m.Set(ctx, "expiring", "value", 50*time.Millisecond)
		assert.NoError(t, err)

		val, err := m.Get(ctx, "expiring")
		assert.NoError(t, err)
		assert.Equal(t, "value", val)

		time.Sleep(100 * time.Millisecond)

		_, err = m.Get(ctx, "expiring")
		assert.Error(t, err)
		assert.Equal(t, redis.Nil, err)
	})
}

func TestMockClient_Del(t *testing.T) {
	m := NewMock()
	ctx := context.Background()

	m.Set(ctx, "key1", "value1", 0)
	m.Set(ctx, "key2", "value2", 0)

	err := m.Del(ctx, "key1", "key2")
	assert.NoError(t, err)

	_, err = m.Get(ctx, "key1")
	assert.Equal(t, redis.Nil, err)

	_, err = m.Get(ctx, "key2")
	assert.Equal(t, redis.Nil, err)
}

func TestMockClient_Exists(t *testing.T) {
	m := NewMock()
	ctx := context.Background()

	m.Set(ctx, "exists1", "value", 0)
	m.Set(ctx, "exists2", "value", 0)

	count, err := m.Exists(ctx, "exists1", "exists2", "notexists")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestMockClient_Expire(t *testing.T) {
	m := NewMock()
	ctx := context.Background()

	m.Set(ctx, "key", "value", 0)

	err := m.Expire(ctx, "key", 50*time.Millisecond)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	_, err = m.Get(ctx, "key")
	assert.Equal(t, redis.Nil, err)
}

func TestMockClient_TTL(t *testing.T) {
	m := NewMock()
	ctx := context.Background()

	t.Run("key with expiration", func(t *testing.T) {
		m.Set(ctx, "key", "value", time.Hour)

		ttl, err := m.TTL(ctx, "key")
		assert.NoError(t, err)
		assert.True(t, ttl > 0)
	})

	t.Run("key without expiration", func(t *testing.T) {
		m.Set(ctx, "persistent", "value", 0)

		ttl, err := m.TTL(ctx, "persistent")
		assert.NoError(t, err)
		assert.Equal(t, -1*time.Second, ttl)
	})

	t.Run("nonexistent key", func(t *testing.T) {
		ttl, err := m.TTL(ctx, "nonexistent")
		assert.NoError(t, err)
		assert.Equal(t, -2*time.Second, ttl)
	})

	t.Run("expired key", func(t *testing.T) {
		// Set a key with a very short expiration
		m.Set(ctx, "expiring", "value", time.Nanosecond)
		// Wait for it to expire
		time.Sleep(time.Millisecond)

		ttl, err := m.TTL(ctx, "expiring")
		assert.NoError(t, err)
		assert.Equal(t, -2*time.Second, ttl) // Expired returns -2
	})
}

func TestMockClient_Keys(t *testing.T) {
	m := NewMock()
	ctx := context.Background()

	m.Set(ctx, "key1", "value1", 0)
	m.Set(ctx, "key2", "value2", 0)

	keys, err := m.Keys(ctx, "*")
	assert.NoError(t, err)
	assert.Len(t, keys, 2)
}

func TestMockClient_IncrDecr(t *testing.T) {
	m := NewMock()
	ctx := context.Background()

	t.Run("incr new key", func(t *testing.T) {
		val, err := m.Incr(ctx, "counter")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), val)
	})

	t.Run("incr existing key", func(t *testing.T) {
		val, err := m.Incr(ctx, "counter")
		assert.NoError(t, err)
		assert.Equal(t, int64(2), val)
	})

	t.Run("decr", func(t *testing.T) {
		val, err := m.Decr(ctx, "counter")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), val)
	})

	t.Run("decr new key", func(t *testing.T) {
		val, err := m.Decr(ctx, "newcounter")
		assert.NoError(t, err)
		assert.Equal(t, int64(-1), val)
	})
}

func TestMockClient_Hash(t *testing.T) {
	m := NewMock()
	ctx := context.Background()

	t.Run("hset and hget", func(t *testing.T) {
		err := m.HSet(ctx, "hash", "field1", "value1")
		assert.NoError(t, err)

		val, err := m.HGet(ctx, "hash", "field1")
		assert.NoError(t, err)
		assert.Equal(t, "value1", val)
	})

	t.Run("hget nonexistent", func(t *testing.T) {
		_, err := m.HGet(ctx, "hash", "nonexistent")
		assert.Equal(t, redis.Nil, err)
	})

	t.Run("hgetall", func(t *testing.T) {
		m.HSet(ctx, "hash2", "f1", "v1", "f2", "v2")

		all, err := m.HGetAll(ctx, "hash2")
		assert.NoError(t, err)
		assert.Equal(t, "v1", all["f1"])
		assert.Equal(t, "v2", all["f2"])
	})

	t.Run("hgetall empty", func(t *testing.T) {
		all, err := m.HGetAll(ctx, "emptyhash")
		assert.NoError(t, err)
		assert.Empty(t, all)
	})

	t.Run("hdel", func(t *testing.T) {
		m.HSet(ctx, "hash3", "field", "value")

		err := m.HDel(ctx, "hash3", "field")
		assert.NoError(t, err)

		_, err = m.HGet(ctx, "hash3", "field")
		assert.Equal(t, redis.Nil, err)
	})
}

func TestMockClient_List(t *testing.T) {
	m := NewMock()
	ctx := context.Background()

	t.Run("lpush and lpop", func(t *testing.T) {
		m.LPush(ctx, "list1", "c", "b", "a")

		val, err := m.LPop(ctx, "list1")
		assert.NoError(t, err)
		assert.Equal(t, "a", val)
	})

	t.Run("rpush and rpop", func(t *testing.T) {
		m.RPush(ctx, "list2", "a", "b", "c")

		val, err := m.RPop(ctx, "list2")
		assert.NoError(t, err)
		assert.Equal(t, "c", val)
	})

	t.Run("lpop empty", func(t *testing.T) {
		_, err := m.LPop(ctx, "emptylist")
		assert.Equal(t, redis.Nil, err)
	})

	t.Run("rpop empty", func(t *testing.T) {
		_, err := m.RPop(ctx, "emptylist")
		assert.Equal(t, redis.Nil, err)
	})

	t.Run("lrange", func(t *testing.T) {
		m.Reset()
		m.RPush(ctx, "list3", "a", "b", "c", "d", "e")

		vals, err := m.LRange(ctx, "list3", 0, 2)
		assert.NoError(t, err)
		assert.Equal(t, []string{"a", "b", "c"}, vals)

		vals, err = m.LRange(ctx, "list3", -2, -1)
		assert.NoError(t, err)
		assert.Equal(t, []string{"d", "e"}, vals)
	})

	t.Run("lrange empty", func(t *testing.T) {
		vals, err := m.LRange(ctx, "nonexistent", 0, -1)
		assert.NoError(t, err)
		assert.Empty(t, vals)
	})

	t.Run("lrange out of bounds", func(t *testing.T) {
		m.Reset()
		m.RPush(ctx, "list4", "a")

		vals, err := m.LRange(ctx, "list4", 10, 20)
		assert.NoError(t, err)
		assert.Empty(t, vals)
	})
}

func TestMockClient_Set(t *testing.T) {
	m := NewMock()
	ctx := context.Background()

	t.Run("sadd and smembers", func(t *testing.T) {
		err := m.SAdd(ctx, "set1", "a", "b", "c")
		assert.NoError(t, err)

		members, err := m.SMembers(ctx, "set1")
		assert.NoError(t, err)
		assert.Len(t, members, 3)
	})

	t.Run("smembers empty", func(t *testing.T) {
		members, err := m.SMembers(ctx, "emptyset")
		assert.NoError(t, err)
		assert.Empty(t, members)
	})

	t.Run("srem", func(t *testing.T) {
		m.SAdd(ctx, "set2", "x", "y", "z")

		err := m.SRem(ctx, "set2", "y")
		assert.NoError(t, err)

		members, _ := m.SMembers(ctx, "set2")
		assert.Len(t, members, 2)
	})

	t.Run("sismember", func(t *testing.T) {
		m.SAdd(ctx, "set3", "member")

		isMember, err := m.SIsMember(ctx, "set3", "member")
		assert.NoError(t, err)
		assert.True(t, isMember)

		isMember, err = m.SIsMember(ctx, "set3", "nonmember")
		assert.NoError(t, err)
		assert.False(t, isMember)
	})

	t.Run("sismember empty set", func(t *testing.T) {
		isMember, err := m.SIsMember(ctx, "emptyset", "member")
		assert.NoError(t, err)
		assert.False(t, isMember)
	})
}

func TestMockClient_ExistsWithAllTypes(t *testing.T) {
	m := NewMock()
	ctx := context.Background()

	m.Set(ctx, "string", "value", 0)
	m.HSet(ctx, "hash", "field", "value")
	m.LPush(ctx, "list", "value")
	m.SAdd(ctx, "set", "value")

	count, err := m.Exists(ctx, "string", "hash", "list", "set", "nonexistent")
	assert.NoError(t, err)
	assert.Equal(t, int64(4), count)
}

func TestMockClient_DelAllTypes(t *testing.T) {
	m := NewMock()
	ctx := context.Background()

	m.Set(ctx, "string", "value", 0)
	m.HSet(ctx, "hash", "field", "value")
	m.LPush(ctx, "list", "value")
	m.SAdd(ctx, "set", "value")

	err := m.Del(ctx, "string", "hash", "list", "set")
	assert.NoError(t, err)

	count, _ := m.Exists(ctx, "string", "hash", "list", "set")
	assert.Equal(t, int64(0), count)
}
