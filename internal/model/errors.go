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

var ErrDeletedURL = errors.New("URL with has been deleted")

type DeletedURLError struct {
}

func (e *DeletedURLError) Error() string {
	return ErrDeletedURL.Error()
}

func (e *DeletedURLError) Unwrap() error {
	return ErrDeletedURL
}
