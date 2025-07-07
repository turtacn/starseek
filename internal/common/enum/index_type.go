package enum

// IndexType defines the type of full-text index used by the underlying OLAP database.
type IndexType int

const (
	None     IndexType = iota // No specific full-text index
	Inverted                  // Inverted index (e.g., StarRocks MATCH)
	NgramBF                   // N-gram Bloom Filter index (e.g., StarRocks NGRAM_BF)
	// Add other index types as needed, e.g., Bitmap, TextIndex (ClickHouse)
)

// String returns the string representation of IndexType.
func (it IndexType) String() string {
	switch it {
	case Inverted:
		return "inverted"
	case NgramBF:
		return "ngram_bf"
	case None:
		return "none"
	default:
		return "unknown"
	}
}

// ParseIndexType parses a string into an IndexType.
func ParseIndexType(s string) IndexType {
	switch s {
	case "inverted":
		return Inverted
	case "ngram_bf":
		return NgramBF
	case "none":
		return None
	default:
		return None // Default to None or return an error/unknown type
	}
}

// TokenizerType defines the type of tokenizer to be used for text processing.
type TokenizerType int

const (
	English       TokenizerType = iota // English tokenizer (e.g., standard analyzer)
	Chinese                            // Chinese tokenizer (e.g., Jieba)
	CrossLanguage                      // Multi-language tokenizer
	NoTokenize                         // No tokenization, exact match
)

// String returns the string representation of TokenizerType.
func (tt TokenizerType) String() string {
	switch tt {
	case English:
		return "english"
	case Chinese:
		return "chinese"
	case CrossLanguage:
		return "cross_language"
	case NoTokenize:
		return "no_tokenize"
	default:
		return "unknown"
	}
}

// ParseTokenizerType parses a string into a TokenizerType.
func ParseTokenizerType(s string) TokenizerType {
	switch s {
	case "english":
		return English
	case "chinese":
		return Chinese
	case "cross_language":
		return CrossLanguage
	case "no_tokenize":
		return NoTokenize
	default:
		return NoTokenize // Default to NoTokenize or return an error/unknown type
	}
}

//Personal.AI order the ending
