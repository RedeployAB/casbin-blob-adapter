package blobadapter

import (
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
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

// checkAccountCredentialsArguments checks if the provided account and credentials are not empty.
func checkAccountCredentialsArguments(account string, cred azcore.TokenCredential) error {
	if len(account) == 0 {
		return ErrInvalidAccount
	}
	if cred == nil {
		return ErrInvalidCredential
	}
	return nil
}

// checkContainerBlobArguments checks if the provided container and blob are not empty.
func checkContainerBlobArguments(container, blob string) error {
	if len(container) == 0 {
		return ErrInvalidContainer
	}
	if len(blob) == 0 {
		return ErrInvalidBlob
	}
	return nil
}

// checkAccountKeyArguments checks if the provided account and key are not empty.
func checkAccountKeyArguments(account, key string) error {
	if len(account) == 0 {
		return ErrInvalidAccount
	}
	if len(key) == 0 {
		return ErrInvalidKey
	}
	return nil
}
