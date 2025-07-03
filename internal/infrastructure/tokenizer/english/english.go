package english

import (
	"fmt"
	"regexp"
	"strings"

	"go.uber.org/zap" // 用于日志字段

	ferrors "github.com/turtacn/starseek/internal/common/errors" // 引入自定义错误包
	"github.com/turtacn/starseek/internal/domain/tokenizer"      // 引入 Tokenizer 接口
	"github.com/turtacn/starseek/internal/infrastructure/logger" // 引入日志接口
)

// wordRegex 是用于匹配英文单词的正则表达式。
// 它匹配一个或多个字母数字字符（包括下划线）。
// 为了简单起见，这里不区分大小写，通常在搜索场景下，会将所有词元转为小写。
// 但是，本适配器只负责分词，不做归一化，归一化通常在分词后的后处理阶段进行。
var wordRegex = regexp.MustCompile(`[a-zA-Z0-9]+`)

// englishTokenizer 是 Tokenizer 接口的英文实现。
// 它使用正则表达式来提取英文文本中的词元。
type englishTokenizer struct {
	log logger.Logger
}

// NewEnglishTokenizer 创建一个新的 englishTokenizer 实例。
func NewEnglishTokenizer(log logger.Logger) (tokenizer.Tokenizer, error) {
	if log == nil {
		return nil, ferrors.NewInternalError("logger is nil", nil)
	}
	log.Info("Initialized English tokenizer.")
	return &englishTokenizer{log: log}, nil
}

// Tokenize 方法将输入的英文文本字符串分词为一系列的词元。
// 它通过正则表达式匹配单词，并过滤掉空字符串。
func (et *englishTokenizer) Tokenize(text string) ([]string, error) {
	if text == "" {
		et.log.Debug("Tokenizing empty string, returning empty slice.")
		return []string{}, nil
	}

	// 使用正则表达式查找所有匹配的单词
	matches := wordRegex.FindAllString(text, -1)

	// 过滤掉空字符串（理论上wordRegex不会产生空字符串，但作为防御性编程）
	var tokens []string
	for _, match := range matches {
		if strings.TrimSpace(match) != "" { // 确保不是纯空格的匹配（虽然正则已经排除）
			tokens = append(tokens, match)
		}
	}

	et.log.Debug("English tokenization complete", zap.String("text", text), zap.Strings("tokens", tokens))
	return tokens, nil
}
