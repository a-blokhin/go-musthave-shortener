package shorterfilerepository

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestFileRepo(t *testing.T) (*FileRepo, string) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test_urls.json")
	repo := New(filePath)
	return repo, filePath
}

func TestFileRepo_AddBatch(t *testing.T) {
	repo, _ := setupTestFileRepo(t)

	result, err := repo.AddBatch([]string{})
	assert.NoError(t, err)
	assert.Empty(t, result)

	urls := []string{"https://example.com"}
	result, err = repo.AddBatch(urls)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.NotEmpty(t, result[0])

	retrievedURL, err := repo.Get(result[0])
	assert.NoError(t, err)
	assert.Equal(t, urls[0], retrievedURL)

	urls = []string{
		"https://practicum.yandex.ru",
		"https://example.com",
		"https://google.com",
	}
	result, err = repo.AddBatch(urls)
	assert.NoError(t, err)
	assert.Len(t, result, 3)

	for i, alias := range result {
		assert.NotEmpty(t, alias)
		retrievedURL, err := repo.Get(alias)
		assert.NoError(t, err)
		assert.Equal(t, urls[i], retrievedURL)
	}

	aliasSet := make(map[string]bool)
	for _, alias := range result {
		assert.False(t, aliasSet[alias], "Alias should be unique: %s", alias)
		aliasSet[alias] = true
	}
}

func TestFileRepo_AddBatch_ExistingURLs(t *testing.T) {
	repo, _ := setupTestFileRepo(t)

	originalAlias, err := repo.Add("https://example.com")
	assert.NoError(t, err)

	urls := []string{"https://example.com", "https://new-url.com"}
	result, err := repo.AddBatch(urls)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	assert.Equal(t, originalAlias, result[0])

	assert.NotEmpty(t, result[1])
	assert.NotEqual(t, originalAlias, result[1])

	retrievedURL1, err := repo.Get(result[0])
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com", retrievedURL1)

	retrievedURL2, err := repo.Get(result[1])
	assert.NoError(t, err)
	assert.Equal(t, "https://new-url.com", retrievedURL2)
}

func TestFileRepo_AddBatch_DuplicateURLsInBatch(t *testing.T) {
	repo, _ := setupTestFileRepo(t)

	urls := []string{
		"https://example.com",
		"https://example.com",
		"https://google.com",
		"https://example.com",
	}
	result, err := repo.AddBatch(urls)
	assert.NoError(t, err)
	assert.Len(t, result, 4)

	assert.Equal(t, result[0], result[1])
	assert.Equal(t, result[1], result[3])
	assert.NotEqual(t, result[0], result[2])

	for _, alias := range result {
		retrievedURL, err := repo.Get(alias)
		assert.NoError(t, err)
		assert.True(t, retrievedURL == "https://example.com" || retrievedURL == "https://google.com")
	}
}

func TestFileRepo_AddBatch_LargeBatch(t *testing.T) {
	repo, _ := setupTestFileRepo(t)

	urls := make([]string, 100)
	for i := 0; i < 100; i++ {
		urls[i] = fmt.Sprintf("https://example%d.com", i)
	}

	result, err := repo.AddBatch(urls)
	assert.NoError(t, err)
	assert.Len(t, result, 100)

	for i, alias := range result {
		assert.NotEmpty(t, alias)
		retrievedURL, err := repo.Get(alias)
		assert.NoError(t, err)
		assert.Equal(t, urls[i], retrievedURL)
	}

	aliasSet := make(map[string]bool)
	for _, alias := range result {
		assert.False(t, aliasSet[alias], "Alias should be unique: %s", alias)
		aliasSet[alias] = true
	}
}

func TestFileRepo_AddBatch_RollbackOnFailure(t *testing.T) {

	tempDir := t.TempDir()

	invalidPath := filepath.Join(tempDir, "subdir")
	err := os.Mkdir(invalidPath, 0755)
	require.NoError(t, err)

	repo := New(invalidPath)

	urls := []string{"https://new1.com", "https://new2.com"}
	result, err := repo.AddBatch(urls)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestFileRepo_AddBatch_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "empty_urls.json")

	file, err := os.Create(filePath)
	require.NoError(t, err)
	file.Close()

	repo := New(filePath)
	urls := []string{"https://emptyfile.com"}
	result, err := repo.AddBatch(urls)

	assert.NoError(t, err)
	assert.Len(t, result, 1)

	retrievedURL, err := repo.Get(result[0])
	assert.NoError(t, err)
	assert.Equal(t, urls[0], retrievedURL)
}

func TestFileRepo_AddBatch_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "nonexistent_urls.json")

	repo := New(filePath)
	urls := []string{"https://nonexistent.com"}
	result, err := repo.AddBatch(urls)

	assert.NoError(t, err)
	assert.Len(t, result, 1)

	retrievedURL, err := repo.Get(result[0])
	assert.NoError(t, err)
	assert.Equal(t, urls[0], retrievedURL)

	_, err = os.Stat(filePath)
	assert.NoError(t, err)
}
