package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// TokenCache represents a cached access token
type TokenCache struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scopes       []string  `json:"scopes"`
}

// IsExpired checks if the cached token is expired
func (tc *TokenCache) IsExpired() bool {
	return time.Now().After(tc.ExpiresAt.Add(-time.Minute * 5)) // 5-minute buffer
}

// TokenManager handles token caching operations
type TokenManager struct {
	cacheDir string
	filename string
}

// NewTokenManager creates a new token manager
func NewTokenManager(cacheDir string) *TokenManager {
	return &TokenManager{
		cacheDir: cacheDir,
		filename: "azure_tokens.json",
	}
}

// SaveToken saves the token to cache
func (tm *TokenManager) SaveToken(token *TokenCache) error {
	cacheDir, err := tm.getFullCacheDir()
	if err != nil {
		return fmt.Errorf("failed to get cache directory: %w", err)
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	cachePath := filepath.Join(cacheDir, tm.filename)
	if err := os.WriteFile(cachePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write token cache: %w", err)
	}

	return nil
}

// LoadToken loads the token from cache
func (tm *TokenManager) LoadToken() (*TokenCache, error) {
	cacheDir, err := tm.getFullCacheDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache directory: %w", err)
	}

	cachePath := filepath.Join(cacheDir, tm.filename)

	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrTokenNotFound
		}
		return nil, fmt.Errorf("failed to read token cache: %w", err)
	}

	var token TokenCache
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	return &token, nil
}

// ClearToken removes the cached token
func (tm *TokenManager) ClearToken() error {
	cacheDir, err := tm.getFullCacheDir()
	if err != nil {
		return fmt.Errorf("failed to get cache directory: %w", err)
	}

	cachePath := filepath.Join(cacheDir, tm.filename)
	if err := os.Remove(cachePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove token cache: %w", err)
	}
	return nil
}

// getFullCacheDir returns the full cache directory path
func (tm *TokenManager) getFullCacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, tm.cacheDir), nil
}
