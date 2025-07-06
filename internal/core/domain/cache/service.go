package cache

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/turtacn/starseek/internal/core/domain/cache/errors"
	"github.com/turtacn/starseek/internal/core/domain/cache/events"
	"github.com/turtacn/starseek/internal/core/domain/cache/repository"
	"github.com/turtacn/starseek/internal/core/ports"
	"github.com/turtacn/starseek/internal/infrastructure/logging"
)

// =============================================================================
// Cache Domain Service Interface
// =============================================================================

// CacheDomainService defines the cache domain service interface
type CacheDomainService interface {
	// Core cache operations
	Get(ctx context.Context, key string, namespace CacheNamespace) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration, namespace CacheNamespace) error
	Delete(ctx context.Context, key string, namespace CacheNamespace) error
	Clear(ctx context.Context, namespace CacheNamespace) error

	// Batch operations
	GetMulti(ctx context.Context, keys []string, namespace CacheNamespace) (map[string]interface{}, error)
	SetMulti(ctx context.Context, data map[string]interface{}, ttl time.Duration, namespace CacheNamespace) error
	DeleteMulti(ctx context.Context, keys []string, namespace CacheNamespace) error

	// Cache strategy management
	SetStrategy(ctx context.Context, namespace CacheNamespace, strategy *CacheStrategy) error
	GetStrategy(ctx context.Context, namespace CacheNamespace) (*CacheStrategy, error)
	OptimizeStrategy(ctx context.Context, namespace CacheNamespace) error

	// Cache warming
	WarmCache(ctx context.Context, namespace CacheNamespace, strategy CacheWarmingStrategy) error
	PreloadHotData(ctx context.Context, namespace CacheNamespace, keys []string) error

	// Cache monitoring
	GetStatistics(ctx context.Context, namespace CacheNamespace) (*CacheStatistics, error)
	GetHealth(ctx context.Context, namespace CacheNamespace) (*CacheHealth, error)
	GetMetrics(ctx context.Context, namespace CacheNamespace) (*CacheMetrics, error)

	// Cache cleanup
	CleanupExpired(ctx context.Context, namespace CacheNamespace) error
	CompactCache(ctx context.Context, namespace CacheNamespace) error
	OptimizeMemory(ctx context.Context, namespace CacheNamespace) error

	// Cache consistency
	InvalidatePattern(ctx context.Context, pattern string, namespace CacheNamespace) error
	InvalidateByTags(ctx context.Context, tags []string, namespace CacheNamespace) error
	SyncCache(ctx context.Context, namespace CacheNamespace) error

	// Multi-level cache coordination
	PromoteToL1(ctx context.Context, key string, namespace CacheNamespace) error
	DemoteToL2(ctx context.Context, key string, namespace CacheNamespace) error
	RebalanceLevels(ctx context.Context, namespace CacheNamespace) error

	// Cache management
	CreateNamespace(ctx context.Context, namespace CacheNamespace, config *CacheConfiguration) error
	DeleteNamespace(ctx context.Context, namespace CacheNamespace) error
	ListNamespaces(ctx context.Context) ([]CacheNamespace, error)

	// Event handling
	Subscribe(ctx context.Context, eventType CacheEventType, handler CacheEventHandler) error
	Unsubscribe(ctx context.Context, eventType CacheEventType) error
	PublishEvent(ctx context.Context, event *CacheEvent) error
}

// =============================================================================
// Cache Domain Service Implementation
// =============================================================================

// cacheDomainService implements the CacheDomainService interface
type cacheDomainService struct {
	// Dependencies
	repository repository.CacheRepository
	eventBus   events.EventBus
	logger     logging.Logger

	// Cache instances
	l1Cache       L1Cache // Memory cache
	l2Cache       L2Cache // Redis cache
	cacheRegistry *CacheRegistry

	// Strategy management
	strategyManager *CacheStrategyManager

	// Monitoring and metrics
	metricsCollector *CacheMetricsCollector
	healthChecker    *CacheHealthChecker

	// Background workers
	cleanupWorker *CacheCleanupWorker
	monitorWorker *CacheMonitorWorker
	warmingWorker *CacheWarmingWorker

	// Event handlers
	eventHandlers map[CacheEventType][]CacheEventHandler

	// Configuration
	config *CacheDomainServiceConfig

	// Synchronization
	mu        sync.RWMutex
	stopCh    chan struct{}
	workersWG sync.WaitGroup

	// State
	isRunning bool
	startTime time.Time
}

// CacheDomainServiceConfig represents the configuration for cache domain service
type CacheDomainServiceConfig struct {
	// General settings
	DefaultTTL    time.Duration
	MaxCacheSize  int64
	MaxMemorySize int64

	// Strategy settings
	EnableAutoOptimization bool
	StrategyUpdateInterval time.Duration

	// Monitoring settings
	MetricsCollectionInterval time.Duration
	HealthCheckInterval       time.Duration

	// Cleanup settings
	CleanupInterval    time.Duration
	CompactionInterval time.Duration

	// Warming settings
	WarmingBatchSize   int
	WarmingConcurrency int
	WarmingTimeout     time.Duration

	// Multi-level cache settings
	L1CacheSize        int64
	L2CacheSize        int64
	PromotionThreshold int64
	DemotionThreshold  int64

	// Event settings
	EventBufferSize      int
	EventProcessingDelay time.Duration

	// Error handling
	MaxRetries              int
	RetryDelay              time.Duration
	CircuitBreakerThreshold int
}

// NewCacheDomainService creates a new cache domain service
func NewCacheDomainService(
	repository repository.CacheRepository,
	eventBus events.EventBus,
	logger logging.Logger,
	config *CacheDomainServiceConfig,
) CacheDomainService {
	service := &cacheDomainService{
		repository:    repository,
		eventBus:      eventBus,
		logger:        logger,
		config:        config,
		eventHandlers: make(map[CacheEventType][]CacheEventHandler),
		stopCh:        make(chan struct{}),
		startTime:     time.Now(),
	}

	// Initialize components
	service.initializeComponents()

	// Start background workers
	service.startBackgroundWorkers()

	return service
}

// initializeComponents initializes all service components
func (s *cacheDomainService) initializeComponents() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Initialize cache registry
	s.cacheRegistry = NewCacheRegistry()

	// Initialize strategy manager
	s.strategyManager = NewCacheStrategyManager(s.config)

	// Initialize metrics collector
	s.metricsCollector = NewCacheMetricsCollector(s.config)

	// Initialize health checker
	s.healthChecker = NewCacheHealthChecker(s.config)

	// Initialize workers
	s.cleanupWorker = NewCacheCleanupWorker(s.config)
	s.monitorWorker = NewCacheMonitorWorker(s.config)
	s.warmingWorker = NewCacheWarmingWorker(s.config)

	// Initialize L1 and L2 caches
	s.l1Cache = NewL1Cache(s.config.L1CacheSize)
	s.l2Cache = NewL2Cache(s.config.L2CacheSize)

	s.isRunning = true
	s.logger.Info("Cache domain service components initialized")
}

// startBackgroundWorkers starts all background workers
func (s *cacheDomainService) startBackgroundWorkers() {
	s.workersWG.Add(4)

	// Start cleanup worker
	go s.runCleanupWorker()

	// Start monitoring worker
	go s.runMonitorWorker()

	// Start warming worker
	go s.runWarmingWorker()

	// Start event processor
	go s.runEventProcessor()

	s.logger.Info("Cache domain service background workers started")
}

// =============================================================================
// Core Cache Operations
// =============================================================================

// Get retrieves a value from cache
func (s *cacheDomainService) Get(ctx context.Context, key string, namespace CacheNamespace) (interface{}, error) {
	// Validate input
	if err := s.validateKey(key); err != nil {
		return nil, err
	}

	// Create cache key
	cacheKey, err := NewCacheKey(namespace, key)
	if err != nil {
		return nil, err
	}

	// Try L1 cache first
	if value, found := s.l1Cache.Get(ctx, cacheKey.String()); found {
		s.recordCacheHit(namespace, "L1")
		return value, nil
	}

	// Try L2 cache
	if value, found := s.l2Cache.Get(ctx, cacheKey.String()); found {
		s.recordCacheHit(namespace, "L2")

		// Promote to L1 if it's a hot key
		if s.shouldPromoteToL1(cacheKey.String(), namespace) {
			s.l1Cache.Set(ctx, cacheKey.String(), value, s.config.DefaultTTL)
		}

		return value, nil
	}

	// Cache miss
	s.recordCacheMiss(namespace)
	return nil, errors.NewCacheNotFoundError(key)
}

// Set stores a value in cache
func (s *cacheDomainService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration, namespace CacheNamespace) error {
	// Validate input
	if err := s.validateKey(key); err != nil {
		return err
	}

	if err := s.validateValue(value); err != nil {
		return err
	}

	// Create cache key
	cacheKey, err := NewCacheKey(namespace, key)
	if err != nil {
		return err
	}

	// Create cache entry
	entry, err := NewCacheEntry(cacheKey, value, ttl)
	if err != nil {
		return err
	}

	// Apply caching strategy
	strategy := s.strategyManager.GetStrategy(namespace)
	if err := s.applySetStrategy(ctx, entry, strategy); err != nil {
		return err
	}

	// Store in appropriate cache levels
	if err := s.setInCacheLevels(ctx, entry); err != nil {
		return err
	}

	// Publish cache set event
	s.publishCacheEvent(ctx, &CacheEvent{
		Type:      CacheEventTypeSet,
		Key:       key,
		Namespace: namespace,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"ttl":  ttl,
			"size": entry.Size(),
		},
	})

	s.recordCacheSet(namespace)
	return nil
}

// Delete removes a value from cache
func (s *cacheDomainService) Delete(ctx context.Context, key string, namespace CacheNamespace) error {
	// Validate input
	if err := s.validateKey(key); err != nil {
		return err
	}

	// Create cache key
	cacheKey, err := NewCacheKey(namespace, key)
	if err != nil {
		return err
	}

	// Delete from all cache levels
	s.l1Cache.Delete(ctx, cacheKey.String())
	s.l2Cache.Delete(ctx, cacheKey.String())

	// Delete from repository
	if err := s.repository.Delete(ctx, cacheKey); err != nil {
		s.logger.Error("Failed to delete from repository", "error", err, "key", key)
		// Don't return error as cache levels are already cleared
	}

	// Publish cache delete event
	s.publishCacheEvent(ctx, &CacheEvent{
		Type:      CacheEventTypeDelete,
		Key:       key,
		Namespace: namespace,
		Timestamp: time.Now(),
	})

	s.recordCacheDelete(namespace)
	return nil
}

// Clear removes all values from a namespace
func (s *cacheDomainService) Clear(ctx context.Context, namespace CacheNamespace) error {
	// Clear from all cache levels
	s.l1Cache.ClearNamespace(ctx, namespace)
	s.l2Cache.ClearNamespace(ctx, namespace)

	// Clear from repository
	if err := s.repository.DeleteByNamespace(ctx, namespace); err != nil {
		s.logger.Error("Failed to clear namespace from repository", "error", err, "namespace", namespace)
		return err
	}

	// Publish cache clear event
	s.publishCacheEvent(ctx, &CacheEvent{
		Type:      CacheEventTypeClear,
		Namespace: namespace,
		Timestamp: time.Now(),
	})

	s.recordCacheClear(namespace)
	return nil
}

// =============================================================================
// Batch Operations
// =============================================================================

// GetMulti retrieves multiple values from cache
func (s *cacheDomainService) GetMulti(ctx context.Context, keys []string, namespace CacheNamespace) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Process keys in batches
	batchSize := 100
	for i := 0; i < len(keys); i += batchSize {
		end := i + batchSize
		if end > len(keys) {
			end = len(keys)
		}

		batch := keys[i:end]
		batchResult, err := s.getMultiBatch(ctx, batch, namespace)
		if err != nil {
			return nil, err
		}

		for k, v := range batchResult {
			result[k] = v
		}
	}

	return result, nil
}

// getMultiBatch processes a batch of keys for GetMulti
func (s *cacheDomainService) getMultiBatch(ctx context.Context, keys []string, namespace CacheNamespace) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	missingKeys := make([]string, 0)

	// Try L1 cache first
	for _, key := range keys {
		cacheKey, err := NewCacheKey(namespace, key)
		if err != nil {
			continue
		}

		if value, found := s.l1Cache.Get(ctx, cacheKey.String()); found {
			result[key] = value
			s.recordCacheHit(namespace, "L1")
		} else {
			missingKeys = append(missingKeys, key)
		}
	}

	// Try L2 cache for missing keys
	if len(missingKeys) > 0 {
		l2Results := s.l2Cache.GetMulti(ctx, missingKeys, namespace)
		for key, value := range l2Results {
			result[key] = value
			s.recordCacheHit(namespace, "L2")

			// Promote to L1 if it's a hot key
			if s.shouldPromoteToL1(key, namespace) {
				cacheKey, _ := NewCacheKey(namespace, key)
				s.l1Cache.Set(ctx, cacheKey.String(), value, s.config.DefaultTTL)
			}
		}

		// Update missing keys list
		newMissingKeys := make([]string, 0)
		for _, key := range missingKeys {
			if _, found := l2Results[key]; !found {
				newMissingKeys = append(newMissingKeys, key)
			}
		}
		missingKeys = newMissingKeys
	}

	// Record cache misses
	for _, key := range missingKeys {
		s.recordCacheMiss(namespace)
	}

	return result, nil
}

// SetMulti stores multiple values in cache
func (s *cacheDomainService) SetMulti(ctx context.Context, data map[string]interface{}, ttl time.Duration, namespace CacheNamespace) error {
	// Process data in batches
	batchSize := 100
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}

	for i := 0; i < len(keys); i += batchSize {
		end := i + batchSize
		if end > len(keys) {
			end = len(keys)
		}

		batch := make(map[string]interface{})
		for j := i; j < end; j++ {
			key := keys[j]
			batch[key] = data[key]
		}

		if err := s.setMultiBatch(ctx, batch, ttl, namespace); err != nil {
			return err
		}
	}

	return nil
}

// setMultiBatch processes a batch of key-value pairs for SetMulti
func (s *cacheDomainService) setMultiBatch(ctx context.Context, data map[string]interface{}, ttl time.Duration, namespace CacheNamespace) error {
	entries := make([]*CacheEntry, 0, len(data))

	// Create cache entries
	for key, value := range data {
		if err := s.validateKey(key); err != nil {
			s.logger.Warn("Invalid key in batch", "key", key, "error", err)
			continue
		}

		if err := s.validateValue(value); err != nil {
			s.logger.Warn("Invalid value in batch", "key", key, "error", err)
			continue
		}

		cacheKey, err := NewCacheKey(namespace, key)
		if err != nil {
			s.logger.Warn("Failed to create cache key", "key", key, "error", err)
			continue
		}

		entry, err := NewCacheEntry(cacheKey, value, ttl)
		if err != nil {
			s.logger.Warn("Failed to create cache entry", "key", key, "error", err)
			continue
		}

		entries = append(entries, entry)
	}

	// Apply caching strategy
	strategy := s.strategyManager.GetStrategy(namespace)
	for _, entry := range entries {
		if err := s.applySetStrategy(ctx, entry, strategy); err != nil {
			s.logger.Warn("Failed to apply set strategy", "key", entry.Key().String(), "error", err)
		}
	}

	// Store in cache levels
	if err := s.setMultiInCacheLevels(ctx, entries); err != nil {
		return err
	}

	// Publish batch set event
	s.publishCacheEvent(ctx, &CacheEvent{
		Type:      CacheEventTypeSetMulti,
		Namespace: namespace,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"count": len(entries),
			"ttl":   ttl,
		},
	})

	s.recordCacheSetMulti(namespace, len(entries))
	return nil
}

// DeleteMulti removes multiple values from cache
func (s *cacheDomainService) DeleteMulti(ctx context.Context, keys []string, namespace CacheNamespace) error {
	// Process keys in batches
	batchSize := 100
	for i := 0; i < len(keys); i += batchSize {
		end := i + batchSize
		if end > len(keys) {
			end = len(keys)
		}

		batch := keys[i:end]
		if err := s.deleteMultiBatch(ctx, batch, namespace); err != nil {
			return err
		}
	}

	return nil
}

// deleteMultiBatch processes a batch of keys for DeleteMulti
func (s *cacheDomainService) deleteMultiBatch(ctx context.Context, keys []string, namespace CacheNamespace) error {
	cacheKeys := make([]*CacheKey, 0, len(keys))

	// Create cache keys and delete from cache levels
	for _, key := range keys {
		if err := s.validateKey(key); err != nil {
			s.logger.Warn("Invalid key in delete batch", "key", key, "error", err)
			continue
		}

		cacheKey, err := NewCacheKey(namespace, key)
		if err != nil {
			s.logger.Warn("Failed to create cache key for delete", "key", key, "error", err)
			continue
		}

		cacheKeys = append(cacheKeys, cacheKey)

		// Delete from cache levels
		s.l1Cache.Delete(ctx, cacheKey.String())
		s.l2Cache.Delete(ctx, cacheKey.String())
	}

	// Delete from repository
	if err := s.repository.DeleteMulti(ctx, cacheKeys); err != nil {
		s.logger.Error("Failed to delete batch from repository", "error", err, "count", len(cacheKeys))
		// Don't return error as cache levels are already cleared
	}

	// Publish batch delete event
	s.publishCacheEvent(ctx, &CacheEvent{
		Type:      CacheEventTypeDeleteMulti,
		Namespace: namespace,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"count": len(cacheKeys),
		},
	})

	s.recordCacheDeleteMulti(namespace, len(cacheKeys))
	return nil
}

