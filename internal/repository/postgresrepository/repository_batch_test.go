package postgresrepository

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"go-musthave-shortener/internal/database"
)

func setupTestPostgresRepo(t *testing.T) (*PostgresRepo, func()) {
	t.Helper()
	testDSN := "postgres://postgres:postgres@localhost:5432/test_shortener?sslmode=disable"

	db, err := database.New(testDSN)
	if err != nil {
		t.Skipf("Skipping PostgreSQL test: cannot connect to database: %v", err)
	}

	ctx := context.Background()
	_, err = db.Pool().Exec(ctx, "TRUNCATE TABLE urls RESTART IDENTITY CASCADE")
	if err != nil {
		t.Skipf("Skipping PostgreSQL test: cannot prepare database: %v", err)
	}

	logger := zap.NewNop()
	repo := New(db.Pool(), logger)

	cleanup := func() {
		db.Close()
	}

	return repo, cleanup
}

func TestPostgresRepo_AddBatch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL integration test in short mode")
	}

	repo, cleanup := setupTestPostgresRepo(t)
	defer cleanup()

	result, err := repo.AddBatch(context.TODO(), []string{})
	assert.NoError(t, err)
	assert.Empty(t, result)

	urls := []string{"https://example.com"}
	result, err = repo.AddBatch(context.TODO(), urls)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.NotEmpty(t, result[0])

	retrievedURL, err := repo.Get(context.TODO(), result[0])
	assert.NoError(t, err)
	assert.Equal(t, urls[0], retrievedURL)

	urls = []string{
		"https://practicum.yandex.ru",
		"https://example.com",
		"https://google.com",
	}
	result, err = repo.AddBatch(context.TODO(), urls)
	assert.NoError(t, err)
	assert.Len(t, result, 3)

	for i, alias := range result {
		assert.NotEmpty(t, alias)
		retrievedURL, err := repo.Get(context.TODO(), alias)
		assert.NoError(t, err)
		assert.Equal(t, urls[i], retrievedURL)
	}

	aliasSet := make(map[string]bool)
	for _, alias := range result {
		assert.False(t, aliasSet[alias], "Alias should be unique: %s", alias)
		aliasSet[alias] = true
	}
}

func TestPostgresRepo_AddBatch_ExistingURLs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL integration test in short mode")
	}

	repo, cleanup := setupTestPostgresRepo(t)
	defer cleanup()

	originalAlias, err := repo.Add(context.TODO(), "https://example.com")
	assert.NoError(t, err)

	urls := []string{"https://example.com", "https://new-url.com"}
	result, err := repo.AddBatch(context.TODO(), urls)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	assert.Equal(t, originalAlias, result[0])

	assert.NotEmpty(t, result[1])
	assert.NotEqual(t, originalAlias, result[1])

	retrievedURL1, err := repo.Get(context.TODO(), result[0])
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com", retrievedURL1)

	retrievedURL2, err := repo.Get(context.TODO(), result[1])
	assert.NoError(t, err)
	assert.Equal(t, "https://new-url.com", retrievedURL2)
}

func TestPostgresRepo_AddBatch_DuplicateURLsInBatch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL integration test in short mode")
	}

	repo, cleanup := setupTestPostgresRepo(t)
	defer cleanup()

	urls := []string{
		"https://example.com",
		"https://example.com",
		"https://google.com",
		"https://example.com",
	}
	result, err := repo.AddBatch(context.TODO(), urls)
	assert.NoError(t, err)
	assert.Len(t, result, 4)

	assert.Equal(t, result[0], result[1])
	assert.Equal(t, result[1], result[3])
	assert.NotEqual(t, result[0], result[2])

	for _, alias := range result {
		retrievedURL, err := repo.Get(context.TODO(), alias)
		assert.NoError(t, err)
		assert.True(t, retrievedURL == "https://example.com" || retrievedURL == "https://google.com")
	}
}

func TestPostgresRepo_AddBatch_LargeBatch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL integration test in short mode")
	}

	repo, cleanup := setupTestPostgresRepo(t)
	defer cleanup()

	urls := make([]string, 100)
	for i := 0; i < 100; i++ {
		urls[i] = fmt.Sprintf("https://example%d.com", i)
	}

	result, err := repo.AddBatch(context.TODO(), urls)
	assert.NoError(t, err)
	assert.Len(t, result, 100)

	for i, alias := range result {
		assert.NotEmpty(t, alias)
		retrievedURL, err := repo.Get(context.TODO(), alias)
		assert.NoError(t, err)
		assert.Equal(t, urls[i], retrievedURL)
	}

	aliasSet := make(map[string]bool)
	for _, alias := range result {
		assert.False(t, aliasSet[alias], "Alias should be unique: %s", alias)
		aliasSet[alias] = true
	}
}

