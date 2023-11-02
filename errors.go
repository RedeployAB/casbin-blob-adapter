package blobadapter

import (
	"errors"
)

var (
	// ErrInvalidAccount is returned when the account is invalid.
	ErrInvalidAccount = errors.New("invalid account")
	// ErrInvalidCredential is returned when the credentials are invald.
	ErrInvalidCredential = errors.New("invalid credentials")
	// ErrInvalidConnectionString is returned when the connection string is invalid.
	ErrInvalidConnectionString = errors.New("invalid connection string")
	// ErrInvalidKey is returned when the key is invalid.
	ErrInvalidKey = errors.New("invalid key")
	// ErrInvalidContainer is returned when the container is invalid.
	ErrInvalidContainer = errors.New("invalid container")
	// ErrInvalidBlob is returned when the blob is invalid.
	ErrInvalidBlob = errors.New("invalid blob")
)
