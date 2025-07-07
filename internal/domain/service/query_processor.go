package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/turtacn/starseek/internal/common/enum"
	"github.com/turtacn/starseek/internal/common/errorx"
	"github.com/turtacn/starseek/internal/domain/model"
	"github.com/turtacn/starseek/internal/domain/repository"
	"go.uber.org/zap"
)

// QueryProcessor is a domain service responsible for parsing and structuring raw search queries.
type QueryProcessor struct {
	logger    *zap.Logger
	indexRepo repository.IndexMetadataRepository // To get tokenizer info for fields
	// Add a tokenizer factory/service here if actual tokenization is needed
}

// NewQueryProcessor creates a new QueryProcessor instance.
func NewQueryProcessor(logger *zap.Logger, indexRepo repository.IndexMetadataRepository) *QueryProcessor {
	return &QueryProcessor{
		logger:    logger.Named("QueryProcessor"),
		indexRepo: indexRepo,
	}
}

// Parse takes a raw query string and a list of target fields,
// then returns a structured ParsedQuery object.
//
// Syntax examples to support:
// - "keyword1 keyword2"           => Keywords
// - "title:golang content:concurrency" => FieldSpecificQuery
// - "golang AND (title:并发 OR content:编程) NOT rust" => MustWords, ShouldWords, NotWords, FieldSpecific
// - "phrase query" => IsPhrase
func (qp *QueryProcessor) Parse(rawQuery string, fields []string) (*model.ParsedQuery, error) {
	parsed := &model.ParsedQuery{
		OriginalQuery: rawQuery,
		TargetFields:  fields,
	}

	// Step 1: Extract field-specific queries (e.g., "title:golang")
	fieldQueryRegex := regexp.MustCompile(`(\w+):(".*?"|\S+)`) // Matches field:value or field:"phrase value"
	rawQuery = fieldQueryRegex.ReplaceAllStringFunc(rawQuery, func(match string) string {
		parts := strings.SplitN(match, ":", 2)
		fieldName := strings.TrimSpace(parts[0])
		fieldValue := strings.Trim(strings.TrimSpace(parts[1]), `"`) // Remove quotes for phrase

		fq := model.FieldSpecificQuery{
			FieldName: fieldName,
			Keywords:  qp.tokenizeString(fieldValue, enum.CrossLanguage), // Use a default tokenizer for parsing
			IsPhrase:  strings.HasPrefix(parts[1], `"`),
		}
		parsed.FieldQueries = append(parsed.FieldQueries, fq)
		return "" // Remove this part from the raw query for general keyword parsing
	})

	// Step 2: Handle boolean operators (AND, OR, NOT) for remaining general keywords
	// This is a simplified boolean parser. A real one might use a more robust parser library.
	queryWithoutFields := strings.TrimSpace(rawQuery)

	// NOT operator first
	notRegex := regexp.MustCompile(`\bNOT\s+(\S+)`) // Basic NOT keyword
	queryWithoutFields = notRegex.ReplaceAllStringFunc(queryWithoutFields, func(match string) string {
		keyword := strings.TrimSpace(strings.TrimPrefix(match, "NOT"))
		parsed.NotWords = append(parsed.NotWords, qp.tokenizeString(keyword, enum.CrossLanguage)...)
		return ""
	})

	// AND operator
	andRegex := regexp.MustCompile(`\bAND\s+(\S+)`)
	queryWithoutFields = andRegex.ReplaceAllStringFunc(queryWithoutFields, func(match string) string {
		keyword := strings.TrimSpace(strings.TrimPrefix(match, "AND"))
		parsed.MustWords = append(parsed.MustWords, qp.tokenizeString(keyword, enum.CrossLanguage)...)
		return ""
	})

	// Remaining words are treated as general keywords (implicitly OR) or SHOULD
	generalWords := strings.Fields(queryWithoutFields)
	for _, word := range generalWords {
		if strings.EqualFold(word, "AND") || strings.EqualFold(word, "OR") || strings.EqualFold(word, "NOT") {
			// Skip actual boolean operators if they are left over
			continue
		}
		parsed.Keywords = append(parsed.Keywords, qp.tokenizeString(word, enum.CrossLanguage)...)
	}

	qp.logger.Debug("Parsed query",
		zap.String("original", parsed.OriginalQuery),
		zap.Strings("keywords", parsed.Keywords),
		zap.Strings("mustWords", parsed.MustWords),
		zap.Strings("notWords", parsed.NotWords),
		zap.Any("fieldQueries", parsed.FieldQueries),
	)

	return parsed, nil
}

// tokenizeString is a placeholder for actual tokenization logic.
// In a real system, this would involve calling a specific tokenizer based on the `tokenizerType`.
func (qp *QueryProcessor) tokenizeString(text string, tokenizerType enum.TokenizerType) []string {
	// Simple space-based tokenizer for demonstration.
	// For Chinese, it would use a library like "go-jieba".
	// For English, it might normalize, stem, etc.
	if text == "" {
		return []string{}
	}
	return strings.Fields(text)
}

// GetTokenizerForField fetches the configured tokenizer for a specific table and column.
// This would be used internally by a real tokenizer service.
func (qp *QueryProcessor) GetTokenizerForField(tableName, columnName string) (enum.TokenizerType, error) {
	ctx, cancel := context.WithCancel(context.Background()) // Use a short-lived context
	defer cancel()

	meta, err := qp.indexRepo.FindByTableAndColumn(ctx, tableName, columnName)
	if err != nil {
		if errorx.IsNotFound(err) {
			qp.logger.Warn("No index metadata found for field, using default tokenizer",
				zap.String("table", tableName), zap.String("column", columnName))
			return enum.NoTokenize, nil // Or a default cross-language tokenizer
		}
		return enum.NoTokenize, errorx.NewError(errorx.ErrInternalServer.Code, "failed to retrieve index metadata for tokenizer", err)
	}
	return meta.Tokenizer, nil
}

//Personal.AI order the ending
