// test/e2e/search_test.go
package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/turtacn/starseek/internal/api/http" // Adjust module path
	"github.com/turtacn/starseek/internal/application"
	"github.com/turtacn/starseek/internal/common/config"
	domainindex "github.com/turtacn/starseek/internal/domain/index"
	domainsearch "github.com/turtacn/starseek/internal/domain/search"
	"github.com/turtacn/starseek/internal/domain/tokenizer"
	"github.com/turtacn/starseek/internal/infrastructure/cache"
	"github.com/turtacn/starseek/internal/infrastructure/database"
	"github.com/turtacn/starseek/internal/infrastructure/logger"
	"github.com/turtacn/starseek/internal/infrastructure/metrics"
	"github.com/turtacn/starseek/internal/infrastructure/persistence"
	"github.com/turtacn/starseek/internal/infrastructure/server"
	"github.com/turtacn/starseek/internal/infrastructure/tracing" // Ensure tracing is also imported for its CloseTracerProvider
)

var (
	testServer        *server.Server
	baseURL           string
	appConfig         *config.AppConfig
	appLogger         logger.Logger
	appTracerProvider *tracing.TracerProvider
	once              sync.Once
)

// TestMain sets up and tears down the entire application for E2E tests.
func TestMain(m *testing.M) {
	// Configure Gin for testing
	gin.SetMode(gin.ReleaseMode)

	// One-time setup
	once.Do(func() {
		setupE2ETestEnvironment()
	})

	// Run all tests
	code := m.Run()

	// One-time teardown
	teardownE2ETestEnvironment()

	os.Exit(code)
}

func setupE2ETestEnvironment() {
	fmt.Println("Setting up E2E test environment...")

	// 1. Load Configuration
	// Assuming config files are in the parent directory's 'configs' folder relative to this test file.
	// Adjust path as needed for your project structure.
	configPath := filepath.Join("..", "..", "configs")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Failed to load configuration for E2E tests: %v\n", err)
		os.Exit(1)
	}
	appConfig = cfg // Store for global access if needed

	// Ensure the test uses a unique port
	// You can make this dynamic if desired, but for E2E, a fixed one is often fine if known not to conflict
	if appConfig.Server.Port == 0 {
		appConfig.Server.Port = 8081 // Default E2E port
	}

	// 2. Initialize Logger (discard output for clean test logs)
	// For actual E2E debugging, you might want to switch to console output
	log, err := logger.NewZapLogger(appConfig.Logger.Level, "console") // Use console format for better readability in test logs
	if err != nil {
		fmt.Printf("Failed to initialize logger for E2E tests: %v\n", err)
		os.Exit(1)
	}
	appLogger = log
	log.Info("E2E Test: Application starting up for testing...")

	// 3. Initialize Metrics (disabled in config for simplicity)
	appMetrics := metrics.InitMetrics(&appConfig.Metrics, appLogger) // This starts its own goroutine if enabled

	// 4. Initialize Tracing (disabled in config for simplicity)
	tp, err := tracing.InitOpenTelemetry(&appConfig.Tracing, appLogger)
	if err != nil {
		appLogger.Errorf("E2E Test: Failed to initialize tracing: %v", err)
	}
	appTracerProvider = tp

	// 5. Initialize Cache Client (Redis)
	redisClient, err := cache.NewRedisClient(&appConfig.Redis, appLogger)
	if err != nil {
		appLogger.Fatalf("E2E Test: Failed to initialize Redis client: %v. Is Redis running at %s?", err, appConfig.Redis.Addr)
	}
	// No defer close here, handled in teardown

	// 6. Initialize Metadata Database Client (MySQL)
	metadataDBClient, err := database.NewMySQLClient(&appConfig.MySQL, appLogger)
	if err != nil {
		appLogger.Fatalf("E2E Test: Failed to initialize metadata database client: %v. Is MySQL running?", err)
	}
	// No defer close here, handled in teardown

	// 7. Initialize Main Business Database Client
	mainDBClient, err := database.NewDatabaseClient(&appConfig.MySQL, appLogger) // Using same MySQL config
	if err != nil {
		appLogger.Fatalf("E2E Test: Failed to initialize main database client: %v", err)
	}
	// No defer close here, handled in teardown

	// --- Dependency Injection / Application Assembly ---
	indexMetadataRepo := persistence.NewSQLIndexMetadataRepository(metadataDBClient.DB, appLogger)
	tokenizerService, err := tokenizer.NewTokenizerService(&appConfig.Tokenizer, appLogger)
	if err != nil {
		appLogger.Fatalf("E2E Test: Failed to initialize tokenizer service: %v", err)
	}

	domainIndexService := domainindex.NewDomainIndexService(indexMetadataRepo, appLogger)

	// IMPORTANT: For true E2E with StarRocks, replace DummySearchEngine here.
	// You would have `starrocks.NewStarRocksClient` or similar here.
	// This DummySearchEngine simply simulates success/failure without actual DB interaction.
	dummySearchEngine := domainsearch.NewDummySearchEngine(appLogger)
	domainSearchService := domainsearch.NewDomainSearchService(domainIndexService, dummySearchEngine, tokenizerService, appLogger)

	appIndexService := application.NewIndexApplicationService(domainIndexService, appLogger)
	appSearchService := application.NewSearchApplicationService(domainSearchService, appLogger)

	// API Layer (HTTP Handlers and Routers)
	httpHandler := http.NewHandler(appIndexService, appSearchService)
	httpRouter := http.NewRouter(httpHandler, appLogger, appMetrics, appTracerProvider)

	// Initialize HTTP Server
	testServer = server.NewServer(httpRouter, &appConfig.Server, appLogger)
	baseURL = fmt.Sprintf("http://localhost%s", testServer.GetAddr())

	// Start server in a goroutine
	go func() {
		if err := testServer.Run(); err != nil && err != http.ErrServerClosed {
			appLogger.Errorf("E2E Test: HTTP server failed to start or stopped unexpectedly: %v", err)
			os.Exit(1) // Fatal error if server doesn't start
		}
	}()

	// Wait for the server to be ready
	waitForServerReady(baseURL+"/health", 5*time.Second)

	appLogger.Infof("E2E Test: StarSeek server started at %s", baseURL)
	appLogger.Info("E2E Test: Environment setup complete.")
}

