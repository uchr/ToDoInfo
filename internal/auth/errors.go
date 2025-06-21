package auth

import "errors"

var (
	// ErrTokenNotFound is returned when no cached token is found
	ErrTokenNotFound = errors.New("no cached token found")

	// ErrInvalidConfig is returned when the configuration is invalid
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrAuthenticationFailed is returned when authentication fails
	ErrAuthenticationFailed = errors.New("authentication failed")

	// ErrTokenExpired is returned when the token is expired and cannot be refreshed
	ErrTokenExpired = errors.New("token expired")
)
