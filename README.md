# StarSeek: A Unified Full-Text Search Mid-Platform for Data Warehouses and Lakehouses


## 🌟 Project Overview

StarSeek is an innovative open-source mid-platform designed to empower columnar analytical databases like StarRocks, Apache Doris, and ClickHouse with advanced full-text search capabilities typically found in dedicated search engines such as Elasticsearch. Leveraging the robust inverted index features of these OLAP databases, StarSeek provides a unified, intelligent, and highly scalable search service layer, significantly enhancing their utility for document retrieval, content search, and complex data exploration.

This project addresses the inherent limitations of analytical databases in handling sophisticated full-text search requirements, transforming them into powerful search backends without the overhead of maintaining a separate search cluster.

## 🎯 Key Pain Points & Core Value

### Pain Points Addressed

*   **Limited Native Full-Text Search Capabilities**: While StarRocks, Doris, and ClickHouse offer basic inverted index functionalities, they lack advanced search features such as cross-table/cross-column queries, relevance scoring (TF-IDF/BM25), synonym expansion, result highlighting, and unified metadata management.
*   **Complex Development & Maintenance**: Developers often need to manually construct intricate SQL queries (e.g., `MATCH_AGAINST`, `UNION ALL`), handle tokenization consistency, and implement ranking logic in the application layer, leading to increased complexity and reduced efficiency.
*   **Suboptimal User Experience**: End-users often face less precise search results and a lack of advanced interactive search features, which are commonplace in modern search applications.
*   **Data Source Coupling**: Applications become tightly coupled with the specific SQL dialect and indexing mechanisms of the underlying analytical database, hindering flexibility and future migration.

### Core Value Proposition

*   **Unified Search Gateway**: StarSeek acts as a single, standardized API endpoint for all full-text search operations across your analytical data stores, simplifying client-side integration.
*   **Feature Augmentation**: It bridges the functional gap, introducing Elasticsearch-like capabilities such as:
    *   **Unified Index Metadata Management**: Centralized registry for all indexed columns across multiple tables and databases.
    *   **Advanced Query Processing**: Support for natural language queries, multi-field search (`title:keyword AND content:another`), synonym expansion, and intelligent query translation to optimized SQL.
    *   **Simulated Relevance Ranking**: Implements server-side TF-IDF/BM25-like scoring, providing highly relevant search results.
    *   **Highlighting & Pagination**: Essential features for a rich user search experience.
    *   **Cross-Table/Cross-Column Search**: Seamlessly search across multiple tables and columns with a single query.
*   **Enhanced Developer Efficiency**: Abstract away the complexities of SQL generation, tokenization, and ranking, allowing developers to focus on application logic.
*   **Data Source Agnosticism**: Provides an adapter layer to support StarRocks, Apache Doris, and ClickHouse, ensuring future extensibility to other similar databases.
*   **Performance Optimization**: Incorporates caching mechanisms (e.g., Redis for hot keywords) and concurrent query execution management to optimize search performance.

## ✨ Main Features

*   **Index Registry Service**:
    *   Centralized management of indexed column metadata (table name, column name, index type, tokenizer, data type).
    *   Supports various index types: English, Chinese, Multilingual, No-tokenizer.
