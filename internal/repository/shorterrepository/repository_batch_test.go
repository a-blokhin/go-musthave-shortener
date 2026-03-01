package shorterrepository

import (
	"testing"
)

func TestRepo_AddBatch(t *testing.T) {
	repo := New()
	urls := []string{"https://example1.com", "https://example2.com", "https://example3.com"}
	userID := "user123"
	ctx := t.Context()

	aliases, err := repo.AddBatch(ctx, urls, userID)
	if err != nil {
		t.Fatalf("AddBatch() returned an error: %v", err)
	}

	if len(aliases) != len(urls) {
		t.Errorf("expected %d aliases, got %d", len(urls), len(aliases))
	}

	for i, alias := range aliases {
		if len(alias) != repo.aliasLength {
			t.Errorf("expected alias length to be %d, got %d", repo.aliasLength, len(alias))
		}

		if !repo.hasAlias(alias) {
			t.Errorf("alias %s not stored in repository", alias)
		}

		if !repo.hasLink(urls[i]) {
			t.Errorf("URL %s not stored in repository", urls[i])
		}
	}

	userURLs, err := repo.GetByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("GetByUserID() returned an error: %v", err)
	}

	if len(userURLs) != len(urls) {
		t.Errorf("expected %d user URLs, got %d", len(urls), len(userURLs))
	}

	for i, userURL := range userURLs {
		if userURL.ShortURL != aliases[i] || userURL.OriginalURL != urls[i] {
			t.Errorf("user URL mismatch at index %d: got %s -> %s, expected %s -> %s",
				i, userURL.ShortURL, userURL.OriginalURL, aliases[i], urls[i])
		}
	}
}

func TestRepo_AddBatchWithEmptyURLs(t *testing.T) {
	repo := New()
	urls := []string{}
	userID := "user123"
	ctx := t.Context()

	aliases, err := repo.AddBatch(ctx, urls, userID)
	if err != nil {
		t.Fatalf("AddBatch() returned an error: %v", err)
	}

	if len(aliases) != 0 {
		t.Errorf("expected 0 aliases, got %d", len(aliases))
	}
}

func TestRepo_AddBatchWithDuplicateURLs(t *testing.T) {
	repo := New()
	urls := []string{"https://example1.com", "https://example1.com", "https://example2.com"}
	userID := "user123"
	ctx := t.Context()

	aliases, err := repo.AddBatch(ctx, urls, userID)
	if err != nil {
		t.Fatalf("AddBatch() returned an error: %v", err)
	}

	if len(aliases) != len(urls) {
		t.Errorf("expected %d aliases, got %d", len(urls), len(aliases))
	}

	if aliases[0] != aliases[1] {
		t.Errorf("expected same alias for duplicate URLs, got %s and %s", aliases[0], aliases[1])
	}

	if aliases[0] == aliases[2] {
		t.Errorf("expected different alias for different URLs, got %s for both", aliases[0])
	}
}

func TestRepo_AddBatchWithoutUser(t *testing.T) {
	repo := New()
	urls := []string{"https://example1.com", "https://example2.com"}
	ctx := t.Context()

	aliases, err := repo.AddBatch(ctx, urls, "")
	if err != nil {
		t.Fatalf("AddBatch() returned an error: %v", err)
	}

	if len(aliases) != len(urls) {
		t.Errorf("expected %d aliases, got %d", len(urls), len(aliases))
	}

	userURLs, err := repo.GetByUserID(ctx, "anyuser")
	if err != nil {
		t.Fatalf("GetByUserID() returned an error: %v", err)
	}

	if len(userURLs) != 0 {
		t.Errorf("expected 0 user URLs, got %d", len(userURLs))
	}
}
