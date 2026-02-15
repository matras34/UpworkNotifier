package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
}

func NewCache(addr, password string, db int) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Cache{client: client}, nil
}

func (c *Cache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *Cache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, key).Result()
	return n > 0, err
}

func (c *Cache) Increment(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

func (c *Cache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.client.Expire(ctx, key, expiration).Err()
}

func (c *Cache) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.client.SetNX(ctx, key, value, expiration).Result()
}

func (c *Cache) Close() error {
	return c.client.Close()
}

// Rate limiting helpers
func (c *Cache) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	count, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if count == 1 {
		c.client.Expire(ctx, key, window)
	}

	return count <= int64(limit), nil
}

// Session helpers
func (c *Cache) StoreSession(ctx context.Context, sessionID string, data map[string]interface{}, expiration time.Duration) error {
	return c.client.HSet(ctx, "session:"+sessionID, data).Err()
}

func (c *Cache) GetSession(ctx context.Context, sessionID string) (map[string]string, error) {
	return c.client.HGetAll(ctx, "session:"+sessionID).Result()
}

func (c *Cache) DeleteSession(ctx context.Context, sessionID string) error {
	return c.client.Del(ctx, "session:"+sessionID).Err()
}
