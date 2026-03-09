//go:build example

package usecase_test

import (
	"context"
	"fmt"

	"go-musthave-shortener/internal/repository"
)

// ExampleLinkRepository_Add demonstrates how to add a single URL to the repository
func ExampleLinkRepository_Add() {
	// Create a mock repository
	repo := &mockLinkRepository{
		storage: make(map[string]string),
	}

	// Add a URL to the repository
	ctx := context.Background()
	shortURL, err := repo.Add(ctx, "https://example.com", "user123")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Short URL created: %s\n", shortURL)
	// Output: Short URL created: abc123
}

// ExampleLinkRepository_AddBatch demonstrates how to add multiple URLs to the repository
func ExampleLinkRepository_AddBatch() {
	// Create a mock repository
	repo := &mockLinkRepository{
		storage: make(map[string]string),
	}

	// Add multiple URLs to the repository
	ctx := context.Background()
	urls := []string{
		"https://example.com/1",
		"https://example.com/2",
		"https://example.com/3",
	}
	shortURLs, err := repo.AddBatch(ctx, urls, "user123")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for i, shortURL := range shortURLs {
		fmt.Printf("URL %d: %s -> %s\n", i+1, urls[i], shortURL)
	}
	// Output:
	// URL 1: https://example.com/1 -> 1
	// URL 2: https://example.com/2 -> 2
	// URL 3: https://example.com/3 -> 3
}

// ExampleLinkRepository_Get demonstrates how to retrieve an original URL from a short alias
func ExampleLinkRepository_Get() {
	// Create a mock repository with a stored URL
	repo := &mockLinkRepository{
		storage: map[string]string{
			"abc123": "https://example.com",
		},
	}

	// Retrieve the original URL
	ctx := context.Background()
	originalURL, err := repo.Get(ctx, "abc123")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Original URL: %s\n", originalURL)
	// Output: Original URL: https://example.com
}

// ExampleLinkRepository_GetByUserID demonstrates how to retrieve all URLs for a user
func ExampleLinkRepository_GetByUserID() {
	// Create a mock repository with user URLs
	repo := &mockLinkRepository{
		userStorage: map[string][]repository.UserURL{
			"user123": {
				{
					ShortURL:    "abc123",
					OriginalURL: "https://example.com/1",
				},
				{
					ShortURL:    "xyz789",
					OriginalURL: "https://example.com/2",
				},
			},
		},
	}

	// Retrieve all URLs for the user
	ctx := context.Background()
	userURLs, err := repo.GetByUserID(ctx, "user123")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, url := range userURLs {
		fmt.Printf("%s -> %s\n", url.ShortURL, url.OriginalURL)
	}
	// Output:
	// abc123 -> https://example.com/1
	// xyz789 -> https://example.com/2
}

// ExampleLinkRepository_BatchDelete demonstrates how to delete multiple URLs for a user
func ExampleLinkRepository_BatchDelete() {
	// Create a mock repository
	repo := &mockLinkRepository{
		storage: make(map[string]string),
	}

	// Add some URLs first
	ctx := context.Background()
	repo.Add(ctx, "https://example.com/1", "user123")
	repo.Add(ctx, "https://example.com/2", "user123")

	// Delete the URLs
	shortURLs := []string{"abc123", "xyz789"}
	err := repo.BatchDelete(ctx, shortURLs, "user123")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("URLs deleted successfully")
	// Output: URLs deleted successfully
}

// Mock implementation of LinkRepository for testing

type mockLinkRepository struct {
	storage     map[string]string
	userStorage map[string][]repository.UserURL
}

func (m *mockLinkRepository) Add(ctx context.Context, url string, userID string) (string, error) {
	shortURL := "abc123"
	if m.storage == nil {
		m.storage = make(map[string]string)
	}
	m.storage[shortURL] = url
	return shortURL, nil
}

func (m *mockLinkRepository) AddBatch(ctx context.Context, urls []string, userID string) ([]string, error) {
	shortURLs := make([]string, len(urls))
	for i, url := range urls {
		shortURL := fmt.Sprintf("%d", i+1)
		if m.storage == nil {
			m.storage = make(map[string]string)
		}
		m.storage[shortURL] = url
		shortURLs[i] = shortURL
	}
	return shortURLs, nil
}

func (m *mockLinkRepository) Get(ctx context.Context, alias string) (string, error) {
	if m.storage == nil {
		return "", fmt.Errorf("URL not found")
	}
	originalURL, ok := m.storage[alias]
	if !ok {
		return "", fmt.Errorf("URL not found")
	}
	return originalURL, nil
}

func (m *mockLinkRepository) GetByUserID(ctx context.Context, userID string) ([]repository.UserURL, error) {
	if m.userStorage == nil {
		return []repository.UserURL{}, nil
	}
	return m.userStorage[userID], nil
}

func (m *mockLinkRepository) BatchDelete(ctx context.Context, shortURLs []string, userID string) error {
	// In a real implementation, this would mark URLs as deleted
	return nil
}
