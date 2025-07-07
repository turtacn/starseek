package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/turtacn/starseek/internal/common/errorx"
)

// Cache defines the interface for a generic caching service.
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Incr(ctx context.Context, key string) (int64, error)
	// Additional methods like SAdd, SMembers, ZAdd, ZRange could be added for more complex caching needs
}

// RedisCache implements the Cache interface using Redis.
type RedisCache struct {
	client *redis.Client
}

// Config holds the configuration for the RedisCache.
type Config struct {
	Addr     string `mapstructure:"addr"`     // Redis server address (e.g., "localhost:6379")
	Password string `mapstructure:"password"` // Redis password
	DB       int    `mapstructure:"db"`       // Redis DB number
}

// NewRedisCache creates a new RedisCache instance.
func NewRedisCache(cfg Config) (Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{client: client}, nil
}

// Get retrieves a value from Redis by key.
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", errorx.ErrNotFound.Wrap(fmt.Errorf("cache key %s not found", key))
		}
		return "", errorx.NewError(errorx.ErrCache.Code, fmt.Sprintf("failed to get from cache for key %s", key), err)
	}
	return val, nil
}

// Set stores a key-value pair in Redis with a specified TTL.
func (r *RedisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return errorx.NewError(errorx.ErrCache.Code, fmt.Sprintf("failed to set cache for key %s", key), err)
	}
	return nil
}

// Delete removes a key from Redis.
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return errorx.NewError(errorx.ErrCache.Code, fmt.Sprintf("failed to delete from cache for key %s", key), err)
	}
	return nil
}

// Exists checks if a key exists in Redis.
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, errorx.NewError(errorx.ErrCache.Code, fmt.Sprintf("failed to check existence for key %s", key), err)
	}
	return count > 0, nil
}

// Incr increments the integer value of a key by one. If the key does not exist, it is set to 0 before performing the operation.
func (r *RedisCache) Incr(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, errorx.NewError(errorx.ErrCache.Code, fmt.Sprintf("failed to increment key %s", key), err)
	}
	return val, nil
}

//Personal.AI order the ending
