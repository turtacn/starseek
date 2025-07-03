# StarSeek: 面向数仓与数据湖的统一全文检索中台服务

## 🌟 项目简介

StarSeek 是一个创新的开源中台服务，旨在为 StarRocks、Apache Doris 和 ClickHouse 等列式分析型数据库赋能，提供类似 Elasticsearch 的高级全文检索能力。通过充分利用这些 OLAP 数据库底层的倒排索引特性，StarSeek 构建了一个统一、智能且高度可扩展的搜索服务层，显著增强了它们在文档检索、内容搜索和复杂数据探索方面的效用。

本项目旨在解决分析型数据库在处理复杂全文检索需求时的固有局限性，将其转变为强大的搜索后端，同时避免了维护一个独立搜索集群的额外开销。

## 🎯 核心痛点与核心价值

### 解决的痛点

*   **原生全文检索能力有限**: 尽管 StarRocks、Doris 和 ClickHouse 提供了基础的倒排索引功能，但它们缺乏高级搜索特性，如跨表/跨列查询、相关度评分（TF-IDF/BM25）、同义词扩展、结果高亮显示以及统一的元数据管理。
*   **开发与维护复杂**: 开发者常常需要手动构建复杂的 SQL 查询（例如 `MATCH_AGAINST`、`UNION ALL`），处理分词一致性，并在应用层实现排名逻辑，这导致了开发复杂性增加和效率降低。
*   **用户体验不佳**: 最终用户常常面临搜索结果不够精确，且缺乏现代搜索应用中常见的先进交互式搜索功能。
*   **数据源耦合**: 应用程序与底层分析型数据库的特定 SQL 方言和索引机制紧密耦合，阻碍了灵活性和未来的迁移。

### 核心价值主张

*   **统一搜索入口**: StarSeek 为您在分析型数据存储上的所有全文检索操作提供一个统一、标准化的 API 端点，简化了客户端集成。
*   **功能增强**: 它弥补了底层数据库的功能空白，引入了类似 Elasticsearch 的丰富搜索能力，例如：
    *   **统一索引元数据管理**: 集中注册和管理多表、多数据库中所有已索引列的元数据。
    *   **高级查询处理**: 支持自然语言查询、多字段搜索（如 `title:关键词 AND content:另一个`）、同义词扩展，以及智能地将查询转换为优化的 SQL。
    *   **模拟相关度排名**: 在服务端实现类似 TF-IDF/BM25 的评分机制，提供高度相关的搜索结果。
    *   **高亮与分页**: 提供了丰富的用户搜索体验所必需的功能。
    *   **跨表/跨列搜索**: 通过一个查询无缝地在多个表和列中进行搜索。
*   **提升开发者效率**: 抽象了复杂的 SQL 生成、分词和排名逻辑，使开发者能够专注于应用业务逻辑。
*   **数据源无关性**: 提供了一个适配器层，以支持 StarRocks、Apache Doris 和 ClickHouse 作为主要后端，并以 StarRocks 为核心焦点。
*   **性能优化**: 集成了缓存机制（例如 Redis 用于热门关键词）和并发查询执行管理，以优化搜索性能。

## ✨ 主要功能特性

*   **索引注册服务 (Index Registry)**:
    *   集中管理已索引列的元数据（表名、列名、索引类型、分词器、数据类型）。
    *   支持多种索引类型：英文、中文、多语言、不分词。
*   **查询处理引擎 (Query Processing Engine)**:
    *   将自然语言查询转换为优化的 SQL。
    *   集成多语言分词（与 StarRocks 索引分词保持一致）。
    *   可选的同义词扩展。
    *   支持字段筛选和布尔逻辑（`field:keyword`）。
    *   模拟关键词打分机制用于排名。
    *   自动生成跨表 SQL 查询。
*   **查询优化模块 (Query Optimization Module)**:
    *   对热门关键词的查询结果进行缓存（例如，使用 Redis）。
    *   可能对结果集进行缓存以加速后续请求。
    *   充分利用底层数据库的列式存储优化。
*   **排名模块 (Elasticsearch 模拟)**:
    *   在服务端计算伪 TF-IDF/BM25 分数。
    *   根据文本内容计算词频（Term Frequency, TF）。
    *   根据预计算的统计数据计算逆文档频率（Inverse Document Frequency, IDF）。
*   **任务调度与并发控制**:
    *   管理并发 SQL 执行任务，防止数据库过载。
    *   可配置的并发度限制。
*   **多数据库支持**:
    *   采用可扩展的适配器模式，主要支持 StarRocks、Apache Doris 和 ClickHouse 作为后端，以 StarRocks 为核心。

## 🏛️ 架构概览

StarSeek 的详细架构，包括其分层设计、模块交互和部署考量，已在 `docs/architecture.md` 中详细阐述。

[查看完整的架构设计文档](docs/architecture.md)

## 🚀 构建与运行指南

### 前置条件

*   Go (版本 1.20.2 或更高)
*   Git
*   Docker (用于本地数据库设置，可选但推荐)
*   StarRocks, Doris 或 ClickHouse 实例 (或 Docker 化设置)
*   Redis 实例 (用于缓存)

