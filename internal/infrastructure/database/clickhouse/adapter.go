package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2" // 引入 ClickHouse 驱动
	"go.uber.org/zap"                          // 用于日志字段

	ferrors "github.com/turtacn/starseek/internal/common/errors"   // 引入自定义错误包
	"github.com/turtacn/starseek/internal/config"                  // 引入配置包
	"github.com/turtacn/starseek/internal/infrastructure/database" // 引入 DBClient 接口
	"github.com/turtacn/starseek/internal/infrastructure/logger"   // 引入日志接口
)

// clickhouseAdapter 是 DBClient 接口的 ClickHouse 实现。
// 它通过标准库的 *sql.DB 包装了与 ClickHouse 的连接。
type clickhouseAdapter struct {
	db  *sql.DB
	log logger.Logger
}

// NewClickHouseAdapter 创建一个新的 ClickHouse 客户端适配器实例。
// 它根据提供的配置初始化 ClickHouse 数据库连接池，并进行连接测试。
// ClickHouse 默认原生协议端口是 9000。
func NewClickHouseAdapter(cfg *config.DatabaseConfig, log logger.Logger) (database.DBClient, error) {
	if cfg == nil {
		return nil, ferrors.NewInternalError("database config is nil", nil)
	}
	if log == nil {
		return nil, ferrors.NewInternalError("logger is nil", nil)
	}
	// 确保端口设置正确，ClickHouse 默认原生协议端口是 9000
	if cfg.Port == "" {
		cfg.Port = "9000" // ClickHouse 默认原生协议端口
		log.Warn("ClickHouse port not specified in config, defaulting to 9000")
	}

	// ClickHouse DSN (Data Source Name) 格式
	// clickhouse://[user:password@]host:port[/database][?param1=value1&...]
	// 这里不直接设置超时参数在DSN中，而是通过SetConnMaxLifetime等
	dsn := fmt.Sprintf("clickhouse://%s:%s@%s:%s/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)
	// 可以添加更多参数，例如：
	// dsn := fmt.Sprintf("clickhouse://%s:%s@%s:%s/%s?dial_timeout=%d&read_timeout=%d&write_timeout=%d",
	//    cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName,
	//    cfg.ConnectTimeoutSeconds, cfg.ReadTimeoutSeconds, cfg.WriteTimeoutSeconds)

	// 使用 "clickhouse" 驱动来连接 ClickHouse
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		log.Error("Failed to open ClickHouse connection", zap.Error(err),
			zap.String("host", cfg.Host), zap.String("port", cfg.Port), zap.String("db_name", cfg.DBName))
		return nil, ferrors.NewExternalServiceError(fmt.Sprintf("failed to open ClickHouse connection: %v", err), err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetimeSeconds) * time.Second)
	db.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTimeSeconds) * time.Second)

	// 尝试连接数据库，验证配置是否正确
	// 注意：clickhouse-go/v2 驱动的 PingContext 可能会更慢，因为它涉及到网络往返
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ConnectTimeoutSeconds)*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		db.Close() // 如果 ping 失败，立即关闭连接
		log.Error("Failed to ping ClickHouse database", zap.Error(err),
			zap.String("host", cfg.Host), zap.String("port", cfg.Port), zap.String("db_name", cfg.DBName))
		return nil, ferrors.NewExternalServiceError(fmt.Sprintf("failed to ping ClickHouse database: %v", err), err)
	}

	log.Info("Successfully connected to ClickHouse database",
		zap.String("db_name", cfg.DBName),
		zap.String("host", cfg.Host),
		zap.String("port", cfg.Port),
		zap.Int("max_open_conns", cfg.MaxOpenConns),
		zap.Int("max_idle_conns", cfg.MaxIdleConns))

	return &clickhouseAdapter{db: db, log: log}, nil
}

// Query implements DBClient.Query for ClickHouse.
func (c *clickhouseAdapter) Query(ctx context.Context, sql string, args ...interface{}) (*sql.Rows, error) {
	c.log.Debug("Executing ClickHouse Query", zap.String("sql", sql), zap.Any("args", args))
	return c.db.QueryContext(ctx, sql, args...)
}

// Exec implements DBClient.Exec for ClickHouse.
// 注意：ClickHouse 对某些操作（如 INSERT INTO SELECT）可能不会返回 RowsAffected。
// clickhouse-go/v2 驱动在执行 INSERT/SELECT 等操作时，RowsAffected 可能会是 0，
// 但对于 DDL (CREATE/DROP TABLE) 或 ALTER 等操作，通常能正确反映。
func (c *clickhouseAdapter) Exec(ctx context.Context, sql string, args ...interface{}) (sql.Result, error) {
	c.log.Debug("Executing ClickHouse Exec", zap.String("sql", sql), zap.Any("args", args))
	return c.db.ExecContext(ctx, sql, args...)
}

// Ping implements DBClient.Ping for ClickHouse.
func (c *clickhouseAdapter) Ping(ctx context.Context) error {
	c.log.Debug("Pinging ClickHouse database")
	return c.db.PingContext(ctx)
}

// Close implements DBClient.Close for ClickHouse.
func (c *clickhouseAdapter) Close() error {
	c.log.Info("Closing ClickHouse database connection...")
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}
