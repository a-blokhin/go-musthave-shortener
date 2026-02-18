package shorterrepository

import (
	"context"
	"errors"
	"fmt"
	"go-musthave-shortener/internal/model"
	"go-musthave-shortener/internal/repository"
	"math/rand/v2"
	"sync"
)

type Repo struct {
	shortToLink map[string]string
	linkToShort map[string]string
	userToURLs  map[string][]repository.UserURL
	deletedURLs map[string]bool
	urlOwners   map[string]string
	mutex       sync.RWMutex
	aliasLength int
}

func New() *Repo {
	return &Repo{
		shortToLink: map[string]string{},
		linkToShort: map[string]string{},
		userToURLs:  map[string][]repository.UserURL{},
		deletedURLs: map[string]bool{},
		urlOwners:   map[string]string{},
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
			r.urlOwners[alias] = userID

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
				r.urlOwners[alias] = userID

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

	if r.deletedURLs[alias] {
		return "", &model.DeletedURLError{}
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

	var result []repository.UserURL
	for _, userURL := range userURLs {
		if !r.deletedURLs[userURL.ShortURL] {
			result = append(result, userURL)
		}
	}

	return result, nil
}

func (r *Repo) BatchDelete(ctx context.Context, shortURLs []string, userID string) error {
	if len(shortURLs) == 0 {
		return nil
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, shortURL := range shortURLs {
		owner, exists := r.urlOwners[shortURL]
		if !exists {
			continue
		}
		if owner != userID {
			continue
		}
		r.deletedURLs[shortURL] = true
	}

	return nil
}

func generateAlias(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.IntN(len(charset))]
	}
	return string(result)
}
