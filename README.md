# StarSeek - Full-Text Search Platform for Data Warehouses

[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](https://opensource.org/licenses/Apache-2.0)

[🇨🇳 中文版本](./README-zh.md)

## Overview

StarSeek is an enterprise-grade full-text search middleware service designed to unify and enhance full-text search capabilities across multiple columnar databases including **StarRocks**, **ClickHouse**, and **Doris**. By centralizing inverted index management and providing advanced search features, StarSeek bridges the gap between traditional data warehouses and modern search engines.

## Key Pain Points & Core Value

### Pain Points We Solve

- **Fragmented Search**: Each table's inverted indexes are managed independently, lacking unified search capabilities
- **Limited Features**: Native databases lack advanced search features like ranking, highlighting, and cross-table searches
- **Complex Query Construction**: Building efficient full-text queries across multiple tables requires deep SQL expertise
- **Performance Bottlenecks**: Large UNION ALL queries and lack of search-specific optimizations

### Core Value Proposition

- **🎯 Unified Search Interface**: Single API endpoint for searching across all indexed columns and tables
- **🚀 Enhanced Performance**: Intelligent query optimization, caching, and concurrent execution
- **⭐ Advanced Features**: Ranking algorithms, pagination, highlighting, and relevance scoring
- **🔧 Easy Integration**: RESTful API with minimal learning curve
- **📊 Multi-Database Support**: Native support for StarRocks, ClickHouse, and Doris

## Key Features

### 🔍 Full-Text Search Engine
- Cross-table and cross-column search capabilities
- Multi-language tokenization (Chinese, English, multilingual)
- Synonym expansion and field-specific filtering
- Boolean query support (`title:AI AND content:machine learning`)

### 📈 Intelligent Ranking
- TF-IDF simulation for relevance scoring
- Configurable ranking algorithms
- Custom scoring factors and boost parameters

### ⚡ Performance Optimization
- Redis-based inverted index caching
- Bitmap-accelerated row filtering
- Concurrent query execution with flow control
- Hot keyword preloading

### 🎛️ Management & Monitoring
- Centralized index registry and metadata management
- Real-time query analytics and performance metrics
- Comprehensive logging and distributed tracing

## Quick Start

### Prerequisites
- Go 1.20.2 or later
- Redis 6.0+
- One or more supported databases (StarRocks/ClickHouse/Doris)

### Installation

```bash
# Clone the repository
git clone https://github.com/turtacn/starseek.git
cd starseek

# Install dependencies
go mod download

# Build the application
make build

# Start the service
./bin/starseek --config=config/config.yaml
````

### Configuration Example

```yaml
# config/config.yaml
server:
  http_port: 8080
  grpc_port: 9090

databases:
  starrocks:
    host: "localhost"
    port: 9030
    user: "root"
    password: ""
    database: "search_db"
  
redis:
  addr: "localhost:6379"
  password: ""
  db: 0

search:
  max_concurrent_queries: 10
  cache_ttl: "1h"
  default_page_size: 20
```

## API Examples

### Basic Search

```bash
# Search across all indexed columns
curl "http://localhost:8080/api/v1/search?q=artificial intelligence&limit=10"
```

```json
{
  "query": "artificial intelligence",
  "total": 156,
  "results": [
    {
      "table": "articles",
      "row_id": 12345,
      "score": 0.95,
      "highlights": {
        "title": "The Future of <em>Artificial Intelligence</em>",
        "content": "AI and <em>machine learning</em> are transforming..."
      },
      "fields": {
        "title": "The Future of Artificial Intelligence",
        "content": "AI and machine learning are transforming industries...",
        "created_at": "2024-01-15T10:30:00Z"
      }
    }
  ],
  "execution_time": "45ms"
}
```

### Advanced Field-Specific Search

```bash
# Search with field filters and boolean operators
curl "http://localhost:8080/api/v1/search?q=title:AI AND content:deep learning&fields=title,content,author"
```

### Go SDK Usage

```go
package main

import (
    "context"
    "fmt"
    "github.com/turtacn/starseek/pkg/client"
)

func main() {
    // Initialize StarSeek client
    client := starseek.NewClient("http://localhost:8080")
    
    // Perform search
    req := &starseek.SearchRequest{
        Query:     "machine learning",
        Fields:    []string{"title", "content"},
        Limit:     10,
        Highlight: true,
    }
    
    resp, err := client.Search(context.Background(), req)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Found %d results in %s\n", resp.Total, resp.ExecutionTime)
    for _, result := range resp.Results {
        fmt.Printf("Table: %s, Score: %.2f\n", result.Table, result.Score)
        fmt.Printf("Title: %s\n", result.Highlights["title"])
    }
}
```

## Architecture Overview

StarSeek follows a layered architecture with clear separation of concerns:

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   HTTP/gRPC     │    │   Web Dashboard  │    │   Go SDK        │
│   REST API      │    │   (Optional)     │    │   Client        │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                        │                        │
         └────────────────────────┼────────────────────────┘
                                  │
┌─────────────────────────────────┴─────────────────────────────────┐
│                        Application Layer                          │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │ Query Processor │  │ Index Registry  │  │ Task Scheduler  │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
└───────────────────────────────────────────────────────────────────┘
         │                        │                        │
┌─────────────────────────────────┴─────────────────────────────────┐
│                         Domain Layer                               │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │ Search Engine   │  │ Ranking Engine  │  │ Cache Manager   │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
└───────────────────────────────────────────────────────────────────┘
         │                        │                        │
┌─────────────────────────────────┴─────────────────────────────────┐
│                    Infrastructure Layer                            │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │   StarRocks     │  │   ClickHouse    │  │     Redis       │  │
│  │    Adapter      │  │    Adapter      │  │     Cache       │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
└───────────────────────────────────────────────────────────────────┘
```

For detailed architecture documentation, see [docs/architecture.md](docs/architecture.md).

## Building and Running

### Development Environment

```bash
# Install development dependencies
make dev-deps

# Run tests
make test

# Run with hot reload
make dev

# Format code
make fmt

# Generate API documentation
make docs
```

### Production Deployment

```bash
# Build optimized binary
make build-prod

# Build Docker image
docker build -t starseek:latest .

# Deploy with Docker Compose
docker-compose up -d
```

### Performance Benchmarks

| Scenario              | QPS   | Avg Latency | P99 Latency |
| --------------------- | ----- | ----------- | ----------- |
| Simple keyword search | 1,000 | 15ms        | 45ms        |
| Cross-table search    | 500   | 35ms        | 120ms       |
| Complex boolean query | 200   | 85ms        | 250ms       |

*Benchmarks run on 4-core, 16GB RAM with StarRocks cluster*

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Standards

* Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
* Add tests for new features
* Update documentation for API changes
* Ensure all tests pass (`make test`)

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Community

* **Documentation**: [docs.starseek.io](https://docs.starseek.io)
* **Issues**: [GitHub Issues](https://github.com/turtacn/starseek/issues)
* **Discussions**: [GitHub Discussions](https://github.com/turtacn/starseek/discussions)

## Acknowledgments

* [StarRocks](https://github.com/StarRocks/StarRocks) - High-performance analytical database
* [ClickHouse](https://github.com/ClickHouse/ClickHouse) - Fast columnar database
* [Apache Doris](https://github.com/apache/doris) - Real-time analytical database
* [Elasticsearch](https://github.com/elastic/elasticsearch) - Search engine inspiration

---

**Star ⭐ this repository if you find it useful!**