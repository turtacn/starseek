# StarSeek 项目总体架构设计

## 1. 引言

`StarSeek` 项目旨在构建一个统一的全文检索中台服务，以增强 StarRocks、Apache Doris 和 ClickHouse 等列式分析型数据库的全文检索能力。本项目将弥补这些数据库在高级搜索功能（如跨表查询、相关度排序、高亮显示、同义词扩展等）方面的不足，提供类似 Elasticsearch 的搜索体验，同时避免引入和维护独立的搜索集群所带来的复杂性与成本。本文档将详细阐述 `StarSeek` 的总体架构设计、核心模块、数据流转、部署策略以及代码生成蓝图，旨在构建一个高性能、高可用、可扩展且易于维护的系统。

## 2. 领域内的 DFX 问题全景

在设计 `StarSeek` 之前，我们深入分析了当前分析型数据库在全文检索领域面临的各项设计（DFX）问题：

### 2.1 功能性（Functionality）问题

*   **功能受限**: 现有分析型数据库的倒排索引主要用于加速精确匹配查询，缺乏对高级搜索语义的支持，例如短语搜索、模糊搜索、拼写纠错等。
*   **跨表/跨列搜索缺失**: 无法原生支持在一个查询中同时搜索多个表或多个列的倒排索引，需要手动编写复杂的 `UNION ALL` SQL。
*   **相关性排序缺失**: 缺乏内置的文本相关性排序算法（如 TF-IDF、BM25），导致搜索结果无法按相关度排序，影响用户体验。
*   **元数据管理分散**: 倒排索引的元信息（如分词器、索引类型、字段类型）分散在各个表的 DDL 中，难以统一管理和发现。
*   **同义词/联想词缺失**: 缺乏同义词扩展和搜索联想功能，降低了搜索召回率和用户便利性。

### 2.2 性能（Performance）问题

*   **应用层二次处理开销**: 若在应用层实现排名、高亮等功能，需要将大量原始数据从数据库拉取到服务层进行处理，增加网络 I/O 和 CPU 负担。
*   **高并发压力**: 在高并发搜索场景下，直接向底层数据库发起大量复杂 SQL 查询可能导致数据库过载。
*   **SQL 生成复杂度**: 动态构建复杂的跨表/跨列 SQL 语句本身存在性能开销，且易出错。
*   **缓存机制不足**: 缺乏针对全文检索场景的特定缓存策略，导致重复查询性能低下。

### 2.3 可靠性与容错性（Reliability & Fault Tolerance）问题

*   **数据库单点故障**: 搜索服务强依赖底层数据库，数据库故障可能导致整个搜索服务不可用。
*   **网络分区**: 服务与数据库之间的网络不稳定可能导致查询失败。
*   **并发控制风险**: 未经优化的并发查询可能耗尽数据库连接或资源，导致雪崩效应。
*   **数据一致性**: 索引元数据与实际数据库索引可能不一致，导致查询错误。

### 2.4 可伸缩性（Scalability）问题

*   **垂直扩展限制**: 单个 `StarSeek` 服务实例可能成为瓶颈，需要支持水平扩展。
*   **底层数据库扩展挑战**: 依赖底层数据库的扩展能力，但 `StarSeek` 服务本身也需能处理更多并发和数据量。
*   **元数据同步**: 随着表和索引数量的增长，元数据的发现和管理机制需要具备良好的可伸缩性。

### 2.5 可观测性（Observability）问题

*   **缺乏统一日志**: 各模块日志分散，难以追踪请求的完整生命周期。
*   **监控指标不足**: 缺乏关键性能指标（QPS、延迟、错误率等）的收集和展示。
*   **链路追踪缺失**: 难以诊断跨模块、跨服务的请求调用路径和性能瓶颈。
*   **报警机制不完善**: 无法及时发现和响应服务异常。

### 2.6 可维护性（Maintainability）与开发效率（Developer Productivity）问题

*   **代码结构混乱**: 若无清晰的架构指导，代码可能耦合严重，难以理解和修改。
*   **SQL 方言差异**: 支持多数据库意味着需要处理不同 SQL 方言的兼容性问题，增加了开发复杂性。
*   **分词器管理**: 不同的分词器实现和版本管理可能导致混乱。
*   **测试覆盖**: 缺乏有效的测试策略，难以保证功能变更的正确性。
*   **依赖管理**: Go 模块依赖管理不当可能导致版本冲突和构建问题。

## 3. 解决方案全景

针对上述 DFX 问题，`StarSeek` 提出了以下解决方案策略：

### 3.1 架构策略