// =============================================================================
// Cache Strategy Management
// =============================================================================

// CacheStrategy represents a caching strategy
type CacheStrategy struct {
	Name               string
	Namespace          CacheNamespace
	DefaultTTL         time.Duration
	MaxSize            int64
	EvictionPolicy     EvictionPolicy
	L1Enabled          bool
	L2Enabled          bool
	CompressionEnabled bool
	SerializationType  string
	AccessPattern      AccessPattern
	DataType           DataType
	Priorities         map[string]int
	Rules              []CacheRule
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// AccessPattern represents data access patterns
type AccessPattern string

const (
	AccessPatternReadHeavy  AccessPattern = "READ_HEAVY"
	AccessPatternWriteHeavy AccessPattern = "WRITE_HEAVY"
	AccessPatternMixed      AccessPattern = "MIXED"
	AccessPatternBurst      AccessPattern = "BURST"
	AccessPatternSequential AccessPattern = "SEQUENTIAL"
	AccessPatternRandom     AccessPattern = "RANDOM"
)

// DataType represents data types for optimization
type DataType string

const (
	DataTypeString  DataType = "STRING"
	DataTypeJSON    DataType = "JSON"
	DataTypeBinary  DataType = "BINARY"
	DataTypeNumber  DataType = "NUMBER"
	DataTypeBoolean DataType = "BOOLEAN"
	DataTypeList    DataType = "LIST"
	DataTypeMap     DataType = "MAP"
	DataTypeCustom  DataType = "CUSTOM"
)

// CacheRule represents a caching rule
type CacheRule struct {
	ID         string
	Name       string
	Condition  string
	Action     string
	Priority   int
	Enabled    bool
	Parameters map[string]interface{}
	CreatedAt  time.Time
}

// CacheStrategyManager manages cache strategies
type CacheStrategyManager struct {
	strategies map[CacheNamespace]*CacheStrategy
	mu         sync.RWMutex
	config     *CacheDomainServiceConfig
}

// NewCacheStrategyManager creates a new cache strategy manager
func NewCacheStrategyManager(config *CacheDomainServiceConfig) *CacheStrategyManager {
	return &CacheStrategyManager{
		strategies: make(map[CacheNamespace]*CacheStrategy),
		config:     config,
	}
}

// SetStrategy sets a cache strategy for a namespace
func (s *cacheDomainService) SetStrategy(ctx context.Context, namespace CacheNamespace, strategy *CacheStrategy) error {
	if err := s.validateStrategy(strategy); err != nil {
		return err
	}

	s.strategyManager.SetStrategy(namespace, strategy)

	// Publish strategy change event
	s.publishCacheEvent(ctx, &CacheEvent{
		Type:      CacheEventTypeStrategyChange,
		Namespace: namespace,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"strategy_name": strategy.Name,
		},
	})

	s.logger.Info("Cache strategy set", "namespace", namespace, "strategy", strategy.Name)
	return nil
}

// GetStrategy gets a cache strategy for a namespace
func (s *cacheDomainService) GetStrategy(ctx context.Context, namespace CacheNamespace) (*CacheStrategy, error) {
	strategy := s.strategyManager.GetStrategy(namespace)
	if strategy == nil {
		return nil, errors.NewStrategyNotFoundError(string(namespace))
	}

	return strategy, nil
}

// OptimizeStrategy optimizes cache strategy based on usage patterns
func (s *cacheDomainService) OptimizeStrategy(ctx context.Context, namespace CacheNamespace) error {
	// Get current statistics
	stats, err := s.GetStatistics(ctx, namespace)
	if err != nil {
		return err
	}

	// Analyze usage patterns
	analysis := s.analyzeUsagePatterns(stats)

	// Create optimized strategy
	optimizedStrategy := s.createOptimizedStrategy(namespace, analysis)

	// Apply optimized strategy
	if err := s.SetStrategy(ctx, namespace, optimizedStrategy); err != nil {
		return err
	}

	s.logger.Info("Cache strategy optimized", "namespace", namespace, "analysis", analysis)
	return nil
}

// SetStrategy sets a strategy in the manager
func (csm *CacheStrategyManager) SetStrategy(namespace CacheNamespace, strategy *CacheStrategy) {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	strategy.UpdatedAt = time.Now()
	csm.strategies[namespace] = strategy
}

// GetStrategy gets a strategy from the manager
func (csm *CacheStrategyManager) GetStrategy(namespace CacheNamespace) *CacheStrategy {
	csm.mu.RLock()
	defer csm.mu.RUnlock()

	strategy, exists := csm.strategies[namespace]
	if !exists {
		return csm.getDefaultStrategy(namespace)
	}

	return strategy
}

// getDefaultStrategy returns a default strategy for a namespace
func (csm *CacheStrategyManager) getDefaultStrategy(namespace CacheNamespace) *CacheStrategy {
	return &CacheStrategy{
		Name:               "default",
		Namespace:          namespace,
		DefaultTTL:         csm.config.DefaultTTL,
		MaxSize:            csm.config.MaxCacheSize,
		EvictionPolicy:     EvictionPolicyLRU,
		L1Enabled:          true,
		L2Enabled:          true,
		CompressionEnabled: false,
		SerializationType:  "json",
		AccessPattern:      AccessPatternMixed,
		DataType:           DataTypeJSON,
		Priorities:         make(map[string]int),
		Rules:              make([]CacheRule, 0),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}

// =============================================================================
// Cache Warming Logic
// =============================================================================

// CacheWarmingWorker handles cache warming operations
type CacheWarmingWorker struct {
	config    *CacheDomainServiceConfig
	isRunning bool
	mu        sync.RWMutex
}

// NewCacheWarmingWorker creates a new cache warming worker
func NewCacheWarmingWorker(config *CacheDomainServiceConfig) *CacheWarmingWorker {
	return &CacheWarmingWorker{
		config:    config,
		isRunning: false,
	}
}

// WarmCache warms the cache with data according to the specified strategy
func (s *cacheDomainService) WarmCache(ctx context.Context, namespace CacheNamespace, strategy CacheWarmingStrategy) error {
	s.logger.Info("Starting cache warming", "namespace", namespace, "strategy", strategy)

	switch strategy {
	case CacheWarmingStrategyPreload:
		return s.warmCachePreload(ctx, namespace)
	case CacheWarmingStrategyBackground:
		return s.warmCacheBackground(ctx, namespace)
	case CacheWarmingStrategyLazy:
		return s.warmCacheLazy(ctx, namespace)
	case CacheWarmingStrategyScheduled:
		return s.warmCacheScheduled(ctx, namespace)
	default:
		return errors.NewInvalidStrategyError(string(strategy))
	}
}

// warmCachePreload implements preload warming strategy
func (s *cacheDomainService) warmCachePreload(ctx context.Context, namespace CacheNamespace) error {
	// Get hot keys from statistics
	hotKeys, err := s.getHotKeys(ctx, namespace)
	if err != nil {
		return err
	}

	// Preload hot data
	return s.PreloadHotData(ctx, namespace, hotKeys)
}

// warmCacheBackground implements background warming strategy
func (s *cacheDomainService) warmCacheBackground(ctx context.Context, namespace CacheNamespace) error {
	// Start background warming process
	go s.backgroundWarmingProcess(ctx, namespace)
	return nil
}

// warmCacheLazy implements lazy warming strategy
func (s *cacheDomainService) warmCacheLazy(ctx context.Context, namespace CacheNamespace) error {
	// Enable lazy warming for the namespace
	s.setLazyWarmingEnabled(namespace, true)
	return nil
}

// warmCacheScheduled implements scheduled warming strategy
func (s *cacheDomainService) warmCacheScheduled(ctx context.Context, namespace CacheNamespace) error {
	// Schedule periodic warming
	go s.scheduledWarmingProcess(ctx, namespace)
	return nil
}

// PreloadHotData preloads hot data into cache
func (s *cacheDomainService) PreloadHotData(ctx context.Context, namespace CacheNamespace, keys []string) error {
	s.logger.Info("Preloading hot data", "namespace", namespace, "key_count", len(keys))

	// Process keys in batches
	batchSize := s.config.WarmingBatchSize
	concurrency := s.config.WarmingConcurrency

	// Create worker pool
	keysChan := make(chan string, len(keys))
	errorsChan := make(chan error, len(keys))

	// Start workers
	for i := 0; i < concurrency; i++ {
		go s.preloadWorker(ctx, namespace, keysChan, errorsChan)
	}

	// Send keys to workers
	for _, key := range keys {
		keysChan <- key
	}
	close(keysChan)

	// Collect results
	var errors []error
	for i := 0; i < len(keys); i++ {
		if err := <-errorsChan; err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		s.logger.Warn("Some keys failed to preload", "error_count", len(errors))
		return errors[0] // Return first error
	}

	s.logger.Info("Hot data preloaded successfully", "namespace", namespace, "key_count", len(keys))
	return nil
}

// preloadWorker is a worker for preloading data
func (s *cacheDomainService) preloadWorker(ctx context.Context, namespace CacheNamespace, keysChan <-chan string, errorsChan chan<- error) {
	for key := range keysChan {
		// Load data from repository
		data, err := s.loadDataFromRepository(ctx, key, namespace)
		if err != nil {
			errorsChan <- err
			continue
		}

		// Store in cache
		if err := s.Set(ctx, key, data, s.config.DefaultTTL, namespace); err != nil {
			errorsChan <- err
			continue
		}

		errorsChan <- nil
	}
}

// backgroundWarmingProcess runs background warming
func (s *cacheDomainService) backgroundWarmingProcess(ctx context.Context, namespace CacheNamespace) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			// Get keys that need warming
			keys, err := s.getKeysForWarming(ctx, namespace)
			if err != nil {
				s.logger.Error("Failed to get keys for warming", "error", err)
				continue
			}

			if len(keys) > 0 {
				if err := s.PreloadHotData(ctx, namespace, keys); err != nil {
					s.logger.Error("Background warming failed", "error", err)
				}
			}
		}
	}
}

// scheduledWarmingProcess runs scheduled warming
func (s *cacheDomainService) scheduledWarmingProcess(ctx context.Context, namespace CacheNamespace) {
	// Schedule warming at specific times (e.g., before peak hours)
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			// Check if it's time for warming
			if s.isWarmingTime() {
				if err := s.warmCachePreload(ctx, namespace); err != nil {
					s.logger.Error("Scheduled warming failed", "error", err)
				}
			}
		}
	}
}

// =============================================================================
// Cache Monitoring Logic
// =============================================================================

// CacheMetrics represents cache metrics
type CacheMetrics struct {
	Namespace       CacheNamespace
	TotalRequests   int64
	CacheHits       int64
	CacheMisses     int64
	L1Hits          int64
	L2Hits          int64
	ErrorCount      int64
	AverageLatency  time.Duration
	ThroughputQPS   float64
	MemoryUsage     int64
	StorageUsage    int64
	EvictionCount   int64
	ExpirationCount int64
	LastUpdated     time.Time
	HitRate         float64
	MissRate        float64
}

// CacheMetricsCollector collects and aggregates cache metrics
type CacheMetricsCollector struct {
	metrics map[CacheNamespace]*CacheMetrics
	mu      sync.RWMutex
	config  *CacheDomainServiceConfig
}

// NewCacheMetricsCollector creates a new metrics collector
func NewCacheMetricsCollector(config *CacheDomainServiceConfig) *CacheMetricsCollector {
	return &CacheMetricsCollector{
		metrics: make(map[CacheNamespace]*CacheMetrics),
		config:  config,
	}
}

// GetStatistics returns cache statistics for a namespace
func (s *cacheDomainService) GetStatistics(ctx context.Context, namespace CacheNamespace) (*CacheStatistics, error) {
	// Get metrics from collector
	metrics := s.metricsCollector.GetMetrics(namespace)

	// Get repository statistics
	repoStats, err := s.repository.GetStatistics(ctx, namespace)
	if err != nil {
		s.logger.Error("Failed to get repository statistics", "error", err, "namespace", namespace)
		// Continue with local metrics only
	}

	// Combine statistics
	stats := &CacheStatistics{
		Namespace:       namespace,
		TotalRequests:   metrics.TotalRequests,
		CacheHits:       metrics.CacheHits,
		CacheMisses:     metrics.CacheMisses,
		HitRate:         metrics.HitRate,
		MissRate:        metrics.MissRate,
		AverageLatency:  metrics.AverageLatency,
		ThroughputQPS:   metrics.ThroughputQPS,
		MemoryUsage:     metrics.MemoryUsage,
		StorageUsage:    metrics.StorageUsage,
		EvictionCount:   metrics.EvictionCount,
		ExpirationCount: metrics.ExpirationCount,
		LastUpdated:     metrics.LastUpdated,
		ErrorCount:      metrics.ErrorCount,
	}

	// Merge with repository stats if available
	if repoStats != nil {
		stats.TotalSize = repoStats.TotalSize
		stats.EntryCount = repoStats.EntryCount
		stats.AverageEntrySize = repoStats.AverageEntrySize
		stats.LastAccess = repoStats.LastAccess
	}

	return stats, nil
}

// GetHealth returns cache health information
func (s *cacheDomainService) GetHealth(ctx context.Context, namespace CacheNamespace) (*CacheHealth, error) {
	health := s.healthChecker.GetHealth(namespace)

	// Get additional health info from repository
	repoHealth, err := s.repository.GetHealth(ctx, namespace)
	if err != nil {
		s.logger.Error("Failed to get repository health", "error", err, "namespace", namespace)
		// Continue with local health only
	}

	// Combine health information
	if repoHealth != nil {
		health.Status = s.combineHealthStatus(health.Status, repoHealth.Status)
		health.Issues = append(health.Issues, repoHealth.Issues...)
	}

	return health, nil
}

// GetMetrics returns cache metrics for a namespace
func (s *cacheDomainService) GetMetrics(ctx context.Context, namespace CacheNamespace) (*CacheMetrics, error) {
	metrics := s.metricsCollector.GetMetrics(namespace)
	if metrics == nil {
		return &CacheMetrics{
			Namespace:   namespace,
			LastUpdated: time.Now(),
		}, nil
	}

	return metrics, nil
}

// GetMetrics returns metrics for a namespace
func (cmc *CacheMetricsCollector) GetMetrics(namespace CacheNamespace) *CacheMetrics {
	cmc.mu.RLock()
	defer cmc.mu.RUnlock()

	metrics, exists := cmc.metrics[namespace]
	if !exists {
		return &CacheMetrics{
			Namespace:   namespace,
			LastUpdated: time.Now(),
		}
	}

	return metrics
}

// UpdateMetrics updates metrics for a namespace
func (cmc *CacheMetricsCollector) UpdateMetrics(namespace CacheNamespace, update func(*CacheMetrics)) {
	cmc.mu.Lock()
	defer cmc.mu.Unlock()

	metrics, exists := cmc.metrics[namespace]
	if !exists {
		metrics = &CacheMetrics{
			Namespace:   namespace,
			LastUpdated: time.Now(),
		}
		cmc.metrics[namespace] = metrics
	}

	update(metrics)
	metrics.LastUpdated = time.Now()

	// Calculate derived metrics
	total := metrics.CacheHits + metrics.CacheMisses
	if total > 0 {
		metrics.HitRate = float64(metrics.CacheHits) / float64(total)
		metrics.MissRate = float64(metrics.CacheMisses) / float64(total)
	}
}

// CacheHealthChecker monitors cache health
type CacheHealthChecker struct {
	healthData map[CacheNamespace]*CacheHealth
	mu         sync.RWMutex
	config     *CacheDomainServiceConfig
}

// NewCacheHealthChecker creates a new health checker
func NewCacheHealthChecker(config *CacheDomainServiceConfig) *CacheHealthChecker {
	return &CacheHealthChecker{
		healthData: make(map[CacheNamespace]*CacheHealth),
		config:     config,
	}
}

// GetHealth returns health information for a namespace
func (chc *CacheHealthChecker) GetHealth(namespace CacheNamespace) *CacheHealth {
	chc.mu.RLock()
	defer chc.mu.RUnlock()

	health, exists := chc.healthData[namespace]
	if !exists {
		return &CacheHealth{
			Namespace: namespace,
			Status:    CacheHealthStatusHealthy,
			Issues:    make([]CacheHealthIssue, 0),
			LastCheck: time.Now(),
		}
	}

	return health
}

// UpdateHealth updates health information
func (chc *CacheHealthChecker) UpdateHealth(namespace CacheNamespace, update func(*CacheHealth)) {
	chc.mu.Lock()
	defer chc.mu.Unlock()

	health, exists := chc.healthData[namespace]
	if !exists {
		health = &CacheHealth{
			Namespace: namespace,
			Status:    CacheHealthStatusHealthy,
			Issues:    make([]CacheHealthIssue, 0),
			LastCheck: time.Now(),
		}
		chc.healthData[namespace] = health
	}

	update(health)
	health.LastCheck = time.Now()
}

// =============================================================================
// Cache Cleanup Logic
// =============================================================================

// CacheCleanupWorker handles cache cleanup operations
type CacheCleanupWorker struct {
	config    *CacheDomainServiceConfig
	isRunning bool
	mu        sync.RWMutex
}

