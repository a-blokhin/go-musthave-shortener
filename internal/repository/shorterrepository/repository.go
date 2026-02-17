package shorterrepository

import (
	"context"
	"errors"
	"fmt"
	"go-musthave-shortener/internal/repository"
	"math/rand/v2"
	"sync"
)

type Repo struct {
	shortToLink map[string]string
	linkToShort map[string]string
	userToURLs  map[string][]repository.UserURL
	mutex       sync.RWMutex
	aliasLength int
}

func New() *Repo {
	return &Repo{
		shortToLink: map[string]string{},
		linkToShort: map[string]string{},
		userToURLs:  map[string][]repository.UserURL{},
		aliasLength: 8,
	}
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func (r *Repo) Add(ctx context.Context, url string, userID string) (string, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.hasLink(url) {
		return r.linkToShort[url], nil
	}

	const maxAttempts = 10

	for range maxAttempts {
		alias := generateAlias(r.aliasLength)
		if !r.hasAlias(alias) {
			r.shortToLink[alias] = url
			r.linkToShort[url] = alias
			
			// Add to user URLs if userID is provided
			if userID != "" {
				userURL := repository.UserURL{
					ShortURL:    alias,
					OriginalURL: url,
				}
				r.userToURLs[userID] = append(r.userToURLs[userID], userURL)
			}
			
			return alias, nil
		}
	}

	return "", errors.New("failed to generate unique alias")
}

func (r *Repo) AddBatch(ctx context.Context, urls []string, userID string) ([]string, error) {
	if len(urls) == 0 {
		return []string{}, nil
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	result := make([]string, len(urls))
	const maxAttempts = 10

	for i, url := range urls {

		if r.hasLink(url) {
			result[i] = r.linkToShort[url]
			continue
		}

		for range maxAttempts {
			alias := generateAlias(r.aliasLength)
			if !r.hasAlias(alias) {
				r.shortToLink[alias] = url
				r.linkToShort[url] = alias
				
				// Add to user URLs if userID is provided
				if userID != "" {
					userURL := repository.UserURL{
						ShortURL:    alias,
						OriginalURL: url,
					}
					r.userToURLs[userID] = append(r.userToURLs[userID], userURL)
				}
				
				result[i] = alias
				break
			}
		}

		if result[i] == "" {
			return nil, errors.New("failed to generate unique alias for one or more URLs")
		}
	}

	return result, nil
}

func (r *Repo) Get(ctx context.Context, alias string) (string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if !r.hasAlias(alias) {
		return "", fmt.Errorf("can't find requested alias %s", alias)
	}

	return r.shortToLink[alias], nil
}

func (r *Repo) hasLink(link string) bool {
	if _, ok := r.linkToShort[link]; ok {
		return ok
	}

	return false
}

func (r *Repo) hasAlias(alias string) bool {
	if _, ok := r.shortToLink[alias]; ok {
		return ok
	}

	return false
}

func (r *Repo) GetByUserID(ctx context.Context, userID string) ([]repository.UserURL, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	userURLs, exists := r.userToURLs[userID]
	if !exists {
		return []repository.UserURL{}, nil
	}

	// Return a copy to avoid external modification
	result := make([]repository.UserURL, len(userURLs))
	copy(result, userURLs)
	
	return result, nil
}

func generateAlias(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.IntN(len(charset))]
	}
	return string(result)
}
