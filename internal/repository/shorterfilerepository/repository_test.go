package shorterfilerepository

import (
	"os"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {

	tmpFile, err := os.CreateTemp("", "test_shortener_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	repo := New(tmpFile.Name())

	if repo == nil {
		t.Fatal("New() returned nil")
	}

	if repo.filePath != tmpFile.Name() {
		t.Errorf("expected filepath to be %s, got %s", tmpFile.Name(), repo.filePath)
	}

	if repo.shortToLink == nil {
		t.Error("shortToLink map not initialized")
	}

	if repo.userToURLs == nil {
		t.Error("userToURLs map not initialized")
	}
}

func TestRepo_Add(t *testing.T) {
	ctx := t.Context()

	tmpFile, err := os.CreateTemp("", "test_shortener_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	repo := New(tmpFile.Name())
	url := "https://example.com"
	userID := "user123"

	alias, err := repo.Add(ctx, url, userID)
	if err != nil {
		t.Fatalf("Add() returned an error: %v", err)
	}

	if len(alias) == 0 {
		t.Error("expected non-empty alias")
	}

	storedAlias, exists := repo.shortToLink[alias]
	if !exists {
		t.Error("alias not stored in repository")
	}

	if storedAlias != url {
		t.Errorf("expected URL to be %s, got %s", url, storedAlias)
	}

	userURLs, err := repo.GetByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("GetByUserID() returned an error: %v", err)
	}

	if len(userURLs) != 1 {
		t.Errorf("expected 1 user URL, got %d", len(userURLs))
	}

	if userURLs[0].ShortURL != alias || userURLs[0].OriginalURL != url {
		t.Errorf("user URL mismatch: got %s -> %s, expected %s -> %s",
			userURLs[0].ShortURL, userURLs[0].OriginalURL, alias, url)
	}

	sameAlias, err := repo.Add(ctx, url, userID)
	if err != nil {
		t.Fatalf("Add() for same URL returned an error: %v", err)
	}

	if sameAlias != alias {
		t.Errorf("expected same alias for same URL, got different: %s vs %s", alias, sameAlias)
	}
}

func TestRepo_AddWithoutUser(t *testing.T) {
	ctx := t.Context()

	tmpFile, err := os.CreateTemp("", "test_shortener_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	repo := New(tmpFile.Name())
	url := "https://example.com"

	alias, err := repo.Add(ctx, url, "")
	if err != nil {
		t.Fatalf("Add() returned an error: %v", err)
	}

	if len(alias) == 0 {
		t.Error("expected non-empty alias")
	}

	storedAlias, exists := repo.linkToShort[url]
	if !exists {
		t.Error("URL not stored in repository")
	}

	if storedAlias == "" {
		t.Error("expected non-empty alias")
	}

	userURLs, err := repo.GetByUserID(ctx, "anyuser")
	if err != nil {
		t.Fatalf("GetByUserID() returned an error: %v", err)
	}

	if len(userURLs) != 0 {
		t.Errorf("expected 0 user URLs, got %d", len(userURLs))
	}
}

func TestRepo_Get(t *testing.T) {
	ctx := t.Context()

	tmpFile, err := os.CreateTemp("", "test_shortener_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	repo := New(tmpFile.Name())
	url := "https://example.com"

	alias, err := repo.Add(ctx, url, "")
	if err != nil {
		t.Fatalf("Add() returned an error: %v", err)
	}

	retrievedURL, err := repo.Get(ctx, alias)
	if err != nil {
		t.Fatalf("Get() returned an error: %v", err)
	}

	if retrievedURL != url {
		t.Errorf("expected URL %s, got %s", url, retrievedURL)
	}

	_, err = repo.Get(ctx, "nonexistent-alias")
	if err == nil {
		t.Error("expected error for non-existent alias, got nil")
	}
}

func TestRepo_GetByUserID(t *testing.T) {
	ctx := t.Context()

	tmpFile, err := os.CreateTemp("", "test_shortener_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	repo := New(tmpFile.Name())
	userID1 := "user1"
	userID2 := "user2"
	url1 := "https://example1.com"
	url2 := "https://example2.com"
	url3 := "https://example3.com"

	_, err = repo.Add(ctx, url1, userID1)
	if err != nil {
		t.Fatalf("Add() returned an error: %v", err)
	}

	_, err = repo.Add(ctx, url2, userID1)
	if err != nil {
		t.Fatalf("Add() returned an error: %v", err)
	}

	_, err = repo.Add(ctx, url3, userID2)
	if err != nil {
		t.Fatalf("Add() returned an error: %v", err)
	}

	userURLs1, err := repo.GetByUserID(ctx, userID1)
	if err != nil {
		t.Fatalf("GetByUserID() returned an error: %v", err)
	}

	if len(userURLs1) != 2 {
		t.Errorf("expected 2 user URLs for user1, got %d", len(userURLs1))
	}

	userURLs2, err := repo.GetByUserID(ctx, userID2)
	if err != nil {
		t.Fatalf("GetByUserID() returned an error: %v", err)
	}

	if len(userURLs2) != 1 {
		t.Errorf("expected 1 user URL for user2, got %d", len(userURLs2))
	}

	userURLs3, err := repo.GetByUserID(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("GetByUserID() returned an error: %v", err)
	}

	if len(userURLs3) != 0 {
		t.Errorf("expected 0 user URLs for non-existent user, got %d", len(userURLs3))
	}
}

func TestRepo_saveToFile(t *testing.T) {
	ctx := t.Context()

	tmpFile, err := os.CreateTemp("", "test_shortener_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	repo := New(tmpFile.Name())
	url := "https://example.com"
	userID := "user123"

	_, err = repo.Add(ctx, url, userID)
	if err != nil {
		t.Fatalf("Add() returned an error: %v", err)
	}

	err = repo.saveToFile()
	if err != nil {
		t.Fatalf("saveToFile() returned an error: %v", err)
	}

	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if len(content) == 0 {
		t.Error("expected file to have content")
	}

	if !strings.Contains(string(content), url) {
		t.Error("expected file to contain the URL")
	}

	if !strings.Contains(string(content), userID) {
		t.Error("expected file to contain the user ID")
	}
}

func TestRepo_loadFromFile(t *testing.T) {
	ctx := t.Context()

	tmpFile, err := os.CreateTemp("", "test_shortener_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	testData := `[
		{
			"uuid": "1",
			"short_url": "abc123",
			"original_url": "https://example1.com",
			"user_id": "user1"
		},
		{
			"uuid": "2",
			"short_url": "def456",
			"original_url": "https://example2.com",
			"user_id": "user2"
		}
	]`
	err = os.WriteFile(tmpFile.Name(), []byte(testData), 0644)
	if err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}

	repo := New(tmpFile.Name())

	if len(repo.shortToLink) != 2 {
		t.Errorf("expected 2 URLs in shortToLink, got %d", len(repo.shortToLink))
	}

	userURLs1, err := repo.GetByUserID(ctx, "user1")
	if err != nil {
		t.Fatalf("GetByUserID() returned an error: %v", err)
	}

	if len(userURLs1) != 1 {
		t.Errorf("expected 1 user URL for user1, got %d", len(userURLs1))
	}

	if userURLs1[0].ShortURL != "abc123" || userURLs1[0].OriginalURL != "https://example1.com" {
		t.Errorf("user URL mismatch: got %s -> %s, expected abc123 -> https://example1.com",
			userURLs1[0].ShortURL, userURLs1[0].OriginalURL)
	}
}
