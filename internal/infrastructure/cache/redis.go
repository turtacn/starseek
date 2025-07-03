package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"                                                              // Redis 客户端库
	otelredis "go.opentelemetry.io/contrib/instrumentation/github.com/go-redis/redis/otelredis" // OpenTelemetry Redis Hook
	"go.opentelemetry.io/otel"                                                                  // OpenTelemetry Global API

	"github.com/turtacn/starseek/internal/common/config"         // 配置包
	ferrors "github.com/turtacn/starseek/internal/common/errors" // 自定义错误包
	"github.com/turtacn/starseek/internal/infrastructure/logger" // 日志包
)

// redisCache 实现了 Cache 接口，使用 Redis 作为底层存储。
type redisCache struct {
	client *redis.Client
	log    logger.Logger
}

// NewRedisClient 初始化 Redis 客户端连接，并返回一个 Cache 接口实例。
// 它接收 Redis 配置和日志接口。
func NewRedisClient(cfg *config.RedisConfig, log logger.Logger) (Cache, error) {
	if cfg == nil {
		return nil, ferrors.NewInternalError("Redis configuration is nil", nil)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,      // 密码，如果没有设置则为空字符串
		DB:           cfg.DB,            // 默认数据库，默认为 0
		PoolSize:     cfg.PoolSize,      // 连接池大小
		MinIdleConns: cfg.MinIdleConons, // 最小空闲连接数

		// 超时设置
		DialTimeout:  cfg.DialTimeout,  // 连接超时时间
		ReadTimeout:  cfg.ReadTimeout,  // 读取超时时间
		WriteTimeout: cfg.WriteTimeout, // 写入超时时间
		PoolTimeout:  cfg.PoolTimeout,  // 连接池获取连接超时时间
	})

	// 添加 OpenTelemetry Redis Hook，使其自动追踪 Redis 操作
	// 这里使用全局的 TracerProvider。确保在 InitOpenTelemetry 之后调用此函数。
	rdb.AddHook(otelredis.NewTracingHook(
		otelredis.WithTracerProvider(otel.GetTracerProvider()),
		otelredis.WithDBStatement(true), // 记录 Redis 命令作为 db.statement 属性
	))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Ping 操作超时
	defer cancel()

	// 尝试连接并 Ping Redis 服务器，确保连接可用
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Error(fmt.Sprintf("Failed to connect to Redis: %v", err),
			zap.String("addr", cfg.Addr),
			zap.Int("db", cfg.DB))
		return nil, ferrors.NewInternalError(fmt.Sprintf("failed to connect to Redis: %v", err), err)
	}

	log.Info(fmt.Sprintf("Redis client initialized: %s (DB: %d)", cfg.Addr, cfg.DB))
	return &redisCache{
		client: rdb,
		log:    log,
	}, nil
}

// Set 将一个键值对存储到 Redis 中。
// value 会被序列化为 JSON 字符串。
func (r *redisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	// 将 value 序列化为 JSON 字符串
	serializedValue, err := json.Marshal(value)
	if err != nil {
		r.log.Error(fmt.Sprintf("Failed to marshal value for Redis key %s: %v", key, err))
		return ferrors.NewInternalError(fmt.Sprintf("failed to serialize value for cache key %s", key), err)
	}

	cmd := r.client.Set(ctx, key, serializedValue, expiration)
	if err := cmd.Err(); err != nil {
		r.log.Error(fmt.Sprintf("Failed to set Redis key %s: %v", key, err))
		return ferrors.NewExternalServiceError(fmt.Sprintf("failed to set cache key %s", key), err)
	}
	return nil
}

// Get 从 Redis 中检索与给定键关联的值，并返回为字符串。
func (r *redisCache) Get(ctx context.Context, key string) (string, error) {
	cmd := r.client.Get(ctx, key)
	if err := cmd.Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			// 如果键不存在，返回 ErrNotFound
			r.log.Debug(fmt.Sprintf("Redis key %s not found", key))
			return "", ferrors.NewNotFoundError(fmt.Sprintf("cache key %s not found", key), err)
		}
		r.log.Error(fmt.Sprintf("Failed to get Redis key %s: %v", key, err))
		return "", ferrors.NewExternalServiceError(fmt.Sprintf("failed to get cache key %s", key), err)
	}
	return cmd.Val(), nil
}

// Delete 从 Redis 中移除与给定键关联的项。
func (r *redisCache) Delete(ctx context.Context, key string) error {
	cmd := r.client.Del(ctx, key)
	if err := cmd.Err(); err != nil {
		r.log.Error(fmt.Sprintf("Failed to delete Redis key %s: %v", key, err))
		return ferrors.NewExternalServiceError(fmt.Sprintf("failed to delete cache key %s", key), err)
	}
	return nil
}