*   **分层架构（Layered Architecture）**: 采用经典的四层架构：展现层（Presentation Layer）、应用层（Application Layer）、领域层（Domain Layer）和基础设施层（Infrastructure Layer）。
    *   **展现层**: 负责处理 HTTP 请求和响应，进行参数校验和格式转换。
    *   **应用层**: 协调领域服务，处理业务流程，但不包含业务逻辑本身。
    *   **领域层**: 包含核心业务逻辑和领域模型，定义服务接口和实体。
    *   **基础设施层**: 提供技术支持，如数据库访问、缓存、日志、监控、分词器集成等。
*   **模块化与高内聚低耦合**: 将系统划分为独立的业务模块（如索引元信息、查询处理、排名、任务调度等），每个模块内部高内聚，模块间通过清晰的接口进行低耦合通信。
*   **面向接口编程（Interface-Oriented Programming）**: 广泛使用 Go 语言的接口特性，定义服务契约，实现依赖倒置原则，便于模块替换和单元测试。
*   **适配器模式（Adapter Pattern）**: 在基础设施层为不同的数据库（StarRocks、Doris、ClickHouse）提供统一的数据库接口，将它们的SQL方言差异封装起来。
*   **统一配置管理**: 使用 Viper 等库实现灵活的配置加载，支持环境变量、配置文件等多种方式。

### 3.2 功能性解决方案

*   **索引元信息管理**:
    *   **Registry 服务**: 建立独立的 `IndexRegistry` 服务，集中存储和管理所有表、列的索引元信息。
    *   **元数据发现/注册**: 支持通过 API 手动注册，或未来考虑通过解析数据库 `SHOW CREATE TABLE` 语句或监听 DDL 事件进行自动发现。
    *   **元数据存储**: 使用关系型数据库（如 MySQL/PostgreSQL）或 NoSQL 数据库（如 MongoDB）存储元数据，或简单起见先使用内存或文件存储。
*   **查询处理**:
    *   **统一查询解析器**: 将用户请求解析为内部统一的查询表达式树（Query Expression Tree）。
    *   **多语言分词器集成**: 内置或集成多种分词库（如 GoJieba、segment），根据索引元数据中的分词器信息选择对应分词器，确保与底层数据库索引分词策略一致。
    *   **同义词词典服务**: 引入可配置的同义词词典，在分词后进行词汇扩展。
    *   **SQL 生成器**: 基于查询表达式树和索引元数据，动态生成优化的 SQL 语句，包括 `MATCH_AGAINST`、`UNION ALL`、`WHERE` 子句等。
*   **排名与高亮**:
    *   **离线统计服务**: 独立服务或脚本定期扫描数据库，统计 TF 和 IDF 所需的文档频率信息，并存储到 Redis 或独立数据库。
    *   **实时排名计算**: 在查询结果返回后，`StarSeek` 在应用层对每条记录进行二次处理，结合 TF（通过再次分词计算）和 IDF（从缓存/存储中获取）模拟计算 BM25/TF-IDF 分数。
    *   **高亮处理器**: 在应用层根据原始查询关键词，对返回的文本字段进行高亮处理（例如，插入 `<b>` 标签）。

### 3.3 性能解决方案

*   **缓存机制**:
    *   **热点关键词查询缓存**: 使用 Redis 缓存热门关键词的查询结果（例如，SQL 查询结果集或行ID），减少对底层数据库的重复查询。
    *   **元数据缓存**: 缓存索引元数据，避免频繁查询元数据存储。
*   **并发控制**:
    *   **协程池/连接池**: 使用 Go 协程池和数据库连接池，限制对后端数据库的并发访问，防止过载。
    *   **QPS 限流**: 在 API 网关层或应用层实现 QPS 限流，保护服务本身和后端数据库。
*   **SQL 优化**: `StarSeek` 生成的 SQL 语句将尽可能利用底层数据库的索引和优化器。

### 3.4 可靠性与容错性解决方案

*   **错误集中定义**: 统一的错误类型和错误码管理，提高错误处理的一致性。
*   **重试与熔断**: 对于外部服务调用（如数据库连接），引入重试和熔断机制。
*   **超时控制**: 对所有外部调用和耗时操作设置合理的超时。
*   **日志告警**: 结合监控系统，对关键错误和异常日志触发告警。

### 3.5 可伸缩性解决方案

*   **无状态服务设计**: `StarSeek` 服务本身设计为无状态，便于水平扩展，通过负载均衡器分发请求。
*   **缓存层可伸缩**: Redis 等缓存服务本身具备良好的伸缩性。
*   **数据库层可伸缩**: 依赖底层 StarRocks 等 MPP 数据库自身的分布式和伸缩能力。

### 3.6 可观测性解决方案

