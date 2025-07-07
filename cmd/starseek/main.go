package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/turtacn/starseek/internal/common/errorx"
	"github.com/turtacn/starseek/internal/domain/model"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "go.uber.org/automaxprocs" // Automatically sets GOMAXPROCS

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/turtacn/starseek/internal/adapter/starrocks"
	"github.com/turtacn/starseek/internal/api/handler"
	"github.com/turtacn/starseek/internal/api/router"
	"github.com/turtacn/starseek/internal/application/service"
	"github.com/turtacn/starseek/internal/domain/repository"
	"github.com/turtacn/starseek/internal/infrastructure/cache"
	"github.com/turtacn/starseek/internal/infrastructure/log"
	"github.com/turtacn/starseek/internal/infrastructure/persistence/mysql"
)

// Config holds all application configurations.
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Log       log.Config      `mapstructure:"log"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     cache.Config    `mapstructure:"redis"`
	StarRocks StarRocksConfig `mapstructure:"starrocks"`
}

// ServerConfig holds HTTP server configurations.
type ServerConfig struct {
	Port int `mapstructure:"port"`
}

// DatabaseConfig holds relational database configurations for Starseek's metadata.
type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}

// StarRocksConfig holds StarRocks specific configurations.
type StarRocksConfig struct {
	DSN string `mapstructure:"dsn"`
}

func main() {
	// 1. Load Configuration
	var cfg Config
	viper.SetConfigName("config") // config.yaml
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs") // Search path for config file
	viper.AddConfigPath(".")         // Fallback to current directory

	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error reading config file: %s \n", err)
		os.Exit(1)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error unmarshalling config: %s \n", err)
		os.Exit(1)
	}

	// 2. Initialize Logger
	appLogger, err := log.NewLogger(cfg.Log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error initializing logger: %s \n", err)
		os.Exit(1)
	}
	defer func() {
		_ = appLogger.(*log.ZapLogger).Sync() // Flushes any buffered log entries
	}()

	appLogger.Info("Configuration loaded successfully", zap.Any("config", cfg))

	// 3. Initialize Infrastructure Clients
	// Redis Cache
	redisCache, err := cache.NewRedisCache(cfg.Redis)
	if err != nil {
		appLogger.Fatal("Failed to initialize Redis cache", zap.Error(err))
	}
	appLogger.Info("Redis cache initialized")

	// Metadata Database (MySQL)
	mysqlRepo, err := mysql.NewMySQLIndexRepo(cfg.Database.DSN)
	if err != nil {
		appLogger.Fatal("Failed to initialize MySQL metadata repository", zap.Error(err))
	}
	appLogger.Info("MySQL metadata repository initialized")

	// StarRocks DB Client (for search queries)
	srDB, err := sql.Open("mysql", cfg.StarRocks.DSN) // StarRocks uses MySQL protocol
	if err != nil {
		appLogger.Fatal("Failed to open StarRocks database connection", zap.Error(err))
	}
	srDB.SetMaxIdleConns(10)
	srDB.SetMaxOpenConns(100)
	srDB.SetConnMaxLifetime(time.Hour) // Set connection max lifetime
	if err = srDB.Ping(); err != nil {
		appLogger.Fatal("Failed to connect to StarRocks database", zap.Error(err))
	}
	appLogger.Info("StarRocks database client initialized")

	// 4. Initialize Repositories (implementations of domain interfaces)
	indexRepo := mysqlRepo
	starrocksAdapter := starrocks.NewStarRocksAdapter(srDB)

	// Create a generic SearchRepository that uses the appropriate adapter
	// For simplicity, we'll create a basic SearchRepoImpl that wraps the StarRocks adapter.
	// In a multi-adapter scenario, this would be a more complex factory/dispatcher.
	searchRepo := repository.NewSearchRepositoryImpl(starrocksAdapter, appLogger) // This needs to be implemented.

	// 5. Initialize Domain Services and Application Services (Dependency Injection)
	queryProcessor := service.NewQueryProcessor(appLogger.Named("QueryProcessor"), indexRepo)
	rankingService := service.NewRankingService(appLogger.Named("RankingService"))
	taskScheduler := service.NewTaskScheduler(searchRepo, indexRepo, redisCache, appLogger.Named("TaskScheduler"))

	searchAppService := service.NewSearchAppService(queryProcessor, taskScheduler, rankingService, searchRepo, indexRepo, redisCache, appLogger.Named("SearchAppService"))
	indexAppService := service.NewIndexAppService(indexRepo, redisCache, appLogger.Named("IndexAppService"))

	// 6. Initialize HTTP Handlers
	searchHandler := handler.NewSearchHandler(searchAppService, appLogger.Named("SearchHandler"))
	indexHandler := handler.NewIndexHandler(indexAppService, appLogger.Named("IndexHandler"))

	// 7. Initialize and Run Router (Gin Engine)
	gin.SetMode(gin.ReleaseMode) // Set Gin mode based on environment or config
	r := router.NewRouter(searchHandler, indexHandler, appLogger)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: r,
	}

	// Goroutine to start the server
	go func() {
		appLogger.Info(fmt.Sprintf("Starseek server starting on port %d", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Server failed to listen", zap.Error(err))
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // Capture Ctrl+C and Kubernetes termination signals
	<-quit                                               // Block until a signal is received
	appLogger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Give 5 seconds for graceful shutdown
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Error("Server forced to shutdown", zap.Error(err))
	} else {
		appLogger.Info("Server exited gracefully")
	}

	// Close database connections
	if err := srDB.Close(); err != nil {
		appLogger.Error("Failed to close StarRocks DB connection", zap.Error(err))
	}
	appLogger.Info("StarRocks DB connection closed.")

	// Close metadata DB connection
	if closer, ok := mysqlRepo.(interface{ Close() error }); ok { // Assuming MySQLIndexRepo has a Close method
		if err := closer.Close(); err != nil {
			appLogger.Error("Failed to close MySQL metadata DB connection", zap.Error(err))
		}
		appLogger.Info("MySQL metadata DB connection closed.")
	}
	appLogger.Info("Starseek application terminated.")
}

// SearchRepositoryImpl is a basic implementation of SearchRepository.
// In a real scenario, this would likely be a more complex dispatcher if supporting multiple DBAdapters dynamically.
type SearchRepositoryImpl struct {
	starrocksAdapter starrocks.StarRocksAdapter // For simplicity, direct access for now
	logger           *zap.Logger
}

// NewSearchRepositoryImpl creates a new SearchRepositoryImpl.
func NewSearchRepositoryImpl(adapter starrocks.StarRocksAdapter, logger *zap.Logger) repository.SearchRepository {
	return &SearchRepositoryImpl{
		starrocksAdapter: adapter,
		logger:           logger.Named("SearchRepository"),
	}
}

// Search implements the Search method for SearchRepository.
// This example currently only supports StarRocksAdapter.
func (s *SearchRepositoryImpl) Search(ctx context.Context, parsedQuery *model.ParsedQuery, indexMetadata []*model.IndexMetadata) (*model.SearchResult, error) {
	// For each target field in parsedQuery, generate SQL based on its indexMetadata
	// and execute concurrently. Then merge results (RowIDs).

	var (
		wg           sync.WaitGroup
		resultsChan  = make(chan []map[string]interface{}, len(indexMetadata)) // Channel to collect results from concurrent queries
		errChan      = make(chan error, 1)                                     // Channel to propagate first error
		uniqueRowIDs = make(map[string]struct{})                               // To store unique RowIDs
		mu           sync.Mutex                                                // Mutex for uniqueRowIDs
	)

	// In a realistic scenario, parsedQuery.TargetFields would specify the columns to search,
	// and we'd look up the `IndexMetadata` for each `table.column` combination.
	// For simplicity, let's assume `indexMetadata` passed here are already filtered to be relevant.

	for _, meta := range indexMetadata {
		wg.Add(1)
		go func(colMeta *model.IndexMetadata) {
			defer wg.Done()

			// Build SQL for this specific column
			sqlStr, args, err := s.starrocksAdapter.BuildSearchSQL(colMeta.TableName, colMeta, parsedQuery)
			if err != nil {
				s.logger.Error("Failed to build search SQL", zap.Error(err),
					zap.String("table", colMeta.TableName), zap.String("column", colMeta.ColumnName))
				select {
				case errChan <- err:
				default:
				}
				return
			}
			if sqlStr == "" { // No valid SQL generated (e.g., no matching query terms)
				return
			}

			// Execute the SQL
			rows, err := s.starrocksAdapter.ExecuteSearch(ctx, sqlStr, args)
			if err != nil {
				s.logger.Error("Failed to execute search query against StarRocks", zap.Error(err),
					zap.String("sql", sqlStr), zap.Any("args", args))
				select {
				case errChan <- err:
				default:
				}
				return
			}
			resultsChan <- rows
		}(meta)
	}

	wg.Wait()
	close(resultsChan)

	select {
	case err := <-errChan:
		return nil, err // Return the first error encountered
	default:
		// No errors
	}

	totalHits := 0
	rowIDColumn, err := s.starrocksAdapter.GetRowIDColumn("") // Get generic ID column, needs table-specific for robust
	if err != nil {
		s.logger.Error("Failed to get row ID column name", zap.Error(err))
		return nil, errorx.NewError(errorx.ErrInternalServer.Code, "Failed to identify row ID column", err)
	}

	for rowBatch := range resultsChan {
		for _, row := range rowBatch {
			if idVal, ok := row[rowIDColumn]; ok {
				if idStr, isString := idVal.(string); isString {
					mu.Lock()
					if _, exists := uniqueRowIDs[idStr]; !exists {
						uniqueRowIDs[idStr] = struct{}{}
						totalHits++
					}
					mu.Unlock()
				}
			}
		}
	}

	// Convert unique RowIDs to SearchResultItem slice (only RowID for now)
	items := make([]*model.SearchResultItem, 0, len(uniqueRowIDs))
	for id := range uniqueRowIDs {
		items = append(items, &model.SearchResultItem{RowID: id})
	}

	s.logger.Info("Initial search completed", zap.Int("unique_row_ids", len(uniqueRowIDs)))

	return &model.SearchResult{
		Items:     items,
		TotalHits: int64(totalHits),
	}, nil
}

// FetchOriginalData implements the FetchOriginalData method for SearchRepository.
func (s *SearchRepositoryImpl) FetchOriginalData(ctx context.Context, tableName string, rowIDs []string) ([]map[string]interface{}, error) {
	if len(rowIDs) == 0 {
		return []map[string]interface{}{}, nil
	}

	// Get the row ID column name for the given table
	rowIDColumn, err := s.starrocksAdapter.GetRowIDColumn(tableName)
	if err != nil {
		s.logger.Error("Failed to get row ID column name for fetching original data", zap.String("table", tableName), zap.Error(err))
		return nil, errorx.NewError(errorx.ErrInternalServer.Code, "Failed to identify row ID column for fetching original data", err)
	}

	// Prepare SQL for batch fetching
	placeholders := make([]string, len(rowIDs))
	args := make([]interface{}, len(rowIDs))
	for i, id := range rowIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	sqlStr := fmt.Sprintf("SELECT * FROM %s WHERE %s IN (%s)",
		tableName,
		rowIDColumn,
		strings.Join(placeholders, ","),
	)

	s.logger.Debug("Executing FetchOriginalData query", zap.String("sql", sqlStr), zap.Int("num_ids", len(rowIDs)))

	results, err := s.starrocksAdapter.ExecuteSearch(ctx, sqlStr, args)
	if err != nil {
		s.logger.Error("Failed to execute query for fetching original data", zap.Error(err),
			zap.String("table", tableName), zap.Any("row_ids", rowIDs))
		return nil, errorx.NewError(errorx.ErrDatabase.Code, "Failed to fetch original data from database", err)
	}

	return results, nil
}

//Personal.AI order the ending
