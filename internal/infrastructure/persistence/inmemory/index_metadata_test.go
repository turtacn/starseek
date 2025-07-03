package inmemory_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ferrors "github.com/turtacn/starseek/internal/common/errors"
	"github.com/turtacn/starseek/internal/domain/enum"
	"github.com/turtacn/starseek/internal/domain/index"
	"github.com/turtacn/starseek/internal/infrastructure/persistence/inmemory" // 被测试的内存仓库
)

// Helper function to create a new repository for each test
func setupRepository() index.IndexMetadataRepository {
	return inmemory.NewInMemoryIndexMetadataRepository()
}

func TestInMemoryIndexMetadataRepository_SaveAndFind(t *testing.T) {
	repo := setupRepository()
	ctx := context.Background()

	// Test 1: Save new metadata (Insert)
	newMetadata := &index.IndexMetadata{
		ID:          uuid.New().String(),
		TableName:   "users",
		ColumnName:  "username",
		IndexType:   enum.IndexTypeFullText,
		Tokenizer:   "standard",
		DataType:    "VARCHAR(255)",
		Description: "Full-text index on user usernames",
		// CreatedAt and UpdatedAt will be set by the Save method
	}

	err := repo.Save(ctx, newMetadata)
	require.NoError(t, err, "Save (insert) should not return an error")
	assert.False(t, newMetadata.CreatedAt.IsZero(), "CreatedAt should be set")
	assert.False(t, newMetadata.UpdatedAt.IsZero(), "UpdatedAt should be set")
	assert.WithinDuration(t, 1*time.Second, time.Now(), newMetadata.CreatedAt, "CreatedAt should be around now")
	assert.WithinDuration(t, 1*time.Second, time.Now(), newMetadata.UpdatedAt, "UpdatedAt should be around now")

	// Test 2: Find by ID
	foundMetadata, err := repo.FindByID(ctx, newMetadata.ID)
	require.NoError(t, err, "FindByID should not return an error for existing ID")
	require.NotNil(t, foundMetadata)
	assert.Equal(t, newMetadata.ID, foundMetadata.ID)
	assert.Equal(t, newMetadata.TableName, foundMetadata.TableName)
	assert.Equal(t, newMetadata.ColumnName, foundMetadata.ColumnName)
	assert.Equal(t, newMetadata.IndexType, foundMetadata.IndexType)
	assert.Equal(t, newMetadata.Tokenizer, foundMetadata.Tokenizer)
	assert.Equal(t, newMetadata.DataType, foundMetadata.DataType)
	assert.Equal(t, newMetadata.Description, foundMetadata.Description)
	assert.Equal(t, newMetadata.CreatedAt, foundMetadata.CreatedAt) // Should be exact same time
	assert.Equal(t, newMetadata.UpdatedAt, foundMetadata.UpdatedAt) // Should be exact same time

	// Test 3: Find by TableColumn
	foundByTC, err := repo.FindByTableColumn(ctx, newMetadata.TableName, newMetadata.ColumnName)
	require.NoError(t, err, "FindByTableColumn should not return an error for existing table/column")
	require.NotNil(t, foundByTC)
	assert.Equal(t, newMetadata.ID, foundByTC.ID)

	// Test 4: Update existing metadata
	updatedDescription := "Updated description for user usernames index"
	newMetadata.Description = updatedDescription
	newMetadata.IndexType = enum.IndexTypeBTree // Change type

	// Simulate time passing to verify UpdatedAt changes
	originalUpdatedAt := newMetadata.UpdatedAt
	time.Sleep(50 * time.Millisecond)

	err = repo.Save(ctx, newMetadata)
	require.NoError(t, err, "Save (update) should not return an error")

	foundUpdated, err := repo.FindByID(ctx, newMetadata.ID)
	require.NoError(t, err)
	assert.Equal(t, updatedDescription, foundUpdated.Description)
	assert.Equal(t, enum.IndexTypeBTree, foundUpdated.IndexType)
	assert.True(t, foundUpdated.UpdatedAt.After(originalUpdatedAt), "UpdatedAt should be newer after update")
	assert.Equal(t, newMetadata.CreatedAt, foundUpdated.CreatedAt, "CreatedAt should not change after update")

	// Test 5: Find non-existent ID
	_, err = repo.FindByID(ctx, uuid.New().String())
	assert.Error(t, err)
	assert.True(t, ferrors.IsNotFoundError(err), "Error should be NotFoundError")

	// Test 6: Find non-existent TableColumn
	_, err = repo.FindByTableColumn(ctx, "nonexistent", "column")
	assert.Error(t, err)
	assert.True(t, ferrors.IsNotFoundError(err), "Error should be NotFoundError")
}

