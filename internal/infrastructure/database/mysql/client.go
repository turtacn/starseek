package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql" // 引入 MySQL 驱动
	"go.uber.org/zap"                  // 用于日志字段

	"github.com/turtacn/starseek/internal/config"                  // 引入配置包
	"github.com/turtacn/starseek/internal/infrastructure/database" // 引入 DBClient 接口
	"github.com/turtacn/starseek/internal/infrastructure/logger"   // 引入日志接口
)

// mysqlClient 是 DBClient 接口的 MySQL 实现。
// 它包装了标准库的 *sql.DB，并提供了日志功能。
type mysqlClient struct {
	db  *sql.DB
	log logger.Logger
}

// NewMySQLClient 创建一个新的 MySQLClient 实例。
// 它根据提供的配置初始化 MySQL 数据库连接池，并进行连接测试。
func NewMySQLClient(cfg *config.DatabaseConfig, log logger.Logger) (database.DBClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("database config is nil")
	}
	if log == nil {
		return nil, fmt.Errorf("logger is nil")
	}

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

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Error("Failed to open MySQL connection", zap.Error(err))
		return nil, fmt.Errorf("failed to open MySQL connection: %w", err)
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
		log.Error("Failed to ping MySQL database", zap.Error(err),
			zap.String("host", cfg.Host), zap.String("port", cfg.Port), zap.String("db_name", cfg.DBName))
		return nil, fmt.Errorf("failed to ping MySQL database: %w", err)
	}

	log.Info("Successfully connected to MySQL database",
		zap.String("db_name", cfg.DBName),
		zap.String("host", cfg.Host),
		zap.String("port", cfg.Port),
		zap.Int("max_open_conns", cfg.MaxOpenConns),
		zap.Int("max_idle_conns", cfg.MaxIdleConns))

	return &mysqlClient{db: db, log: log}, nil
}

// Query implements DBClient.Query for MySQL.
func (c *mysqlClient) Query(ctx context.Context, sql string, args ...interface{}) (*sql.Rows, error) {
	c.log.Debug("Executing MySQL Query", zap.String("sql", sql), zap.Any("args", args))
	return c.db.QueryContext(ctx, sql, args...)
}

// Exec implements DBClient.Exec for MySQL.
func (c *mysqlClient) Exec(ctx context.Context, sql string, args ...interface{}) (sql.Result, error) {
	c.log.Debug("Executing MySQL Exec", zap.String("sql", sql), zap.Any("args", args))
	return c.db.ExecContext(ctx, sql, args...)
}

// Ping implements DBClient.Ping for MySQL.
func (c *mysqlClient) Ping(ctx context.Context) error {
	c.log.Debug("Pinging MySQL database")
	return c.db.PingContext(ctx)
}

// Close implements DBClient.Close for MySQL.
func (c *mysqlClient) Close() error {
	c.log.Info("Closing MySQL database connection...")
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}