*   **Query Processing Engine**:
    *   Translates natural language queries into optimized SQL.
    *   Integrated multi-language tokenization (consistent with StarRocks' indexing).
    *   Optional synonym expansion.
    *   Field filtering and boolean logic support (`field:keyword`).
    *   Simulated keyword scoring for ranking.
    *   Automated cross-table SQL query generation.
*   **Query Optimization Module**:
    *   Caching of query results for hot keywords (e.g., in Redis).
    *   Potential for result set caching to accelerate subsequent requests.
    *   Leveraging underlying database's column-store optimization.
*   **Ranking Module (Elasticsearch-like Simulation)**:
    *   Server-side calculation of pseudo TF-IDF/BM25 scores.
    *   Term Frequency (TF) calculation based on text content.
    *   Inverse Document Frequency (IDF) based on pre-computed statistics.
*   **Task Scheduling & Concurrency Control**:
    *   Manages concurrent SQL execution tasks to prevent database overload.
    *   Configurable concurrency limits.
*   **Multi-Database Support**:
    *   Designed with an extensible adapter pattern to support StarRocks, Apache Doris, and ClickHouse as primary backends, with StarRocks as the core focus.

## 🏛️ Architecture Overview

The detailed architecture of StarSeek, including its layered design, module interactions, and deployment considerations, is thoroughly documented in `docs/architecture.md`.

[Explore the comprehensive Architecture Design](docs/architecture.md)

## 🚀 Building and Running Guide

### Prerequisites

*   Go (version 1.20.2 or later)
*   Git
*   Docker (for local database setup, optional but recommended)
*   StarRocks, Doris, or ClickHouse instance (or a Dockerized setup)
*   Redis instance (for caching)

### Getting Started

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/turtacn/starseek.git
    cd starseek
    ```

2.  **Install dependencies:**
    ```bash
    go mod tidy
    ```

3.  **Configure Environment Variables (Example):**
    Create a `.env` file or set environment variables:
    ```
    STARSEEK_DB_TYPE=starrocks
    STARSEEK_DB_HOST=localhost
    STARSEEK_DB_PORT=9030
    STARSEEK_DB_USER=root
    STARSEEK_DB_PASSWORD=
    STARSEEK_DB_NAME=your_database
    STARSEEK_REDIS_ADDR=localhost:6379
    STARSEEK_LISTEN_ADDR=:8080
    STARSEEK_LOG_LEVEL=info
    # More configurations will be detailed in docs/configuration.md
    ```

4.  **Build the application:**
    ```bash
    go build -o starseek ./cmd/starseek/main.go
    ```

5.  **Run the application:**
    ```bash
    ./starseek
    ```
    The service will start listening on the configured address (e.g., `http://localhost:8080`).

### Docker Setup (Example for StarRocks)

For development and testing, you can use Docker Compose to spin up a StarRocks instance.
```bash
# Example docker-compose.yml (simplified)
version: '3.8'
services:
  starrocks-fe:
    image: starrocks/starrocks:latest
    ports:
      - "9030:9030" # FE Query Port
      - "8030:8030" # HTTP Port
    environment:
      - FE_SERVERS="127.0.0.1:9030"
    command: ["/opt/starrocks/bin/start_fe.sh"]
  redis:
    image: redis:6-alpine
    ports:
      - "6379:6379"
````

Run: `docker-compose up -d`

## 🧪 Code Snippets & Capabilities Demo

### 1. Index Registration (API Endpoint)

Registering an inverted index column for a table. This would typically be done via an admin API.

```go
// Example of how the internal service might interact with the Index Registry
// (Actual API might be HTTP POST to /api/v1/indexes/register)
package main

import (
	"context"
	"fmt"
	"github.com/turtacn/starseek/internal/application"
	"github.com/turtacn/starseek/internal/common/types/enum"
	"github.com/turtacn/starseek/internal/domain/index"
	"github.com/turtacn/starseek/internal/infrastructure/persistence/inmemory"
	"github.com/turtacn/starseek/internal/infrastructure/logger"
)

func main() {
	// Initialize logger
	l := logger.NewZapLogger()
	// Using in-memory for demo; persistence would be a database
	repo := inmemory.NewIndexMetadataRepository()
	service := application.NewIndexService(repo, l)

	ctx := context.Background()

	// Register a document title column
	err := service.RegisterIndex(ctx, index.RegisterIndexCommand{
		TableName:   "documents",
		ColumnName:  "title",
		IndexType:   enum.IndexTypeChinese,
		Tokenizer:   "jieba",
		DataType:    "VARCHAR",
		Description: "Document title for full-text search",
	})
	if err != nil {
		fmt.Printf("Failed to register index for documents.title: %v\n", err)
	} else {
		fmt.Println("Successfully registered index for documents.title")
	}

	// Register a product description column
	err = service.RegisterIndex(ctx, index.RegisterIndexCommand{
		TableName:   "products",
		ColumnName:  "description",
		IndexType:   enum.IndexTypeEnglish,
		Tokenizer:   "standard",
		DataType:    "TEXT",
		Description: "Product description for full-text search",
	})
	if err != nil {
		fmt.Printf("Failed to register index for products.description: %v\n", err)
	} else {
		fmt.Println("Successfully registered index for products.description")
	}

	// Example of fetching registered indexes
	allIndexes, err := service.ListIndexes(ctx, index.ListIndexesQuery{})
	if err != nil {
		fmt.Printf("Failed to list indexes: %v\n", err)
	} else {
		fmt.Println("\nRegistered Indexes:")
		for _, idx := range allIndexes {
			fmt.Printf("- Table: %s, Column: %s, Type: %s, Tokenizer: %s\n",
				idx.TableName, idx.ColumnName, idx.IndexType.String(), idx.Tokenizer)
		}
	}
}
```

### 2. Cross-Table Full-Text Search (HTTP API Example)

Simulating a search request via a `curl` command. The StarSeek service processes this.

```bash
# Search for "人工智能" (Artificial Intelligence) across 'documents' and 'articles' tables
# in 'title' and 'content' fields.
# Request might look like:
curl -X GET "http://localhost:8080/api/v1/search?q=人工智能&fields=title,content&tables=documents,articles&page=1&pageSize=10" \
     -H "Content-Type: application/json"

# Example Response (simplified):
# {
#   "query": "人工智能",
#   "results": [
#     {
#       "table": "documents",
#       "rowId": "doc_123",
#       "score": 0.85,
#       "highlight": {
#         "title": "关于<B>人工智能</B>在教育领域的应用",
#         "content": "近年来，<B>人工智能</B>技术飞速发展..."
#       },
#       "data": { /* Original row data */ }
#     },
#     {
#       "table": "articles",
#       "rowId": "art_456",
#       "score": 0.72,
#       "highlight": {
#         "content": "深度学习是<B>人工智能</B>的核心分支..."
#       },
#       "data": { /* Original row data */ }
#     }
#   ],
#   "totalHits": 250,
#   "currentPage": 1,
#   "pageSize": 10
# }
```

### 3. Field-Specific Search

```bash
# Search for keyword "starrocks" specifically in 'title' and "performance" in 'content'
curl -X GET "http://localhost:8080/api/v1/search?q=title:starrocks AND content:performance&tables=documents&page=1&pageSize=5" \
     -H "Content-Type: application/json"
```

## 🤝 Contribution Guide

We welcome contributions from the community! If you're interested in improving StarSeek, please check out our:

* [Contribution Guidelines](CONTRIBUTING.md)
* [Code of Conduct](CODE_OF_CONDUCT.md)

---

[中文版 README-zh.md](README-zh.md)