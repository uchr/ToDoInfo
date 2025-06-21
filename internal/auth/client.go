package auth

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
)

// AuthClient provides Azure authentication and Microsoft Graph access
type AuthClient struct {
	config       *Config
	credential   azcore.TokenCredential
	graphClient  *msgraphsdk.GraphServiceClient
	tokenManager *TokenManager
}

// CachedTokenCredential wraps a cached token as an azcore.TokenCredential
type CachedTokenCredential struct {
	token *TokenCache
}

// GetToken returns the cached token
func (c *CachedTokenCredential) GetToken(ctx context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
	if c.token.IsExpired() {
		return azcore.AccessToken{}, fmt.Errorf("cached token is expired")
	}
	
	return azcore.AccessToken{
		Token:     c.token.AccessToken,
		ExpiresOn: c.token.ExpiresAt,
	}, nil
}

// NewAuthClient creates a new Azure authentication client
func NewAuthClient(config *Config) (*AuthClient, error) {
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}

	tokenManager := NewTokenManager(config.CacheDir)

	client := &AuthClient{
		config:       config,
		tokenManager: tokenManager,
	}

	return client, nil
}

// Authenticate performs browser-based Azure AD authentication
func (ac *AuthClient) Authenticate(ctx context.Context) error {
	// Try to load cached token first
	cachedToken, err := ac.tokenManager.LoadToken()
	if err == nil && !cachedToken.IsExpired() {
		log.Println("✅ Using cached authentication token")
		
		// Create credential from cached token
		ac.credential = &CachedTokenCredential{token: cachedToken}
		
		// Initialize graph client and test the token
		if err := ac.initializeGraphClient(); err != nil {
			log.Printf("Warning: cached token failed, falling back to browser auth: %v", err)
			// Clear invalid cached token and fall through to browser auth
			if clearErr := ac.tokenManager.ClearToken(); clearErr != nil {
				log.Printf("Warning: failed to clear invalid token cache: %v", clearErr)
			}
		} else {
			// Cached token works, we're done
			return nil
		}
	}

	log.Println("No valid cached token found, starting browser authentication...")
	fmt.Println("🌐 Opening browser for authentication...")

	// Create browser credential with proper redirect URI
	options := &azidentity.InteractiveBrowserCredentialOptions{
		ClientID:    ac.config.ClientID,
		TenantID:    ac.config.TenantID,
		RedirectURL: fmt.Sprintf("http://localhost:%d", ac.config.Port),
	}

	credential, err := azidentity.NewInteractiveBrowserCredential(options)
	if err != nil {
		return fmt.Errorf("failed to create browser credential: %w", err)
	}

	ac.credential = credential

	// Test the credential by getting a token
	tokenOptions := policy.TokenRequestOptions{
		Scopes: ac.config.Scopes,
	}

	token, err := credential.GetToken(ctx, tokenOptions)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	// Cache the token
	cachedToken = &TokenCache{
		AccessToken: token.Token,
		ExpiresAt:   token.ExpiresOn,
		Scopes:      ac.config.Scopes,
	}

	if err := ac.tokenManager.SaveToken(cachedToken); err != nil {
		log.Printf("Warning: failed to cache token: %v", err)
	}

	log.Println("✅ Authentication successful!")

	// Initialize graph client
	return ac.initializeGraphClient()
}

// GetGraphClient returns the configured Microsoft Graph client
func (ac *AuthClient) GetGraphClient() *msgraphsdk.GraphServiceClient {
	return ac.graphClient
}

// Logout clears the cached authentication token
func (ac *AuthClient) Logout() error {
	log.Println("Clearing cached authentication token...")
	return ac.tokenManager.ClearToken()
}

// IsAuthenticated checks if the client is currently authenticated
func (ac *AuthClient) IsAuthenticated() bool {
	cachedToken, err := ac.tokenManager.LoadToken()
	return err == nil && !cachedToken.IsExpired()
}

// GetAccessToken returns the current access token for HTTP requests
func (ac *AuthClient) GetAccessToken(ctx context.Context) (string, error) {
	if ac.credential == nil {
		return "", fmt.Errorf("no credential available")
	}
	
	tokenOptions := policy.TokenRequestOptions{
		Scopes: ac.config.Scopes,
	}
	
	token, err := ac.credential.GetToken(ctx, tokenOptions)
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}
	
	return token.Token, nil
}


// initializeGraphClient creates and configures the Microsoft Graph client
func (ac *AuthClient) initializeGraphClient() error {
	if ac.credential == nil {
		return fmt.Errorf("no credential available for graph client initialization")
	}

	graphClient, err := msgraphsdk.NewGraphServiceClientWithCredentials(
		ac.credential,
		ac.config.Scopes,
	)
	if err != nil {
		return fmt.Errorf("failed to create Graph service client: %w", err)
	}

	ac.graphClient = graphClient
	return nil
}

// validateConfig validates the authentication configuration
func validateConfig(config *Config) error {
	if config.ClientID == "" {
		return fmt.Errorf("client ID is required")
	}
	if config.TenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}
	if config.Port <= 0 {
		return fmt.Errorf("port must be positive")
	}
	if len(config.Scopes) == 0 {
		return fmt.Errorf("at least one scope is required")
	}
	return nil
}
