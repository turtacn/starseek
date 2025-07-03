package metrics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	ferrors "github.com/turtacn/starseek/internal/common/errors" // 引入自定义错误包
	"github.com/turtacn/starseek/internal/infrastructure/logger" // 引入日志包
)

// Metrics 结构体定义了应用需要暴露的所有 Prometheus 指标。
type Metrics struct {
	// HTTPRequestTotal 是一个 Counter，用于记录 HTTP 请求的总数。
	// Labels: method (GET, POST等), path (请求路径，如 /api/v1/search), code (HTTP状态码)
	HTTPRequestTotal *prometheus.CounterVec

	// HTTPRequestDuration 是一个 Histogram，用于记录 HTTP 请求的处理延迟。
	// Labels: method, path
	HTTPRequestDuration *prometheus.HistogramVec

	// HTTPInFlightRequests 是一个 Gauge，用于记录当前正在处理的 HTTP 请求数量。
	HTTPInFlightRequests prometheus.Gauge

	// DBQueryTotal 是一个 Counter，用于记录数据库查询的总数。
	// Labels: operation (查询类型，如 SELECT, INSERT), table (表名), status (success, error)
	DBQueryTotal *prometheus.CounterVec

	// DBQueryDuration 是一个 Histogram，用于记录数据库查询的处理延迟。
	// Labels: operation, table
	DBQueryDuration *prometheus.HistogramVec

	// Business specific metrics could be added here, e.g.:
	// SearchQueryTotal *prometheus.CounterVec // Label: status (success, no_result, error)
	// CacheHitRatio    prometheus.Gauge
}

// NewMetrics 初始化并注册所有 Prometheus 指标。
// 它接收一个 logger 接口，用于在指标注册失败时记录日志。
func NewMetrics(log logger.Logger) (*Metrics, error) {
	m := &Metrics{
		HTTPRequestTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests.",
			},
			[]string{"method", "path", "code"},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds.",
				Buckets: prometheus.DefBuckets, // 默认的度量桶
			},
			[]string{"method", "path"},
		),
		HTTPInFlightRequests: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "http_in_flight_requests",
			Help: "Current number of in-flight HTTP requests.",
		}),
		DBQueryTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "db_queries_total",
				Help: "Total number of database queries.",
			},
			[]string{"operation", "table", "status"},
		),
		DBQueryDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "db_query_duration_seconds",
				Help:    "Duration of database queries in seconds.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation", "table"},
		),
	}

	// 注册所有指标到默认的 Prometheus 注册表。
	// 如果注册失败，通常是因为指标名称冲突，这在大型应用中可能会发生。
	// 通常在应用启动时一次性注册，失败则视为严重错误。
	collectors := []prometheus.Collector{
		m.HTTPRequestTotal,
		m.HTTPRequestDuration,
		m.HTTPInFlightRequests,
		m.DBQueryTotal,
		m.DBQueryDuration,
	}

	for _, collector := range collectors {
		if err := prometheus.DefaultRegisterer.Register(collector); err != nil {
			// 如果注册失败，记录日志并返回错误
			log.Error(fmt.Sprintf("Failed to register Prometheus collector: %v", err), zap.String("collector_name", collector.Describe()[0].GetName()))
			return nil, ferrors.NewInternalError(fmt.Sprintf("failed to register metrics: %v", err), err)
		}
	}

	log.Info("Prometheus metrics initialized and registered.")
	return m, nil
}

// RegisterMetricsEndpoint 暴露 /metrics HTTP 端点，用于 Prometheus 抓取。
func RegisterMetricsEndpoint(router *gin.Engine) {
	// 使用 promhttp.Handler() 提供 Prometheus 指标的 HTTP 处理器。
	// 建议在独立的端口或单独的路由组下暴露 /metrics 接口，以避免与业务接口混淆。
	// 这里为了简单，直接添加到主路由。
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
}

// --- Helper methods to update metrics ---

// IncRequestCount 增加 HTTP 请求总数。
func (m *Metrics) IncRequestCount(method, path string, statusCode int) {
	m.HTTPRequestTotal.WithLabelValues(method, path, fmt.Sprintf("%d", statusCode)).Inc()
}

// ObserveRequestDuration 记录 HTTP 请求处理延迟。
func (m *Metrics) ObserveRequestDuration(method, path string, duration time.Duration) {
	m.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

// IncInFlightRequests 增加正在处理的 HTTP 请求数量。
func (m *Metrics) IncInFlightRequests() {
	m.HTTPInFlightRequests.Inc()
}

// DecInFlightRequests 减少正在处理的 HTTP 请求数量。
func (m *Metrics) DecInFlightRequests() {
	m.HTTPInFlightRequests.Dec()
}

// IncDBQueryCount 增加数据库查询总数。
func (m *Metrics) IncDBQueryCount(operation, table, status string) {
	m.DBQueryTotal.WithLabelValues(operation, table, status).Inc()
}

// ObserveDBQueryDuration 记录数据库查询处理延迟。
func (m *Metrics) ObserveDBQueryDuration(operation, table string, duration time.Duration) {
	m.DBQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}
