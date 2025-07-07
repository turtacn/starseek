package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/turtacn/starseek/internal/common/enum"
	"github.com/turtacn/starseek/internal/domain/model"
)

// MockIndexMetadataRepository for QueryProcessorTest
type MockIndexMetadataRepository struct{}

func (m *MockIndexMetadataRepository) Save(ctx context.Context, metadata *model.IndexMetadata) error {
	return nil
}
func (m *MockIndexMetadataRepository) FindByID(ctx context.Context, id uint) (*model.IndexMetadata, error) {
	return nil, nil
}
func (m *MockIndexMetadataRepository) FindByTableAndColumn(ctx context.Context, tableName, columnName string) (*model.IndexMetadata, error) {
	// Simulate some existing index metadata for tokenization info
	if tableName == "articles" {
		if columnName == "title" {
			return &model.IndexMetadata{TableName: tableName, ColumnName: columnName, Tokenizer: enum.Chinese}, nil
		}
		if columnName == "content" {
			return &model.IndexMetadata{TableName: tableName, ColumnName: columnName, Tokenizer: enum.English}, nil
		}
	}
	return nil, nil // Simulate not found
}
func (m *MockIndexMetadataRepository) ListByTable(ctx context.Context, tableName string) ([]*model.IndexMetadata, error) {
	return nil, nil
}
func (m *MockIndexMetadataRepository) ListAll(ctx context.Context) ([]*model.IndexMetadata, error) {
	return nil, nil
}
func (m *MockIndexMetadataRepository) DeleteByID(ctx context.Context, id uint) error { return nil }

func TestQueryProcessor_Parse(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	qp := NewQueryProcessor(logger, &MockIndexMetadataRepository{})

	tests := []struct {
		name        string
		rawQuery    string
		fields      []string
		expected    *model.ParsedQuery
		expectError bool
	}{
		{
			name:     "Simple single keyword",
			rawQuery: "golang",
			fields:   []string{},
			expected: &model.ParsedQuery{
				OriginalQuery: "golang",
				Keywords:      []string{"golang"},
				TargetFields:  []string{},
			},
		},
		{
			name:     "Multiple keywords",
			rawQuery: "go语言 编程",
			fields:   []string{},
			expected: &model.ParsedQuery{
				OriginalQuery: "go语言 编程",
				Keywords:      []string{"go语言", "编程"},
				TargetFields:  []string{},
			},
		},
		{
			name:     "Field specific query",
			rawQuery: "title:分布式",
			fields:   []string{},
			expected: &model.ParsedQuery{
				OriginalQuery: "title:分布式",
				Keywords:      []string{},
				TargetFields:  []string{},
				FieldQueries: []model.FieldSpecificQuery{
					{FieldName: "title", Keywords: []string{"分布式"}},
				},
			},
		},
		{
			name:     "Field specific phrase query",
			rawQuery: `content:"微服务架构"`,
			fields:   []string{},
			expected: &model.ParsedQuery{
				OriginalQuery: `content:"微服务架构"`,
				Keywords:      []string{},
				TargetFields:  []string{},
				FieldQueries: []model.FieldSpecificQuery{
					{FieldName: "content", Keywords: []string{"微服务架构"}, IsPhrase: true},
				},
			},
		},
		{
			name:     "Mixed keywords and field specific",
			rawQuery: `大数据 title:hadoop`,
			fields:   []string{},
			expected: &model.ParsedQuery{
				OriginalQuery: `大数据 title:hadoop`,
				Keywords:      []string{"大数据"},
				TargetFields:  []string{},
				FieldQueries: []model.FieldSpecificQuery{
					{FieldName: "title", Keywords: []string{"hadoop"}},
				},
			},
		},
		{
			name:     "AND operator",
			rawQuery: `golang AND channel`,
			fields:   []string{},
			expected: &model.ParsedQuery{
				OriginalQuery: `golang AND channel`,
				Keywords:      []string{"golang"},
				MustWords:     []string{"channel"},
				TargetFields:  []string{},
			},
		},
		{
			name:     "NOT operator",
			rawQuery: `编程 NOT java`,
			fields:   []string{},
			expected: &model.ParsedQuery{
				OriginalQuery: `编程 NOT java`,
				Keywords:      []string{"编程"},
				NotWords:      []string{"java"},
				TargetFields:  []string{},
			},
		},
		{
			name:     "Complex query 1",
			rawQuery: `go concurrency AND "web assembly" NOT rust title:"深入理解"`,
			fields:   []string{},
			expected: &model.ParsedQuery{
				OriginalQuery: `go concurrency AND "web assembly" NOT rust title:"深入理解"`,
				Keywords:      []string{"go", "concurrency"},
				MustWords:     []string{"web", "assembly"}, // simplified tokenizer doesn't handle phrases here automatically
				NotWords:      []string{"rust"},
				TargetFields:  []string{},
				FieldQueries: []model.FieldSpecificQuery{
					{FieldName: "title", Keywords: []string{"深入理解"}, IsPhrase: true},
				},
			},
		},
		{
			name:     "Complex query 2 with target fields",
			rawQuery: `elastic stack AND "log analysis" NOT kibana`,
			fields:   []string{"title", "content"},
			expected: &model.ParsedQuery{
				OriginalQuery: `elastic stack AND "log analysis" NOT kibana`,
				Keywords:      []string{"elastic", "stack"},
				MustWords:     []string{"log", "analysis"},
				NotWords:      []string{"kibana"},
				TargetFields:  []string{"title", "content"},
			},
		},
		{
			name:     "Query with only boolean operators (should result in no keywords)",
			rawQuery: `AND OR NOT`,
			fields:   []string{},
			expected: &model.ParsedQuery{
				OriginalQuery: `AND OR NOT`,
				Keywords:      []string{},
				MustWords:     []string{},
				NotWords:      []string{},
				TargetFields:  []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := qp.Parse(tt.rawQuery, tt.fields)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Assert specific fields for easier debugging
				assert.Equal(t, tt.expected.OriginalQuery, actual.OriginalQuery, "OriginalQuery mismatch")
				assert.ElementsMatch(t, tt.expected.Keywords, actual.Keywords, "Keywords mismatch")
				assert.ElementsMatch(t, tt.expected.MustWords, actual.MustWords, "MustWords mismatch")
				assert.ElementsMatch(t, tt.expected.NotWords, actual.NotWords, "NotWords mismatch")
				assert.ElementsMatch(t, tt.expected.TargetFields, actual.TargetFields, "TargetFields mismatch")

				// For FieldQueries, deep comparison is needed due to struct equality
				assert.Len(t, actual.FieldQueries, len(tt.expected.FieldQueries), "FieldQueries length mismatch")
				for i, expectedFQ := range tt.expected.FieldQueries {
					actualFQ := actual.FieldQueries[i]
					assert.Equal(t, expectedFQ.FieldName, actualFQ.FieldName, "FieldQuery FieldName mismatch")
					assert.ElementsMatch(t, expectedFQ.Keywords, actualFQ.Keywords, "FieldQuery Keywords mismatch")
					assert.Equal(t, expectedFQ.IsPhrase, actualFQ.IsPhrase, "FieldQuery IsPhrase mismatch")
				}
			}
		})
	}
}

//Personal.AI order the ending
