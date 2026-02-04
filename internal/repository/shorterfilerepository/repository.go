package shorterfilerepository

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"
	"sync"
)

type URLData struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileRepo struct {
	shortToLink map[string]string
	linkToShort map[string]string
	rwMutex     sync.RWMutex
	aliasLength int
	filePath    string
}

func New(filePath string) *FileRepo {
	repo := &FileRepo{
		shortToLink: map[string]string{},
		linkToShort: map[string]string{},
		aliasLength: 8,
		filePath:    filePath,
	}

	repo.loadFromFile()

	return repo
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func (r *FileRepo) Add(url string) (string, error) {
	r.rwMutex.Lock()
	defer r.rwMutex.Unlock()

	if r.hasLink(url) {
		return r.linkToShort[url], nil
	}

	const maxAttempts = 100

	for range maxAttempts {
		alias := generateAlias(r.aliasLength)
		if !r.hasAlias(alias) {
			r.shortToLink[alias] = url
			r.linkToShort[url] = alias

			if err := r.saveToFile(); err != nil {

				delete(r.shortToLink, alias)
				delete(r.linkToShort, url)
				return "", fmt.Errorf("failed to save to file: %w", err)
			}

			return alias, nil
		}
	}

	return "", errors.New("failed to generate unique alias")
}

func (r *FileRepo) Get(alias string) (string, error) {
	r.rwMutex.RLock()
	defer r.rwMutex.RUnlock()

	if !r.hasAlias(alias) {
		return "", fmt.Errorf("can't find requested alias %s", alias)
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
	}

	return nil
}

func (r *FileRepo) saveToFile() error {
	var urlDataList []URLData
	uuid := 1

	for shortURL, originalURL := range r.shortToLink {
		urlDataList = append(urlDataList, URLData{
			UUID:        strconv.Itoa(uuid),
			ShortURL:    shortURL,
			OriginalURL: originalURL,
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

func generateAlias(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.IntN(len(charset))]
	}
	return string(result)
}