// NewCacheCleanupWorker creates a new cleanup worker
func NewCacheCleanupWorker(config *CacheDomainServiceConfig) *CacheCleanupWorker {
	return &CacheCleanupWorker{
		config:    config,
		isRunning: false,
	}
}

// CleanupExpired removes expired cache entries
func (s *cacheDomainService) CleanupExpired(ctx context.Context, namespace CacheNamespace) error {
	s.logger.Info("Starting expired cache cleanup", "namespace", namespace)

	// Cleanup L1 cache
	if err := s.l1Cache.CleanupExpired(ctx, namespace); err != nil {
		s.logger.Error("Failed to cleanup L1 cache", "error", err, "namespace", namespace)
	}

	// Cleanup L2 cache
	if err := s.l2Cache.CleanupExpired(ctx, namespace); err != nil {
		s.logger.Error("Failed to cleanup L2 cache", "error", err, "namespace", namespace)
	}

	// Cleanup repository
	if err := s.repository.CleanupExpired(ctx, namespace); err != nil {
		s.logger.Error("Failed to cleanup repository", "error", err, "namespace", namespace)
		return err
	}

	// Publish cleanup event
	s.publishCacheEvent(ctx, &CacheEvent{
		Type:      CacheEventTypeCleanup,
		Namespace: namespace,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"cleanup_type": "expired",
		},
	})

	s.logger.Info("Expired cache cleanup completed", "namespace", namespace)
	return nil
}

// CompactCache performs cache compaction
func (s *cacheDomainService) CompactCache(ctx context.Context, namespace CacheNamespace) error {
	s.logger.Info("Starting cache compaction", "namespace", namespace)

	// Compact L1 cache
	if err := s.l1Cache.Compact(ctx, namespace); err != nil {
		s.logger.Error("Failed to compact L1 cache", "error", err, "namespace", namespace)
	}

	// Compact L2 cache
	if err := s.l2Cache.Compact(ctx, namespace); err != nil {
		s.logger.Error("Failed to compact L2 cache", "error", err, "namespace", namespace)
	}

	// Compact repository
	if err := s.repository.Compact(ctx, namespace); err != nil {
		s.logger.Error("Failed to compact repository", "error", err, "namespace", namespace)
		return err
	}

	// Publish compaction event
	s.publishCacheEvent(ctx, &CacheEvent{
		Type:      CacheEventTypeCompaction,
		Namespace: namespace,
		Timestamp: time.Now(),
	})

	s.logger.Info("Cache compaction completed", "namespace", namespace)
	return nil
}

// OptimizeMemory optimizes memory usage
func (s *cacheDomainService) OptimizeMemory(ctx context.Context, namespace CacheNamespace) error {
	s.logger.Info("Starting memory optimization", "namespace", namespace)

	// Get current memory usage
	metrics, err := s.GetMetrics(ctx, namespace)
	if err != nil {
		return err
	}

	// Check if optimization is needed
	memoryThreshold := s.config.MaxMemorySize * 80 / 100 // 80% threshold
	if metrics.MemoryUsage < memoryThreshold {
		s.logger.Debug("Memory optimization not needed", "namespace", namespace, "usage", metrics.MemoryUsage)
		return nil
	}

	// Perform optimization
	if err := s.optimizeMemoryUsage(ctx, namespace); err != nil {
		return err
	}

	// Publish optimization event
	s.publishCacheEvent(ctx, &CacheEvent{
		Type:      CacheEventTypeOptimization,
		Namespace: namespace,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"optimization_type": "memory",
		},
	})

	s.logger.Info("Memory optimization completed", "namespace", namespace)
	return nil
}

// optimizeMemoryUsage performs actual memory optimization
func (s *cacheDomainService) optimizeMemoryUsage(ctx context.Context, namespace CacheNamespace) error {
	// Get least recently used entries
	lruEntries, err := s.getLRUEntries(ctx, namespace, 100)
	if err != nil {
		return err
	}

	// Evict LRU entries
	for _, entry := range lruEntries {
		if err := s.Delete(ctx, entry.Key().Key(), namespace); err != nil {
			s.logger.Error("Failed to evict LRU entry", "error", err, "key", entry.Key().Key())
		}
	}

	// Compact after eviction
	return s.CompactCache(ctx, namespace)
}

// =============================================================================
// Cache Consistency Logic
// =============================================================================

// InvalidatePattern invalidates cache entries matching a pattern
func (s *cacheDomainService) InvalidatePattern(ctx context.Context, pattern string, namespace CacheNamespace) error {
	s.logger.Info("Invalidating cache pattern", "pattern", pattern, "namespace", namespace)

	// Get matching entries
	entries, err := s.repository.GetByPattern(ctx, pattern)
	if err != nil {
		return err
	}

	// Invalidate each entry
	for _, entry := range entries {
		if err := s.Delete(ctx, entry.Key().Key(), namespace); err != nil {
			s.logger.Error("Failed to invalidate entry", "error", err, "key", entry.Key().Key())
		}
	}

	// Publish invalidation event
	s.publishCacheEvent(ctx, &CacheEvent{
		Type:      CacheEventTypeInvalidation,
		Namespace: namespace,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"pattern":     pattern,
			"entry_count": len(entries),
		},
	})

	s.logger.Info("Cache pattern invalidation completed", "pattern", pattern, "count", len(entries))
	return nil
}

// InvalidateByTags invalidates cache entries by tags
func (s *cacheDomainService) InvalidateByTags(ctx context.Context, tags []string, namespace CacheNamespace) error {
	s.logger.Info("Invalidating cache by tags", "tags", tags, "namespace", namespace)

	var allEntries []*CacheEntry

	// Get entries for each tag
	for _, tag := range tags {
		entries, err := s.getEntriesByTag(ctx, tag, namespace)
		if err != nil {
			s.logger.Error("Failed to get entries by tag", "error", err, "tag", tag)
			continue
		}
		allEntries = append(allEntries, entries...)
	}

	// Remove duplicates
	uniqueEntries := s.removeDuplicateEntries(allEntries)

	// Invalidate each entry
	for _, entry := range uniqueEntries {
		if err := s.Delete(ctx, entry.Key().Key(), namespace); err != nil {
			s.logger.Error("Failed to invalidate entry", "error", err, "key", entry.Key().Key())
		}
	}

	// Publish invalidation event
	s.publishCacheEvent(ctx, &CacheEvent{
		Type:      CacheEventTypeInvalidation,
		Namespace: namespace,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"tags":        tags,
			"entry_count": len(uniqueEntries),
		},
	})

	s.logger.Info("Cache tag invalidation completed", "tags", tags, "count", len(uniqueEntries))
	return nil
}

// SyncCache synchronizes cache with the underlying data store
func (s *cacheDomainService) SyncCache(ctx context.Context, namespace CacheNamespace) error {
	s.logger.Info("Starting cache synchronization", "namespace", namespace)

	// Get all cache entries
	entries, err := s.repository.GetByNamespace(ctx, namespace)
	if err != nil {
		return err
	}

	// Sync each entry
	syncedCount := 0
	for _, entry := range entries {
		if err := s.syncCacheEntry(ctx, entry); err != nil {
			s.logger.Error("Failed to sync cache entry", "error", err, "key", entry.Key().Key())
			continue
		}
		syncedCount++
	}

	// Publish sync event
	s.publishCacheEvent(ctx, &CacheEvent{
		Type:      CacheEventTypeSync,
		Namespace: namespace,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"total_entries":  len(entries),
			"synced_entries": syncedCount,
		},
	})

	s.logger.Info("Cache synchronization completed", "namespace", namespace, "synced", syncedCount)
	return nil
}

// syncCacheEntry synchronizes a single cache entry
func (s *cacheDomainService) syncCacheEntry(ctx context.Context, entry *CacheEntry) error {
	// Check if entry exists in data store
	exists, err := s.checkDataStoreEntry(ctx, entry.Key())
	if err != nil {
		return err
	}

	if !exists {
		// Entry doesn't exist in data store, remove from cache
		return s.Delete(ctx, entry.Key().Key(), entry.Key().Namespace())
	}

	// Entry exists, check if it's up to date
	upToDate, err := s.checkEntryUpToDate(ctx, entry)
	if err != nil {
		return err
	}

	if !upToDate {
		// Entry is stale, remove from cache
		return s.Delete(ctx, entry.Key().Key(), entry.Key().Namespace())
	}

	return nil
}

// =============================================================================
// Multi-Level Cache Coordination
// =============================================================================

// L1Cache represents the L1 (memory) cache interface
type L1Cache interface {
	Get(ctx context.Context, key string) (interface{}, bool)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	ClearNamespace(ctx context.Context, namespace CacheNamespace) error
	CleanupExpired(ctx context.Context, namespace CacheNamespace) error
	Compact(ctx context.Context, namespace CacheNamespace) error
	GetSize() int64
	GetStats() map[string]interface{}
}

// L2Cache represents the L2 (Redis) cache interface
type L2Cache interface {
	Get(ctx context.Context, key string) (interface{}, bool)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	ClearNamespace(ctx context.Context, namespace CacheNamespace) error
	CleanupExpired(ctx context.Context, namespace CacheNamespace) error
	Compact(ctx context.Context, namespace CacheNamespace) error
	GetMulti(ctx context.Context, keys []string, namespace CacheNamespace) map[string]interface{}
	SetMulti(ctx context.Context, data map[string]interface{}, ttl time.Duration) error
	DeleteMulti(ctx context.Context, keys []string) error
	GetSize() int64
	GetStats() map[string]interface{}
}

// NewL1Cache creates a new L1 cache instance
func NewL1Cache(maxSize int64) L1Cache {
	return &memoryCache{
		maxSize:    maxSize,
		data:       make(map[string]*memoryCacheEntry),
		accessTime: make(map[string]time.Time),
	}
}

// NewL2Cache creates a new L2 cache instance
func NewL2Cache(maxSize int64) L2Cache {
	return &redisCache{
		maxSize: maxSize,
		// Redis client initialization would go here
	}
}

// PromoteToL1 promotes a cache entry to L1 cache
func (s *cacheDomainService) PromoteToL1(ctx context.Context, key string, namespace CacheNamespace) error {
	cacheKey, err := NewCacheKey(namespace, key)
	if err != nil {
		return err
	}

	// Get value from L2 cache
	value, found := s.l2Cache.Get(ctx, cacheKey.String())
	if !found {
		return errors.NewCacheNotFoundError(key)
	}

	// Set in L1 cache
	if err := s.l1Cache.Set(ctx, cacheKey.String(), value, s.config.DefaultTTL); err != nil {
		return err
	}

	// Record promotion
	s.recordCachePromotion(namespace, key)

	s.logger.Debug("Cache entry promoted to L1", "key", key, "namespace", namespace)
	return nil
}

// DemoteToL2 demotes a cache entry from L1 to L2 cache
func (s *cacheDomainService) DemoteToL2(ctx context.Context, key string, namespace CacheNamespace) error {
	cacheKey, err := NewCacheKey(namespace, key)
	if err != nil {
		return err
	}

	// Get value from L1 cache
	value, found := s.l1Cache.Get(ctx, cacheKey.String())
	if !found {
		return errors.NewCacheNotFoundError(key)
	}

	// Set in L2 cache
	if err := s.l2Cache.Set(ctx, cacheKey.String(), value, s.config.DefaultTTL); err != nil {
		return err
	}

	// Remove from L1 cache
	if err := s.l1Cache.Delete(ctx, cacheKey.String()); err != nil {
		s.logger.Error("Failed to remove from L1 cache during demotion", "error", err)
	}

	// Record demotion
	s.recordCacheDemotion(namespace, key)

	s.logger.Debug("Cache entry demoted to L2", "key", key, "namespace", namespace)
	return nil
}

// RebalanceLevels rebalances cache levels based on access patterns
func (s *cacheDomainService) RebalanceLevels(ctx context.Context, namespace CacheNamespace) error {
	s.logger.Info("Starting cache level rebalancing", "namespace", namespace)

	// Get access statistics
	l1Stats := s.l1Cache.GetStats()
	l2Stats := s.l2Cache.GetStats()

	// Determine rebalancing strategy
	strategy := s.determineRebalancingStrategy(l1Stats, l2Stats)

	// Apply rebalancing
	switch strategy {
	case "promote_hot":
		return s.promoteHotEntries(ctx, namespace)
	case "demote_cold":
		return s.demoteColdEntries(ctx, namespace)
	case "balanced":
		return s.balancedRebalancing(ctx, namespace)
	default:
		s.logger.Debug("No rebalancing needed", "namespace", namespace)
		return nil
	}
}

// promoteHotEntries promotes frequently accessed entries to L1
func (s *cacheDomainService) promoteHotEntries(ctx context.Context, namespace CacheNamespace) error {
	// Get hot keys from L2
	hotKeys, err := s.getHotKeysFromL2(ctx, namespace)
	if err != nil {
		return err
	}

	// Promote hot keys to L1
	for _, key := range hotKeys {
		if err := s.PromoteToL1(ctx, key, namespace); err != nil {
			s.logger.Error("Failed to promote hot key", "error", err, "key", key)
		}
	}

	s.logger.Info("Hot entries promoted to L1", "namespace", namespace, "count", len(hotKeys))
	return nil
}

// demoteColdEntries demotes rarely accessed entries from L1
func (s *cacheDomainService) demoteColdEntries(ctx context.Context, namespace CacheNamespace) error {
	// Get cold keys from L1
	coldKeys, err := s.getColdKeysFromL1(ctx, namespace)
	if err != nil {
		return err
	}

	// Demote cold keys to L2
	for _, key := range coldKeys {
		if err := s.DemoteToL2(ctx, key, namespace); err != nil {
			s.logger.Error("Failed to demote cold key", "error", err, "key", key)
		}
	}

	s.logger.Info("Cold entries demoted to L2", "namespace", namespace, "count", len(coldKeys))
	return nil
}

// balancedRebalancing performs balanced rebalancing
func (s *cacheDomainService) balancedRebalancing(ctx context.Context, namespace CacheNamespace) error {
	// Promote some hot entries
	if err := s.promoteHotEntries(ctx, namespace); err != nil {
		s.logger.Error("Failed to promote hot entries", "error", err)
	}

	// Demote some cold entries
	if err := s.demoteColdEntries(ctx, namespace); err != nil {
		s.logger.Error("Failed to demote cold entries", "error", err)
	}

	return nil
}

// =============================================================================
// Cache Level Implementation (Memory Cache)
// =============================================================================

// memoryCacheEntry represents a memory cache entry
type memoryCacheEntry struct {
	Value       interface{}
	Expiration  time.Time
	AccessCount int64
	LastAccess  time.Time
}

// memoryCache implements L1Cache interface
type memoryCache struct {
	maxSize    int64
	data       map[string]*memoryCacheEntry
	accessTime map[string]time.Time
	mu         sync.RWMutex
}

// Get retrieves a value from memory cache
func (mc *memoryCache) Get(ctx context.Context, key string) (interface{}, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	entry, exists := mc.data[key]
	if !exists {
		return nil, false
	}

	// Check expiration
	if time.Now().After(entry.Expiration) {
		// Entry expired, remove it
		delete(mc.data, key)
		delete(mc.accessTime, key)
		return nil, false
	}

	// Update access statistics
	entry.AccessCount++
	entry.LastAccess = time.Now()
	mc.accessTime[key] = time.Now()

	return entry.Value, true
}

// Set stores a value in memory cache
func (mc *memoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Check if we need to evict entries
	if int64(len(mc.data)) >= mc.maxSize {
		mc.evictLRU()
	}

	// Create new entry
	entry := &memoryCacheEntry{
		Value:       value,
		Expiration:  time.Now().Add(ttl),
		AccessCount: 1,
		LastAccess:  time.Now(),
	}

	mc.data[key] = entry
	mc.accessTime[key] = time.Now()

	return nil
}

// Delete removes a value from memory cache
func (mc *memoryCache) Delete(ctx context.Context, key string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	delete(mc.data, key)
	delete(mc.accessTime, key)

	return nil
}

// Clear removes all values from memory cache
func (mc *memoryCache) Clear(ctx context.Context) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.data = make(map[string]*memoryCacheEntry)
	mc.accessTime = make(map[string]time.Time)

	return nil
}

// ClearNamespace removes all values from a specific namespace
func (mc *memoryCache) ClearNamespace(ctx context.Context, namespace CacheNamespace) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	prefix := string(namespace) + ":"

	// Find keys to delete
	keysToDelete := make([]string, 0)
	for key := range mc.data {
		if strings.HasPrefix(key, prefix) {
			keysToDelete = append(keysToDelete, key)
		}
	}

	// Delete keys
	for _, key := range keysToDelete {
		delete(mc.data, key)
		delete(mc.accessTime, key)
	}

	return nil
}

// CleanupExpired removes expired entries
func (mc *memoryCache) CleanupExpired(ctx context.Context, namespace CacheNamespace) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	now := time.Now()
	expiredKeys := make([]string, 0)

	for key, entry := range mc.data {
		if now.After(entry.Expiration) {
			expiredKeys = append(expiredKeys, key)
		}
	}

	// Remove expired keys
	for _, key := range expiredKeys {
		delete(mc.data, key)
		delete(mc.accessTime, key)
	}

	return nil
}

// Compact compacts the memory cache
func (mc *memoryCache) Compact(ctx context.Context, namespace CacheNamespace) error {
	// For memory cache, compact is essentially cleanup
	return mc.CleanupExpired(ctx, namespace)
}