func teardownE2ETestEnvironment() {
	appLogger.Info("Tearing down E2E test environment...")

	// Create a context with a timeout for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Send an interrupt signal to trigger graceful shutdown
	// This simulates what orchestrators like Kubernetes do.
	// This will trigger the defer functions for Redis, MySQL, Logger, Tracing etc.
	// (Note: In this specific setup with `testServer.Run()` in goroutine,
	// testServer.Shutdown() must be called directly. `os.Interrupt` is for main())
	if testServer != nil {
		if err := testServer.Shutdown(shutdownCtx); err != nil {
			appLogger.Errorf("E2E Test: HTTP server forced to shutdown: %v", err)
		} else {
			appLogger.Info("E2E Test: HTTP server shut down gracefully.")
		}
	}

	// Close other resources explicitly if not handled by HTTP server shutdown or `TestMain`'s defer
	// In a real application, these should be handled by central lifecycle management or main's defer.
	// For this E2E structure, we're relying on main-like setup.
	// Explicitly closing TracerProvider. Logger sync is often implicitly handled by Zap.
	if appTracerProvider != nil {
		tracing.CloseTracerProvider(shutdownCtx, appTracerProvider, appLogger)
	}

	appLogger.Info("E2E Test: Environment teardown complete.")
}

// waitForServerReady pings the health endpoint until it gets a 200 OK or timeout.
func waitForServerReady(url string, timeout time.Duration) {
	client := &http.Client{Timeout: 1 * time.Second} // Short timeout for each ping
	start := time.Now()
	for {
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return // Server is ready
		}
		if time.Since(start) > timeout {
			panic(fmt.Sprintf("E2E Test: Server did not become ready at %s within %v", url, timeout))
		}
		appLogger.Debugf("E2E Test: Waiting for server to be ready at %s (error: %v, status: %d)", url, err, resp.StatusCode)
		time.Sleep(100 * time.Millisecond)
	}
}

// makeRequest is a helper to send HTTP requests and parse responses.
func makeRequest(t *testing.T, method, path string, body interface{}) (*http.Response, map[string]interface{}) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		assert.NoError(t, err, "Failed to marshal request body")
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, baseURL+path, reqBody)
	assert.NoError(t, err, "Failed to create HTTP request")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	assert.NoError(t, err, "Failed to send HTTP request")
	assert.NotNil(t, resp, "HTTP response should not be nil")

	var responseData map[string]interface{}
	if resp.Body != nil {
		defer resp.Body.Close()
		bodyBytes, err := io.ReadAll(resp.Body)
		assert.NoError(t, err, "Failed to read response body")
		if len(bodyBytes) > 0 {
			err = json.Unmarshal(bodyBytes, &responseData)
			if err != nil {
				// If not JSON, return nil for responseData, but log for debugging
				t.Logf("Response body is not JSON or empty: %s", string(bodyBytes))
				responseData = nil
			}
		}
	}
	return resp, responseData
}

