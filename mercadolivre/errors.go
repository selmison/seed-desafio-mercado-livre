package mercadolivre

import "errors"

var (
	ErrAlreadyExists    = errors.New("already exists")
	ErrIsNotValid       = errors.New("is not valid")
	ErrNotFound         = errors.New("not found")
	ErrShouldBeFuture   = errors.New("should be in the future")
	ErrValidationFailed = errors.New("validation failed")
)
