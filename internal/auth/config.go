package auth

import (
	"time"
)

// Config holds the configuration for Azure authentication
type Config struct {
	ClientID string
	TenantID string
	Scopes   []string
	CacheDir string
	Port     int
	Timeout  time.Duration
}

// DefaultConfig returns a default configuration for MS TODO access with personal accounts
func DefaultConfig() *Config {
	return &Config{
		TenantID: "consumers", // Use consumers for personal Microsoft accounts
		Port:     8080,        // Local server port
		Scopes: []string{
			"https://graph.microsoft.com/Tasks.ReadWrite",
			"https://graph.microsoft.com/User.Read",
		},
		CacheDir: ".azure-cli-cache",
		Timeout:  time.Minute * 5,
	}
}

// WithClientID sets the Azure AD application client ID
func (c *Config) WithClientID(clientID string) *Config {
	c.ClientID = clientID
	return c
}

// WithScopes sets custom scopes for authentication
func (c *Config) WithScopes(scopes []string) *Config {
	c.Scopes = scopes
	return c
}

// WithCacheDir sets custom cache directory for token storage
func (c *Config) WithCacheDir(dir string) *Config {
	c.CacheDir = dir
	return c
}

// WithPort sets custom port for local callback server
func (c *Config) WithPort(port int) *Config {
	c.Port = port
	return c
}
