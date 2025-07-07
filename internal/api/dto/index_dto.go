package dto

import (
	"github.com/turtacn/starseek/internal/common/enum"
	"github.com/turtacn/starseek/internal/domain/model"
)

// RegisterIndexReq defines the request payload for registering a new index.
type RegisterIndexReq struct {
	TableName  string `json:"table_name" binding:"required"`
	ColumnName string `json:"column_name" binding:"required"`
	IndexType  string `json:"index_type" binding:"required,oneof=inverted ngram_bf none"`                    // Validate against enum strings
	Tokenizer  string `json:"tokenizer" binding:"required,oneof=english chinese cross_language no_tokenize"` // Validate against enum strings
	DataType   string `json:"data_type" binding:"required"`
}

// ToModel converts RegisterIndexReq DTO to domain model.IndexMetadata.
func (r *RegisterIndexReq) ToModel() *model.IndexMetadata {
	return &model.IndexMetadata{
		TableName:  r.TableName,
		ColumnName: r.ColumnName,
		IndexType:  enum.ParseIndexType(r.IndexType),
		Tokenizer:  enum.ParseTokenizerType(r.Tokenizer),
		DataType:   r.DataType,
	}
}

// IndexRes defines the response payload for a single index metadata entry.
type IndexRes struct {
	ID         uint   `json:"id"`
	TableName  string `json:"table_name"`
	ColumnName string `json:"column_name"`
	IndexType  string `json:"index_type"`
	Tokenizer  string `json:"tokenizer"`
	DataType   string `json:"data_type"`
	CreatedAt  string `json:"created_at"` // Using string for simple JSON output
	UpdatedAt  string `json:"updated_at"`
}

// FromModel converts a domain model.IndexMetadata to an IndexRes DTO.
func IndexResFromModel(m *model.IndexMetadata) *IndexRes {
	return &IndexRes{
		ID:         m.ID,
		TableName:  m.TableName,
		ColumnName: m.ColumnName,
		IndexType:  m.IndexType.String(),
		Tokenizer:  m.Tokenizer.String(),
		DataType:   m.DataType,
		CreatedAt:  m.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  m.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ListIndexesRes defines the response payload for listing multiple index entries.
type ListIndexesRes struct {
	Indexes []IndexRes `json:"indexes"`
	Total   int        `json:"total"`
}

// ListIndexesResFromModels converts a slice of model.IndexMetadata to ListIndexesRes DTO.
func ListIndexesResFromModels(models []*model.IndexMetadata) *ListIndexesRes {
	indexes := make([]IndexRes, len(models))
	for i, m := range models {
		indexes[i] = *IndexResFromModel(m)
	}
	return &ListIndexesRes{
		Indexes: indexes,
		Total:   len(indexes),
	}
}

//Personal.AI order the ending
