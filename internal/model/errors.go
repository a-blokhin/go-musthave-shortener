package model

import "errors"

var ErrDuplicateURL = errors.New("URL already exists")

type DuplicateURLError struct {
	ExistingShortURL string
}

func (e *DuplicateURLError) Error() string {
	return ErrDuplicateURL.Error()
}

func (e *DuplicateURLError) Unwrap() error {
	return ErrDuplicateURL
}