func TestInMemoryIndexMetadataRepository_Save_DuplicateTableColumn(t *testing.T) {
	repo := setupRepository()
	ctx := context.Background()

	meta1 := &index.IndexMetadata{
		ID:         uuid.New().String(),
		TableName:  "orders",
		ColumnName: "order_id",
		IndexType:  enum.IndexTypeBTree,
		DataType:   "INT",
	}
	err := repo.Save(ctx, meta1)
	require.NoError(t, err)

	meta2 := &index.IndexMetadata{
		ID:         uuid.New().String(), // Different ID
		TableName:  "orders",            // Same TableName
		ColumnName: "order_id",          // Same ColumnName
		IndexType:  enum.IndexTypeHash,  // Different Type
		DataType:   "INT",
	}
	err = repo.Save(ctx, meta2)
	assert.Error(t, err)
	assert.True(t, ferrors.IsAlreadyExistsError(err), "Error should be AlreadyExistsError due to unique constraint")
	assert.Contains(t, err.Error(), "index for table orders, column order_id already exists")
}

func TestInMemoryIndexMetadataRepository_ListAll(t *testing.T) {
	repo := setupRepository()
	ctx := context.Background()

	// Initially empty
	allIndexes, err := repo.ListAll(ctx)
	require.NoError(t, err)
	assert.Empty(t, allIndexes)

	// Insert some test data
	meta1 := &index.IndexMetadata{ID: uuid.New().String(), TableName: "t1", ColumnName: "c1", IndexType: enum.IndexTypeFullText, DataType: "TEXT"}
	meta2 := &index.IndexMetadata{ID: uuid.New().String(), TableName: "t2", ColumnName: "c2", IndexType: enum.IndexTypeBTree, DataType: "INT"}
	meta3 := &index.IndexMetadata{ID: uuid.New().String(), TableName: "t1", ColumnName: "c3", IndexType: enum.IndexTypeHash, DataType: "VARCHAR"}

	require.NoError(t, repo.Save(ctx, meta1))
	require.NoError(t, repo.Save(ctx, meta2))
	require.NoError(t, repo.Save(ctx, meta3))

	// Test ListAll
	allIndexes, err = repo.ListAll(ctx)
	require.NoError(t, err)
	assert.Len(t, allIndexes, 3)

	// Check if all elements are present (order might vary)
	expectedIDs := map[string]bool{meta1.ID: true, meta2.ID: true, meta3.ID: true}
	foundIDs := make(map[string]bool)
	for _, idx := range allIndexes {
		foundIDs[idx.ID] = true
	}
	assert.Equal(t, expectedIDs, foundIDs)
}