*   **统一日志**: 使用 Zap 等高性能日志库，输出结构化日志，并统一到 ELK Stack 或 Loki 等日志系统中。
*   **指标采集**: 使用 Prometheus SDK 暴露服务运行指标（QPS、延迟、错误率、CPU/内存使用等），通过 Prometheus 抓取并由 Grafana 展示。
*   **链路追踪**: 集成 OpenTelemetry，对请求进行分布式链路追踪，以便诊断跨模块、跨服务的调用链。

### 3.7 可维护性与开发效率解决方案

*   **Clean Code 原则**: 遵循 Go 语言的最佳实践，编写清晰、简洁、可读性强的代码。
*   **自动化测试**: 强制要求单元测试、集成测试，确保代码质量。
*   **Go Modules**: 规范化依赖管理。
*   **清晰的接口定义**: 模块间通过清晰的接口进行交互，降低理解和修改的复杂性。

## 4. 预期效果全景及其展望

### 4.1 预期效果

*   **提升用户搜索体验**: 用户可以通过更自然、更精准的关键词进行搜索，获得高相关度的结果，并支持分页和高亮显示。
*   **降低开发复杂度**: 应用开发者无需关心底层数据库的复杂 SQL 语法和分词细节，只需调用 `StarSeek` 的统一 API 即可实现高级搜索功能。
*   **数据平台能力增强**: 将 StarRocks 等分析型数据库的价值从纯粹的 OLAP 扩展到兼具实时分析和强大的全文检索能力。
*   **系统稳定性与性能提升**: 通过内置的缓存、并发控制和任务调度，有效提升搜索服务的响应速度和稳定性。
*   **未来可扩展性**: 良好的分层和模块化设计使得未来支持更多数据源、集成更多高级搜索功能（如向量搜索、知识图谱）变得容易。

### 4.2 展望

*   **更智能的查询理解**: 引入更先进的自然语言处理（NLP）技术，如查询意图识别、实体抽取，进一步提升搜索智能性。
*   **多模态搜索**: 结合向量数据库或向量嵌入技术，支持图片、视频等多媒体内容的语义搜索。
*   **实时索引同步**: 更紧密地与底层数据库集成，实现索引元数据和 IDF 统计的近实时同步。
*   **自助化配置界面**: 提供 Web UI 界面，方便用户管理索引元信息、配置分词器和同义词。
*   **A/B 测试框架**: 内置 A/B 测试支持，用于迭代优化排名算法和搜索策略。

## 5. 项目总体架构图

以下是 `StarSeek` 的总体架构图，展示了主要模块及其相互关系。

