package mercadolivre

import (
	"errors"
	"net/http"
)

var (
	ErrAlreadyExists    = errors.New("already exists")
	ErrAuthFailed       = errors.New("authentication failed")
	ErrInternalServer   = errors.New(http.StatusText(http.StatusInternalServerError))
	ErrIsNotValid       = errors.New("is not valid")
	ErrMissingToken     = errors.New("missing token")
	ErrNotFound         = errors.New("not found")
	ErrShouldBeFuture   = errors.New("should be in the future")
	ErrShouldBeUnique   = errors.New("should be unique")
	ErrValidationFailed = errors.New("validation failed")
)
