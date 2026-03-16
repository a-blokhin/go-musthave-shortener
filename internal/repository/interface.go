// Package repository provides interfaces for URL storage and retrieval.
// It defines the contract for different storage implementations (in-memory, file, PostgreSQL).
package repository

import "context"

// UserURL represents a shortened URL with its original URL.
// Used for returning user's URLs in the API response.
type UserURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// Stats represents statistics about the URL shortener service.
type Stats struct {
	URLs  int `json:"urls"`  // Number of shortened URLs
	Users int `json:"users"` // Number of users
}

// LinkRepository defines the interface for URL storage operations.
// Implementations can use in-memory storage, file storage, or database storage.
type LinkRepository interface {
	// Add stores a new URL and returns its short alias.
	// If the URL already exists for the user, it returns the existing alias.
	//
	// Parameters:
	//   - ctx: context for the operation
	//   - url: the original URL to shorten
	//   - userID: the user ID who owns the URL
	//
	// Returns the short alias for the URL or an error if storage fails.
	Add(ctx context.Context, url string, userID string) (string, error)

	// AddBatch stores multiple URLs and returns their short aliases.
	// All URLs are stored atomically in a single operation.
	//
	// Parameters:
	//   - ctx: context for the operation
	//   - urls: slice of original URLs to shorten
	//   - userID: the user ID who owns the URLs
	//
	// Returns a slice of short aliases in the same order as input URLs or an error if storage fails.
	AddBatch(ctx context.Context, urls []string, userID string) ([]string, error)

	// Get retrieves the original URL for a given short alias.
	//
	// Parameters:
	//   - ctx: context for the operation
	//   - alias: the short alias to look up
	//
	// Returns the original URL or an error if not found or retrieval fails.
	Get(ctx context.Context, alias string) (string, error)

	// GetByUserID retrieves all URLs for a specific user.
	//
	// Parameters:
	//   - ctx: context for the operation
	//   - userID: the user ID to look up URLs for
	//
	// Returns a slice of UserURL containing all URLs for the user or an error if retrieval fails.
	GetByUserID(ctx context.Context, userID string) ([]UserURL, error)

	// BatchDelete marks multiple URLs as deleted for a specific user.
	// This is a soft delete operation; URLs are marked as deleted but not removed from storage.
	//
	// Parameters:
	//   - ctx: context for the operation
	//   - shortURLs: slice of short aliases to delete
	//   - userID: the user ID who owns the URLs
	//
	// Returns an error if the deletion fails.
	BatchDelete(ctx context.Context, shortURLs []string, userID string) error

	// GetStats returns statistics about the URL shortener service.
	//
	// Parameters:
	//   - ctx: context for the operation
	//
	// Returns a Stats struct containing the number of URLs and users, or an error if retrieval fails.
	GetStats(ctx context.Context) (Stats, error)
}
