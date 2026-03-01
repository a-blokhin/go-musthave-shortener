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

type DeletedURLError struct{}

func (DeletedURLError) Error() string { return "url has been deleted" }

var ErrDeletedURL error = DeletedURLError{}
