// Package model provides data structures for the URL shortener service.
package model

import (
	"time"

	"github.com/google/uuid"
)

// URL represents a shortened URL with its metadata.
type URL struct {
	ID          uuid.UUID `json:"id"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
	UserID      string    `json:"user_id,omitempty"`
	DeletedFlag bool      `json:"is_deleted,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// NewURL creates a new URL instance without a user ID.
// This is used for anonymous URL creation.
//
// Parameters:
//   - shortURL: the short alias for the URL
//   - originalURL: the original URL to shorten
//
// Returns a new URL instance with generated ID and creation timestamp.
func NewURL(shortURL, originalURL string) *URL {
	return &URL{
		ID:          uuid.New(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		CreatedAt:   time.Now(),
	}
}

// NewURLWithUser creates a new URL instance with a user ID.
// This is used for authenticated URL creation.
//
// Parameters:
//   - shortURL: the short alias for the URL
//   - originalURL: the original URL to shorten
//   - userID: the user ID who owns the URL
//
// Returns a new URL instance with generated ID, user ID, and creation timestamp.
func NewURLWithUser(shortURL, originalURL, userID string) *URL {
	return &URL{
		ID:          uuid.New(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		UserID:      userID,
		CreatedAt:   time.Now(),
	}
}
