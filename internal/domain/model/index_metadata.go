package model

import (
	"github.com/turtacn/starseek/internal/common/enum"
	"time"
)

// IndexMetadata represents the metadata of a full-text index configured for a specific column in an OLAP table.
type IndexMetadata struct {
	ID         uint               `gorm:"primaryKey" json:"id"`        // Unique ID of the index metadata
	TableName  string             `gorm:"not null" json:"table_name"`  // Name of the table
	ColumnName string             `gorm:"not null" json:"column_name"` // Name of the column with the index
	IndexType  enum.IndexType     `gorm:"not null" json:"index_type"`  // Type of index (e.g., Inverted, NgramBF)
	Tokenizer  enum.TokenizerType `gorm:"not null" json:"tokenizer"`   // Tokenizer used for the column (e.g., Chinese, English)
	DataType   string             `gorm:"not null" json:"data_type"`   // Data type of the column (e.g., TEXT, VARCHAR)
	CreatedAt  time.Time          `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time          `gorm:"autoUpdateTime" json:"updated_at"`
}

//Personal.AI order the ending
