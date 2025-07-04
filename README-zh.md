# StarSeek - 数据仓库全文检索平台

[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](https://opensource.org/licenses/Apache-2.0)

[🇺🇸 English Version](./README.md)

## 项目简介

StarSeek 是一个企业级全文检索中台服务，专为 **StarRocks**、**ClickHouse** 和 **Doris** 等列式数据库设计。通过统一管理倒排索引并提供高级搜索功能，StarSeek 在传统数据仓库和现代搜索引擎之间架起了桥梁。

## 核心痛点与价值

### 解决的核心痛点

- **检索分散化**：各表倒排索引独立管理，缺乏统一的搜索能力
- **功能局限性**：原生数据库缺乏排序、高亮、跨表搜索等高级功能
- **查询构建复杂**：跨多表高效全文查询需要深度SQL专业知识
- **性能瓶颈**：大型UNION ALL查询和缺乏搜索专用优化

### 核心价值主张

- **🎯 统一搜索接口**：单一API端点搜索所有索引列和表
- **🚀 性能增强**：智能查询优化、缓存和并发执行
- **⭐ 高级特性**：排序算法、分页、高亮和相关度评分
- **🔧 易于集成**：RESTful API，学习成本最小
- **📊 多数据库支持**：原生支持StarRocks、ClickHouse和Doris

## 主要功能特性

### 🔍 全文搜索引擎
- 跨表、跨列搜索能力
- 多语言分词（中文、英文、多语言）
- 同义词扩展和字段特定过滤
- 布尔查询支持（`title:AI AND content:机器学习`）

### 📈 智能排序
- TF-IDF相关度评分模拟
- 可配置排序算法
- 自定义评分因子和权重参数

### ⚡ 性能优化
- 基于Redis的倒排索引缓存
- Bitmap加速行过滤
- 并发查询执行与流量控制
- 热门关键词预加载

### 🎛️ 管理与监控
- 集中化索引注册表和元数据管理
- 实时查询分析和性能指标
- 全面的日志记录和分布式追踪

## 快速开始

### 前置条件
- Go 1.20.2 或更高版本
- Redis 6.0+
- 支持的数据库之一（StarRocks/ClickHouse/Doris）

### 安装部署

```bash
# 克隆仓库
git clone https://github.com/turtacn/starseek.git
cd starseek

# 安装依赖
go mod download

# 构建应用
make build

# 启动服务
./bin/starseek --config=config/config.yaml
````

### 配置示例

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

## API 使用示例

### 基础搜索

```bash
# 搜索所有索引列
curl "http://localhost:8080/api/v1/search?q=人工智能&limit=10"
```

```json
{
  "query": "人工智能",
  "total": 156,
  "results": [
    {
      "table": "articles",
      "row_id": 12345,
      "score": 0.95,
      "highlights": {
        "title": "<em>人工智能</em>的未来发展",
        "content": "AI和<em>机器学习</em>正在改变..."
      },
      "fields": {
        "title": "人工智能的未来发展",
        "content": "AI和机器学习正在改变各个行业...",
        "created_at": "2024-01-15T10:30:00Z"
      }
    }
  ],
  "execution_time": "45ms"
}
```

### 高级字段搜索

```bash
# 带字段过滤和布尔操作符的搜索
curl "http://localhost:8080/api/v1/search?q=title:AI AND content:深度学习&fields=title,content,author"
```

### Go SDK 使用

```go
package main

import (
    "context"
    "fmt"
    "github.com/turtacn/starseek/pkg/client"
)

func main() {
    // 初始化StarSeek客户端
    client := starseek.NewClient("http://localhost:8080")
    
    // 执行搜索
    req := &starseek.SearchRequest{
        Query:     "机器学习",
        Fields:    []string{"title", "content"},
        Limit:     10,
        Highlight: true,
    }
    
    resp, err := client.Search(context.Background(), req)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("找到 %d 条结果，耗时 %s\n", resp.Total, resp.ExecutionTime)
    for _, result := range resp.Results {
        fmt.Printf("表: %s, 评分: %.2f\n", result.Table, result.Score)
        fmt.Printf("标题: %s\n", result.Highlights["title"])
    }
}
```

## 架构概览

StarSeek 采用分层架构，职责清晰分离：

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   HTTP/gRPC     │    │   Web 控制台     │    │   Go SDK        │
│   REST API      │    │   (可选)         │    │   客户端        │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                        │                        │
         └────────────────────────┼────────────────────────┘
                                  │
┌─────────────────────────────────┴─────────────────────────────────┐
│                          应用层                                   │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │   查询处理器    │  │   索引注册表    │  │   任务调度器    │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
└───────────────────────────────────────────────────────────────────┘
         │                        │                        │
┌─────────────────────────────────┴─────────────────────────────────┐
│                          领域层                                   │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │   搜索引擎      │  │   排序引擎      │  │   缓存管理器    │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
└───────────────────────────────────────────────────────────────────┘
         │                        │                        │
┌─────────────────────────────────┴─────────────────────────────────┐
│                        基础设施层                                 │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │   StarRocks     │  │   ClickHouse    │  │     Redis       │  │
│  │     适配器      │  │     适配器      │  │     缓存        │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
└───────────────────────────────────────────────────────────────────┘
```

详细架构文档请参见 [docs/architecture.md](docs/architecture.md)。

## 构建和运行

### 开发环境

```bash
# 安装开发依赖
make dev-deps

# 运行测试
make test

# 热重载运行
make dev

# 代码格式化
make fmt

# 生成API文档
make docs
```

### 生产部署

```bash
# 构建优化二进制文件
make build-prod

# 构建Docker镜像
docker build -t starseek:latest .

# 使用Docker Compose部署
docker-compose up -d
```

### 性能基准测试

| 场景      | QPS   | 平均延迟 | P99延迟 |
| ------- | ----- | ---- | ----- |
| 简单关键词搜索 | 1,000 | 15ms | 45ms  |
| 跨表搜索    | 500   | 35ms | 120ms |
| 复杂布尔查询  | 200   | 85ms | 250ms |

*基准测试环境：4核16GB内存，StarRocks集群*

## 贡献指南

我们欢迎贡献！详情请参见 [贡献指南](CONTRIBUTING.md)。

### 开发流程

1. Fork 仓库
2. 创建功能分支（`git checkout -b feature/amazing-feature`）
3. 提交更改（`git commit -m 'Add amazing feature'`）
4. 推送到分支（`git push origin feature/amazing-feature`）
5. 创建 Pull Request

### 代码规范

* 遵循 [Go代码审查注意事项](https://github.com/golang/go/wiki/CodeReviewComments)
* 为新功能添加测试
* 更新API变更的文档
* 确保所有测试通过（`make test`）

## 开源协议

本项目采用 Apache License 2.0 许可协议 - 详见 [LICENSE](LICENSE) 文件。

## 社区

* **文档**：[docs.starseek.io](https://docs.starseek.io)
* **问题反馈**：[GitHub Issues](https://github.com/turtacn/starseek/issues)
* **讨论交流**：[GitHub Discussions](https://github.com/turtacn/starseek/discussions)

## 致谢

* [StarRocks](https://github.com/StarRocks/StarRocks) - 高性能分析数据库
* [ClickHouse](https://github.com/ClickHouse/ClickHouse) - 快速列式数据库
* [Apache Doris](https://github.com/apache/doris) - 实时分析数据库
* [Elasticsearch](https://github.com/elastic/elasticsearch) - 搜索引擎设计参考

---

**如果觉得有用，请给我们一个 ⭐ Star！**