// --- E2E Test Scenarios ---

func TestE2E_HealthCheck(t *testing.T) {
	resp, data := makeRequest(t, http.MethodGet, "/health", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "ok", data["status"])
}

func TestE2E_IndexRegistrationAndSearch(t *testing.T) {
	indexName := "products_e2e"
	indexSchema := map[string]interface{}{
		"name":        map[string]string{"type": "text"},
		"description": map[string]string{"type": "text"},
		"price":       map[string]string{"type": "float"},
		"category":    map[string]string{"type": "keyword"},
		"created_at":  map[string]string{"type": "date"},
	}
	indexSettings := map[string]interface{}{
		"number_of_shards":   1,
		"number_of_replicas": 0,
	}

	// 1. Register Index
	t.Run("RegisterIndex", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name":     indexName,
			"schema":   indexSchema,
			"settings": indexSettings,
		}
		resp, data := makeRequest(t, http.MethodPost, "/api/v1/indices", reqBody)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Contains(t, data, "id")
		assert.Contains(t, data, "name")
		assert.Equal(t, indexName, data["name"])
		t.Logf("Index '%s' registered with ID: %s", indexName, data["id"])
	})

	// 1a. Try to register the same index again (should fail with 409 Conflict)
	t.Run("RegisterExistingIndex", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name":     indexName,
			"schema":   indexSchema,
			"settings": indexSettings,
		}
		resp, data := makeRequest(t, http.MethodPost, "/api/v1/indices", reqBody)
		assert.Equal(t, http.StatusConflict, resp.StatusCode)
		assert.Contains(t, data, "error")
		assert.Contains(t, data["error"].(string), "already exists")
		t.Logf("Attempted to re-register existing index: %s", data["error"])
	})

	// 2. Simple Keyword Search
	t.Run("SimpleKeywordSearch", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"query": map[string]interface{}{
				"match": map[string]string{
					"name": "example product",
				},
			},
		}
		resp, data := makeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/indices/%s/_search", indexName), reqBody)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, data, "results")
		results, ok := data["results"].([]interface{})
		assert.True(t, ok, "results should be an array")
		// Since we have a DummySearchEngine, we expect fixed results
		assert.Len(t, results, 2)
		assert.Contains(t, results[0].(map[string]interface{})["title"], "Result One")
		t.Logf("Simple search returned %d results: %v", len(results), results)
	})

	// 3. Search with Field Filtering (e.g., price range)
	// NOTE: DummySearchEngine does not implement filtering, so it will return all dummy docs.
	// In a real E2E, this would return filtered results.
	t.Run("FieldFilteringSearch", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"query": map[string]interface{}{
				"range": map[string]interface{}{
					"price": map[string]float64{
						"gte": 10.0,
						"lte": 50.0,
					},
				},
			},
		}
		resp, data := makeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/indices/%s/_search", indexName), reqBody)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		results, ok := data["results"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, results, 2, "Dummy engine returns all results; real engine would filter")
		t.Logf("Field filtering search (dummy) returned %d results: %v", len(results), results)
	})

	// 4. Pagination
	// NOTE: DummySearchEngine does not implement pagination, so it will return all dummy docs.
	// In a real E2E, this would return paginated results.
	t.Run("PaginationSearch", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
			"from": 0,
			"size": 1, // Request only 1 result
		}
		resp, data := makeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/indices/%s/_search", indexName), reqBody)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		results, ok := data["results"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, results, 2, "Dummy engine returns all results; real engine would paginate")
		t.Logf("Pagination search (dummy) returned %d results: %v", len(results), results)
	})

	// 5. Search on a non-existent index (should fail with 404 Not Found)
	t.Run("SearchNonExistentIndex", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}
		resp, data := makeRequest(t, http.MethodPost, "/api/v1/indices/non_existent_index/_search", reqBody)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		assert.Contains(t, data, "error")
		assert.Contains(t, data["error"].(string), "not found")
		t.Logf("Attempted search on non-existent index: %s", data["error"])
	})

	// TODO: Add more advanced scenarios if DummySearchEngine is replaced:
	// - Cross-index search (if implemented in application layer)
	// - Highlighting (requires search engine support)
	// - Relevance sorting (requires search engine support)
	// - Complex boolean queries
	// - Aggregations (if supported by StarRocks and exposed via API)
}

// TODO: Add a test for document indexing/ingestion if that API is available.
// E.g., func TestE2E_DocumentIngestion(t *testing.T) { ... }