// GetSize returns the current size of the cache
func (mc *memoryCache) GetSize() int64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return int64(len(mc.data))
}

// GetStats returns cache statistics
func (mc *memoryCache) GetStats() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	totalAccess := int64(0)
	for _, entry := range mc.data {
		totalAccess += entry.AccessCount
	}

	return map[string]interface{}{
		"entry_count":  len(mc.data),
		"total_access": totalAccess,
		"max_size":     mc.maxSize,
	}
}

// evictLRU evicts the least recently used entry
func (mc *memoryCache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time

	for key, accessTime := range mc.accessTime {
		if oldestKey == "" || accessTime.Before(oldestTime) {
			oldestKey = key
			oldestTime = accessTime
		}
	}

	if oldestKey != "" {
		delete(mc.data, oldestKey)
		delete(mc.accessTime, oldestKey)
	}
}

// =============================================================================
// Cache Level Implementation (Redis Cache)
// =============================================================================

// redisCache implements L2Cache interface
type redisCache struct {
	maxSize int64
	// Redis client would be initialized here
}

// Get retrieves a value from Redis cache
func (rc *redisCache) Get(ctx context.Context, key string) (interface{}, bool) {
	// Redis implementation would go here
	return nil, false
}

// Set stores a value in Redis cache
func (rc *redisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Redis implementation would go here
	return nil
}

// Delete removes a value from Redis cache
func (rc *redisCache) Delete(ctx context.Context, key string) error {
	// Redis implementation would go here
	return nil
}

// Clear removes all values from Redis cache
func (rc *redisCache) Clear(ctx context.Context) error {
	// Redis implementation would go here
	return nil
}

// ClearNamespace removes all values from a specific namespace
func (rc *redisCache) ClearNamespace(ctx context.Context, namespace CacheNamespace) error {
	// Redis implementation would go here
	return nil
}

// CleanupExpired removes expired entries
func (rc *redisCache) CleanupExpired(ctx context.Context, namespace CacheNamespace) error {
	// Redis handles expiration automatically
	return nil
}

// Compact compacts the Redis cache
func (rc *redisCache) Compact(ctx context.Context, namespace CacheNamespace) error {
	// Redis implementation would go here
	return nil
}

// GetMulti retrieves multiple values from Redis cache
func (rc *redisCache) GetMulti(ctx context.Context, keys []string, namespace CacheNamespace) map[string]interface{} {
	// Redis implementation would go here
	return make(map[string]interface{})
}

// SetMulti stores multiple values in Redis cache
func (rc *redisCache) SetMulti(ctx context.Context, data map[string]interface{}, ttl time.Duration) error {
	// Redis implementation would go here
	return nil
}

// DeleteMulti removes multiple values from Redis cache
func (rc *redisCache) DeleteMulti(ctx context.Context, keys []string) error {
	// Redis implementation would go here
	return nil
}

// GetSize returns the current size of the cache
func (rc *redisCache) GetSize() int64 {
	// Redis implementation would go here
	return 0
}

// GetStats returns cache statistics
func (rc *redisCache) GetStats() map[string]interface{} {
	// Redis implementation would go here
	return make(map[string]interface{})
}

// =============================================================================
// Cache Management Operations
// =============================================================================

// CreateNamespace creates a new cache namespace
func (s *cacheDomainService) CreateNamespace(ctx context.Context, namespace CacheNamespace, config *CacheConfiguration) error {
	s.logger.Info("Creating cache namespace", "namespace", namespace)

	// Validate namespace
	if err := s.validateNamespace(namespace); err != nil {
		return err
	}

	// Check if namespace already exists
	if s.cacheRegistry.HasNamespace(namespace) {
		return errors.NewNamespaceExistsError(string(namespace))
	}

	// Create namespace configuration
	namespaceConfig := &NamespaceConfiguration{
		Namespace: namespace,
		Config:    config,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
	}

	// Register namespace
	if err := s.cacheRegistry.RegisterNamespace(namespace, namespaceConfig); err != nil {
		return err
	}

	// Initialize namespace in repository
	if err := s.repository.CreateNamespace(ctx, namespace, config); err != nil {
		s.cacheRegistry.UnregisterNamespace(namespace)
		return err
	}

	// Set default strategy
	defaultStrategy := s.createDefaultStrategy(namespace, config)
	if err := s.SetStrategy(ctx, namespace, defaultStrategy); err != nil {
		s.logger.Error("Failed to set default strategy", "error", err, "namespace", namespace)
	}

	// Publish namespace creation event
	s.publishCacheEvent(ctx, &CacheEvent{
		Type:      CacheEventTypeNamespaceCreated,
		Namespace: namespace,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"config": config,
		},
	})

	s.logger.Info("Cache namespace created successfully", "namespace", namespace)
	return nil
}

// DeleteNamespace deletes a cache namespace
func (s *cacheDomainService) DeleteNamespace(ctx context.Context, namespace CacheNamespace) error {
	s.logger.Info("Deleting cache namespace", "namespace", namespace)

	// Check if namespace exists
	if !s.cacheRegistry.HasNamespace(namespace) {
		return errors.NewNamespaceNotFoundError(string(namespace))
	}

	// Clear all cache data
	if err := s.Clear(ctx, namespace); err != nil {
		s.logger.Error("Failed to clear namespace data", "error", err, "namespace", namespace)
		return err
	}

	// Delete from repository
	if err := s.repository.DeleteNamespace(ctx, namespace); err != nil {
		return err
	}

	// Unregister namespace
	s.cacheRegistry.UnregisterNamespace(namespace)

	// Remove strategy
	s.strategyManager.RemoveStrategy(namespace)

	// Publish namespace deletion event
	s.publishCacheEvent(ctx, &CacheEvent{
		Type:      CacheEventTypeNamespaceDeleted,
		Namespace: namespace,
		Timestamp: time.Now(),
	})

	s.logger.Info("Cache namespace deleted successfully", "namespace", namespace)
	return nil
}

// ListNamespaces returns all cache namespaces
func (s *cacheDomainService) ListNamespaces(ctx context.Context) ([]CacheNamespace, error) {
	return s.cacheRegistry.ListNamespaces(), nil
}

// =============================================================================
// Event Handling
// =============================================================================

// Subscribe subscribes to cache events
func (s *cacheDomainService) Subscribe(ctx context.Context, eventType CacheEventType, handler CacheEventHandler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	handlers, exists := s.eventHandlers[eventType]
	if !exists {
		handlers = make([]CacheEventHandler, 0)
	}

	handlers = append(handlers, handler)
	s.eventHandlers[eventType] = handlers

	s.logger.Debug("Event handler subscribed", "event_type", eventType)
	return nil
}

// Unsubscribe unsubscribes from cache events
func (s *cacheDomainService) Unsubscribe(ctx context.Context, eventType CacheEventType) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.eventHandlers, eventType)

	s.logger.Debug("Event handler unsubscribed", "event_type", eventType)
	return nil
}

// PublishEvent publishes a cache event
func (s *cacheDomainService) PublishEvent(ctx context.Context, event *CacheEvent) error {
	return s.publishCacheEvent(ctx, event)
}

// publishCacheEvent publishes a cache event internally
func (s *cacheDomainService) publishCacheEvent(ctx context.Context, event *CacheEvent) error {
	// Publish to event bus
	if err := s.eventBus.Publish(ctx, event); err != nil {
		s.logger.Error("Failed to publish event to event bus", "error", err, "event", event)
	}

	// Notify local handlers
	s.notifyEventHandlers(ctx, event)

	return nil
}

// notifyEventHandlers notifies local event handlers
func (s *cacheDomainService) notifyEventHandlers(ctx context.Context, event *CacheEvent) {
	s.mu.RLock()
	handlers, exists := s.eventHandlers[event.Type]
	s.mu.RUnlock()

	if !exists {
		return
	}

	// Notify handlers asynchronously
	go func() {
		for _, handler := range handlers {
			func() {
				defer func() {
					if r := recover(); r != nil {
						s.logger.Error("Event handler panicked", "error", r, "event_type", event.Type)
					}
				}()

				if err := handler(ctx, event); err != nil {
					s.logger.Error("Event handler failed", "error", err, "event_type", event.Type)
				}
			}()
		}
	}()
}

// =============================================================================
// Background Workers
// =============================================================================

// runCleanupWorker runs the cleanup worker
func (s *cacheDomainService) runCleanupWorker() {
	defer s.workersWG.Done()

	ticker := time.NewTicker(s.config.CleanupInterval)
	defer ticker.Stop()

	s.logger.Info("Cleanup worker started")

	for {
		select {
		case <-s.stopCh:
			s.logger.Info("Cleanup worker stopping")
			return
		case <-ticker.C:
			s.runCleanupCycle()
		}
	}
}

// runCleanupCycle runs a cleanup cycle
func (s *cacheDomainService) runCleanupCycle() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	namespaces := s.cacheRegistry.ListNamespaces()

	for _, namespace := range namespaces {
		if err := s.CleanupExpired(ctx, namespace); err != nil {
			s.logger.Error("Cleanup failed", "error", err, "namespace", namespace)
		}
	}

	s.logger.Debug("Cleanup cycle completed", "namespaces", len(namespaces))
}

// runMonitorWorker runs the monitoring worker
func (s *cacheDomainService) runMonitorWorker() {
	defer s.workersWG.Done()

	ticker := time.NewTicker(s.config.MetricsCollectionInterval)
	defer ticker.Stop()

	s.logger.Info("Monitor worker started")

	for {
		select {
		case <-s.stopCh:
			s.logger.Info("Monitor worker stopping")
			return
		case <-ticker.C:
			s.runMonitoringCycle()
		}
	}
}

// runMonitoringCycle runs a monitoring cycle
func (s *cacheDomainService) runMonitoringCycle() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	namespaces := s.cacheRegistry.ListNamespaces()

	for _, namespace := range namespaces {
		// Update metrics
		s.updateMetrics(ctx, namespace)

		// Check health
		s.checkHealth(ctx, namespace)

		// Auto-optimize if enabled
		if s.config.EnableAutoOptimization {
			if err := s.OptimizeStrategy(ctx, namespace); err != nil {
				s.logger.Error("Auto-optimization failed", "error", err, "namespace", namespace)
			}
		}
	}
}

// runWarmingWorker runs the warming worker
func (s *cacheDomainService) runWarmingWorker() {
	defer s.workersWG.Done()

	s.logger.Info("Warming worker started")

	for {
		select {
		case <-s.stopCh:
			s.logger.Info("Warming worker stopping")
			return
		case <-time.After(1 * time.Minute):
			s.runWarmingCycle()
		}
	}
}

// runWarmingCycle runs a warming cycle
func (s *cacheDomainService) runWarmingCycle() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	namespaces := s.cacheRegistry.ListNamespaces()

	for _, namespace := range namespaces {
		// Check if warming is needed
		if s.isWarmingNeeded(ctx, namespace) {
			if err := s.WarmCache(ctx, namespace, CacheWarmingStrategyBackground); err != nil {
				s.logger.Error("Background warming failed", "error", err, "namespace", namespace)
			}
		}
	}
}

// runEventProcessor runs the event processor
func (s *cacheDomainService) runEventProcessor() {
	defer s.workersWG.Done()

	s.logger.Info("Event processor started")

	for {
		select {
		case <-s.stopCh:
			s.logger.Info("Event processor stopping")
			return
		case <-time.After(s.config.EventProcessingDelay):
			s.processEvents()
		}
	}
}

// processEvents processes pending events
func (s *cacheDomainService) processEvents() {
	// Process any pending events
	// This is a placeholder for event processing logic
}

// =============================================================================
// Helper Methods
// =============================================================================

// validateKey validates a cache key
func (s *cacheDomainService) validateKey(key string) error {
	if key == "" {
		return errors.NewInvalidKeyError("key cannot be empty")
	}

	if len(key) > 250 {
		return errors.NewInvalidKeyError("key too long")
	}

	return nil
}

// validateValue validates a cache value
func (s *cacheDomainService) validateValue(value interface{}) error {
	if value == nil {
		return errors.NewInvalidValueError("value cannot be nil")
	}

	return nil
}

// validateNamespace validates a cache namespace
func (s *cacheDomainService) validateNamespace(namespace CacheNamespace) error {
	if namespace == "" {
		return errors.NewInvalidNamespaceError("namespace cannot be empty")
	}

	return nil
}

// validateStrategy validates a cache strategy
func (s *cacheDomainService) validateStrategy(strategy *CacheStrategy) error {
	if strategy == nil {
		return errors.NewInvalidStrategyError("strategy cannot be nil")
	}

	if strategy.Name == "" {
		return errors.NewInvalidStrategyError("strategy name cannot be empty")
	}

	return nil
}

// shouldPromoteToL1 determines if a key should be promoted to L1 cache
func (s *cacheDomainService) shouldPromoteToL1(key string, namespace CacheNamespace) bool {
	// Check access frequency
	accessCount := s.getAccessCount(key, namespace)
	return accessCount >= s.config.PromotionThreshold
}

// getAccessCount gets the access count for a key
func (s *cacheDomainService) getAccessCount(key string, namespace CacheNamespace) int64 {
	// Implementation would track access counts
	return 0
}

// applySetStrategy applies caching strategy when setting a value
func (s *cacheDomainService) applySetStrategy(ctx context.Context, entry *CacheEntry, strategy *CacheStrategy) error {
	// Apply compression if enabled
	if strategy.CompressionEnabled {
		if err := s.compressEntry(entry); err != nil {
			s.logger.Error("Failed to compress entry", "error", err)
		}
	}

	// Apply serialization
	if err := s.serializeEntry(entry, strategy.SerializationType); err != nil {
		return err
	}

	// Apply rules
	for _, rule := range strategy.Rules {
		if rule.Enabled {
			if err := s.applyRule(ctx, entry, &rule); err != nil {
				s.logger.Error("Failed to apply rule", "error", err, "rule", rule.Name)
			}
		}
	}

	return nil
}

// setInCacheLevels stores an entry in appropriate cache levels
func (s *cacheDomainService) setInCacheLevels(ctx context.Context, entry *CacheEntry) error {
	strategy := s.strategyManager.GetStrategy(entry.Key().Namespace())

	// Set in L1 cache if enabled
	if strategy.L1Enabled {
		if err := s.l1Cache.Set(ctx, entry.Key().String(), entry.Value(), entry.TTL()); err != nil {
			s.logger.Error("Failed to set in L1 cache", "error", err)
		}
	}

	// Set in L2 cache if enabled
	if strategy.L2Enabled {
		if err := s.l2Cache.Set(ctx, entry.Key().String(), entry.Value(), entry.TTL()); err != nil {
			s.logger.Error("Failed to set in L2 cache", "error", err)
		}
	}

	// Set in repository
	if err := s.repository.Set(ctx, entry); err != nil {
		return err
	}

	return nil
}

// setMultiInCacheLevels stores multiple entries in cache levels
func (s *cacheDomainService) setMultiInCacheLevels(ctx context.Context, entries []*CacheEntry) error {
	if len(entries) == 0 {
		return nil
	}

	namespace := entries[0].Key().Namespace()
	strategy := s.strategyManager.GetStrategy(namespace)

	// Prepare data for batch operations
	data := make(map[string]interface{})
	for _, entry := range entries {
		data[entry.Key().String()] = entry.Value()
	}

	// Set in L1 cache if enabled
	if strategy.L1Enabled {
		for key, value := range data {
			if err := s.l1Cache.Set(ctx, key, value, strategy.DefaultTTL); err != nil {
				s.logger.Error("Failed to set in L1 cache", "error", err, "key", key)
			}
		}
	}

	// Set in L2 cache if enabled
	if strategy.L2Enabled {
		if err := s.l2Cache.SetMulti(ctx, data, strategy.DefaultTTL); err != nil {
			s.logger.Error("Failed to set batch in L2 cache", "error", err)
		}
	}

	// Set in repository
	if err := s.repository.SetMulti(ctx, entries); err != nil {
		return err
	}

	return nil
}

// recordCacheHit records a cache hit metric
func (s *cacheDomainService) recordCacheHit(namespace CacheNamespace, level string) {
	s.metricsCollector.UpdateMetrics(namespace, func(metrics *CacheMetrics) {
		metrics.CacheHits++
		metrics.TotalRequests++
		if level == "L1" {
			metrics.L1Hits++
		} else if level == "L2" {
			metrics.L2Hits++
		}
	})
}

// recordCacheMiss records a cache miss metric
func (s *cacheDomainService) recordCacheMiss(namespace CacheNamespace) {
	s.metricsCollector.UpdateMetrics(namespace, func(metrics *CacheMetrics) {
		metrics.CacheMisses++
		metrics.TotalRequests++
	})
}

// recordCacheSet records a cache set metric
func (s *cacheDomainService) recordCacheSet(namespace CacheNamespace) {
	s.metricsCollector.UpdateMetrics(namespace, func(metrics *CacheMetrics) {
		// Update set-related metrics
	})
}

// recordCacheDelete records a cache delete metric
func (s *cacheDomainService) recordCacheDelete(namespace CacheNamespace) {
	s.metricsCollector.UpdateMetrics(namespace, func(metrics *CacheMetrics) {
		// Update delete-related metrics
	})
}

// recordCacheClear records a cache clear metric
func (s *cacheDomainService) recordCacheClear(namespace CacheNamespace) {
	s.metricsCollector.UpdateMetrics(namespace, func(metrics *CacheMetrics) {
		// Update clear-related metrics
	})
}

