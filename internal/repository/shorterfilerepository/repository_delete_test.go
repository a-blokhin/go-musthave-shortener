package shorterfilerepository

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteURLs(t *testing.T) {
	ctx := t.Context()

	tmpFile, err := os.CreateTemp("", "test_repo_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	repo := New(tmpFile.Name())

	userID := "test-user"
	otherUserID := "other-user"

	shortURL1, err := repo.Add(ctx, "https://example1.com", userID)
	require.NoError(t, err)

	shortURL2, err := repo.Add(ctx, "https://example2.com", userID)
	require.NoError(t, err)

	shortURL3, err := repo.Add(ctx, "https://example3.com", otherUserID)
	require.NoError(t, err)

	err = repo.BatchDelete(ctx, []string{shortURL1, shortURL2}, userID)
	require.NoError(t, err)

	_, err = repo.Get(ctx, shortURL1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "has been deleted")

	_, err = repo.Get(ctx, shortURL2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "has been deleted")

	originalURL, err := repo.Get(ctx, shortURL3)
	assert.NoError(t, err)
	assert.Equal(t, "https://example3.com", originalURL)

	userURLs, err := repo.GetByUserID(ctx, userID)
	require.NoError(t, err)
	assert.Empty(t, userURLs)

	otherUserURLs, err := repo.GetByUserID(ctx, otherUserID)
	require.NoError(t, err)
	assert.Len(t, otherUserURLs, 1)
	assert.Equal(t, shortURL3, otherUserURLs[0].ShortURL)

	repo2 := New(tmpFile.Name())

	_, err = repo2.Get(ctx, shortURL1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "has been deleted")

	_, err = repo2.Get(ctx, shortURL2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "has been deleted")

	originalURL, err = repo2.Get(ctx, shortURL3)
	assert.NoError(t, err)
	assert.Equal(t, "https://example3.com", originalURL)
}

func TestDeleteURLs_EmptyList(t *testing.T) {
	ctx := t.Context()

	tmpFile, err := os.CreateTemp("", "test_repo_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	repo := New(tmpFile.Name())

	userID := "test-user"

	err = repo.BatchDelete(ctx, []string{}, userID)
	assert.NoError(t, err)
}

func TestDeleteURLs_NonExistentURLs(t *testing.T) {
	ctx := t.Context()

	tmpFile, err := os.CreateTemp("", "test_repo_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	repo := New(tmpFile.Name())

	userID := "test-user"

	err = repo.BatchDelete(ctx, []string{"nonexistent1", "nonexistent2"}, userID)
	assert.NoError(t, err)
}

func TestDeleteURLs_MixedOwnership(t *testing.T) {
	ctx := t.Context()

	tmpFile, err := os.CreateTemp("", "test_repo_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	repo := New(tmpFile.Name())

	userID := "test-user"
	otherUserID := "other-user"

	shortURL1, err := repo.Add(ctx, "https://example1.com", userID)
	require.NoError(t, err)

	shortURL2, err := repo.Add(ctx, "https://example2.com", otherUserID)
	require.NoError(t, err)

	shortURL3, err := repo.Add(ctx, "https://example3.com", userID)
	require.NoError(t, err)

	err = repo.BatchDelete(ctx, []string{shortURL1, shortURL2, shortURL3, "nonexistent"}, userID)
	require.NoError(t, err)

	_, err = repo.Get(ctx, shortURL1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "has been deleted")

	_, err = repo.Get(ctx, shortURL3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "has been deleted")

	originalURL, err := repo.Get(ctx, shortURL2)
	assert.NoError(t, err)
	assert.Equal(t, "https://example2.com", originalURL)

	userURLs, err := repo.GetByUserID(ctx, userID)
	require.NoError(t, err)
	assert.Empty(t, userURLs)
}

func TestDeleteURLs_UserWithoutURLs(t *testing.T) {
	ctx := t.Context()

	tmpFile, err := os.CreateTemp("", "test_repo_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	repo := New(tmpFile.Name())

	userID := "test-user"
	otherUserID := "other-user"

	shortURL1, err := repo.Add(ctx, "https://example1.com", otherUserID)
	require.NoError(t, err)

	err = repo.BatchDelete(ctx, []string{shortURL1}, userID)
	require.NoError(t, err)

	originalURL, err := repo.Get(ctx, shortURL1)
	assert.NoError(t, err)
	assert.Equal(t, "https://example1.com", originalURL)
}

func TestDeleteURLs_FilePersistence(t *testing.T) {
	ctx := t.Context()

	tmpFile, err := os.CreateTemp("", "test_repo_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	repo := New(tmpFile.Name())

	userID := "test-user"

	shortURL1, err := repo.Add(ctx, "https://example1.com", userID)
	require.NoError(t, err)

	shortURL2, err := repo.Add(ctx, "https://example2.com", userID)
	require.NoError(t, err)

	err = repo.BatchDelete(ctx, []string{shortURL1}, userID)
	require.NoError(t, err)

	repo2 := New(tmpFile.Name())

	_, err = repo2.Get(ctx, shortURL1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "has been deleted")

	originalURL, err := repo2.Get(ctx, shortURL2)
	assert.NoError(t, err)
	assert.Equal(t, "https://example2.com", originalURL)

	userURLs, err := repo2.GetByUserID(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, userURLs, 1)
	assert.Equal(t, shortURL2, userURLs[0].ShortURL)
}
