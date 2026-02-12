package shorterrepository

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"sync"
)

type Repo struct {
	shortToLink map[string]string
	linkToShort map[string]string
	rwMutex     sync.RWMutex
	aliasLength int
}

func New() *Repo {
	return &Repo{
		shortToLink: map[string]string{},
		linkToShort: map[string]string{},
		aliasLength: 8,
	}
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func (r *Repo) Add(url string) (string, error) {
	r.rwMutex.Lock()
	defer r.rwMutex.Unlock()

	if r.hasLink(url) {
		return r.linkToShort[url], nil
	}

	const maxAttempts = 10

	for range maxAttempts {
		alias := generateAlias(r.aliasLength)
		if !r.hasAlias(alias) {
			r.shortToLink[alias] = url
			r.linkToShort[url] = alias
			return alias, nil
		}
	}

	return "", errors.New("failed to generate unique alias")
}

func (r *Repo) AddBatch(urls []string) ([]string, error) {
	if len(urls) == 0 {
		return []string{}, nil
	}

	r.rwMutex.Lock()
	defer r.rwMutex.Unlock()

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

func (r *Repo) Get(alias string) (string, error) {
	r.rwMutex.RLock()
	defer r.rwMutex.RUnlock()

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

func generateAlias(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.IntN(len(charset))]
	}
	return string(result)
}
