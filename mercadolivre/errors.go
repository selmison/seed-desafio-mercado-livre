package mercadolivre

import (
	"errors"
	"net/http"
)

var (
	ErrAlreadyExists    = errors.New("already exists")
	ErrInternalServer   = errors.New(http.StatusText(http.StatusInternalServerError))
	ErrIsNotValid       = errors.New("is not valid")
	ErrNotFound         = errors.New("not found")
	ErrShouldBeFuture   = errors.New("should be in the future")
	ErrShouldBeUnique   = errors.New("should be unique")
	ErrLoginFailed      = errors.New("login failed")
	ErrValidationFailed = errors.New("validation failed")
)
