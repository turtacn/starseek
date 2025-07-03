package cache

import (
	"context"
	"time"
)

// Cache 定义了项目通用的缓存机制接口。
// 这个接口的目的是提供一个抽象层，允许底层缓存实现（如 Redis、Memcached、In-memory cache）
// 在不影响上层业务逻辑的情况下进行切换。
type Cache interface {
	// Set 将一个键值对存储到缓存中，并指定一个过期时间。
	//
	// ctx: 用于操作的上下文，可以包含超时、取消信号或追踪信息。
	// key: 缓存项的唯一标识符。
	// value: 要存储的数据。可以是任何 Go 类型，具体的序列化逻辑由实现负责（例如，序列化为 JSON 字符串）。
	// expiration: 缓存项的过期时间。如果为 0，则表示永不过期（如果底层缓存支持）。
	//
	// 返回错误，如果操作失败。
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error

	// Get 从缓存中检索与给定键关联的值。
	//
	// ctx: 用于操作的上下文。
	// key: 要检索的缓存项的键。
	//
	// 返回值作为字符串，以及一个错误。
	// 如果键不存在，应返回一个特定的错误（例如，在 common/errors 中定义的 ErrNotFound）。
	// 实现负责将存储的数据反序列化为字符串。
	Get(ctx context.Context, key string) (string, error)

	// Delete 从缓存中移除与给定键关联的项。
	//
	// ctx: 用于操作的上下文。
	// key: 要删除的缓存项的键。
	//
	// 返回错误，如果操作失败。即使键不存在，也不应返回错误。
	Delete(ctx context.Context, key string) error
}

// 注意: 为了 Get 方法能更通用地反序列化到指定类型，
// 更好的做法可能是 Get(ctx context.Context, key string, dest interface{}) error，
// 或者 Get(ctx context.Context, key string) ([]byte, error) 然后由调用方处理反序列化。
// 但是，根据实现要求，这里 Get 方法返回 string。