// recordCacheSetMulti records a cache set multi metric
func (s *cacheDomainService) recordCacheSetMulti(namespace CacheNamespace, count int) {
	s.metricsCollector.UpdateMetrics(namespace, func(metrics *CacheMetrics) {
		// Update set multi-related metrics
	})
}

// recordCacheDeleteMulti records a cache delete multi metric
func (s *cacheDomainService) recordCacheDeleteMulti(namespace CacheNamespace, count int) {
	s.metricsCollector.UpdateMetrics(namespace, func(metrics *CacheMetrics) {
		// Update delete multi-related metrics
	})
}

// recordCachePromotion records a cache promotion metric
func (s *cacheDomainService) recordCachePromotion(namespace CacheNamespace, key string) {
	s.metricsCollector.UpdateMetrics(namespace, func(metrics *CacheMetrics) {
		// Update promotion-related metrics
	})
}

// recordCacheDemotion records a cache demotion metric
func (s *cacheDomainService) recordCacheDemotion(namespace CacheNamespace, key string) {
	s.metricsCollector.UpdateMetrics(namespace, func(metrics *CacheMetrics) {
		// Update demotion-related metrics
	})
}

// analyzeUsagePatterns analyzes cache usage patterns
func (s *cacheDomainService) analyzeUsagePatterns(stats *CacheStatistics) map[string]interface{} {
	analysis := make(map[string]interface{})

	// Analyze hit rate
	if stats.HitRate > 0.8 {
		analysis["hit_rate"] = "high"
	} else if stats.HitRate > 0.5 {
		analysis["hit_rate"] = "medium"
	} else {
		analysis["hit_rate"] = "low"
	}

	// Analyze access pattern
	if stats.ThroughputQPS > 1000 {
		analysis["access_pattern"] = "high_throughput"
	} else {
		analysis["access_pattern"] = "normal"
	}

	// Analyze memory usage
	if stats.MemoryUsage > s.config.MaxMemorySize*80/100 {
		analysis["memory_usage"] = "high"
	} else {
		analysis["memory_usage"] = "normal"
	}

	return analysis
}

// createOptimizedStrategy creates an optimized cache strategy
func (s *cacheDomainService) createOptimizedStrategy(namespace CacheNamespace, analysis map[string]interface{}) *CacheStrategy {
	strategy := s.strategyManager.GetStrategy(namespace)

	// Clone existing strategy
	optimized := &CacheStrategy{
		Name:               strategy.Name + "_optimized",
		Namespace:          namespace,
		DefaultTTL:         strategy.DefaultTTL,
		MaxSize:            strategy.MaxSize,
		EvictionPolicy:     strategy.EvictionPolicy,
		L1Enabled:          strategy.L1Enabled,
		L2Enabled:          strategy.L2Enabled,
		CompressionEnabled: strategy.CompressionEnabled,
		SerializationType:  strategy.SerializationType,
		AccessPattern:      strategy.AccessPattern,
		DataType:           strategy.DataType,
		Priorities:         strategy.Priorities,
		Rules:              strategy.Rules,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Apply optimizations based on analysis
	if analysis["hit_rate"] == "low" {
		optimized.DefaultTTL = optimized.DefaultTTL * 2 // Increase TTL
	}

	if analysis["memory_usage"] == "high" {
		optimized.CompressionEnabled = true
		optimized.MaxSize = optimized.MaxSize * 80 / 100 // Reduce max size
	}

	return optimized
}

// getHotKeys gets hot keys for a namespace
func (s *cacheDomainService) getHotKeys(ctx context.Context, namespace CacheNamespace) ([]string, error) {
	// Implementation would analyze access patterns to find hot keys
	return []string{}, nil
}

// getHotKeysFromL2 gets hot keys from L2 cache
func (s *cacheDomainService) getHotKeysFromL2(ctx context.Context, namespace CacheNamespace) ([]string, error) {
	// Implementation would analyze L2 cache access patterns
	return []string{}, nil
}

// getColdKeysFromL1 gets cold keys from L1 cache
func (s *cacheDomainService) getColdKeysFromL1(ctx context.Context, namespace CacheNamespace) ([]string, error) {
	// Implementation would analyze L1 cache access patterns
	return []string{}, nil
}

// determineRebalancingStrategy determines the rebalancing strategy
func (s *cacheDomainService) determineRebalancingStrategy(l1Stats, l2Stats map[string]interface{}) string {
	// Implementation would analyze cache statistics to determine strategy
	return "balanced"
}

// loadDataFromRepository loads data from repository
func (s *cacheDomainService) loadDataFromRepository(ctx context.Context, key string, namespace CacheNamespace) (interface{}, error) {
	cacheKey, err := NewCacheKey(namespace, key)
	if err != nil {
		return nil, err
	}

	return s.repository.Get(ctx, cacheKey)
}

// getKeysForWarming gets keys that need warming
func (s *cacheDomainService) getKeysForWarming(ctx context.Context, namespace CacheNamespace) ([]string, error) {
	// Implementation would determine which keys need warming
	return []string{}, nil
}

// isWarmingTime checks if it's time for warming
func (s *cacheDomainService) isWarmingTime() bool {
	// Implementation would check if it's an appropriate time for warming
	return false
}

// setLazyWarmingEnabled enables/disables lazy warming
func (s *cacheDomainService) setLazyWarmingEnabled(namespace CacheNamespace, enabled bool) {
	// Implementation would set lazy warming flag
}

// isWarmingNeeded checks if warming is needed for a namespace
func (s *cacheDomainService) isWarmingNeeded(ctx context.Context, namespace CacheNamespace) bool {
	// Implementation would check if warming is needed
	return false
}

// updateMetrics updates metrics for a namespace
func (s *cacheDomainService) updateMetrics(ctx context.Context, namespace CacheNamespace) {
	// Implementation would update metrics
}

// checkHealth checks health for a namespace
func (s *cacheDomainService) checkHealth(ctx context.Context, namespace CacheNamespace) {
	// Implementation would check health
}

// getLRUEntries gets least recently used entries
func (s *cacheDomainService) getLRUEntries(ctx context.Context, namespace CacheNamespace, limit int) ([]*CacheEntry, error) {
	// Implementation would get LRU entries
	return []*CacheEntry{}, nil
}

// getEntriesByTag gets entries by tag
func (s *cacheDomainService) getEntriesByTag(ctx context.Context, tag string, namespace CacheNamespace) ([]*CacheEntry, error) {
	// Implementation would get entries by tag
	return []*CacheEntry{}, nil
}

// removeDuplicateEntries removes duplicate entries
func (s *cacheDomainService) removeDuplicateEntries(entries []*CacheEntry) []*CacheEntry {
	seen := make(map[string]bool)
	result := make([]*CacheEntry, 0)

	for _, entry := range entries {
		key := entry.Key().String()
		if !seen[key] {
			seen[key] = true
			result = append(result, entry)
		}
	}

	return result
}

// checkDataStoreEntry checks if an entry exists in the data store
func (s *cacheDomainService) checkDataStoreEntry(ctx context.Context, key *CacheKey) (bool, error) {
	// Implementation would check data store
	return true, nil
}

// checkEntryUpToDate checks if an entry is up to date
func (s *cacheDomainService) checkEntryUpToDate(ctx context.Context, entry *CacheEntry) (bool, error) {
	// Implementation would check if entry is up to date
	return true, nil
}

// combineHealthStatus combines health statuses
func (s *cacheDomainService) combineHealthStatus(status1, status2 CacheHealthStatus) CacheHealthStatus {
	if status1 == CacheHealthStatusUnhealthy || status2 == CacheHealthStatusUnhealthy {
		return CacheHealthStatusUnhealthy
	}

	if status1 == CacheHealthStatusDegraded || status2 == CacheHealthStatusDegraded {
		return CacheHealthStatusDegraded
	}

	return CacheHealthStatusHealthy
}

// compressEntry compresses a cache entry
func (s *cacheDomainService) compressEntry(entry *CacheEntry) error {
	// Implementation would compress entry
	return nil
}

// serializeEntry serializes a cache entry
func (s *cacheDomainService) serializeEntry(entry *CacheEntry, serializationType string) error {
	// Implementation would serialize entry
	return nil
}

// applyRule applies a cache rule
func (s *cacheDomainService) applyRule(ctx context.Context, entry *CacheEntry, rule *CacheRule) error {
	// Implementation would apply rule
	return nil
}

// createDefaultStrategy creates a default strategy
func (s *cacheDomainService) createDefaultStrategy(namespace CacheNamespace, config *CacheConfiguration) *CacheStrategy {
	return &CacheStrategy{
		Name:               "default",
		Namespace:          namespace,
		DefaultTTL:         config.DefaultTTL,
		MaxSize:            config.MaxSize,
		EvictionPolicy:     EvictionPolicyLRU,
		L1Enabled:          true,
		L2Enabled:          true,
		CompressionEnabled: false,
		SerializationType:  "json",
		AccessPattern:      AccessPatternMixed,
		DataType:           DataTypeJSON,
		Priorities:         make(map[string]int),
		Rules:              make([]CacheRule, 0),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}

// RemoveStrategy removes a strategy from the manager
func (csm *CacheStrategyManager) RemoveStrategy(namespace CacheNamespace) {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	delete(csm.strategies, namespace)
}

// =============================================================================
// Service Lifecycle Management
// =============================================================================

// Start starts the cache domain service
func (s *cacheDomainService) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return errors.NewServiceAlreadyRunningError("cache domain service")
	}

	s.logger.Info("Starting cache domain service")

	// Initialize components
	s.initializeComponents()

	// Start background workers
	s.startBackgroundWorkers()

	s.isRunning = true
	s.logger.Info("Cache domain service started successfully")

	return nil
}

// Stop stops the cache domain service
func (s *cacheDomainService) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return nil
	}

	s.logger.Info("Stopping cache domain service")

	// Stop background workers
	close(s.stopCh)
	s.workersWG.Wait()

	// Cleanup resources
	s.cleanup()

	s.isRunning = false
	s.logger.Info("Cache domain service stopped successfully")

	return nil
}

// cleanup cleans up service resources
func (s *cacheDomainService) cleanup() {
	// Clear all caches
	if s.l1Cache != nil {
		s.l1Cache.Clear(context.Background())
	}

	if s.l2Cache != nil {
		s.l2Cache.Clear(context.Background())
	}

	// Clear registry
	if s.cacheRegistry != nil {
		s.cacheRegistry.Clear()
	}

	s.logger.Info("Cache domain service cleanup completed")
}

// IsRunning returns whether the service is running
func (s *cacheDomainService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.isRunning
}

// GetUptime returns the service uptime
func (s *cacheDomainService) GetUptime() time.Duration {
	return time.Since(s.startTime)
}

// =============================================================================
// Cache Configuration and Registry
// =============================================================================

// CacheConfiguration represents cache configuration
type CacheConfiguration struct {
	DefaultTTL         time.Duration
	MaxSize            int64
	MaxMemorySize      int64
	EvictionPolicy     EvictionPolicy
	CompressionEnabled bool
	SerializationType  string
	L1Enabled          bool
	L2Enabled          bool
	MonitoringEnabled  bool
	WarmingEnabled     bool
	Tags               []string
	Metadata           map[string]interface{}
}

// NamespaceConfiguration represents namespace configuration
type NamespaceConfiguration struct {
	Namespace CacheNamespace
	Config    *CacheConfiguration
	CreatedAt time.Time
	UpdatedAt time.Time
	IsActive  bool
}

// CacheRegistry manages cache namespaces
type CacheRegistry struct {
	namespaces map[CacheNamespace]*NamespaceConfiguration
	mu         sync.RWMutex
}

// NewCacheRegistry creates a new cache registry
func NewCacheRegistry() *CacheRegistry {
	return &CacheRegistry{
		namespaces: make(map[CacheNamespace]*NamespaceConfiguration),
	}
}

// RegisterNamespace registers a cache namespace
func (cr *CacheRegistry) RegisterNamespace(namespace CacheNamespace, config *NamespaceConfiguration) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if _, exists := cr.namespaces[namespace]; exists {
		return errors.NewNamespaceExistsError(string(namespace))
	}

	cr.namespaces[namespace] = config
	return nil
}

// UnregisterNamespace unregisters a cache namespace
func (cr *CacheRegistry) UnregisterNamespace(namespace CacheNamespace) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	delete(cr.namespaces, namespace)
}

// HasNamespace checks if a namespace exists
func (cr *CacheRegistry) HasNamespace(namespace CacheNamespace) bool {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	_, exists := cr.namespaces[namespace]
	return exists
}

// GetNamespace gets a namespace configuration
func (cr *CacheRegistry) GetNamespace(namespace CacheNamespace) (*NamespaceConfiguration, bool) {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	config, exists := cr.namespaces[namespace]
	return config, exists
}

// ListNamespaces lists all registered namespaces
func (cr *CacheRegistry) ListNamespaces() []CacheNamespace {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	namespaces := make([]CacheNamespace, 0, len(cr.namespaces))
	for namespace := range cr.namespaces {
		namespaces = append(namespaces, namespace)
	}

	return namespaces
}

// Clear clears all namespaces
func (cr *CacheRegistry) Clear() {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	cr.namespaces = make(map[CacheNamespace]*NamespaceConfiguration)
}

// =============================================================================
// Advanced Cache Operations
// =============================================================================

// GetOrSet gets a value from cache or sets it if not found
func (s *cacheDomainService) GetOrSet(ctx context.Context, key string, namespace CacheNamespace, valueFactory func() (interface{}, error), ttl time.Duration) (interface{}, error) {
	// Try to get from cache first
	value, err := s.Get(ctx, key, namespace)
	if err == nil {
		return value, nil
	}

	// If not found, use factory to create value
	if errors.IsCacheNotFoundError(err) {
		newValue, factoryErr := valueFactory()
		if factoryErr != nil {
			return nil, factoryErr
		}

		// Set in cache
		if setErr := s.Set(ctx, key, newValue, namespace, ttl); setErr != nil {
			s.logger.Error("Failed to set value in cache", "error", setErr, "key", key)
			// Return the value even if caching failed
		}

		return newValue, nil
	}

	return nil, err
}

// GetAndRefresh gets a value and refreshes its TTL
func (s *cacheDomainService) GetAndRefresh(ctx context.Context, key string, namespace CacheNamespace, ttl time.Duration) (interface{}, error) {
	// Get current value
	value, err := s.Get(ctx, key, namespace)
	if err != nil {
		return nil, err
	}

	// Refresh TTL
	if refreshErr := s.Refresh(ctx, key, namespace, ttl); refreshErr != nil {
		s.logger.Error("Failed to refresh cache entry", "error", refreshErr, "key", key)
		// Return the value even if refresh failed
	}

	return value, nil
}

// Increment increments a numeric value in cache
func (s *cacheDomainService) Increment(ctx context.Context, key string, namespace CacheNamespace, delta int64) (int64, error) {
	s.logger.Debug("Incrementing cache value", "key", key, "namespace", namespace, "delta", delta)

	// Validate input
	if err := s.validateKey(key); err != nil {
		return 0, err
	}

	// Try to increment in repository
	result, err := s.repository.Increment(ctx, key, namespace, delta)
	if err != nil {
		return 0, err
	}

	// Update cache levels
	cacheKey, keyErr := NewCacheKey(namespace, key)
	if keyErr != nil {
		return 0, keyErr
	}

	// Update L1 cache
	if err := s.l1Cache.Set(ctx, cacheKey.String(), result, s.config.DefaultTTL); err != nil {
		s.logger.Error("Failed to update L1 cache after increment", "error", err)
	}

	// Update L2 cache
	if err := s.l2Cache.Set(ctx, cacheKey.String(), result, s.config.DefaultTTL); err != nil {
		s.logger.Error("Failed to update L2 cache after increment", "error", err)
	}

	// Record metrics
	s.recordCacheIncrement(namespace, key, delta)

	return result, nil
}

// Decrement decrements a numeric value in cache
func (s *cacheDomainService) Decrement(ctx context.Context, key string, namespace CacheNamespace, delta int64) (int64, error) {
	return s.Increment(ctx, key, namespace, -delta)
}

// Touch updates the last access time of a cache entry
func (s *cacheDomainService) Touch(ctx context.Context, key string, namespace CacheNamespace) error {
	s.logger.Debug("Touching cache entry", "key", key, "namespace", namespace)

	// Validate input
	if err := s.validateKey(key); err != nil {
		return err
	}

	// Touch in repository
	if err := s.repository.Touch(ctx, key, namespace); err != nil {
		return err
	}

	// Record metrics
	s.recordCacheTouch(namespace, key)

	return nil
}

// Lock acquires a distributed lock
func (s *cacheDomainService) Lock(ctx context.Context, key string, namespace CacheNamespace, ttl time.Duration) (*CacheLock, error) {
	s.logger.Debug("Acquiring cache lock", "key", key, "namespace", namespace, "ttl", ttl)

	// Validate input
	if err := s.validateKey(key); err != nil {
		return nil, err
	}

	// Create lock key
	lockKey := "lock:" + key

	// Try to acquire lock
	lock, err := s.repository.AcquireLock(ctx, lockKey, namespace, ttl)
	if err != nil {
		return nil, err
	}

	// Record metrics
	s.recordCacheLock(namespace, key)

	return lock, nil
}

