package shorterfilerepository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.json")

	repo := New(filePath)

	if repo == nil {
		t.Fatal("New() returned nil")
	}

	if repo.shortToLink == nil {
		t.Error("shortToLink map not initialized")
	}

	if repo.linkToShort == nil {
		t.Error("linkToShort map not initialized")
	}

	if repo.aliasLength != 8 {
		t.Errorf("expected aliasLength to be 8, got %d", repo.aliasLength)
	}

	if repo.filePath != filePath {
		t.Errorf("expected filePath to be %s, got %s", filePath, repo.filePath)
	}
}

func TestFileRepo_Add(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.json")

	repo := New(filePath)
	url := "https://example.com"

	alias, err := repo.Add(url)
	if err != nil {
		t.Fatalf("Add() returned an error: %v", err)
	}

	if len(alias) != repo.aliasLength {
		t.Errorf("expected alias length to be %d, got %d", repo.aliasLength, len(alias))
	}

	if !repo.hasAlias(alias) {
		t.Error("alias not stored in repository")
	}

	if !repo.hasLink(url) {
		t.Error("URL not stored in repository")
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("file was not created after Add()")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	var urlDataList []URLData
	if err := json.Unmarshal(data, &urlDataList); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if len(urlDataList) != 1 {
		t.Errorf("expected 1 entry in file, got %d", len(urlDataList))
	}

	if urlDataList[0].ShortURL != alias {
		t.Errorf("expected ShortURL to be %s, got %s", alias, urlDataList[0].ShortURL)
	}

	if urlDataList[0].OriginalURL != url {
		t.Errorf("expected OriginalURL to be %s, got %s", url, urlDataList[0].OriginalURL)
	}

	sameAlias, err := repo.Add(url)
	if err != nil {
		t.Fatalf("Add() for same URL returned an error: %v", err)
	}

	if sameAlias != alias {
		t.Errorf("expected same alias for same URL, got different: %s vs %s", alias, sameAlias)
	}
}

func TestFileRepo_Get(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.json")

	repo := New(filePath)
	url := "https://example.com"

	alias, err := repo.Add(url)
	if err != nil {
		t.Fatalf("Add() returned an error: %v", err)
	}

	retrievedURL, err := repo.Get(alias)
	if err != nil {
		t.Fatalf("Get() returned an error: %v", err)
	}

	if retrievedURL != url {
		t.Errorf("expected URL %s, got %s", url, retrievedURL)
	}

	_, err = repo.Get("nonexistent-alias")
	if err == nil {
		t.Error("expected error for non-existent alias, got nil")
	}
}

func TestFileRepo_Persistence(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.json")

	repo1 := New(filePath)

	urls := map[string]string{
		"https://example.com": "",
		"https://google.com":  "",
		"https://github.com":  "",
	}

	for url := range urls {
		alias, err := repo1.Add(url)
		if err != nil {
			t.Fatalf("Add() returned an error: %v", err)
		}
		urls[url] = alias
	}

	repo2 := New(filePath)

	for url, alias := range urls {
		retrievedURL, err := repo2.Get(alias)
		if err != nil {
			t.Errorf("Get() returned an error for alias %s: %v", alias, err)
		}

		if retrievedURL != url {
			t.Errorf("expected URL %s, got %s", url, retrievedURL)
		}

		sameAlias, err := repo2.Add(url)
		if err != nil {
			t.Fatalf("Add() returned an error: %v", err)
		}

		if sameAlias != alias {
			t.Errorf("expected same alias %s for restored URL, got %s", alias, sameAlias)
		}
	}

	newURL := "https://stackoverflow.com"
	newAlias, err := repo2.Add(newURL)
	if err != nil {
		t.Fatalf("Add() returned an error: %v", err)
	}

	repo3 := New(filePath)

	retrievedURL, err := repo3.Get(newAlias)
	if err != nil {
		t.Fatalf("Get() returned an error: %v", err)
	}

	if retrievedURL != newURL {
		t.Errorf("expected URL %s, got %s", newURL, retrievedURL)
	}

	for url, alias := range urls {
		retrievedURL, err := repo3.Get(alias)
		if err != nil {
			t.Errorf("Get() returned an error for alias %s: %v", alias, err)
		}

		if retrievedURL != url {
			t.Errorf("expected URL %s, got %s", url, retrievedURL)
		}
	}
}

func TestFileRepo_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "empty.json")

	if err := os.WriteFile(filePath, []byte{}, 0644); err != nil {
		t.Fatalf("failed to create empty file: %v", err)
	}

	repo := New(filePath)

	if repo == nil {
		t.Fatal("New() returned nil for empty file")
	}

	url := "https://example.com"
	alias, err := repo.Add(url)
	if err != nil {
		t.Fatalf("Add() returned an error: %v", err)
	}

	retrievedURL, err := repo.Get(alias)
	if err != nil {
		t.Fatalf("Get() returned an error: %v", err)
	}

	if retrievedURL != url {
		t.Errorf("expected URL %s, got %s", url, retrievedURL)
	}
}

