package shorterrepository

import (
	"context"
	"testing"
)

func TestRepo_AddBatch(t *testing.T) {
	repo := New()

	aliases, err := repo.AddBatch(context.TODO(), []string{})
	if err != nil {
		t.Fatalf("AddBatch() with empty slice returned an error: %v", err)
	}
	if len(aliases) != 0 {
		t.Errorf("expected empty result for empty input, got %d aliases", len(aliases))
	}

	urls := []string{"https://example.com"}
	aliases, err = repo.AddBatch(context.TODO(), urls)
	if err != nil {
		t.Fatalf("AddBatch() returned an error: %v", err)
	}
	if len(aliases) != 1 {
		t.Errorf("expected 1 alias, got %d", len(aliases))
	}
	if len(aliases[0]) != repo.aliasLength {
		t.Errorf("expected alias length to be %d, got %d", repo.aliasLength, len(aliases[0]))
	}

	retrievedURL, err := repo.Get(context.TODO(), aliases[0])
	if err != nil {
		t.Fatalf("Get() returned an error: %v", err)
	}
	if retrievedURL != urls[0] {
		t.Errorf("expected URL %s, got %s", urls[0], retrievedURL)
	}

	urls = []string{
		"https://example1.com",
		"https://example2.com",
		"https://example3.com",
	}
	aliases, err = repo.AddBatch(context.TODO(), urls)
	if err != nil {
		t.Fatalf("AddBatch() returned an error: %v", err)
	}
	if len(aliases) != len(urls) {
		t.Errorf("expected %d aliases, got %d", len(urls), len(aliases))
	}

	for i, url := range urls {
		retrievedURL, err := repo.Get(context.TODO(), aliases[i])
		if err != nil {
			t.Fatalf("Get() returned an error: %v", err)
		}
		if retrievedURL != url {
			t.Errorf("expected URL %s, got %s", url, retrievedURL)
		}
	}

	urls = []string{
		"https://duplicate1.com",
		"https://duplicate1.com",
		"https://duplicate2.com",
	}
	aliases, err = repo.AddBatch(context.TODO(), urls)
	if err != nil {
		t.Fatalf("AddBatch() with duplicates returned an error: %v", err)
	}
	if len(aliases) != len(urls) {
		t.Errorf("expected %d aliases, got %d", len(urls), len(aliases))
	}

	if aliases[0] != aliases[1] {
		t.Errorf("expected same alias for duplicate URLs, got %s and %s", aliases[0], aliases[1])
	}

	existingURL := "https://existing.com"
	existingAlias, err := repo.Add(context.TODO(), existingURL)
	if err != nil {
		t.Fatalf("Add() returned an error: %v", err)
	}

	urls = []string{existingURL, "https://new.com"}
	aliases, err = repo.AddBatch(context.TODO(), urls)
	if err != nil {
		t.Fatalf("AddBatch() with existing URL returned an error: %v", err)
	}
	if len(aliases) != len(urls) {
		t.Errorf("expected %d aliases, got %d", len(urls), len(aliases))
	}

	if aliases[0] != existingAlias {
		t.Errorf("expected existing alias %s, got %s", existingAlias, aliases[0])
	}

	urls = make([]string, 100)
	for i := range urls {
		urls[i] = "https://example" + string(rune(i)) + ".com"
	}
	aliases, err = repo.AddBatch(context.TODO(), urls)
	if err != nil {
		t.Fatalf("AddBatch() with large batch returned an error: %v", err)
	}
	if len(aliases) != len(urls) {
		t.Errorf("expected %d aliases, got %d", len(urls), len(aliases))
	}
}
