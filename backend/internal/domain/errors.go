package domain

import "errors"

// Errors shared by repositories and services. Wrap them with fmt.Errorf("%w: ...")
// to add detail; the HTTP layer maps them to status codes with errors.Is.
var (
	ErrNotFound           = errors.New("not found")
	ErrConflict           = errors.New("already exists")
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	// ErrUnavailable: the tool item cannot be reserved or rented for the period.
	ErrUnavailable = errors.New("tool not available")
	// ErrInvalidState: the operation is not allowed in the entity's current status.
	ErrInvalidState = errors.New("invalid state")
)