// Unlock releases a distributed lock
func (s *cacheDomainService) Unlock(ctx context.Context, lock *CacheLock) error {
	s.logger.Debug("Releasing cache lock", "key", lock.Key, "namespace", lock.Namespace)

	// Release lock
	if err := s.repository.ReleaseLock(ctx, lock); err != nil {
		return err
	}

	// Record metrics
	s.recordCacheUnlock(lock.Namespace, lock.Key)

	return nil
}

// =============================================================================
// Cache Analytics and Insights
// =============================================================================

// GetAnalytics returns cache analytics
func (s *cacheDomainService) GetAnalytics(ctx context.Context, namespace CacheNamespace, timeRange TimeRange) (*CacheAnalytics, error) {
	s.logger.Debug("Getting cache analytics", "namespace", namespace, "timeRange", timeRange)

	// Get analytics from repository
	analytics, err := s.repository.GetAnalytics(ctx, namespace, timeRange)
	if err != nil {
		return nil, err
	}

	// Enhance with local metrics
	s.enhanceAnalytics(analytics, namespace)

	return analytics, nil
}

// GetInsights returns cache insights and recommendations
func (s *cacheDomainService) GetInsights(ctx context.Context, namespace CacheNamespace) (*CacheInsights, error) {
	s.logger.Debug("Getting cache insights", "namespace", namespace)

	// Get current statistics
	stats, err := s.GetStatistics(ctx, namespace)
	if err != nil {
		return nil, err
	}

	// Get analytics
	analytics, err := s.GetAnalytics(ctx, namespace, TimeRangeLast24Hours)
	if err != nil {
		s.logger.Error("Failed to get analytics for insights", "error", err)
		analytics = &CacheAnalytics{} // Use empty analytics
	}

	// Generate insights
	insights := s.generateInsights(stats, analytics)

	return insights, nil
}

// GetTrends returns cache usage trends
func (s *cacheDomainService) GetTrends(ctx context.Context, namespace CacheNamespace, timeRange TimeRange) (*CacheTrends, error) {
	s.logger.Debug("Getting cache trends", "namespace", namespace, "timeRange", timeRange)

	// Get trends from repository
	trends, err := s.repository.GetTrends(ctx, namespace, timeRange)
	if err != nil {
		return nil, err
	}

	// Calculate trend indicators
	s.calculateTrendIndicators(trends)

	return trends, nil
}

// =============================================================================
// Cache Backup and Restore
// =============================================================================

// Backup creates a backup of cache data
func (s *cacheDomainService) Backup(ctx context.Context, namespace CacheNamespace, options *BackupOptions) (*BackupInfo, error) {
	s.logger.Info("Starting cache backup", "namespace", namespace, "options", options)

	// Create backup
	backupInfo, err := s.repository.Backup(ctx, namespace, options)
	if err != nil {
		return nil, err
	}

	// Publish backup event
	s.publishCacheEvent(ctx, &CacheEvent{
		Type:      CacheEventTypeBackup,
		Namespace: namespace,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"backup_id": backupInfo.ID,
			"size":      backupInfo.Size,
		},
	})

	s.logger.Info("Cache backup completed", "namespace", namespace, "backup_id", backupInfo.ID)
	return backupInfo, nil
}

// Restore restores cache data from backup
func (s *cacheDomainService) Restore(ctx context.Context, namespace CacheNamespace, backupID string, options *RestoreOptions) error {
	s.logger.Info("Starting cache restore", "namespace", namespace, "backup_id", backupID)

	// Clear existing data if requested
	if options.ClearExisting {
		if err := s.Clear(ctx, namespace); err != nil {
			return err
		}
	}

	// Restore from backup
	if err := s.repository.Restore(ctx, namespace, backupID, options); err != nil {
		return err
	}

	// Invalidate all cache levels
	if err := s.l1Cache.ClearNamespace(ctx, namespace); err != nil {
		s.logger.Error("Failed to clear L1 cache after restore", "error", err)
	}

	if err := s.l2Cache.ClearNamespace(ctx, namespace); err != nil {
		s.logger.Error("Failed to clear L2 cache after restore", "error", err)
	}

	// Publish restore event
	s.publishCacheEvent(ctx, &CacheEvent{
		Type:      CacheEventTypeRestore,
		Namespace: namespace,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"backup_id": backupID,
		},
	})

	s.logger.Info("Cache restore completed", "namespace", namespace, "backup_id", backupID)
	return nil
}

// ListBackups lists available backups
func (s *cacheDomainService) ListBackups(ctx context.Context, namespace CacheNamespace) ([]*BackupInfo, error) {
	return s.repository.ListBackups(ctx, namespace)
}

// DeleteBackup deletes a backup
func (s *cacheDomainService) DeleteBackup(ctx context.Context, namespace CacheNamespace, backupID string) error {
	s.logger.Info("Deleting cache backup", "namespace", namespace, "backup_id", backupID)

	if err := s.repository.DeleteBackup(ctx, namespace, backupID); err != nil {
		return err
	}

	// Publish backup deletion event
	s.publishCacheEvent(ctx, &CacheEvent{
		Type:      CacheEventTypeBackupDeleted,
		Namespace: namespace,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"backup_id": backupID,
		},
	})

	s.logger.Info("Cache backup deleted", "namespace", namespace, "backup_id", backupID)
	return nil
}

// =============================================================================
// Cache Debugging and Diagnostics
// =============================================================================

// Debug returns debug information for cache
func (s *cacheDomainService) Debug(ctx context.Context, namespace CacheNamespace) (*CacheDebugInfo, error) {
	s.logger.Debug("Getting cache debug info", "namespace", namespace)

	// Get debug info from repository
	debugInfo, err := s.repository.Debug(ctx, namespace)
	if err != nil {
		return nil, err
	}

	// Add service-level debug info
	debugInfo.ServiceInfo = &CacheServiceDebugInfo{
		Version:       s.version,
		StartTime:     s.startTime,
		Uptime:        s.GetUptime(),
		IsRunning:     s.IsRunning(),
		WorkerCount:   s.getWorkerCount(),
		Configuration: s.config,
	}

	// Add L1 cache debug info
	if s.l1Cache != nil {
		debugInfo.L1CacheInfo = s.l1Cache.GetStats()
	}

	// Add L2 cache debug info
	if s.l2Cache != nil {
		debugInfo.L2CacheInfo = s.l2Cache.GetStats()
	}

	return debugInfo, nil
}

// Trace enables or disables cache operation tracing
func (s *cacheDomainService) Trace(ctx context.Context, namespace CacheNamespace, enabled bool) error {
	s.logger.Info("Setting cache tracing", "namespace", namespace, "enabled", enabled)

	// Set tracing in repository
	if err := s.repository.SetTracing(ctx, namespace, enabled); err != nil {
		return err
	}

	// Update local tracing flag
	s.setTracing(namespace, enabled)

	return nil
}

// GetTraces returns recent cache operation traces
func (s *cacheDomainService) GetTraces(ctx context.Context, namespace CacheNamespace, limit int) ([]*CacheTrace, error) {
	return s.repository.GetTraces(ctx, namespace, limit)
}

// =============================================================================
// Cache Security and Access Control
// =============================================================================

// ValidateAccess validates access to cache operations
func (s *cacheDomainService) ValidateAccess(ctx context.Context, namespace CacheNamespace, operation CacheOperation, principal *CachePrincipal) error {
	if s.accessController == nil {
		return nil // No access control configured
	}

	return s.accessController.ValidateAccess(ctx, namespace, operation, principal)
}

// GrantAccess grants access to cache operations
func (s *cacheDomainService) GrantAccess(ctx context.Context, namespace CacheNamespace, operation CacheOperation, principal *CachePrincipal) error {
	if s.accessController == nil {
		return errors.NewAccessControlNotConfiguredError()
	}

	return s.accessController.GrantAccess(ctx, namespace, operation, principal)
}

// RevokeAccess revokes access to cache operations
func (s *cacheDomainService) RevokeAccess(ctx context.Context, namespace CacheNamespace, operation CacheOperation, principal *CachePrincipal) error {
	if s.accessController == nil {
		return errors.NewAccessControlNotConfiguredError()
	}

	return s.accessController.RevokeAccess(ctx, namespace, operation, principal)
}

// =============================================================================
// Cache Integration and Middleware
// =============================================================================

// AddMiddleware adds middleware to cache operations
func (s *cacheDomainService) AddMiddleware(middleware CacheMiddleware) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.middlewares = append(s.middlewares, middleware)
}

// RemoveMiddleware removes middleware from cache operations
func (s *cacheDomainService) RemoveMiddleware(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, middleware := range s.middlewares {
		if middleware.Name() == name {
			s.middlewares = append(s.middlewares[:i], s.middlewares[i+1:]...)
			break
		}
	}
}

// executeMiddleware executes middleware chain
func (s *cacheDomainService) executeMiddleware(ctx context.Context, operation CacheOperation, next func(context.Context) error) error {
	if len(s.middlewares) == 0 {
		return next(ctx)
	}

	// Create middleware chain
	chain := func(ctx context.Context) error {
		return next(ctx)
	}

	// Apply middleware in reverse order
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		middleware := s.middlewares[i]
		currentChain := chain
		chain = func(ctx context.Context) error {
			return middleware.Execute(ctx, operation, currentChain)
		}
	}

	return chain(ctx)
}

// =============================================================================
// Cache Metrics and Monitoring Extensions
// =============================================================================

// EnableMetrics enables metrics collection for a namespace
func (s *cacheDomainService) EnableMetrics(ctx context.Context, namespace CacheNamespace) error {
	s.logger.Info("Enabling metrics for namespace", "namespace", namespace)

	// Enable metrics in repository
	if err := s.repository.EnableMetrics(ctx, namespace); err != nil {
		return err
	}

	// Initialize metrics in collector
	s.metricsCollector.UpdateMetrics(namespace, func(metrics *CacheMetrics) {
		// Initialize metrics
	})

	return nil
}

// DisableMetrics disables metrics collection for a namespace
func (s *cacheDomainService) DisableMetrics(ctx context.Context, namespace CacheNamespace) error {
	s.logger.Info("Disabling metrics for namespace", "namespace", namespace)

	// Disable metrics in repository
	if err := s.repository.DisableMetrics(ctx, namespace); err != nil {
		return err
	}

	return nil
}

// ExportMetrics exports metrics in specified format
func (s *cacheDomainService) ExportMetrics(ctx context.Context, namespace CacheNamespace, format string) ([]byte, error) {
	metrics, err := s.GetMetrics(ctx, namespace)
	if err != nil {
		return nil, err
	}

	return s.metricsExporter.Export(metrics, format)
}

// =============================================================================
// Cache Partition Management
// =============================================================================

// CreatePartition creates a new cache partition
func (s *cacheDomainService) CreatePartition(ctx context.Context, namespace CacheNamespace, partitionID string, config *PartitionConfig) error {
	s.logger.Info("Creating cache partition", "namespace", namespace, "partition", partitionID)

	// Create partition in repository
	if err := s.repository.CreatePartition(ctx, namespace, partitionID, config); err != nil {
		return err
	}

	// Register partition
	if err := s.partitionManager.RegisterPartition(namespace, partitionID, config); err != nil {
		return err
	}

	// Publish partition creation event
	s.publishCacheEvent(ctx, &CacheEvent{
		Type:      CacheEventTypePartitionCreated,
		Namespace: namespace,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"partition_id": partitionID,
			"config":       config,
		},
	})

	s.logger.Info("Cache partition created", "namespace", namespace, "partition", partitionID)
	return nil
}

// DeletePartition deletes a cache partition
func (s *cacheDomainService) DeletePartition(ctx context.Context, namespace CacheNamespace, partitionID string) error {
	s.logger.Info("Deleting cache partition", "namespace", namespace, "partition", partitionID)

	// Delete partition from repository
	if err := s.repository.DeletePartition(ctx, namespace, partitionID); err != nil {
		return err
	}

	// Unregister partition
	s.partitionManager.UnregisterPartition(namespace, partitionID)

	// Publish partition deletion event
	s.publishCacheEvent(ctx, &CacheEvent{
		Type:      CacheEventTypePartitionDeleted,
		Namespace: namespace,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"partition_id": partitionID,
		},
	})

	s.logger.Info("Cache partition deleted", "namespace", namespace, "partition", partitionID)
	return nil
}

// ListPartitions lists all partitions for a namespace
func (s *cacheDomainService) ListPartitions(ctx context.Context, namespace CacheNamespace) ([]string, error) {
	return s.partitionManager.ListPartitions(namespace), nil
}

// =============================================================================
// Helper Methods for New Features
// =============================================================================

// recordCacheIncrement records cache increment metrics
func (s *cacheDomainService) recordCacheIncrement(namespace CacheNamespace, key string, delta int64) {
	s.metricsCollector.UpdateMetrics(namespace, func(metrics *CacheMetrics) {
		// Update increment-related metrics
	})
}

// recordCacheTouch records cache touch metrics
func (s *cacheDomainService) recordCacheTouch(namespace CacheNamespace, key string) {
	s.metricsCollector.UpdateMetrics(namespace, func(metrics *CacheMetrics) {
		// Update touch-related metrics
	})
}

// recordCacheLock records cache lock metrics
func (s *cacheDomainService) recordCacheLock(namespace CacheNamespace, key string) {
	s.metricsCollector.UpdateMetrics(namespace, func(metrics *CacheMetrics) {
		// Update lock-related metrics
	})
}

// recordCacheUnlock records cache unlock metrics
func (s *cacheDomainService) recordCacheUnlock(namespace CacheNamespace, key string) {
	s.metricsCollector.UpdateMetrics(namespace, func(metrics *CacheMetrics) {
		// Update unlock-related metrics
	})
}

// enhanceAnalytics enhances analytics with additional data
func (s *cacheDomainService) enhanceAnalytics(analytics *CacheAnalytics, namespace CacheNamespace) {
	// Add service-level analytics
	metrics := s.metricsCollector.GetMetrics(namespace)
	if metrics != nil {
		analytics.ServiceMetrics = metrics
	}
}

// generateInsights generates cache insights from statistics and analytics
func (s *cacheDomainService) generateInsights(stats *CacheStatistics, analytics *CacheAnalytics) *CacheInsights {
	insights := &CacheInsights{
		Namespace:        stats.Namespace,
		GeneratedAt:      time.Now(),
		Recommendations:  make([]CacheRecommendation, 0),
		PerformanceScore: s.calculatePerformanceScore(stats),
		EfficiencyScore:  s.calculateEfficiencyScore(stats),
	}

	// Generate recommendations
	recommendations := s.generateRecommendations(stats, analytics)
	insights.Recommendations = recommendations

	return insights
}

// calculateTrendIndicators calculates trend indicators
func (s *cacheDomainService) calculateTrendIndicators(trends *CacheTrends) {
	// Calculate trend indicators like growth rate, volatility, etc.
	if len(trends.DataPoints) >= 2 {
		firstPoint := trends.DataPoints[0]
		lastPoint := trends.DataPoints[len(trends.DataPoints)-1]

		// Calculate growth rate
		if firstPoint.Value > 0 {
			trends.GrowthRate = (lastPoint.Value - firstPoint.Value) / firstPoint.Value * 100
		}

		// Calculate volatility
		trends.Volatility = s.calculateVolatility(trends.DataPoints)
	}
}

// calculatePerformanceScore calculates performance score
func (s *cacheDomainService) calculatePerformanceScore(stats *CacheStatistics) float64 {
	// Calculate performance score based on hit rate, latency, etc.
	score := 0.0

	// Hit rate contributes 40%
	score += stats.HitRate * 40

	// Latency contributes 30% (lower is better)
	if stats.AverageLatency > 0 {
		latencyScore := math.Max(0, 30-float64(stats.AverageLatency.Milliseconds())/10)
		score += latencyScore
	}

	// Throughput contributes 30%
	if stats.ThroughputQPS > 0 {
		throughputScore := math.Min(30, stats.ThroughputQPS/100)
		score += throughputScore
	}

	return math.Min(100, score)
}

// calculateEfficiencyScore calculates efficiency score
func (s *cacheDomainService) calculateEfficiencyScore(stats *CacheStatistics) float64 {
	// Calculate efficiency score based on memory usage, eviction rate, etc.
	score := 100.0

	// Memory usage penalty
	if stats.MemoryUsage > 0 {
		memoryUsageRatio := float64(stats.MemoryUsage) / float64(s.config.MaxMemorySize)
		if memoryUsageRatio > 0.8 {
			score -= (memoryUsageRatio - 0.8) * 50
		}
	}

	// Eviction rate penalty
	if stats.EvictionCount > 0 && stats.TotalRequests > 0 {
		evictionRate := float64(stats.EvictionCount) / float64(stats.TotalRequests)
		score -= evictionRate * 30
	}

	return math.Max(0, score)
}

