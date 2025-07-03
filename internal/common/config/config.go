package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
	ferrors "github.com/turtacn/starseek/internal/common/errors" // 引入自定义错误包，使用别名
	"github.com/turtacn/starseek/internal/common/types/enum"     // 引入枚举类型
)

// Config 结构体定义了整个项目的配置
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Logger   LoggerConfig   `mapstructure:"logger"`
}

// ServerConfig 定义服务器相关的配置
type ServerConfig struct {
	ListenAddr string `mapstructure:"listen_addr"` // 监听地址，例如 ":8080"
}

// DatabaseConfig 定义数据库相关的配置
type DatabaseConfig struct {
	Type            enum.DatabaseType `mapstructure:"type"`              // 数据库类型 (StarRocks, Doris, ClickHouse, MySQL)
	Host            string            `mapstructure:"host"`              // 主机地址
	Port            int               `mapstructure:"port"`              // 端口
	User            string            `mapstructure:"user"`              // 用户名
	Password        string            `mapstructure:"password"`          // 密码
	DBName          string            `mapstructure:"db_name"`           // 数据库名
	MaxOpenConns    int               `mapstructure:"max_open_conns"`    // 最大打开连接数
	MaxIdleConns    int               `mapstructure:"max_idle_conns"`    // 最大空闲连接数
	ConnMaxLifetime time.Duration     `mapstructure:"conn_max_lifetime"` // 连接最大生命周期
}

