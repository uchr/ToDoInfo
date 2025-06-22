package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity/cache"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
)

// ---------- public API ----------------------------------------------------

// AuthClient provides Azure authentication and Microsoft Graph access
type AuthClient struct {
	config      *Config
	credential  azcore.TokenCredential
	graphClient *msgraphsdk.GraphServiceClient

	recordPath string // ~/.<cacheDir>/auth_record.json (tiny non-secret file)
}

// NewAuthClient constructs a client and prepares the record-file path.
func NewAuthClient(cfg *Config) (*AuthClient, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot resolve HOME: %w", err)
	}
	cacheDir := filepath.Join(home, cfg.CacheDir)
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return nil, fmt.Errorf("create cache dir: %w", err)
	}
	return &AuthClient{
		config:     cfg,
		recordPath: filepath.Join(cacheDir, "auth_record.json"),
	}, nil
}

// Authenticate guarantees that a working credential & Graph client exist.
func (ac *AuthClient) Authenticate(ctx context.Context, logger *slog.Logger) error {
	if ac.credential != nil {
		// already authenticated in this process
		return nil
	}

	cred, err := ac.buildCredential(ctx, logger)
	if err != nil {
		return fmt.Errorf("build credential: %w", err)
	}
	// quick sanity check: can we get a token?
	if _, err = cred.GetToken(ctx, policy.TokenRequestOptions{Scopes: ac.config.Scopes}); err != nil {
		return fmt.Errorf("token test failed: %w", err)
	}

	ac.credential = cred
	return ac.initializeGraphClient()
}

// IsAuthenticated tells whether GetAccessToken would succeed without UI.
func (ac *AuthClient) IsAuthenticated(ctx context.Context, logger *slog.Logger) bool {
	if ac.credential == nil {
		if err := ac.Authenticate(ctx, logger); err != nil {
			return false
		}
	}
	_, err := ac.credential.GetToken(ctx, policy.TokenRequestOptions{Scopes: ac.config.Scopes})
	return err == nil
}

// GetGraphClient returns the Graph SDK client (Authenticate once first).
func (ac *AuthClient) GetGraphClient() *msgraphsdk.GraphServiceClient { return ac.graphClient }

// GetAccessToken exposes a raw bearer token for HTTP libraries outside MS Graph.
func (ac *AuthClient) GetAccessToken(ctx context.Context) (string, error) {
	if ac.credential == nil {
		return "", fmt.Errorf("not authenticated")
	}
	tok, err := ac.credential.GetToken(ctx, policy.TokenRequestOptions{Scopes: ac.config.Scopes})
	if err != nil {
		return "", fmt.Errorf("get token: %w", err)
	}
	return tok.Token, nil
}

// Logout forgets the browser session by deleting the AuthenticationRecord.
func (ac *AuthClient) Logout(_ context.Context, _ *slog.Logger) error {
	ac.credential = nil
	ac.graphClient = nil
	return os.Remove(ac.recordPath) // ignore “file not found”
}

// ---------- internal helpers ----------------------------------------------

// buildCredential reuses a stored AuthenticationRecord if present;
// otherwise it runs the interactive flow once and persists the record.
func (ac *AuthClient) buildCredential(ctx context.Context, logger *slog.Logger) (azcore.TokenCredential, error) {
	// 1. open the encrypted cross-platform token cache
	pcache, err := cache.New(nil)
	if err != nil {
		return nil, fmt.Errorf("open persistent cache: %w", err)
	}

	// 2. try to load a previously saved record
	var record azidentity.AuthenticationRecord
	if data, err := os.ReadFile(ac.recordPath); err == nil {
		if uErr := json.Unmarshal(data, &record); uErr != nil {
			logger.WarnContext(ctx, "auth record corrupt — starting fresh", slog.Any("error", uErr))
		}
	}

	// 3. build the credential (record may be zero-value = first run)
	cred, err := azidentity.NewInteractiveBrowserCredential(&azidentity.InteractiveBrowserCredentialOptions{
		TenantID:             ac.config.TenantID,
		ClientID:             ac.config.ClientID,
		RedirectURL:          fmt.Sprintf("http://localhost:%d", ac.config.Port),
		Cache:                pcache,
		AuthenticationRecord: record,
		// DisableAutomaticAuthentication avoids double prompts if silent auth fails
		DisableAutomaticAuthentication: true,
	})
	if err != nil {
		return nil, err
	}

	// 4. if the record was empty we must interact once and then persist
	if record.Username == "" {
		logger.DebugContext(ctx, "no stored auth record — launching browser")
		newRec, err := cred.Authenticate(
			ctx,
			&policy.TokenRequestOptions{Scopes: ac.config.Scopes},
		)
		if err != nil {
			return nil, err
		} else if data, err := json.Marshal(newRec); err == nil {
			_ = os.WriteFile(ac.recordPath, data, 0600)
		}
	}
	return cred, nil
}

// initializeGraphClient wires up the Microsoft Graph SDK.
func (ac *AuthClient) initializeGraphClient() error {
	gc, err := msgraphsdk.NewGraphServiceClientWithCredentials(ac.credential, ac.config.Scopes)
	if err != nil {
		return fmt.Errorf("new graph client: %w", err)
	}
	ac.graphClient = gc
	return nil
}

// ---------- tiny validation helper ----------------------------------------

func validateConfig(cfg *Config) error {
	switch {
	case cfg.ClientID == "":
		return fmt.Errorf("client ID required")
	case cfg.TenantID == "":
		return fmt.Errorf("tenant ID required")
	case cfg.Port <= 0:
		return fmt.Errorf("port must be > 0")
	case len(cfg.Scopes) == 0:
		return fmt.Errorf("at least one scope required")
	default:
		return nil
	}
}