func TestFileRepo_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "invalid.json")

	invalidJSON := []byte(`{"invalid": json content}`)
	if err := os.WriteFile(filePath, invalidJSON, 0644); err != nil {
		t.Fatalf("failed to create file with invalid JSON: %v", err)
	}

	repo := New(filePath)

	if repo == nil {
		t.Fatal("New() returned nil for invalid JSON file")
	}

	if len(repo.shortToLink) != 0 {
		t.Error("repository should be empty after loading invalid JSON")
	}
}

func TestFileRepo_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "nonexistent.json")

	repo := New(filePath)

	if repo == nil {
		t.Fatal("New() returned nil for non-existent file")
	}

	url := "https://example.com"
	alias, err := repo.Add(url)
	if err != nil {
		t.Fatalf("Add() returned an error: %v", err)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("file should be created after Add()")
	}

	retrievedURL, err := repo.Get(alias)
	if err != nil {
		t.Fatalf("Get() returned an error: %v", err)
	}

	if retrievedURL != url {
		t.Errorf("expected URL %s, got %s", url, retrievedURL)
	}
}

func TestFileRepo_hasLink(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.json")

	repo := New(filePath)
	url := "https://example.com"

	if repo.hasLink(url) {
		t.Error("hasLink() should return false for non-existent URL")
	}

	_, err := repo.Add(url)
	if err != nil {
		t.Fatalf("Add() returned an error: %v", err)
	}

	if !repo.hasLink(url) {
		t.Error("hasLink() should return true for existing URL")
	}
}

func TestFileRepo_hasAlias(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.json")

	repo := New(filePath)
	url := "https://example.com"

	alias, err := repo.Add(url)
	if err != nil {
		t.Fatalf("Add() returned an error: %v", err)
	}

	if !repo.hasAlias(alias) {
		t.Error("hasAlias() should return true for existing alias")
	}

	if repo.hasAlias("nonexistent-alias") {
		t.Error("hasAlias() should return false for non-existent alias")
	}
}

func TestGenerateAlias(t *testing.T) {
	length := 10
	alias := generateAlias(length)

	if len(alias) != length {
		t.Errorf("expected alias length to be %d, got %d", length, len(alias))
	}

	for _, char := range alias {
		if !strings.ContainsRune(charset, char) {
			t.Errorf("character %c is not from the defined charset", char)
		}
	}

	alias2 := generateAlias(length)
	if alias == alias2 {
		t.Error("generateAlias() should return different values on subsequent calls")
	}
}

func TestFileRepo_JSONFormat(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.json")

	repo := New(filePath)

	urls := []struct {
		url   string
		alias string
	}{
		{url: "http://yandex.ru"},
		{url: "http://ya.ru"},
		{url: "http://practicum.yandex.ru"},
	}

	for i := range urls {
		alias, err := repo.Add(urls[i].url)
		if err != nil {
			t.Fatalf("Add() returned an error: %v", err)
		}
		urls[i].alias = alias
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	var urlDataList []URLData
	if err := json.Unmarshal(data, &urlDataList); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if len(urlDataList) != len(urls) {
		t.Errorf("expected %d entries in file, got %d", len(urls), len(urlDataList))
	}

	uuidMap := make(map[string]bool)
	for _, entry := range urlDataList {
		if entry.UUID == "" {
			t.Error("UUID should not be empty")
		}

		if uuidMap[entry.UUID] {
			t.Errorf("duplicate UUID found: %s", entry.UUID)
		}
		uuidMap[entry.UUID] = true

		if entry.ShortURL == "" {
			t.Error("ShortURL should not be empty")
		}
		if entry.OriginalURL == "" {
			t.Error("OriginalURL should not be empty")
		}
	}
}