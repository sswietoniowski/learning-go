package common

import (
	"fmt"
	"net/http"
)

type Error struct {
	HttpErrorCode int

	PublicError string
	ErrorSlug   string

	InternalError error
	Details       []ErrorDetails
}

type ErrorDetails struct {
	EntityType string
	EntityID   string
	ErrorSlug  string
	Message    string
}

func (c Error) Error() string {
	s := fmt.Sprintf("%s, Slug: %s", c.PublicError, c.ErrorSlug)

	if c.InternalError != nil {
		s = fmt.Sprintf("%s, InternalError: %s", s, c.InternalError.Error())
	}
	if len(c.Details) > 0 {
		s = fmt.Sprintf("%s, DocumentData: %v", s, c.Details)
	}

	return s
}

func (c Error) WithDetails(details []ErrorDetails) Error {
	return Error{
		c.HttpErrorCode,
		c.PublicError,
		c.ErrorSlug,
		c.InternalError,
		append(c.Details, details...),
	}
}

func (c Error) WithInternalError(err error) Error {
	return Error{
		c.HttpErrorCode,
		c.PublicError,
		c.ErrorSlug,
		err,
		c.Details,
	}
}

func NewNotFoundError(slug, publicErrorFormat string, a ...any) Error {
	return Error{
		HttpErrorCode: http.StatusNotFound,
		PublicError:   fmt.Sprintf(publicErrorFormat, a...),
		ErrorSlug:     slug,
	}
}

func NewInvalidInputError(slug, publicErrorFormat string, a ...any) Error {
	return Error{
		HttpErrorCode: http.StatusBadRequest,
		PublicError:   fmt.Sprintf(publicErrorFormat, a...),
		ErrorSlug:     slug,
	}
}

func NewUnauthorizedError(slug, publicErrorFormat string, a ...any) Error {
	return Error{
		HttpErrorCode: http.StatusUnauthorized,
		PublicError:   fmt.Sprintf(publicErrorFormat, a...),
		ErrorSlug:     slug,
	}
}

func NewExpiredError(slug, publicErrorFormat string, a ...any) Error {
	return Error{
		HttpErrorCode: http.StatusGone,
		PublicError:   fmt.Sprintf(publicErrorFormat, a...),
		ErrorSlug:     slug,
	}
}