// generateRecommendations generates cache recommendations
func (s *cacheDomainService) generateRecommendations(stats *CacheStatistics, analytics *CacheAnalytics) []CacheRecommendation {
	recommendations := make([]CacheRecommendation, 0)

	// Low hit rate recommendation
	if stats.HitRate < 0.5 {
		recommendations = append(recommendations, CacheRecommendation{
			Type:        RecommendationTypeConfiguration,
			Priority:    RecommendationPriorityHigh,
			Title:       "Improve Cache Hit Rate",
			Description: "Consider increasing TTL or adjusting cache size",
			Impact:      "High",
		})
	}

	// High memory usage recommendation
	if stats.MemoryUsage > 0 && float64(stats.MemoryUsage)/float64(s.config.MaxMemorySize) > 0.8 {
		recommendations = append(recommendations, CacheRecommendation{
			Type:        RecommendationTypeOptimization,
			Priority:    RecommendationPriorityMedium,
			Title:       "Optimize Memory Usage",
			Description: "Enable compression or reduce cache size",
			Impact:      "Medium",
		})
	}

	// High eviction rate recommendation
	if stats.EvictionCount > 0 && stats.TotalRequests > 0 {
		evictionRate := float64(stats.EvictionCount) / float64(stats.TotalRequests)
		if evictionRate > 0.1 {
			recommendations = append(recommendations, CacheRecommendation{
				Type:        RecommendationTypeConfiguration,
				Priority:    RecommendationPriorityMedium,
				Title:       "Reduce Eviction Rate",
				Description: "Consider increasing cache size or adjusting eviction policy",
				Impact:      "Medium",
			})
		}
	}

	return recommendations
}

// calculateVolatility calculates volatility of trend data
func (s *cacheDomainService) calculateVolatility(dataPoints []TrendDataPoint) float64 {
	if len(dataPoints) < 2 {
		return 0
	}

	// Calculate standard deviation
	mean := 0.0
	for _, point := range dataPoints {
		mean += point.Value
	}
	mean /= float64(len(dataPoints))

	variance := 0.0
	for _, point := range dataPoints {
		variance += math.Pow(point.Value-mean, 2)
	}
	variance /= float64(len(dataPoints))

	return math.Sqrt(variance)
}

// getWorkerCount returns the number of active workers
func (s *cacheDomainService) getWorkerCount() int {
	return 4 // cleanup, monitor, warming, event processor
}

// setTracing sets tracing flag for a namespace
func (s *cacheDomainService) setTracing(namespace CacheNamespace, enabled bool) {
	// Implementation would set tracing flag
}

// =============================================================================
// Cache Service Factory
// =============================================================================

// CacheServiceFactory creates cache service instances
type CacheServiceFactory struct {
	config *CacheDomainServiceConfig
	logger Logger
}

// NewCacheServiceFactory creates a new cache service factory
func NewCacheServiceFactory(config *CacheDomainServiceConfig, logger Logger) *CacheServiceFactory {
	return &CacheServiceFactory{
		config: config,
		logger: logger,
	}
}

// CreateService creates a new cache service instance
func (f *CacheServiceFactory) CreateService(repository CacheRepository, eventBus EventBus) CacheDomainService {
	return NewCacheDomainService(f.config, repository, eventBus, f.logger)
}

// =============================================================================
// Cache Service Builder
// =============================================================================

// CacheServiceBuilder builds cache service with fluent interface
type CacheServiceBuilder struct {
	config     *CacheDomainServiceConfig
	repository CacheRepository
	eventBus   EventBus
	logger     Logger
}

// NewCacheServiceBuilder creates a new cache service builder
func NewCacheServiceBuilder() *CacheServiceBuilder {
	return &CacheServiceBuilder{
		config: &CacheDomainServiceConfig{
			DefaultTTL:                15 * time.Minute,
			MaxMemorySize:             1 << 30, // 1GB
			MaxCacheSize:              1 << 20, // 1M entries
			CleanupInterval:           5 * time.Minute,
			MetricsCollectionInterval: 30 * time.Second,
			PromotionThreshold:        10,
			EnableAutoOptimization:    true,
			EventProcessingDelay:      100 * time.Millisecond,
			MaxConcurrentOperations:   1000,
		},
	}
}

// WithConfig sets the service configuration
func (b *CacheServiceBuilder) WithConfig(config *CacheDomainServiceConfig) *CacheServiceBuilder {
	b.config = config
	return b
}

// WithRepository sets the cache repository
func (b *CacheServiceBuilder) WithRepository(repository CacheRepository) *CacheServiceBuilder {
	b.repository = repository
	return b
}

// WithEventBus sets the event bus
func (b *CacheServiceBuilder) WithEventBus(eventBus EventBus) *CacheServiceBuilder {
	b.eventBus = eventBus
	return b
}

// WithLogger sets the logger
func (b *CacheServiceBuilder) WithLogger(logger Logger) *CacheServiceBuilder {
	b.logger = logger
	return b
}

// Build builds the cache service
func (b *CacheServiceBuilder) Build() CacheDomainService {
	if b.repository == nil {
		panic("repository is required")
	}

	if b.eventBus == nil {
		panic("event bus is required")
	}

	if b.logger == nil {
		panic("logger is required")
	}

	return NewCacheDomainService(b.config, b.repository, b.eventBus, b.logger)
}

// =============================================================================
// Constants and Enums
// =============================================================================

// Cache event types
const (
	CacheEventTypeSet              CacheEventType = "cache.set"
	CacheEventTypeGet              CacheEventType = "cache.get"
	CacheEventTypeDelete           CacheEventType = "cache.delete"
	CacheEventTypeClear            CacheEventType = "cache.clear"
	CacheEventTypeExpired          CacheEventType = "cache.expired"
	CacheEventTypeEvicted          CacheEventType = "cache.evicted"
	CacheEventTypeHit              CacheEventType = "cache.hit"
	CacheEventTypeMiss             CacheEventType = "cache.miss"
	CacheEventTypePromoted         CacheEventType = "cache.promoted"
	CacheEventTypeDemoted          CacheEventType = "cache.demoted"
	CacheEventTypeOptimized        CacheEventType = "cache.optimized"
	CacheEventTypeWarmed           CacheEventType = "cache.warmed"
	CacheEventTypeRebalanced       CacheEventType = "cache.rebalanced"
	CacheEventTypeNamespaceCreated CacheEventType = "cache.namespace.created"
	CacheEventTypeNamespaceDeleted CacheEventType = "cache.namespace.deleted"
	CacheEventTypePartitionCreated CacheEventType = "cache.partition.created"
	CacheEventTypePartitionDeleted CacheEventType = "cache.partition.deleted"
	CacheEventTypeBackup           CacheEventType = "cache.backup"
	CacheEventTypeRestore          CacheEventType = "cache.restore"
	CacheEventTypeBackupDeleted    CacheEventType = "cache.backup.deleted"
	CacheEventTypeHealthCheck      CacheEventType = "cache.health.check"
	CacheEventTypeError            CacheEventType = "cache.error"
)

// Cache operations
const (
	CacheOperationGet      CacheOperation = "get"
	CacheOperationSet      CacheOperation = "set"
	CacheOperationDelete   CacheOperation = "delete"
	CacheOperationClear    CacheOperation = "clear"
	CacheOperationList     CacheOperation = "list"
	CacheOperationStats    CacheOperation = "stats"
	CacheOperationOptimize CacheOperation = "optimize"
	CacheOperationBackup   CacheOperation = "backup"
	CacheOperationRestore  CacheOperation = "restore"
	CacheOperationAdmin    CacheOperation = "admin"
)

// Health status
const (
	CacheHealthStatusHealthy   CacheHealthStatus = "healthy"
	CacheHealthStatusDegraded  CacheHealthStatus = "degraded"
	CacheHealthStatusUnhealthy CacheHealthStatus = "unhealthy"
)

// Eviction policies
const (
	EvictionPolicyLRU    EvictionPolicy = "lru"
	EvictionPolicyLFU    EvictionPolicy = "lfu"
	EvictionPolicyFIFO   EvictionPolicy = "fifo"
	EvictionPolicyRandom EvictionPolicy = "random"
	EvictionPolicyTTL    EvictionPolicy = "ttl"
	EvictionPolicySize   EvictionPolicy = "size"
	EvictionPolicyCustom EvictionPolicy = "custom"
)

// Access patterns
const (
	AccessPatternReadHeavy  AccessPattern = "read_heavy"
	AccessPatternWriteHeavy AccessPattern = "write_heavy"
	AccessPatternMixed      AccessPattern = "mixed"
	AccessPatternBursty     AccessPattern = "bursty"
	AccessPatternSteady     AccessPattern = "steady"
)

// Data types
const (
	DataTypeString  DataType = "string"
	DataTypeJSON    DataType = "json"
	DataTypeBinary  DataType = "binary"
	DataTypeNumber  DataType = "number"
	DataTypeBoolean DataType = "boolean"
	DataTypeList    DataType = "list"
	DataTypeSet     DataType = "set"
	DataTypeHash    DataType = "hash"
)

// Cache warming strategies
const (
	CacheWarmingStrategyEager      CacheWarmingStrategy = "eager"
	CacheWarmingStrategyLazy       CacheWarmingStrategy = "lazy"
	CacheWarmingStrategyBackground CacheWarmingStrategy = "background"
	CacheWarmingStrategyScheduled  CacheWarmingStrategy = "scheduled"
)

// Time ranges
const (
	TimeRangeLast5Minutes  TimeRange = "last_5_minutes"
	TimeRangeLast15Minutes TimeRange = "last_15_minutes"
	TimeRangeLast30Minutes TimeRange = "last_30_minutes"
	TimeRangeLast1Hour     TimeRange = "last_1_hour"
	TimeRangeLast6Hours    TimeRange = "last_6_hours"
	TimeRangeLast12Hours   TimeRange = "last_12_hours"
	TimeRangeLast24Hours   TimeRange = "last_24_hours"
	TimeRangeLast7Days     TimeRange = "last_7_days"
	TimeRangeLast30Days    TimeRange = "last_30_days"
	TimeRangeCustom        TimeRange = "custom"
)

// Recommendation types
const (
	RecommendationTypeConfiguration RecommendationType = "configuration"
	RecommendationTypeOptimization  RecommendationType = "optimization"
	RecommendationTypePerformance   RecommendationType = "performance"
	RecommendationTypeSecurity      RecommendationType = "security"
	RecommendationTypeMaintenance   RecommendationType = "maintenance"
)

// Recommendation priorities
const (
	RecommendationPriorityLow      RecommendationPriority = "low"
	RecommendationPriorityMedium   RecommendationPriority = "medium"
	RecommendationPriorityHigh     RecommendationPriority = "high"
	RecommendationPriorityCritical RecommendationPriority = "critical"
)

// =============================================================================
// Type Definitions
// =============================================================================

// Basic types
type (
	CacheEventType         string
	CacheOperation         string
	CacheHealthStatus      string
	EvictionPolicy         string
	AccessPattern          string
	DataType               string
	CacheWarmingStrategy   string
	TimeRange              string
	RecommendationType     string
	RecommendationPriority string
)

// CacheNamespace represents a cache namespace
type CacheNamespace string

// String returns the string representation
func (cn CacheNamespace) String() string {
	return string(cn)
}

// IsValid checks if the namespace is valid
func (cn CacheNamespace) IsValid() bool {
	return len(cn) > 0 && len(cn) <= 64
}

// =============================================================================
// Data Structures
// =============================================================================

// CacheKey represents a cache key
type CacheKey struct {
	namespace CacheNamespace
	key       string
	tags      []string
	metadata  map[string]interface{}
}

// NewCacheKey creates a new cache key
func NewCacheKey(namespace CacheNamespace, key string) (*CacheKey, error) {
	if !namespace.IsValid() {
		return nil, errors.NewInvalidNamespaceError(string(namespace))
	}

	if key == "" {
		return nil, errors.NewInvalidKeyError("key cannot be empty")
	}

	return &CacheKey{
		namespace: namespace,
		key:       key,
		tags:      make([]string, 0),
		metadata:  make(map[string]interface{}),
	}, nil
}

// Namespace returns the namespace
func (ck *CacheKey) Namespace() CacheNamespace {
	return ck.namespace
}

// Key returns the key
func (ck *CacheKey) Key() string {
	return ck.key
}

// Tags returns the tags
func (ck *CacheKey) Tags() []string {
	return ck.tags
}

// Metadata returns the metadata
func (ck *CacheKey) Metadata() map[string]interface{} {
	return ck.metadata
}

// String returns the string representation
func (ck *CacheKey) String() string {
	return fmt.Sprintf("%s:%s", ck.namespace, ck.key)
}

// WithTags adds tags to the key
func (ck *CacheKey) WithTags(tags ...string) *CacheKey {
	ck.tags = append(ck.tags, tags...)
	return ck
}

// WithMetadata adds metadata to the key
func (ck *CacheKey) WithMetadata(metadata map[string]interface{}) *CacheKey {
	for k, v := range metadata {
		ck.metadata[k] = v
	}
	return ck
}

// CacheEntry represents a cache entry
type CacheEntry struct {
	key         *CacheKey
	value       interface{}
	ttl         time.Duration
	createdAt   time.Time
	updatedAt   time.Time
	accessedAt  time.Time
	accessCount int64
	size        int64
	compressed  bool
	encrypted   bool
	metadata    map[string]interface{}
}

// NewCacheEntry creates a new cache entry
func NewCacheEntry(key *CacheKey, value interface{}, ttl time.Duration) *CacheEntry {
	now := time.Now()
	return &CacheEntry{
		key:         key,
		value:       value,
		ttl:         ttl,
		createdAt:   now,
		updatedAt:   now,
		accessedAt:  now,
		accessCount: 0,
		size:        0,
		compressed:  false,
		encrypted:   false,
		metadata:    make(map[string]interface{}),
	}
}

// Key returns the cache key
func (ce *CacheEntry) Key() *CacheKey {
	return ce.key
}

// Value returns the value
func (ce *CacheEntry) Value() interface{} {
	return ce.value
}

// TTL returns the TTL
func (ce *CacheEntry) TTL() time.Duration {
	return ce.ttl
}

// CreatedAt returns the creation time
func (ce *CacheEntry) CreatedAt() time.Time {
	return ce.createdAt
}

// UpdatedAt returns the last update time
func (ce *CacheEntry) UpdatedAt() time.Time {
	return ce.updatedAt
}

// AccessedAt returns the last access time
func (ce *CacheEntry) AccessedAt() time.Time {
	return ce.accessedAt
}

// AccessCount returns the access count
func (ce *CacheEntry) AccessCount() int64 {
	return ce.accessCount
}

// Size returns the size
func (ce *CacheEntry) Size() int64 {
	return ce.size
}

// IsCompressed returns whether the entry is compressed
func (ce *CacheEntry) IsCompressed() bool {
	return ce.compressed
}

// IsEncrypted returns whether the entry is encrypted
func (ce *CacheEntry) IsEncrypted() bool {
	return ce.encrypted
}

// Metadata returns the metadata
func (ce *CacheEntry) Metadata() map[string]interface{} {
	return ce.metadata
}

// IsExpired checks if the entry is expired
func (ce *CacheEntry) IsExpired() bool {
	if ce.ttl <= 0 {
		return false
	}
	return time.Since(ce.createdAt) > ce.ttl
}

// ExpiresAt returns the expiration time
func (ce *CacheEntry) ExpiresAt() time.Time {
	if ce.ttl <= 0 {
		return time.Time{}
	}
	return ce.createdAt.Add(ce.ttl)
}

// TouchAccess updates the access time and count
func (ce *CacheEntry) TouchAccess() {
	ce.accessedAt = time.Now()
	ce.accessCount++
}

// UpdateValue updates the value and timestamp
func (ce *CacheEntry) UpdateValue(value interface{}) {
	ce.value = value
	ce.updatedAt = time.Now()
}

