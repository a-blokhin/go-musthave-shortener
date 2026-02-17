package shorterrepository

import (
	"context"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	repo := New()

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
}

func TestRepo_Add(t *testing.T) {
	repo := New()
	url := "https://example.com"

	alias, err := repo.Add(context.TODO(), url, "")
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

	sameAlias, err := repo.Add(context.TODO(), url, "")
	if err != nil {
		t.Fatalf("Add() for same URL returned an error: %v", err)
	}

	if sameAlias != alias {
		t.Errorf("expected same alias for same URL, got different: %s vs %s", alias, sameAlias)
	}
}

func TestRepo_Get(t *testing.T) {
	repo := New()
	url := "https://example.com"

	alias, err := repo.Add(context.TODO(), url, "")
	if err != nil {
		t.Fatalf("Add() returned an error: %v", err)
	}

	retrievedURL, err := repo.Get(context.TODO(), alias)
	if err != nil {
		t.Fatalf("Get() returned an error: %v", err)
	}

	if retrievedURL != url {
		t.Errorf("expected URL %s, got %s", url, retrievedURL)
	}

	_, err = repo.Get(context.TODO(), "nonexistent-alias")
	if err == nil {
		t.Error("expected error for non-existent alias, got nil")
	}
}

func TestRepo_hasLink(t *testing.T) {
	repo := New()
	url := "https://example.com"

	if repo.hasLink(url) {
		t.Error("hasLink() should return false for non-existent URL")
	}

	_, err := repo.Add(context.TODO(), url, "")
	if err != nil {
		t.Fatalf("Add() returned an error: %v", err)
	}

	if !repo.hasLink(url) {
		t.Error("hasLink() should return true for existing URL")
	}
}

func TestRepo_hasAlias(t *testing.T) {
	repo := New()
	url := "https://example.com"

	alias, err := repo.Add(context.TODO(), url, "")
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
