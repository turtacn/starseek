package starrocks

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql" // Import for database/sql
	"github.com/stretchr/testify/assert"
	"github.com/turtacn/starseek/internal/common/enum"
	"github.com/turtacn/starseek/internal/domain/model"
)

// MockDB for testing ExecuteSearch (simplified, just to pass `*sql.DB`)
type MockDB struct{}

func (m *MockDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	// Not implementing actual query logic for BuildSearchSQL test,
	// but keeping it for ExecuteSearch if needed later with mock rows.
	return nil, nil
}

// TestBuildSearchSQL tests the BuildSearchSQL method for various query scenarios.
func TestBuildSearchSQL(t *testing.T) {
	adapter := NewStarRocksAdapter(&sql.DB{}) // Pass a dummy *sql.DB for testing BuildSearchSQL

	tests := []struct {
		name         string
		tableName    string
		columnMeta   *model.IndexMetadata
		parsedQuery  *model.ParsedQuery
		expectedSQL  string
		expectedArgs []interface{}
		expectError  bool
	}{
		{
			name:      "Single Keyword Search",
			tableName: "articles",
			columnMeta: &model.IndexMetadata{
				ColumnName: "title",
				IndexType:  enum.Inverted,
				Tokenizer:  enum.Chinese,
			},
			parsedQuery: &model.ParsedQuery{
				Keywords: []string{"golang"},
			},
			expectedSQL:  "SELECT id FROM articles WHERE MATCH_ANY(title, ?)",
			expectedArgs: []interface{}{"golang"},
			expectError:  false,
		},
		{
			name:      "Multiple Keywords (OR logic implied)",
			tableName: "articles",
			columnMeta: &model.IndexMetadata{
				ColumnName: "content",
				IndexType:  enum.Inverted,
				Tokenizer:  enum.English,
			},
			parsedQuery: &model.ParsedQuery{
				Keywords: []string{"microservices", "kubernetes"},
			},
			expectedSQL:  "SELECT id FROM articles WHERE MATCH_ANY(content, ?)",
			expectedArgs: []interface{}{"microservices | kubernetes"},
			expectError:  false,
		},
		{
			name:      "Field Specific Query - Simple",
			tableName: "products",
			columnMeta: &model.IndexMetadata{
				ColumnName: "name",
				IndexType:  enum.Inverted,
				Tokenizer:  enum.English,
			},
			parsedQuery: &model.ParsedQuery{
				FieldQueries: []model.FieldSpecificQuery{
					{FieldName: "name", Keywords: []string{"apple"}},
				},
			},
			expectedSQL:  "SELECT id FROM products WHERE MATCH_ANY(name, ?)",
			expectedArgs: []interface{}{"apple"},
			expectError:  false,
		},
		{
			name:      "Field Specific Query - Phrase",
			tableName: "products",
			columnMeta: &model.IndexMetadata{
				ColumnName: "description",
				IndexType:  enum.Inverted,
				Tokenizer:  enum.English,
			},
			parsedQuery: &model.ParsedQuery{
				FieldQueries: []model.FieldSpecificQuery{
					{FieldName: "description", Keywords: []string{"new", "york"}, IsPhrase: true},
				},
			},
			expectedSQL:  `SELECT id FROM products WHERE MATCH_ANY(description, ?)`,
			expectedArgs: []interface{}{`"new york"`},
			expectError:  false,
		},
		{
			name:      "Mixed Keywords and Field-Specific (for same column)",
			tableName: "articles",
			columnMeta: &model.IndexMetadata{
				ColumnName: "title",
				IndexType:  enum.Inverted,
				Tokenizer:  enum.Chinese,
			},
			parsedQuery: &model.ParsedQuery{
				Keywords: []string{"大数据"},
				FieldQueries: []model.FieldSpecificQuery{
					{FieldName: "title", Keywords: []string{"hadoop"}},
				},
			},
			expectedSQL:  "SELECT id FROM articles WHERE MATCH_ANY(title, ?) OR MATCH_ANY(title, ?)",
			expectedArgs: []interface{}{"大数据", "hadoop"},
			expectError:  false,
		},
		{
			name:      "No Full-Text Index Configured",
			tableName: "users",
			columnMeta: &model.IndexMetadata{
				ColumnName: "email",
				IndexType:  enum.None,
			},
			parsedQuery: &model.ParsedQuery{
				Keywords: []string{"test"},
			},
			expectedSQL:  "",
			expectedArgs: nil,
			expectError:  true,
		},
		{
			name:         "Empty Query",
			tableName:    "articles",
			columnMeta:   &model.IndexMetadata{ColumnName: "title", IndexType: enum.Inverted},
			parsedQuery:  &model.ParsedQuery{},
			expectedSQL:  "",
			expectedArgs: nil,
			expectError:  true, // Should return error or empty if no search terms generated
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args, err := adapter.BuildSearchSQL(tt.tableName, tt.columnMeta, tt.parsedQuery)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSQL, sql)
				assert.Equal(t, tt.expectedArgs, args)
			}
		})
	}
}

//Personal.AI order the ending
