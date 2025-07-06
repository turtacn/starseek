package interfaces

import (
	"context"
	"time"

	"github.com/turtacn/starseek/internal/core/domain"
)

// SearchService 定义搜索业务逻辑接口
type SearchService interface {
	// Search 执行搜索查询
	Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error)

	// SearchWithHighlight 执行搜索并返回高亮结果
	SearchWithHighlight(ctx context.Context, req *SearchRequest) (*SearchResponse, error)

	// SuggestQuery 提供查询建议
	SuggestQuery(ctx context.Context, query string, limit int) ([]string, error)

	// ValidateQuery 验证查询语法
	ValidateQuery(ctx context.Context, query string) error

	// SearchHistory 获取搜索历史
	SearchHistory(ctx context.Context, userID string, limit int) ([]*SearchRecord, error)
}

// IndexService 定义索引管理业务逻辑接口
type IndexService interface {
	// RegisterIndex 注册新索引
	RegisterIndex(ctx context.Context, config *IndexConfig) error

	// UpdateIndexConfig 更新索引配置
	UpdateIndexConfig(ctx context.Context, indexName string, config *IndexConfig) error

	// DeleteIndex 删除索引
	DeleteIndex(ctx context.Context, indexName string) error

	// GetIndexMetadata 获取索引元数据
	GetIndexMetadata(ctx context.Context, indexName string) (*IndexMetadata, error)

	// ListIndexes 列出所有索引
	ListIndexes(ctx context.Context) ([]*IndexMetadata, error)

	// OptimizeIndex 优化索引
	OptimizeIndex(ctx context.Context, indexName string) error

	// GetIndexStats 获取索引统计信息
	GetIndexStats(ctx context.Context, indexName string) (*IndexStats, error)

	// BackupIndex 备份索引
	BackupIndex(ctx context.Context, indexName string, backupPath string) error

	// RestoreIndex 恢复索引
	RestoreIndex(ctx context.Context, backupPath string, indexName string) error
}

// RankingService 定义排名算法逻辑接口
type RankingService interface {
	// CalculateTFIDF 计算TF-IDF评分
	CalculateTFIDF(ctx context.Context, req *TFIDFRequest) (*RankingScore, error)

	// CalculateBM25 计算BM25评分
	CalculateBM25(ctx context.Context, req *BM25Request) (*RankingScore, error)

	// CalculateCustomScore 计算自定义评分
	CalculateCustomScore(ctx context.Context, req *CustomScoreRequest) (*RankingScore, error)

	// RankDocuments 对文档进行排序
	RankDocuments(ctx context.Context, documents []*domain.Document, query string, algorithm RankingAlgorithm) ([]*RankedDocument, error)

	// UpdateRankingModel 更新排名模型
	UpdateRankingModel(ctx context.Context, modelConfig *RankingModelConfig) error

	// GetRankingAlgorithms 获取可用的排名算法
	GetRankingAlgorithms(ctx context.Context) ([]RankingAlgorithm, error)

	// EvaluateRanking 评估排名质量
	EvaluateRanking(ctx context.Context, req *RankingEvaluationRequest) (*RankingEvaluation, error)
}

// TaskService 定义任务调度逻辑接口
type TaskService interface {
	// SubmitTask 提交任务
	SubmitTask(ctx context.Context, task *Task) (*TaskResult, error)

	// SubmitBatchTasks 提交批量任务
	SubmitBatchTasks(ctx context.Context, tasks []*Task) ([]*TaskResult, error)

	// GetTaskStatus 获取任务状态
	GetTaskStatus(ctx context.Context, taskID string) (*TaskStatus, error)

	// CancelTask 取消任务
	CancelTask(ctx context.Context, taskID string) error

	// ListTasks 列出任务
	ListTasks(ctx context.Context, filter *TaskFilter) ([]*TaskInfo, error)

	// ExecuteParallelTasks 并发执行任务
	ExecuteParallelTasks(ctx context.Context, tasks []*Task, maxConcurrency int) ([]*TaskResult, error)

	// MergeTaskResults 合并任务结果
	MergeTaskResults(ctx context.Context, results []*TaskResult, strategy MergeStrategy) (*MergedResult, error)

	// SchedulePeriodicTask 调度周期性任务
	SchedulePeriodicTask(ctx context.Context, task *PeriodicTask) error

	// RemovePeriodicTask 移除周期性任务
	RemovePeriodicTask(ctx context.Context, taskID string) error
}

