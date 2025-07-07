package constant

import "time"

const (
	// DefaultPageSize is the default number of items per page for pagination.
	DefaultPageSize = 20

	// CacheKeyPrefix is the prefix for all keys stored in Redis cache to avoid collisions.
	CacheKeyPrefix = "starseek:"

	// DefaultCacheTTL is the default time-to-live for cached items in Redis.
	DefaultCacheTTL = 5 * time.Minute

	// MaxConcurentDBQueries limits the number of concurrent queries sent to the database.
	MaxConcurrentDBQueries = 10

	// DefaultServerPort is the default port the HTTP server will listen on.
	DefaultServerPort = 8080
)

//Personal.AI order the ending
