package shorterfilerepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-musthave-shortener/internal/model"
	"go-musthave-shortener/internal/repository"
	"maps"
	"math/rand/v2"
	"os"
	"strconv"
	"sync"
)

type URLData struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id,omitempty"`
	DeletedFlag bool   `json:"is_deleted,omitempty"`
}

type FileRepo struct {
	shortToLink map[string]string
	linkToShort map[string]string
	userToURLs  map[string][]repository.UserURL
	deletedURLs map[string]bool
	urlOwners   map[string]string
	mutex       sync.RWMutex
	aliasLength int
	filePath    string
}

func New(filePath string) *FileRepo {
	repo := &FileRepo{
		shortToLink: map[string]string{},
		linkToShort: map[string]string{},
		userToURLs:  map[string][]repository.UserURL{},
		deletedURLs: map[string]bool{},
		urlOwners:   map[string]string{},
		aliasLength: 8,
		filePath:    filePath,
	}

	repo.loadFromFile()

	return repo
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func (r *FileRepo) Add(ctx context.Context, url string, userID string) (string, error) {
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

			if err := r.saveToFile(); err != nil {

				delete(r.shortToLink, alias)
				delete(r.linkToShort, url)
				delete(r.urlOwners, alias)
				if userID != "" {
					r.removeUserURL(userID, alias)
				}
				return "", fmt.Errorf("failed to save to file: %w", err)
			}

			return alias, nil
		}
	}

	return "", errors.New("failed to generate unique alias")
}

func (r *FileRepo) AddBatch(ctx context.Context, urls []string, userID string) ([]string, error) {
	if len(urls) == 0 {
		return []string{}, nil
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	originalShortToLink := make(map[string]string, len(r.shortToLink))
	originalLinkToShort := make(map[string]string, len(r.linkToShort))

	maps.Copy(originalShortToLink, r.shortToLink)
	maps.Copy(originalLinkToShort, r.linkToShort)

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
			r.shortToLink = originalShortToLink
			r.linkToShort = originalLinkToShort
			r.userToURLs = map[string][]repository.UserURL{}
			return nil, errors.New("failed to generate unique alias for one or more URLs")
		}
	}

	if err := r.saveToFile(); err != nil {
		r.shortToLink = originalShortToLink
		r.linkToShort = originalLinkToShort
		r.userToURLs = map[string][]repository.UserURL{}
		return nil, fmt.Errorf("failed to save to file: %w", err)
	}

	return result, nil
}

func (r *FileRepo) Get(ctx context.Context, alias string) (string, error) {
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

func (r *FileRepo) hasLink(link string) bool {
	if _, ok := r.linkToShort[link]; ok {
		return ok
	}
	return false
}

func (r *FileRepo) hasAlias(alias string) bool {
	if _, ok := r.shortToLink[alias]; ok {
		return ok
	}
	return false
}

func (r *FileRepo) loadFromFile() error {

	if _, err := os.Stat(r.filePath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if len(data) == 0 {
		return nil
	}

	var urlDataList []URLData
	if err := json.Unmarshal(data, &urlDataList); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	for _, item := range urlDataList {
		r.shortToLink[item.ShortURL] = item.OriginalURL
		r.linkToShort[item.OriginalURL] = item.ShortURL
		r.urlOwners[item.ShortURL] = item.UserID
		if item.DeletedFlag {
			r.deletedURLs[item.ShortURL] = true
		}
		if item.UserID != "" {
			userURL := repository.UserURL{
				ShortURL:    item.ShortURL,
				OriginalURL: item.OriginalURL,
			}
			r.userToURLs[item.UserID] = append(r.userToURLs[item.UserID], userURL)
		}
	}

	return nil
}

func (r *FileRepo) saveToFile() error {
	var urlDataList []URLData
	uuid := 1

	for shortURL, originalURL := range r.shortToLink {
		userID := r.urlOwners[shortURL]
		isDeleted := r.deletedURLs[shortURL]

		urlDataList = append(urlDataList, URLData{
			UUID:        strconv.Itoa(uuid),
			ShortURL:    shortURL,
			OriginalURL: originalURL,
			UserID:      userID,
			DeletedFlag: isDeleted,
		})
		uuid++
	}

	data, err := json.MarshalIndent(urlDataList, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(r.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (r *FileRepo) GetByUserID(ctx context.Context, userID string) ([]repository.UserURL, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	userURLs, ok := r.userToURLs[userID]
	if !ok {
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

func (r *FileRepo) BatchDelete(ctx context.Context, shortURLs []string, userID string) error {
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

	if err := r.saveToFile(); err != nil {
		for _, shortURL := range shortURLs {
			delete(r.deletedURLs, shortURL)
		}
		return fmt.Errorf("failed to save deletion to file: %w", err)
	}

	return nil
}

func (r *FileRepo) GetStats(ctx context.Context) (repository.Stats, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	users := make(map[string]bool)
	for _, userID := range r.urlOwners {
		if userID != "" {
			users[userID] = true
		}
	}

	return repository.Stats{
		URLs:  len(r.shortToLink),
		Users: len(users),
	}, nil
}

func (r *FileRepo) removeUserURL(userID, shortURL string) {
	if userURLs, exists := r.userToURLs[userID]; exists {
		for i, userURL := range userURLs {
			if userURL.ShortURL == shortURL {
				r.userToURLs[userID] = append(userURLs[:i], userURLs[i+1:]...)
				break
			}
		}
	}
}

func generateAlias(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.IntN(len(charset))]
	}
	return string(result)
}
