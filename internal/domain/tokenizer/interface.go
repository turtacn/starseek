package tokenizer

// Tokenizer 定义了一个通用的分词器接口。
// 任何实现了此接口的类型都能够将给定的文本分词为字符串切片。
type Tokenizer interface {
	// Tokenize 方法将输入的文本字符串分词为一系列的词元（tokens）。
	// 返回一个字符串切片，其中每个元素代表一个词元。
	// 如果在分词过程中发生错误，将返回错误。
	Tokenize(text string) ([]string, error)
}