func TestInMemoryIndexMetadataRepository_ListByIndexedColumns(t *testing.T) {
	repo := setupRepository()
	ctx := context.Background()

	// Insert some test data
	meta1 := &index.IndexMetadata{ID: uuid.New().String(), TableName: "users", ColumnName: "name", IndexType: enum.IndexTypeFullText, DataType: "TEXT"}
	meta2 := &index.IndexMetadata{ID: uuid.New().String(), TableName: "users", ColumnName: "email", IndexType: enum.IndexTypeFullText, DataType: "TEXT"}
	meta3 := &index.IndexMetadata{ID: uuid.New().String(), TableName: "products", ColumnName: "price", IndexType: enum.IndexTypeBTree, DataType: "DECIMAL"}
	meta4 := &index.IndexMetadata{ID: uuid.New().String(), TableName: "orders", ColumnName: "total", IndexType: enum.IndexTypeBTree, DataType: "DECIMAL"}

	require.NoError(t, repo.Save(ctx, meta1))
	require.NoError(t, repo.Save(ctx, meta2))
	require.NoError(t, repo.Save(ctx, meta3))
	require.NoError(t, repo.Save(ctx, meta4))

	// Test 1: Filter by single table name
	results, err := repo.ListByIndexedColumns(ctx, []string{"users"}, nil)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Contains(t, results, meta1)
	assert.Contains(t, results, meta2)

	// Test 2: Filter by multiple table names
	results, err = repo.ListByIndexedColumns(ctx, []string{"users", "products"}, nil)
	require.NoError(t, err)
	assert.Len(t, results, 3)
	assert.Contains(t, results, meta1)
	assert.Contains(t, results, meta2)
	assert.Contains(t, results, meta3)

	// Test 3: Filter by single column name
	results, err = repo.ListByIndexedColumns(ctx, nil, []string{"price"})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Contains(t, results, meta3)

	// Test 4: Filter by table name and column name (AND condition)
	results, err = repo.ListByIndexedColumns(ctx, []string{"users"}, []string{"name"})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Contains(t, results, meta1)

	// Test 5: Filter with no matching data
	results, err = repo.ListByIndexedColumns(ctx, []string{"nonexistent_table"}, nil)
	require.NoError(t, err)
	assert.Len(t, results, 0)

	// Test 6: No filters (should behave like ListAll)
	results, err = repo.ListByIndexedColumns(ctx, nil, nil)
	require.NoError(t, err)
	assert.Len(t, results, 4)
	assert.Contains(t, results, meta1)
	assert.Contains(t, results, meta2)
	assert.Contains(t, results, meta3)
	assert.Contains(t, results, meta4)
}

func TestInMemoryIndexMetadataRepository_DeleteByID(t *testing.T) {
	repo := setupRepository()
	ctx := context.Background()

	// Insert a record to delete
	meta := &index.IndexMetadata{
		ID:         uuid.New().String(),
		TableName:  "temp_table",
		ColumnName: "temp_column",
		IndexType:  enum.IndexTypeFullText,
		DataType:   "TEXT",
	}
	require.NoError(t, repo.Save(ctx, meta))

	// Verify it exists before deletion
	_, err := repo.FindByID(ctx, meta.ID)
	require.NoError(t, err)

	// Test 1: Delete existing record
	err = repo.DeleteByID(ctx, meta.ID)
	assert.NoError(t, err, "DeleteByID for existing record should not return an error")

	// Verify it's deleted
	_, err = repo.FindByID(ctx, meta.ID)
	assert.Error(t, err)
	assert.True(t, ferrors.IsNotFoundError(err), "Record should be not found after deletion")

	// Test 2: Delete non-existent record (should not error)
	nonExistentID := uuid.New().String()
	err = repo.DeleteByID(ctx, nonExistentID)
	assert.NoError(t, err, "DeleteByID for non-existent record should not return an error")
}

func TestInMemoryIndexMetadataRepository_Save_EmptyID(t *testing.T) {
	repo := setupRepository()
	ctx := context.Background()

	metadata := &index.IndexMetadata{
		ID:         "", // Empty ID
		TableName:  "invalid",
		ColumnName: "id",
		IndexType:  enum.IndexTypeBTree,
		DataType:   "INT",
	}

	err := repo.Save(ctx, metadata)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "metadata ID cannot be empty")
	assert.True(t, ferrors.IsInternalError(err))
}