### 快速开始

1.  **克隆仓库:**
    ```bash
    git clone https://github.com/turtacn/starseek.git
    cd starseek
    ```

2.  **安装依赖:**
    ```bash
    go mod tidy
    ```

3.  **配置环境变量 (示例):**
    创建 `.env` 文件或设置环境变量：
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
    # 更多配置将在 docs/configuration.md 中详细说明
    ```

4.  **构建应用程序:**
    ```bash
    go build -o starseek ./cmd/starseek/main.go
    ```

5.  **运行应用程序:**
    ```bash
    ./starseek
    ```
    服务将在配置的地址（例如 `http://localhost:8080`）上监听。

### Docker 设置 (StarRocks 示例)

为了开发和测试，您可以使用 Docker Compose 启动一个 StarRocks 实例。
```bash
# 示例 docker-compose.yml (简化)
version: '3.8'
services:
  starrocks-fe:
    image: starrocks/starrocks:latest
    ports:
      - "9030:9030" # FE 查询端口
      - "8030:8030" # HTTP 端口
    environment:
      - FE_SERVERS="127.0.0.1:9030"
    command: ["/opt/starrocks/bin/start_fe.sh"]
  redis:
    image: redis:6-alpine
    ports:
      - "6379:6379"
````

运行：`docker-compose up -d`

## 🧪 代码片段与能力展示

### 1. 索引注册 (API 端点)

注册表的倒排索引列。这通常通过管理 API 完成。

```go
// 内部服务可能如何与索引注册中心交互的示例
// (实际 API 可能是 HTTP POST 到 /api/v1/indexes/register)
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
	// 初始化日志器
	l := logger.NewZapLogger()
	// 演示使用内存存储；实际持久化将使用数据库
	repo := inmemory.NewIndexMetadataRepository()
	service := application.NewIndexService(repo, l)

	ctx := context.Background()

	// 注册文档标题列
	err := service.RegisterIndex(ctx, index.RegisterIndexCommand{
		TableName:   "documents",
		ColumnName:  "title",
		IndexType:   enum.IndexTypeChinese,
		Tokenizer:   "jieba",
		DataType:    "VARCHAR",
		Description: "用于全文检索的文档标题",
	})
	if err != nil {
		fmt.Printf("注册 documents.title 索引失败: %v\n", err)
	} else {
		fmt.Println("成功注册 documents.title 索引")
	}

	// 注册产品描述列
	err = service.RegisterIndex(ctx, index.RegisterIndexCommand{
		TableName:   "products",
		ColumnName:  "description",
		IndexType:   enum.IndexTypeEnglish,
		Tokenizer:   "standard",
		DataType:    "TEXT",
		Description: "用于全文检索的产品描述",
	})
	if err != nil {
		fmt.Printf("注册 products.description 索引失败: %v\n", err)
	} else {
		fmt.Println("成功注册 products.description 索引")
	}

	// 列出所有已注册索引的示例
	allIndexes, err := service.ListIndexes(ctx, index.ListIndexesQuery{})
	if err != nil {
		fmt.Printf("列出索引失败: %v\n", err)
	} else {
		fmt.Println("\n已注册索引:")
		for _, idx := range allIndexes {
			fmt.Printf("- 表: %s, 列: %s, 类型: %s, 分词器: %s\n",
				idx.TableName, idx.ColumnName, idx.IndexType.String(), idx.Tokenizer)
		}
	}
}
```

### 2. 跨表全文检索 (HTTP API 示例)

通过 `curl` 命令模拟搜索请求。StarSeek 服务将处理此请求。

```bash
# 在 'documents' 和 'articles' 表的 'title' 和 'content' 字段中搜索 "人工智能"
# 请求可能类似于:
curl -X GET "http://localhost:8080/api/v1/search?q=人工智能&fields=title,content&tables=documents,articles&page=1&pageSize=10" \
     -H "Content-Type: application/json"

# 示例响应 (简化):
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
#       "data": { /* 原始行数据 */ }
#     },
#     {
#       "table": "articles",
#       "rowId": "art_456",
#       "score": 0.72,
#       "highlight": {
#         "content": "深度学习是<B>人工智能</B>的核心分支..."
#       },
#       "data": { /* 原始行数据 */ }
#     }
#   ],
#   "totalHits": 250,
#   "currentPage": 1,
#   "pageSize": 10
# }
```

### 3. 特定字段搜索

```bash
# 在 'documents' 表中，在 'title' 字段搜索 "starrocks"，在 'content' 字段搜索 "performance"
curl -X GET "http://localhost:8080/api/v1/search?q=title:starrocks AND content:performance&tables=documents&page=1&pageSize=5" \
     -H "Content-Type: application/json"
```

## 🤝 贡献指南

我们欢迎社区贡献！如果您有兴趣改进 StarSeek，请查阅我们的：

* [贡献指南](CONTRIBUTING.md)
* [行为准则](CODE_OF_CONDUCT.md)

---

[English Version README.md](README.md)