// SearchRequest 搜索请求结构
type SearchRequest struct {
	Query      string            `json:"query"`
	Index      string            `json:"index"`
	Filters    map[string]string `json:"filters,omitempty"`
	SortBy     string            `json:"sort_by,omitempty"`
	SortOrder  string            `json:"sort_order,omitempty"`
	Offset     int               `json:"offset,omitempty"`
	Limit      int               `json:"limit,omitempty"`
	Highlight  bool              `json:"highlight,omitempty"`
	UserID     string            `json:"user_id,omitempty"`
	SearchMode SearchMode        `json:"search_mode,omitempty"`
}

// SearchResponse 搜索响应结构
type SearchResponse struct {
	Results     []*SearchResult `json:"results"`
	Total       int64           `json:"total"`
	Duration    time.Duration   `json:"duration"`
	Suggestions []string        `json:"suggestions,omitempty"`
	Facets      map[string]int  `json:"facets,omitempty"`
}

// SearchResult 搜索结果结构
type SearchResult struct {
	DocumentID string                 `json:"document_id"`
	Title      string                 `json:"title"`
	Content    string                 `json:"content"`
	Score      float64                `json:"score"`
	Highlights map[string][]string    `json:"highlights,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Source     string                 `json:"source,omitempty"`
}

// SearchRecord 搜索记录结构
type SearchRecord struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Query     string    `json:"query"`
	Index     string    `json:"index"`
	Results   int       `json:"results"`
	Duration  int64     `json:"duration"`
	Timestamp time.Time `json:"timestamp"`
}

// IndexConfig 索引配置结构
type IndexConfig struct {
	Name        string            `json:"name"`
	Type        IndexType         `json:"type"`
	Settings    map[string]string `json:"settings"`
	Mappings    map[string]string `json:"mappings"`
	Shards      int               `json:"shards"`
	Replicas    int               `json:"replicas"`
	RefreshTime time.Duration     `json:"refresh_time"`
}

// IndexMetadata 索引元数据结构
type IndexMetadata struct {
	Name          string            `json:"name"`
	Type          IndexType         `json:"type"`
	Status        IndexStatus       `json:"status"`
	DocumentCount int64             `json:"document_count"`
	Size          int64             `json:"size"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	Settings      map[string]string `json:"settings"`
}

// IndexStats 索引统计信息结构
type IndexStats struct {
	DocumentCount int64                  `json:"document_count"`
	IndexSize     int64                  `json:"index_size"`
	QueryCount    int64                  `json:"query_count"`
	AvgQueryTime  time.Duration          `json:"avg_query_time"`
	Performance   map[string]interface{} `json:"performance"`
}

// TFIDFRequest TF-IDF计算请求
type TFIDFRequest struct {
	Query      string   `json:"query"`
	DocumentID string   `json:"document_id"`
	Terms      []string `json:"terms"`
	Index      string   `json:"index"`
}

// BM25Request BM25计算请求
type BM25Request struct {
	Query      string  `json:"query"`
	DocumentID string  `json:"document_id"`
	K1         float64 `json:"k1"`
	B          float64 `json:"b"`
	Index      string  `json:"index"`
}

// CustomScoreRequest 自定义评分请求
type CustomScoreRequest struct {
	Query      string                 `json:"query"`
	DocumentID string                 `json:"document_id"`
	Parameters map[string]interface{} `json:"parameters"`
	Algorithm  string                 `json:"algorithm"`
}

// RankingScore 排名评分结构
type RankingScore struct {
	DocumentID string             `json:"document_id"`
	Score      float64            `json:"score"`
	Details    map[string]float64 `json:"details,omitempty"`
}

// RankedDocument 排序后的文档结构
type RankedDocument struct {
	Document *domain.Document `json:"document"`
	Score    float64          `json:"score"`
	Rank     int              `json:"rank"`
}

// RankingModelConfig 排名模型配置
type RankingModelConfig struct {
	Algorithm  RankingAlgorithm       `json:"algorithm"`
	Parameters map[string]interface{} `json:"parameters"`
	Weights    map[string]float64     `json:"weights,omitempty"`
}

// RankingEvaluationRequest 排名评估请求
type RankingEvaluationRequest struct {
	Query        string             `json:"query"`
	Results      []*RankedDocument  `json:"results"`
	RelevantDocs []string           `json:"relevant_docs"`
	Metrics      []EvaluationMetric `json:"metrics"`
}

// RankingEvaluation 排名评估结果
type RankingEvaluation struct {
	Precision float64            `json:"precision"`
	Recall    float64            `json:"recall"`
	F1Score   float64            `json:"f1_score"`
	NDCG      float64            `json:"ndcg"`
	MAP       float64            `json:"map"`
	Details   map[string]float64 `json:"details"`
}

