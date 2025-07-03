package tokenizer

import (
	"fmt"
	"path/filepath" // 用于处理文件路径，例如字典路径

	"go.uber.org/zap"

	ferrors "github.com/turtacn/starseek/internal/common/errors"
	"github.com/turtacn/starseek/internal/common/enum"           // 引入 enum 包，包含 IndexType
	"github.com/turtacn/starseek/internal/config"                // 引入配置包，用于获取字典路径
	"github.com/turtacn/starseek/internal/infrastructure/logger" // 引入日志接口

	// 导入具体的 tokenizer 实现
	jiebaTokenizer "github.com/turtacn/starseek/internal/infrastructure/tokenizer/jieba"
	englishTokenizer "github.com/turtacn/starseek/internal/infrastructure/tokenizer/english"
)

// TokenizerService 定义了分词服务接口。
// 它负责根据索引类型和分词器名称提供相应的分词器实例。
type TokenizerService interface {
	// GetTokenizer 根据指定的索引类型和分词器名称获取一个分词器实例。
	// 如果找不到对应的分词器，将返回错误。
	GetTokenizer(indexType enum.IndexType, tokenizerName string) (Tokenizer, error)
}

// tokenizerService 是 TokenizerService 接口的实现。
// 内部维护一个嵌套的 map 来存储不同类型和名称的分词器实例。
type tokenizerService struct {
	// key: IndexType, value: map[tokenizerName]Tokenizer
	tokenizers map[enum.IndexType]map[string]Tokenizer
	log        logger.Logger
	cfg        *config.AppConfig // 引入应用配置以获取字典路径
}

// NewTokenizerService 创建并初始化一个新的 TokenizerService 实例。
// 它注册了 Jieba 和 English 分词器实例。
func NewTokenizerService(cfg *config.AppConfig, log logger.Logger) (TokenizerService, error) {
	if log == nil {
		return nil, ferrors.NewInternalError("logger is nil", nil)
	}
	if cfg == nil {
		return nil, ferrors.NewInternalError("app config is nil", nil)
	}

	service := &tokenizerService{
		tokenizers: make(map[enum.IndexType]map[string]Tokenizer),
		log:        log,
		cfg:        cfg,
	}

	// 注册 Jieba 中文分词器
	if err := service.registerJiebaTokenizer(); err != nil {
		return nil, fmt.Errorf("failed to register Jieba tokenizer: %w", err)
	}

	// 注册 English 英文分词器
	if err := service.registerEnglishTokenizer(); err != nil {
		return nil, fmt.Errorf("failed to register English tokenizer: %w", err)
	}

	log.Info("TokenizerService initialized and registered all tokenizers.")
	return service, nil
}

// registerJiebaTokenizer 负责初始化和注册 Jieba 分词器。
func (s *tokenizerService) registerJiebaTokenizer() error {
	// 从配置中获取 Jieba 字典路径
	// 假设配置结构是 AppConfig.Tokenizer.JiebaDictPath
	jiebaDictPath := s.cfg.Tokenizer.JiebaDictPath
	if jiebaDictPath == "" {
		s.log.Warn("JiebaDictPath not specified in config, using gojieba default dictionaries.")
	} else {
		// 如果是相对路径，转换为绝对路径，或者确保应用程序启动目录的正确性
		// 实际项目中可能需要更复杂的路径处理，例如通过环境变量或特定目录查找
		absPath, err := filepath.Abs(jiebaDictPath)
		if err != nil {
			s.log.Error("Failed to get absolute path for Jieba dictionary", zap.String("path", jiebaDictPath), zap.Error(err))
			return ferrors.NewInternalError("invalid Jieba dictionary path", err)
		}
		jiebaDictPath = absPath
		s.log.Info("Using custom Jieba dictionary path", zap.String("path", jiebaDictPath))
	}

	jt, err := jiebaTokenizer.NewJiebaTokenizer(jiebaDictPath, s.log)
	if err != nil {
		s.log.Error("Failed to create Jieba tokenizer", zap.Error(err))
		return ferrors.NewInternalError("failed to create Jieba tokenizer", err)
	}

	// 将 Jieba 分词器注册到 map 中
	if s.tokenizers[enum.CHINESE_INDEX] == nil {
		s.tokenizers[enum.CHINESE_INDEX] = make(map[string]Tokenizer)
	}
	s.tokenizers[enum.CHINESE_INDEX]["jieba"] = jt
	s.log.Info("Jieba tokenizer registered",
		zap.String("index_type", enum.CHINESE_INDEX.String()),
		zap.String("tokenizer_name", "jieba"))
	return nil
}

// registerEnglishTokenizer 负责初始化和注册 English 分词器。
func (s *tokenizerService) registerEnglishTokenizer() error {
	et, err := englishTokenizer.NewEnglishTokenizer(s.log)
	if err != nil {
		s.log.Error("Failed to create English tokenizer", zap.Error(err))
		return ferrors.NewInternalError("failed to create English tokenizer", err)
	}

	// 将 English 分词器注册到 map 中
	if s.tokenizers[enum.ENGLISH_INDEX] == nil {
		s.tokenizers[enum.ENGLISH_INDEX] = make(map[string]Tokenizer)
	}
	s.tokenizers[enum.ENGLISH_INDEX]["standard_english"] = et
	s.log.Info("English tokenizer registered",
		zap.String("index_type", enum.ENGLISH_INDEX.String()),
		zap.String("tokenizer_name", "standard_english"))
	return nil
}

// GetTokenizer 方法根据指定的索引类型和分词器名称获取一个分词器实例。
func (s *tokenizerService) GetTokenizer(indexType enum.IndexType, tokenizerName string) (Tokenizer, error) {
	innerMap, ok := s.tokenizers[indexType]
	if !ok {
		s.log.Warn("Attempted to get tokenizer for unsupported index type",
			zap.String("index_type", indexType.String()),
			zap.String("tokenizer_name", tokenizerName))
		return nil, ferrors.NewNotFoundError(fmt.Sprintf("tokenizer type '%s' not supported", indexType.String()), nil)
	}

	tokenizer, ok := innerMap[tokenizerName]
	if !ok {
		s.log.Warn("Attempted to get unknown tokenizer name for index type",
			zap.String("index_type", indexType.String()),
			zap.String("tokenizer_name", tokenizerName))
		return nil, ferrors.NewNotFoundError(fmt.Sprintf("tokenizer '%s' not found for type '%s'", tokenizerName, indexType.String()), nil)
	}

	s.log.Debug("Retrieved tokenizer",
		zap.String("index_type", indexType.String()),
		zap.String("tokenizer_name", tokenizerName))
	return tokenizer, nil
}