// CacheEvent represents a cache event
type CacheEvent struct {
	Type      CacheEventType         `json:"type"`
	Namespace CacheNamespace         `json:"namespace"`
	Key       string                 `json:"key,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Duration  time.Duration          `json:"duration,omitempty"`
}

// CacheEventHandler handles cache events
type CacheEventHandler func(ctx context.Context, event *CacheEvent) error

// CacheStrategy represents a cache strategy
type CacheStrategy struct {
	Name               string         `json:"name"`
	Namespace          CacheNamespace `json:"namespace"`
	DefaultTTL         time.Duration  `json:"default_ttl"`
	MaxSize            int64          `json:"max_size"`
	EvictionPolicy     EvictionPolicy `json:"eviction_policy"`
	L1Enabled          bool           `json:"l1_enabled"`
	L2Enabled          bool           `json:"l2_enabled"`
	CompressionEnabled bool           `json:"compression_enabled"`
	SerializationType  string         `json:"serialization_type"`
	AccessPattern      AccessPattern  `json:"access_pattern"`
	DataType           DataType       `json:"data_type"`
	Priorities         map[string]int `json:"priorities"`
	Rules              []CacheRule    `json:"rules"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

// CacheRule represents a cache rule
type CacheRule struct {
	Name       string                 `json:"name"`
	Condition  string                 `json:"condition"`
	Action     string                 `json:"action"`
	Parameters map[string]interface{} `json:"parameters"`
	Enabled    bool                   `json:"enabled"`
	Priority   int                    `json:"priority"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// CacheStatistics represents cache statistics
type CacheStatistics struct {
	Namespace       CacheNamespace `json:"namespace"`
	TotalEntries    int64          `json:"total_entries"`
	TotalSize       int64          `json:"total_size"`
	MemoryUsage     int64          `json:"memory_usage"`
	HitRate         float64        `json:"hit_rate"`
	MissRate        float64        `json:"miss_rate"`
	TotalRequests   int64          `json:"total_requests"`
	CacheHits       int64          `json:"cache_hits"`
	CacheMisses     int64          `json:"cache_misses"`
	L1Hits          int64          `json:"l1_hits"`
	L2Hits          int64          `json:"l2_hits"`
	EvictionCount   int64          `json:"eviction_count"`
	ExpirationCount int64          `json:"expiration_count"`
	AverageLatency  time.Duration  `json:"average_latency"`
	MaxLatency      time.Duration  `json:"max_latency"`
	MinLatency      time.Duration  `json:"min_latency"`
	ThroughputQPS   float64        `json:"throughput_qps"`
	ErrorCount      int64          `json:"error_count"`
	LastUpdated     time.Time      `json:"last_updated"`
}

// CacheMetrics represents cache metrics
type CacheMetrics struct {
	Namespace     CacheNamespace `json:"namespace"`
	TotalRequests int64          `json:"total_requests"`
	CacheHits     int64          `json:"cache_hits"`
	CacheMisses   int64          `json:"cache_misses"`
	L1Hits        int64          `json:"l1_hits"`
	L2Hits        int64          `json:"l2_hits"`
	Sets          int64          `json:"sets"`
	Gets          int64          `json:"gets"`
	Deletes       int64          `json:"deletes"`
	Evictions     int64          `json:"evictions"`
	Expirations   int64          `json:"expirations"`
	Errors        int64          `json:"errors"`
	TotalLatency  time.Duration  `json:"total_latency"`
	RequestCount  int64          `json:"request_count"`
	LastUpdated   time.Time      `json:"last_updated"`
}

// CacheHealth represents cache health information
type CacheHealth struct {
	Namespace        CacheNamespace    `json:"namespace"`
	Status           CacheHealthStatus `json:"status"`
	L1CacheHealth    CacheHealthStatus `json:"l1_cache_health"`
	L2CacheHealth    CacheHealthStatus `json:"l2_cache_health"`
	RepositoryHealth CacheHealthStatus `json:"repository_health"`
	Uptime           time.Duration     `json:"uptime"`
	LastCheck        time.Time         `json:"last_check"`
	Issues           []string          `json:"issues,omitempty"`
	Recommendations  []string          `json:"recommendations,omitempty"`
}

// CacheLock represents a distributed cache lock
type CacheLock struct {
	Key       string         `json:"key"`
	Namespace CacheNamespace `json:"namespace"`
	Token     string         `json:"token"`
	TTL       time.Duration  `json:"ttl"`
	CreatedAt time.Time      `json:"created_at"`
	ExpiresAt time.Time      `json:"expires_at"`
}

// IsExpired checks if the lock is expired
func (cl *CacheLock) IsExpired() bool {
	return time.Now().After(cl.ExpiresAt)
}

// TimeLeft returns the time left for the lock
func (cl *CacheLock) TimeLeft() time.Duration {
	if cl.IsExpired() {
		return 0
	}
	return time.Until(cl.ExpiresAt)
}

// CacheAnalytics represents cache analytics
type CacheAnalytics struct {
	Namespace      CacheNamespace         `json:"namespace"`
	TimeRange      TimeRange              `json:"time_range"`
	StartTime      time.Time              `json:"start_time"`
	EndTime        time.Time              `json:"end_time"`
	TotalRequests  int64                  `json:"total_requests"`
	AverageHitRate float64                `json:"average_hit_rate"`
	PeakQPS        float64                `json:"peak_qps"`
	AverageLatency time.Duration          `json:"average_latency"`
	TopKeys        []string               `json:"top_keys"`
	TopErrorTypes  []string               `json:"top_error_types"`
	ServiceMetrics *CacheMetrics          `json:"service_metrics,omitempty"`
	CustomMetrics  map[string]interface{} `json:"custom_metrics,omitempty"`
	GeneratedAt    time.Time              `json:"generated_at"`
}

// CacheInsights represents cache insights
type CacheInsights struct {
	Namespace        CacheNamespace        `json:"namespace"`
	GeneratedAt      time.Time             `json:"generated_at"`
	PerformanceScore float64               `json:"performance_score"`
	EfficiencyScore  float64               `json:"efficiency_score"`
	Recommendations  []CacheRecommendation `json:"recommendations"`
	Trends           []string              `json:"trends"`
	Anomalies        []string              `json:"anomalies"`
}

// CacheRecommendation represents a cache recommendation
type CacheRecommendation struct {
	Type        RecommendationType     `json:"type"`
	Priority    RecommendationPriority `json:"priority"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Impact      string                 `json:"impact"`
	Action      string                 `json:"action,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CacheTrends represents cache trends
type CacheTrends struct {
	Namespace   CacheNamespace   `json:"namespace"`
	TimeRange   TimeRange        `json:"time_range"`
	DataPoints  []TrendDataPoint `json:"data_points"`
	GrowthRate  float64          `json:"growth_rate"`
	Volatility  float64          `json:"volatility"`
	Seasonality string           `json:"seasonality"`
	Forecast    []TrendDataPoint `json:"forecast"`
	GeneratedAt time.Time        `json:"generated_at"`
}

// TrendDataPoint represents a trend data point
type TrendDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Metric    string    `json:"metric"`
}

// BackupOptions represents backup options
type BackupOptions struct {
	IncludeMetadata bool              `json:"include_metadata"`
	Compression     bool              `json:"compression"`
	Encryption      bool              `json:"encryption"`
	Format          string            `json:"format"`
	Filters         map[string]string `json:"filters"`
	MaxSize         int64             `json:"max_size"`
}

// RestoreOptions represents restore options
type RestoreOptions struct {
	ClearExisting   bool              `json:"clear_existing"`
	OverwritePolicy string            `json:"overwrite_policy"`
	Filters         map[string]string `json:"filters"`
	DryRun          bool              `json:"dry_run"`
}

// BackupInfo represents backup information
type BackupInfo struct {
	ID         string                 `json:"id"`
	Namespace  CacheNamespace         `json:"namespace"`
	Size       int64                  `json:"size"`
	EntryCount int64                  `json:"entry_count"`
	CreatedAt  time.Time              `json:"created_at"`
	ExpiresAt  time.Time              `json:"expires_at"`
	Format     string                 `json:"format"`
	Compressed bool                   `json:"compressed"`
	Encrypted  bool                   `json:"encrypted"`
	Checksum   string                 `json:"checksum"`
	Metadata   map[string]interface{} `json:"metadata"`
	Location   string                 `json:"location"`
	Status     string                 `json:"status"`
}

// CacheDebugInfo represents cache debug information
type CacheDebugInfo struct {
	Namespace      CacheNamespace         `json:"namespace"`
	ServiceInfo    *CacheServiceDebugInfo `json:"service_info"`
	L1CacheInfo    interface{}            `json:"l1_cache_info"`
	L2CacheInfo    interface{}            `json:"l2_cache_info"`
	RepositoryInfo interface{}            `json:"repository_info"`
	StrategyInfo   *CacheStrategy         `json:"strategy_info"`
	MetricsInfo    *CacheMetrics          `json:"metrics_info"`
	RecentEvents   []*CacheEvent          `json:"recent_events"`
	ActiveLocks    []*CacheLock           `json:"active_locks"`
	Configuration  map[string]interface{} `json:"configuration"`
	GeneratedAt    time.Time              `json:"generated_at"`
}

// CacheServiceDebugInfo represents cache service debug information
type CacheServiceDebugInfo struct {
	Version        string                    `json:"version"`
	StartTime      time.Time                 `json:"start_time"`
	Uptime         time.Duration             `json:"uptime"`
	IsRunning      bool                      `json:"is_running"`
	WorkerCount    int                       `json:"worker_count"`
	Configuration  *CacheDomainServiceConfig `json:"configuration"`
	MemoryUsage    int64                     `json:"memory_usage"`
	GoroutineCount int                       `json:"goroutine_count"`
}

// CacheTrace represents a cache operation trace
type CacheTrace struct {
	ID         string                 `json:"id"`
	Namespace  CacheNamespace         `json:"namespace"`
	Operation  CacheOperation         `json:"operation"`
	Key        string                 `json:"key"`
	StartTime  time.Time              `json:"start_time"`
	EndTime    time.Time              `json:"end_time"`
	Duration   time.Duration          `json:"duration"`
	Success    bool                   `json:"success"`
	Error      string                 `json:"error,omitempty"`
	Metadata   map[string]interface{} `json:"metadata"`
	StackTrace string                 `json:"stack_trace,omitempty"`
}

// CachePrincipal represents a cache principal for access control
type CachePrincipal struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Name        string            `json:"name"`
	Roles       []string          `json:"roles"`
	Permissions []string          `json:"permissions"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
	ExpiresAt   time.Time         `json:"expires_at"`
}

// IsExpired checks if the principal is expired
func (cp *CachePrincipal) IsExpired() bool {
	return !cp.ExpiresAt.IsZero() && time.Now().After(cp.ExpiresAt)
}

// HasRole checks if the principal has a specific role
func (cp *CachePrincipal) HasRole(role string) bool {
	for _, r := range cp.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasPermission checks if the principal has a specific permission
func (cp *CachePrincipal) HasPermission(permission string) bool {
	for _, p := range cp.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// PartitionConfig represents partition configuration
type PartitionConfig struct {
	MaxSize        int64                  `json:"max_size"`
	TTL            time.Duration          `json:"ttl"`
	EvictionPolicy EvictionPolicy         `json:"eviction_policy"`
	Replicas       int                    `json:"replicas"`
	Shards         int                    `json:"shards"`
	Metadata       map[string]interface{} `json:"metadata"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// =============================================================================
// Interface Definitions
// =============================================================================

// CacheMiddleware represents cache middleware
type CacheMiddleware interface {
	Name() string
	Execute(ctx context.Context, operation CacheOperation, next func(context.Context) error) error
}

// CacheAccessController represents cache access controller
type CacheAccessController interface {
	ValidateAccess(ctx context.Context, namespace CacheNamespace, operation CacheOperation, principal *CachePrincipal) error
	GrantAccess(ctx context.Context, namespace CacheNamespace, operation CacheOperation, principal *CachePrincipal) error
	RevokeAccess(ctx context.Context, namespace CacheNamespace, operation CacheOperation, principal *CachePrincipal) error
}

// CacheMetricsExporter represents cache metrics exporter
type CacheMetricsExporter interface {
	Export(metrics *CacheMetrics, format string) ([]byte, error)
}

// CachePartitionManager represents cache partition manager
type CachePartitionManager interface {
	RegisterPartition(namespace CacheNamespace, partitionID string, config *PartitionConfig) error
	UnregisterPartition(namespace CacheNamespace, partitionID string)
	ListPartitions(namespace CacheNamespace) []string
	GetPartition(namespace CacheNamespace, partitionID string) (*PartitionConfig, bool)
}

// EventBus represents an event bus
type EventBus interface {
	Publish(ctx context.Context, event *CacheEvent) error
	Subscribe(ctx context.Context, eventType CacheEventType, handler CacheEventHandler) error
	Unsubscribe(ctx context.Context, eventType CacheEventType) error
}

// Logger represents a logger interface
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Fatal(msg string, args ...interface{})
}

// =============================================================================
// Cache Repository Interface Extensions
// =============================================================================

// CacheRepository represents the cache repository interface
type CacheRepository interface {
	// Basic operations
	Get(ctx context.Context, key *CacheKey) (interface{}, error)
	Set(ctx context.Context, entry *CacheEntry) error
	Delete(ctx context.Context, key *CacheKey) error
	Exists(ctx context.Context, key *CacheKey) (bool, error)

	// Batch operations
	GetMulti(ctx context.Context, keys []*CacheKey) (map[string]interface{}, error)
	SetMulti(ctx context.Context, entries []*CacheEntry) error
	DeleteMulti(ctx context.Context, keys []*CacheKey) error

	// Advanced operations
	Increment(ctx context.Context, key string, namespace CacheNamespace, delta int64) (int64, error)
	Touch(ctx context.Context, key string, namespace CacheNamespace) error
	Refresh(ctx context.Context, key string, namespace CacheNamespace, ttl time.Duration) error

	// Namespace operations
	CreateNamespace(ctx context.Context, namespace CacheNamespace, config *CacheConfiguration) error
	DeleteNamespace(ctx context.Context, namespace CacheNamespace) error
	ListKeys(ctx context.Context, namespace CacheNamespace, pattern string) ([]string, error)
	Clear(ctx context.Context, namespace CacheNamespace) error

	// Statistics and monitoring
	GetStatistics(ctx context.Context, namespace CacheNamespace) (*CacheStatistics, error)
	GetHealth(ctx context.Context, namespace CacheNamespace) (*CacheHealth, error)
	GetAnalytics(ctx context.Context, namespace CacheNamespace, timeRange TimeRange) (*CacheAnalytics, error)
	GetTrends(ctx context.Context, namespace CacheNamespace, timeRange TimeRange) (*CacheTrends, error)

	// Backup and restore
	Backup(ctx context.Context, namespace CacheNamespace, options *BackupOptions) (*BackupInfo, error)
	Restore(ctx context.Context, namespace CacheNamespace, backupID string, options *RestoreOptions) error
	ListBackups(ctx context.Context, namespace CacheNamespace) ([]*BackupInfo, error)
	DeleteBackup(ctx context.Context, namespace CacheNamespace, backupID string) error

	// Lock operations
	AcquireLock(ctx context.Context, key string, namespace CacheNamespace, ttl time.Duration) (*CacheLock, error)
	ReleaseLock(ctx context.Context, lock *CacheLock) error

	// Debugging and tracing
	Debug(ctx context.Context, namespace CacheNamespace) (*CacheDebugInfo, error)
	SetTracing(ctx context.Context, namespace CacheNamespace, enabled bool) error
	GetTraces(ctx context.Context, namespace CacheNamespace, limit int) ([]*CacheTrace, error)

	// Metrics operations
	EnableMetrics(ctx context.Context, namespace CacheNamespace) error
	DisableMetrics(ctx context.Context, namespace CacheNamespace) error

	// Partition operations
	CreatePartition(ctx context.Context, namespace CacheNamespace, partitionID string, config *PartitionConfig) error
	DeletePartition(ctx context.Context, namespace CacheNamespace, partitionID string) error

	// Cleanup operations
	CleanupExpired(ctx context.Context, namespace CacheNamespace) error
	EvictLRU(ctx context.Context, namespace CacheNamespace, count int) error

	// Lifecycle
	Initialize(ctx context.Context) error
	Shutdown(ctx context.Context) error
	Ping(ctx context.Context) error
}

// =============================================================================
// Cache Level Interfaces
// =============================================================================

// L1Cache represents L1 cache interface
type L1Cache interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	ClearNamespace(ctx context.Context, namespace CacheNamespace) error
	GetStats() interface{}
}

// L2Cache represents L2 cache interface
type L2Cache interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	SetMulti(ctx context.Context, data map[string]interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeleteMulti(ctx context.Context, keys []string) error
	Clear(ctx context.Context) error
	ClearNamespace(ctx context.Context, namespace CacheNamespace) error
	GetStats() interface{}
}

// =============================================================================
// Global Configuration
// =============================================================================

// DefaultCacheConfiguration returns default cache configuration
func DefaultCacheConfiguration() *CacheConfiguration {
	return &CacheConfiguration{
		DefaultTTL:         15 * time.Minute,
		MaxSize:            1000000,
		MaxMemorySize:      1 << 30, // 1GB
		EvictionPolicy:     EvictionPolicyLRU,
		CompressionEnabled: false,
		SerializationType:  "json",
		L1Enabled:          true,
		L2Enabled:          true,
		MonitoringEnabled:  true,
		WarmingEnabled:     false,
		Tags:               make([]string, 0),
		Metadata:           make(map[string]interface{}),
	}
}

// DefaultCacheDomainServiceConfig returns default service configuration
func DefaultCacheDomainServiceConfig() *CacheDomainServiceConfig {
	return &CacheDomainServiceConfig{
		DefaultTTL:                15 * time.Minute,
		MaxMemorySize:             1 << 30, // 1GB
		MaxCacheSize:              1000000,
		CleanupInterval:           5 * time.Minute,
		MetricsCollectionInterval: 30 * time.Second,
		PromotionThreshold:        10,
		EnableAutoOptimization:    false,
		EventProcessingDelay:      100 * time.Millisecond,
		MaxConcurrentOperations:   1000,
	}
}

// =============================================================================
// Utility Functions
// =============================================================================

// IsValidCacheKey checks if a cache key is valid
func IsValidCacheKey(key string) bool {
	return len(key) > 0 && len(key) <= 250
}

// IsValidNamespace checks if a namespace is valid
func IsValidNamespace(namespace CacheNamespace) bool {
	return namespace.IsValid()
}

// GenerateBackupID generates a unique backup ID
func GenerateBackupID() string {
	return fmt.Sprintf("backup_%d_%s", time.Now().Unix(), generateRandomString(8))
}

// generateRandomString generates a random string
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// CalculateSize calculates the size of a cache entry
func CalculateSize(value interface{}) int64 {
	// This is a simplified size calculation
	// In a real implementation, you would use reflection or serialization
	switch v := value.(type) {
	case string:
		return int64(len(v))
	case []byte:
		return int64(len(v))
	case int, int8, int16, int32, int64:
		return 8
	case uint, uint8, uint16, uint32, uint64:
		return 8
	case float32:
		return 4
	case float64:
		return 8
	case bool:
		return 1
	default:
		// For complex types, estimate 1KB
		return 1024
	}
}

// =============================================================================
// Package Initialization
// =============================================================================

func init() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())
}

// =============================================================================
// End of File
// =============================================================================