func TestPostgresRepo_AddBatch_ConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL integration test in short mode")
	}

	repo, cleanup := setupTestPostgresRepo(t)
	defer cleanup()

	var wg sync.WaitGroup
	results := make([][]string, 10)
	errors := make([]error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			urls := make([]string, 5)
			for j := 0; j < 5; j++ {
				urls[j] = fmt.Sprintf("https://concurrent%d-%d.com", index, j)
			}

			results[index], errors[index] = repo.AddBatch(context.TODO(), urls)
		}(i)
	}

	wg.Wait()

	for i := 0; i < 10; i++ {
		assert.NoError(t, errors[i])
		assert.Len(t, results[i], 5)
	}

	for i := 0; i < 10; i++ {
		for j := 0; j < 5; j++ {
			alias := results[i][j]
			expectedURL := fmt.Sprintf("https://concurrent%d-%d.com", i, j)

			retrievedURL, err := repo.Get(context.TODO(), alias)
			assert.NoError(t, err)
			assert.Equal(t, expectedURL, retrievedURL)
		}
	}

	allAliases := make(map[string]bool)
	for i := 0; i < 10; i++ {
		for _, alias := range results[i] {
			assert.False(t, allAliases[alias], "Alias should be unique across all batches: %s", alias)
			allAliases[alias] = true
		}
	}
}

func TestPostgresRepo_AddBatch_ConcurrentSameURLs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL integration test in short mode")
	}

	repo, cleanup := setupTestPostgresRepo(t)
	defer cleanup()

	var wg sync.WaitGroup
	results := make([][]string, 10)
	errors := make([]error, 10)
	testURL := "https://shared-url.com"

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			urls := []string{testURL}
			results[index], errors[index] = repo.AddBatch(context.TODO(), urls)
		}(i)
	}

	wg.Wait()

	for i := 0; i < 10; i++ {
		assert.NoError(t, errors[i])
		assert.Len(t, results[i], 1)
	}

	firstAlias := results[0][0]
	for i := 1; i < 10; i++ {
		assert.Equal(t, firstAlias, results[i][0], "All batches should return the same alias for the same URL")
	}

	retrievedURL, err := repo.Get(context.TODO(), firstAlias)
	assert.NoError(t, err)
	assert.Equal(t, testURL, retrievedURL)
}

func TestPostgresRepo_AddBatch_TransactionHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL integration test in short mode")
	}

	repo, cleanup := setupTestPostgresRepo(t)
	defer cleanup()

	initialAlias, err := repo.Add(context.TODO(), "https://initial.com")
	require.NoError(t, err)

	urls := []string{"https://success1.com", "https://success2.com"}
	result, err := repo.AddBatch(context.TODO(), urls)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	for i, alias := range result {
		retrievedURL, err := repo.Get(context.TODO(), alias)
		assert.NoError(t, err)
		assert.Equal(t, urls[i], retrievedURL)
	}

	retrievedURL, err := repo.Get(context.TODO(), initialAlias)
	assert.NoError(t, err)
	assert.Equal(t, "https://initial.com", retrievedURL)
}

func TestPostgresRepo_AddBatch_Isolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL integration test in short mode")
	}

	repo, cleanup := setupTestPostgresRepo(t)
	defer cleanup()

	var wg sync.WaitGroup
	results := make([][]string, 5)
	errors := make([]error, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			urls := []string{
				fmt.Sprintf("https://isolation%d-1.com", index),
				fmt.Sprintf("https://isolation%d-2.com", index),
			}

			results[index], errors[index] = repo.AddBatch(context.TODO(), urls)
		}(i)
	}

	wg.Wait()

	for i := 0; i < 5; i++ {
		assert.NoError(t, errors[i])
		assert.Len(t, results[i], 2)
	}

	for i := 0; i < 5; i++ {
		for j := 0; j < 2; j++ {
			alias := results[i][j]
			expectedURL := fmt.Sprintf("https://isolation%d-%d.com", i, j+1)

			retrievedURL, err := repo.Get(context.TODO(), alias)
			assert.NoError(t, err)
			assert.Equal(t, expectedURL, retrievedURL)
		}
	}

	allAliases := make(map[string]bool)
	for i := 0; i < 5; i++ {
		for _, alias := range results[i] {
			assert.False(t, allAliases[alias], "Alias should be unique: %s", alias)
			allAliases[alias] = true
		}
	}
}