```mermaid
graph LR

    %% 模块命名规则: M_大写缩写[中文名称（English Term）]
    %% Presentation Layer
    subgraph PL[展现层（Presentation Layer）]
        M_API[REST API（RESTful API）] --> M_AUTH[认证鉴权（Authentication）]
        M_AUTH --> M_VALID[请求校验（Request Validation）]
    end

    %% Application Layer
    subgraph AL[应用层（Application Layer）]
        M_VALID --> A_QP[查询处理服务（Query Processing Service）]
        A_QP --> A_IS[索引服务（Index Service）]
        A_QP --> A_TS[任务调度服务（Task Scheduling Service）]
        A_QP --> A_RO[排名优化服务（Ranking & Optimization Service）]
        A_QP --> A_HL[高亮服务（Highlighting Service）]
    end

    %% Domain Layer
    subgraph DL[领域层（Domain Layer）]
        A_QP --> D_QBE[查询构建引擎（Query Building Engine）]
        A_IS --> D_IMR[索引元数据仓库（Index Metadata Repository）]
        A_TS --> D_TE[任务执行器（Task Executor）]
        A_RO --> D_RMS[排名模型服务（Ranking Model Service）]
        A_HL --> D_HLE[高亮引擎（Highlighting Engine）]

        D_QBE -- 依赖分词器 --> D_TK[分词器（Tokenizer）]
        D_QBE -- 依赖同义词库 --> D_SYN[同义词服务（Synonym Service）]
    end

    %% Infrastructure Layer
    subgraph IL[基础设施层（Infrastructure Layer）]
        D_IMR --> I_DBI[数据库接口（Database Interface）]
        D_TE --> I_DBI
        D_RMS --> I_REDIS[Redis缓存（Redis Cache）]
        D_RMS --> I_DBI
        D_TK --> I_EXTTK[外部分词库（External Tokenizer Libraries）]
        D_SYN --> I_DBI
        I_DBI -- 适配不同数据库 --> I_SR_ADAPTER[StarRocks适配器（StarRocks Adapter）]
        I_DBI -- 适配不同数据库 --> I_DR_ADAPTER[Doris适配器（Doris Adapter）]
        I_DBI -- 适配不同数据库 --> I_CK_ADAPTER[ClickHouse适配器（ClickHouse Adapter）]
        I_REDIS -- 依赖 --> I_REDIS_CLIENT[Redis客户端（Redis Client）]

        subgraph OBSERV[可观测性（Observability）]
            I_LOG[日志（Logging）]
            I_MET[指标（Metrics）]
            I_TRA[追踪（Tracing）]
        end
        I_LOG -- 用于 --> M_API
        I_LOG -- 用于 --> A_QP
        I_LOG -- 用于 --> D_QBE
        I_LOG -- 用于 --> I_DBI
        I_MET -- 用于 --> M_API
        I_MET -- 用于 --> A_QP
        I_TRA -- 用于 --> M_API
        I_TRA -- 用于 --> A_QP
    end

    %% Data Stores
    subgraph DS[数据存储（Data Stores）]
        DB_SR[StarRocks]
        DB_DR[Apache Doris]
        DB_CK[ClickHouse]
        DB_META[元数据DB（Metadata DB）]
        REDIS_CACHE[Redis缓存（Redis Cache Store）]
    end

    %% External Systems (Optional)
    subgraph ES[外部系统（External Systems）]
        PROM[Prometheus]
        GRAF[Grafana]
        ELK[ELK Stack/Loki]
        OTEL[OpenTelemetry Collector]
    end

    %% Connections
    I_SR_ADAPTER --> DB_SR
    I_DR_ADAPTER --> DB_DR
    I_CK_ADAPTER --> DB_CK
    I_DBI --> DB_META
    I_REDIS_CLIENT --> REDIS_CACHE

    I_LOG --> ELK
    I_MET --> PROM
    PROM --> GRAF
    I_TRA --> OTEL
    OTEL --> ELK 
    %% Or Jaeger/Zipkin

    style PL fill:#add8e6,stroke:#333,stroke-width:2px
    style AL fill:#87ceeb,stroke:#333,stroke-width:2px
    style DL fill:#4682b4,stroke:#333,stroke-width:2px
    style IL fill:#00bfff,stroke:#333,stroke-width:2px
    style DS fill:#6b8e23,stroke:#333,stroke-width:2px
    style ES fill:#ffa500,stroke:#333,stroke-width:2px
    style OBSERV fill:#b0c4de,stroke:#333,stroke-width:2px
````

**架构图说明:**

* **展现层（Presentation Layer）**: 对外提供 RESTful API 接口，负责请求的接收、认证鉴权和初步的请求校验。
* **应用层（Application Layer）**: 协调领域层服务，处理高级业务流程，如接收搜索请求，协调查询处理、索引管理、任务调度和结果排名优化等。
* **领域层（Domain Layer）**: 包含 `StarSeek` 的核心业务逻辑和领域模型。

  * **查询构建引擎（Query Building Engine）**: 负责解析用户查询，结合分词器和同义词服务，构建内部查询表达式，并最终生成针对不同数据源的 SQL。
  * **索引元数据仓库（Index Metadata Repository）**: 统一管理所有已配置倒排索引的元信息。
  * **任务执行器（Task Executor）**: 负责并发控制和 SQL 查询的执行。
  * **排名模型服务（Ranking Model Service）**: 实现类似 TF-IDF/BM25 的打分逻辑，结合预计算的 IDF 数据和实时计算的 TF 数据进行排名。
  * **高亮引擎（Highlighting Engine）**: 对搜索结果中的匹配关键词进行高亮处理。
  * **分词器（Tokenizer）**: 封装多种分词算法，根据需求选择。
  * **同义词服务（Synonym Service）**: 提供同义词扩展功能。
* **基础设施层（Infrastructure Layer）**: 提供技术支持和通用能力。

  * **数据库接口（Database Interface）**: 抽象数据库操作，通过适配器模式支持 StarRocks、Doris、ClickHouse。
  * **Redis 缓存（Redis Cache）**: 提供通用缓存能力，用于热点数据、元数据、IDF 统计等。
  * **可观测性（Observability）**: 统一的日志、指标和追踪机制，支持接入外部监控系统。
* **数据存储（Data Stores）**: 包含 `StarSeek` 依赖的各种数据存储。

  * `StarRocks`、`Apache Doris`、`ClickHouse`: 实际存储业务数据的分析型数据库。
  * `元数据 DB（Metadata DB）`: 存储索引注册信息、同义词词典、IDF 统计等 `StarSeek` 自身的元数据。
  * `Redis 缓存（Redis Cache Store）`: 提供高速缓存。
* **外部系统（External Systems）**: 外部监控与日志聚合系统，用于 `StarSeek` 的可观测性。


## 6. 参考项目
* [StarRocks](https://github.com/StarRocks/StarRocks)
* [ClickHouse](https://github.com/ClickHouse/ClickHouse)
* [doris](https://github.com/apache/doris)
* [elasticsearch](https://github.com/elastic/elasticsearch)