// RedisConfig 定义 Redis 相关的配置
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`     // Redis 地址，例如 "localhost:6379"
	Password string `mapstructure:"password"` // 密码
	DB       int    `mapstructure:"db"`       // 数据库索引
}

// LoggerConfig 定义日志相关的配置
type LoggerConfig struct {
	Level      string `mapstructure:"level"`       // 日志级别 (debug, info, warn, error, fatal, panic)
	Format     string `mapstructure:"format"`      // 日志格式 (json, text)
	OutputPath string `mapstructure:"output_path"` // 日志输出路径，例如 "stdout" 或 "logs/app.log"
}

// LoadConfig 从文件和环境变量加载配置
func LoadConfig() (*Config, error) {
	v := viper.New()

	// 1. 设置配置文件的名称和搜索路径
	v.SetConfigName("config")           // 配置文件名 (不带扩展名)，例如 config.yaml, config.json
	v.AddConfigPath(".")                // 当前目录
	v.AddConfigPath("./config")         // config 目录 (常见的配置存放位置)
	v.AddConfigPath("/etc/starseek/")   // Linux 系统级配置路径
	v.AddConfigPath("$HOME/.starseek/") // 用户主目录配置路径

	// 2. 设置环境变量前缀和绑定
	// 例如，STARSEEK_SERVER_LISTEN_ADDR 会映射到 config.Server.ListenAddr
	v.SetEnvPrefix("STARSEEK")                         // 所有环境变量都以 STARSEEK_ 开头
	v.AutomaticEnv()                                   // 自动读取环境变量
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // 允许环境变量使用下划线代替点，例如 DB_NAME 对应 db.name

	// 3. 设置默认值
	setDefaults(v)

	// 4. 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		// 如果配置文件不存在，则忽略错误，继续从环境变量和默认值加载
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, ferrors.NewConfigError(fmt.Sprintf("Failed to read config file: %v", err), err)
		}
		// 如果文件未找到，不报错，表示可以使用纯环境变量或默认值启动
		fmt.Printf("Config file not found, loading from environment variables and defaults.\n")
	}

	// 5. 将配置绑定到结构体
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, ferrors.NewConfigError(fmt.Sprintf("Failed to unmarshal config: %v", err), err)
	}

	// 6. 进行配置校验
	if err := validateConfig(&cfg); err != nil {
		return nil, ferrors.NewConfigError(fmt.Sprintf("Config validation failed: %v", err), err)
	}

	return &cfg, nil
}

// setDefaults 为配置项设置默认值
func setDefaults(v *viper.Viper) {
	// Server
	v.SetDefault("server.listen_addr", ":8080")

	// Database
	v.SetDefault("database.type", enum.DatabaseTypeStarRocks.String()) // 默认数据库类型为 StarRocks
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 9030) // StarRocks 默认端口
	v.SetDefault("database.user", "root")
	v.SetDefault("database.password", "")
	v.SetDefault("database.db_name", "default_db")
	v.SetDefault("database.max_open_conns", 100)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.conn_max_lifetime", 5*time.Minute) // 5分钟

	// Redis
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)

	// Logger
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "text")
	v.SetDefault("logger.output_path", "stdout")
}

// validateConfig 校验配置的合法性
func validateConfig(cfg *Config) error {
	// 校验 Server 配置
	if cfg.Server.ListenAddr == "" {
		return fmt.Errorf("server.listen_addr cannot be empty")
	}

	// 校验 Database 配置
	// 将字符串类型的数据库类型转换为 enum.DatabaseType
	// 注意: viper Unmarshal 对于自定义类型需要特殊处理或者在 Unmarshal 之后手动转换
	// 这里我们假设 viper 能够通过 `mapstructure` 标签和 `String()` 方法在 Unmarshal 时自动处理，
	// 但如果 Unmarshal 失败，需要在此处再次校验其原始值。
	// 更健壮的方式是为 enum.DatabaseType 实现 TextUnmarshaler 接口，
	// 但为了简化，这里直接依赖 viper 的默认行为或进行后置校验。
	// 在 LoadConfig 中 `v.Unmarshal` 已经完成了 `string` 到 `enum.DatabaseType` 的转换，
	// 这里只需要校验转换后的 `enum.DatabaseType` 是否有效。

	if !cfg.Database.Type.IsValid() {
		// 实际上，如果从配置文件或环境变量读取到了一个无效的字符串，
		// viper 会将其映射为该类型的零值 (enum.DatabaseType(0))，
		// 此时 IsValid() 会返回 false。
		// 但为了给出更具体的错误信息，可以考虑在 Unmarshal 之前先读取原始字符串，然后手动转换。
		// 这里简化为只校验转换后的有效性。
		return fmt.Errorf("database.type is invalid: %s", cfg.Database.Type.String())
	}

	if cfg.Database.Host == "" {
		return fmt.Errorf("database.host cannot be empty")
	}
	if cfg.Database.Port <= 0 {
		return fmt.Errorf("database.port must be a positive integer")
	}
	if cfg.Database.User == "" {
		return fmt.Errorf("database.user cannot be empty")
	}
	if cfg.Database.DBName == "" {
		return fmt.Errorf("database.db_name cannot be empty")
	}
	if cfg.Database.MaxOpenConns <= 0 {
		return fmt.Errorf("database.max_open_conns must be positive")
	}
	if cfg.Database.MaxIdleConns <= 0 || cfg.Database.MaxIdleConns > cfg.Database.MaxOpenConns {
		return fmt.Errorf("database.max_idle_conns must be positive and not greater than max_open_conns")
	}
	if cfg.Database.ConnMaxLifetime < 0 {
		return fmt.Errorf("database.conn_max_lifetime cannot be negative")
	}

	// 校验 Redis 配置
	if cfg.Redis.Addr == "" {
		return fmt.Errorf("redis.addr cannot be empty")
	}
	if cfg.Redis.DB < 0 {
		return fmt.Errorf("redis.db cannot be negative")
	}

	// 校验 Logger 配置
	switch strings.ToLower(cfg.Logger.Level) {
	case "debug", "info", "warn", "error", "fatal", "panic":
		// Valid
	default:
		return fmt.Errorf("logger.level '%s' is invalid, must be one of debug, info, warn, error, fatal, panic", cfg.Logger.Level)
	}
	switch strings.ToLower(cfg.Logger.Format) {
	case "json", "text":
		// Valid
	default:
		return fmt.Errorf("logger.format '%s' is invalid, must be json or text", cfg.Logger.Format)
	}
	if cfg.Logger.OutputPath == "" {
		return fmt.Errorf("logger.output_path cannot be empty")
	}

	return nil
}
