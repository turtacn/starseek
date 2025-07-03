package jieba

import (
	"fmt"

	gojieba "github.com/yanyiwu/gojieba" // 引入 gojieba 库
	"go.uber.org/zap"                    // 用于日志字段

	ferrors "github.com/turtacn/starseek/internal/common/errors" // 引入自定义错误包
	"github.com/turtacn/starseek/internal/domain/tokenizer"      // 引入 Tokenizer 接口
	"github.com/turtacn/starseek/internal/infrastructure/logger" // 引入日志接口
)

// jiebaTokenizer 是 Tokenizer 接口的 Jieba 实现。
// 它包装了 gojieba.Jieba 实例进行中文分词。
type jiebaTokenizer struct {
	jieba *gojieba.Jieba
	log   logger.Logger
}

// NewJiebaTokenizer 创建一个新的 jiebaTokenizer 实例。
// 它根据提供的字典路径初始化 gojieba.Jieba 实例。
// 如果 dictPath 为空字符串，gojieba 会使用其默认字典。
func NewJiebaTokenizer(dictPath string, log logger.Logger) (tokenizer.Tokenizer, error) {
	if log == nil {
		return nil, ferrors.NewInternalError("logger is nil", nil)
	}

	var j *gojieba.Jieba
	if dictPath == "" {
		// 使用 gojieba 默认字典路径
		j = gojieba.NewJieba()
		log.Info("Initialized Jieba tokenizer with default dictionaries.")
	} else {
		// 使用指定的字典路径
		// gojieba.NewJieba(dictPath1, dictPath2, ...)
		// 这里假设 dictPath 是主字典路径，如果需要更多字典，需要根据 gojieba 文档调整
		log.Info("Initializing Jieba tokenizer with custom dictionary path", zap.String("dict_path", dictPath))
		j = gojieba.NewJieba(
			dictPath,                // 主词典
			gojieba.DefaultHMMDict,  // HMM 词典
			gojieba.DefaultUserDict, // 用户词典
			gojieba.DefaultIdf,      // IDF 词典
			gojieba.DefaultStopWord, // 停用词词典
		)
	}

	if j == nil {
		// gojieba.NewJieba() 理论上不会返回 nil，除非内存不足等极端情况
		// 但为了安全起见，还是检查一下。
		errMsg := "failed to initialize gojieba instance"
		log.Error(errMsg)
		return nil, ferrors.NewExternalServiceError(errMsg, nil)
	}

	log.Info("Successfully initialized Jieba tokenizer.")
	return &jiebaTokenizer{jieba: j, log: log}, nil
}

// Tokenize 方法将输入的文本字符串分词为一系列的词元。
// 对于中文分词，通常使用 "CutForSearch" 模式以获取更适合搜索的词元。
func (jt *jiebaTokenizer) Tokenize(text string) ([]string, error) {
	if jt.jieba == nil {
		err := ferrors.NewInternalError("jieba tokenizer not initialized", nil)
		jt.log.Error("Attempted to tokenize with uninitialized jieba", zap.Error(err))
		return nil, err
	}

	// CutForSearch 适用于搜索引擎模式，可以识别出更长的词和部分短词
	// 第二个参数 true 表示开启 HMM (Hidden Markov Model) 模式
	tokens := jt.jieba.CutForSearch(text, true)
	jt.log.Debug("Jieba tokenization complete", zap.String("text", text), zap.Strings("tokens", tokens))
	return tokens, nil
}
