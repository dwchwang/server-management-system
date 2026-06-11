package scheduler

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient defines the Redis operations needed by the scheduler.
// Using an interface allows mocking in tests without a real Redis instance.
type RedisClient interface {
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	ReleaseLock(ctx context.Context, key string, value string) error
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
}

// RealRedisClient wraps go-redis Client to implement RedisClient.
type RealRedisClient struct {
	client *redis.Client
}

// NewRealRedisClient creates a RealRedisClient from a go-redis client.
func NewRealRedisClient(client *redis.Client) *RealRedisClient {
	return &RealRedisClient{client: client}
}

func (r *RealRedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, value, expiration).Result()
}

func (r *RealRedisClient) ReleaseLock(ctx context.Context, key string, value string) error {
	const compareAndDelete = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
else
	return 0
end`
	return r.client.Eval(ctx, compareAndDelete, []string{key}, value).Err()
}

func (r *RealRedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RealRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// MockRedisClient is a test mock implementing RedisClient.
type MockRedisClient struct {
	SetNXFunc       func(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	ReleaseLockFunc func(ctx context.Context, key string, value string) error
	GetFunc         func(ctx context.Context, key string) (string, error)
	SetFunc         func(ctx context.Context, key string, value interface{}, expiration time.Duration) error
}

func (m *MockRedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	if m.SetNXFunc != nil {
		return m.SetNXFunc(ctx, key, value, expiration)
	}
	return false, nil
}

func (m *MockRedisClient) ReleaseLock(ctx context.Context, key string, value string) error {
	if m.ReleaseLockFunc != nil {
		return m.ReleaseLockFunc(ctx, key, value)
	}
	return nil
}

func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, key)
	}
	return "", nil
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if m.SetFunc != nil {
		return m.SetFunc(ctx, key, value, expiration)
	}
	return nil
}

// Verify interface satisfaction
var _ RedisClient = (*RealRedisClient)(nil)
var _ RedisClient = (*MockRedisClient)(nil)
