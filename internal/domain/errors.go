package domain

import "errors"

// Sentinel errors for the domain layer.
var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrConflict     = errors.New("conflict")
	ErrInvalidInput = errors.New("invalid input")
)
