package model

import (
	"time"

	"github.com/google/uuid"
)

type URL struct {
	ID          uuid.UUID `json:"id"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
	UserID      string    `json:"user_id,omitempty"`
	DeletedFlag bool      `json:"is_deleted,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func NewURL(shortURL, originalURL string) *URL {
	return &URL{
		ID:          uuid.New(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		CreatedAt:   time.Now(),
	}
}

func NewURLWithUser(shortURL, originalURL, userID string) *URL {
	return &URL{
		ID:          uuid.New(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		UserID:      userID,
		CreatedAt:   time.Now(),
	}
}