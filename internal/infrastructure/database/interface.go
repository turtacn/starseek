package database

import (
	"context"
	"database/sql"
	"fmt"

	ferrors "github.com/turtacn/starseek/internal/common/errors"         // 引入自定义错误包
	"github.com/turtacn/starseek/internal/config"                        // 引入配置包，用于数据库配置
	"github.com/turtacn/starseek/internal/infrastructure/database/mysql" // 引入MySQL适配器，因为它在persistence中被用到
	"github.com/turtacn/starseek/internal/infrastructure/logger"         // 引入日志包
	// TODO: 等待 StarRocks, Doris, ClickHouse 适配器文件创建后，在这里引入
	// "github.com/turtacn/starseek/internal/infrastructure/database/starrocks"
	// "github.com/turtacn/starseek/internal/infrastructure/database/doris"
	// "github.com/turtacn/starseek/internal/infrastructure/database/clickhouse"
)

// DBClient 定义了通用的数据库客户端接口，用于抽象不同数据库的操作。
// 旨在提供与标准库 database/sql 相似的核心功能。
type DBClient interface {
	// Query 执行一个返回行的查询（如 SELECT）。
	// 参数为 SQL 语句和可选的参数。
	// 返回 *sql.Rows 和 error。
	Query(ctx context.Context, sql string, args ...interface{}) (*sql.Rows, error)

	// Exec 执行一个不返回行的命令（如 INSERT, UPDATE, DELETE, CREATE TABLE, DROP TABLE）。
	// 参数为 SQL 语句和可选的参数。
	// 返回 sql.Result 和 error，sql.Result 通常包含受影响的行数或最后插入的 ID。
	Exec(ctx context.Context, sql string, args ...interface{}) (sql.Result, error)

	// Ping 验证数据库连接是否仍然活跃。
	// 如果连接断开或无法建立，则返回错误。
	Ping(ctx context.Context) error

	// Close 关闭数据库连接池。
	// 通常在应用程序退出时调用以释放资源。
	Close() error
}

// NewDatabaseClient 是一个工厂函数，根据配置创建并返回一个 DBClient 实例。
// 它根据 cfg.Type 字段决定创建哪种具体的数据库客户端适配器。
func NewDatabaseClient(cfg *config.DatabaseConfig, log logger.Logger) (DBClient, error) {
	if cfg == nil {
		return nil, ferrors.NewInternalError("database configuration is nil", nil)
	}
	if log == nil {
		return nil, ferrors.NewInternalError("logger is nil", nil)
	}

	switch cfg.Type {
	case "mysql":
		// 使用 internal/infrastructure/database/mysql/client.go 中定义的 NewMySQLClient
		// 注意：这里的 NewMySQLClient 应该返回 DBClient 接口
		client, err := mysql.NewMySQLClient(cfg, log)
		if err != nil {
			return nil, ferrors.NewExternalServiceError(fmt.Sprintf("failed to create MySQL client: %v", err), err)
		}
		return client, nil
	// TODO: 未来添加其他数据库类型支持
	// case "starrocks":
	// 	client, err := starrocks.NewStarRocksClient(cfg, log)
	// 	if err != nil {
	// 		return nil, ferrors.NewExternalServiceError(fmt.Sprintf("failed to create StarRocks client: %v", err), err)
	// 	}
	// 	return client, nil
	// case "doris":
	// 	client, err := doris.NewDorisClient(cfg, log)
	// 	if err != nil {
	// 		return nil, ferrors.NewExternalServiceError(fmt.Sprintf("failed to create Doris client: %v", err), err)
	// 	}
	// 	return client, nil
	// case "clickhouse":
	// 	client, err := clickhouse.NewClickHouseClient(cfg, log)
	// 	if err != nil {
	// 		return nil, ferrors.NewExternalServiceError(fmt.Sprintf("failed to create ClickHouse client: %v", err), err)
	// 	}
	// 	return client, nil
	default:
		return nil, ferrors.NewUnsupportedError(fmt.Sprintf("unsupported database type: %s", cfg.Type), nil)
	}
}