// Task 任务结构
type Task struct {
	ID         string                 `json:"id"`
	Type       TaskType               `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
	Priority   TaskPriority           `json:"priority"`
	Timeout    time.Duration          `json:"timeout"`
	Retries    int                    `json:"retries"`
	Metadata   map[string]string      `json:"metadata,omitempty"`
}

// TaskResult 任务结果结构
type TaskResult struct {
	TaskID    string                 `json:"task_id"`
	Status    TaskResultStatus       `json:"status"`
	Result    interface{}            `json:"result,omitempty"`
	Error     string                 `json:"error,omitempty"`
	StartTime time.Time              `json:"start_time"`
	EndTime   time.Time              `json:"end_time"`
	Duration  time.Duration          `json:"duration"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// TaskStatus 任务状态结构
type TaskStatus struct {
	ID        string           `json:"id"`
	Status    TaskResultStatus `json:"status"`
	Progress  float64          `json:"progress"`
	Message   string           `json:"message,omitempty"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// TaskInfo 任务信息结构
type TaskInfo struct {
	ID        string           `json:"id"`
	Type      TaskType         `json:"type"`
	Status    TaskResultStatus `json:"status"`
	Priority  TaskPriority     `json:"priority"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// TaskFilter 任务过滤器
type TaskFilter struct {
	Status    []TaskResultStatus `json:"status,omitempty"`
	Type      []TaskType         `json:"type,omitempty"`
	Priority  []TaskPriority     `json:"priority,omitempty"`
	StartTime *time.Time         `json:"start_time,omitempty"`
	EndTime   *time.Time         `json:"end_time,omitempty"`
	Limit     int                `json:"limit,omitempty"`
	Offset    int                `json:"offset,omitempty"`
}

// MergedResult 合并结果结构
type MergedResult struct {
	Data     interface{}            `json:"data"`
	Count    int                    `json:"count"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// PeriodicTask 周期性任务结构
type PeriodicTask struct {
	ID       string        `json:"id"`
	Task     *Task         `json:"task"`
	Interval time.Duration `json:"interval"`
	NextRun  time.Time     `json:"next_run"`
	Enabled  bool          `json:"enabled"`
}

// 枚举类型定义
type (
	SearchMode       string
	IndexType        string
	IndexStatus      string
	RankingAlgorithm string
	EvaluationMetric string
	TaskType         string
	TaskPriority     string
	TaskResultStatus string
	MergeStrategy    string
)

// SearchMode 常量
const (
	SearchModeExact    SearchMode = "exact"
	SearchModeFuzzy    SearchMode = "fuzzy"
	SearchModePrefix   SearchMode = "prefix"
	SearchModeWildcard SearchMode = "wildcard"
)

// IndexType 常量
const (
	IndexTypeText     IndexType = "text"
	IndexTypeNumeric  IndexType = "numeric"
	IndexTypeKeyword  IndexType = "keyword"
	IndexTypeGeoPoint IndexType = "geo_point"
)

// IndexStatus 常量
const (
	IndexStatusActive     IndexStatus = "active"
	IndexStatusInactive   IndexStatus = "inactive"
	IndexStatusOptimizing IndexStatus = "optimizing"
	IndexStatusError      IndexStatus = "error"
)

// RankingAlgorithm 常量
const (
	RankingAlgorithmTFIDF  RankingAlgorithm = "tfidf"
	RankingAlgorithmBM25   RankingAlgorithm = "bm25"
	RankingAlgorithmCustom RankingAlgorithm = "custom"
)

// EvaluationMetric 常量
const (
	MetricPrecision EvaluationMetric = "precision"
	MetricRecall    EvaluationMetric = "recall"
	MetricF1Score   EvaluationMetric = "f1_score"
	MetricNDCG      EvaluationMetric = "ndcg"
	MetricMAP       EvaluationMetric = "map"
)

// TaskType 常量
const (
	TaskTypeIndex    TaskType = "index"
	TaskTypeSearch   TaskType = "search"
	TaskTypeOptimize TaskType = "optimize"
	TaskTypeBackup   TaskType = "backup"
	TaskTypeRestore  TaskType = "restore"
	TaskTypeAnalyze  TaskType = "analyze"
)

// TaskPriority 常量
const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityNormal TaskPriority = "normal"
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityUrgent TaskPriority = "urgent"
)

// TaskResultStatus 常量
const (
	TaskStatusPending   TaskResultStatus = "pending"
	TaskStatusRunning   TaskResultStatus = "running"
	TaskStatusCompleted TaskResultStatus = "completed"
	TaskStatusFailed    TaskResultStatus = "failed"
	TaskStatusCancelled TaskResultStatus = "cancelled"
)

// MergeStrategy 常量
const (
	MergeStrategyUnion        MergeStrategy = "union"
	MergeStrategyIntersection MergeStrategy = "intersection"
	MergeStrategyWeighted     MergeStrategy = "weighted"
	MergeStrategyRanked       MergeStrategy = "ranked"
)

//Personal.AI order the ending
