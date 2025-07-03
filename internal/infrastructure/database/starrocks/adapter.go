package starrocks

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql" // StarRocks 使用 MySQL 协议
	"go.uber.org/zap"                  // 用于日志字段

	ferrors "github.com/turtacn/starseek/internal/common/errors"   // 引入自定义错误包
	"github.com/turtacn/starseek/internal/config"                  // 引入配置包
	"github.com/turtacn/starseek/internal/infrastructure/database" // 引入 DBClient 接口
	"github.com/turtacn/starseek/internal/infrastructure/logger"   // 引入日志接口
)

// starrocksAdapter 是 DBClient 接口的 StarRocks 实现。
// 它通过标准库的 *sql.DB 包装了与 StarRocks 的连接。
type starrocksAdapter struct {
	db  *sql.DB
	log logger.Logger
}

// NewStarRocksAdapter 创建一个新的 StarRocks 客户端适配器实例。
// 它根据提供的配置初始化 StarRocks 数据库连接池，并进行连接测试。
// StarRocks 通常监听 9030 端口进行 MySQL 协议连接。
func NewStarRocksAdapter(cfg *config.DatabaseConfig, log logger.Logger) (database.DBClient, error) {
	if cfg == nil {
		return nil, ferrors.NewInternalError("database config is nil", nil)
	}
	if log == nil {
		return nil, ferrors.NewInternalError("logger is nil", nil)
	}
	// 确保端口设置正确，StarRocks 默认查询端口是 9030
	if cfg.Port == "" {
		cfg.Port = "9030" // StarRocks FE 默认查询端口
		log.Warn("StarRocks port not specified in config, defaulting to 9030")
	}

	// DSN (Data Source Name) 格式与 MySQL 相同
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=%s&timeout=%ds&readTimeout=%ds&writeTimeout=%ds",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		"Local", // 设置时区为本地，确保时间类型正确处理
		cfg.ConnectTimeoutSeconds,
		cfg.ReadTimeoutSeconds,
		cfg.WriteTimeoutSeconds,
	)

	// 使用 "mysql" 驱动来连接 StarRocks
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Error("Failed to open StarRocks connection", zap.Error(err),
			zap.String("host", cfg.Host), zap.String("port", cfg.Port), zap.String("db_name", cfg.DBName))
		return nil, ferrors.NewExternalServiceError(fmt.Sprintf("failed to open StarRocks connection: %v", err), err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetimeSeconds) * time.Second)
	db.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTimeSeconds) * time.Second)

	// 尝试连接数据库，验证配置是否正确
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ConnectTimeoutSeconds)*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		db.Close() // 如果 ping 失败，立即关闭连接
		log.Error("Failed to ping StarRocks database", zap.Error(err),
			zap.String("host", cfg.Host), zap.String("port", cfg.Port), zap.String("db_name", cfg.DBName))
		return nil, ferrors.NewExternalServiceError(fmt.Sprintf("failed to ping StarRocks database: %v", err), err)
	}

	log.Info("Successfully connected to StarRocks database",
		zap.String("db_name", cfg.DBName),
		zap.String("host", cfg.Host),
		zap.String("port", cfg.Port),
		zap.Int("max_open_conns", cfg.MaxOpenConns),
		zap.Int("max_idle_conns", cfg.MaxIdleConns))

	return &starrocksAdapter{db: db, log: log}, nil
}

// Query implements DBClient.Query for StarRocks.
func (s *starrocksAdapter) Query(ctx context.Context, sql string, args ...interface{}) (*sql.Rows, error) {
	s.log.Debug("Executing StarRocks Query", zap.String("sql", sql), zap.Any("args", args))
	return s.db.QueryContext(ctx, sql, args...)
}

// Exec implements DBClient.Exec for StarRocks.
func (s *starrocksAdapter) Exec(ctx context.Context, sql string, args ...interface{}) (sql.Result, error) {
	s.log.Debug("Executing StarRocks Exec", zap.String("sql", sql), zap.Any("args", args))
	return s.db.ExecContext(ctx, sql, args...)
}

// Ping implements DBClient.Ping for StarRocks.
func (s *starrocksAdapter) Ping(ctx context.Context) error {
	s.log.Debug("Pinging StarRocks database")
	return s.db.PingContext(ctx)
}

// Close implements DBClient.Close for StarRocks.
func (s *starrocksAdapter) Close() error {
	s.log.Info("Closing StarRocks database connection...")
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
