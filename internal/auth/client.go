package auth

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	extcache "github.com/AzureAD/microsoft-authentication-extensions-for-go/cache"
	"github.com/AzureAD/microsoft-authentication-extensions-for-go/cache/accessor/file"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
)

// ---------- public API ----------------------------------------------------

// AuthClient provides Azure authentication for Microsoft Graph access.
//
// The token cache is an unencrypted JSON file (0600) on disk. This is a
// deliberate choice: the previous implementation used azidentity's keyring-
// backed encrypted cache, whose decryption key lives in the Linux kernel
// keyring and is lost on container shutdown — so refresh tokens could be
// persisted but never decrypted after a redeploy. Relying on file-system
// permissions + Docker volume isolation is the same approach MSAL's own
// headless example uses.
type AuthClient struct {
	config *Config

	client  *public.Client
	account *public.Account

	cacheDir  string
	cachePath string
}

// NewAuthClient constructs a client and prepares the cache directory.
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
		config:    cfg,
		cacheDir:  cacheDir,
		cachePath: filepath.Join(cacheDir, "msal_cache.json"),
	}, nil
}

// Authenticate guarantees that a working account exists, prompting the user
// only if no cached account is available or silent refresh fails.
func (ac *AuthClient) Authenticate(ctx context.Context, logger *slog.Logger) error {
	if ac.account != nil {
		return nil
	}
	if err := ac.buildClient(); err != nil {
		return fmt.Errorf("build client: %w", err)
	}

	// First try silent acquisition from any cached account.
	accounts, err := ac.client.Accounts(ctx)
	if err != nil {
		return fmt.Errorf("read accounts: %w", err)
	}
	if len(accounts) > 0 {
		if _, err := ac.client.AcquireTokenSilent(ctx, ac.config.Scopes, public.WithSilentAccount(accounts[0])); err == nil {
			ac.account = &accounts[0]
			return nil
		} else {
			logger.DebugContext(ctx, "silent auth failed, falling back to interactive", slog.Any("error", err))
		}
	}

	var result public.AuthResult
	if ac.config.Headless {
		dc, err := ac.client.AcquireTokenByDeviceCode(ctx, ac.config.Scopes)
		if err != nil {
			return fmt.Errorf("start device code: %w", err)
		}
		if ac.config.DeviceCodeCallback != nil {
			ac.config.DeviceCodeCallback(ctx, dc.Result.Message)
		} else {
			logger.InfoContext(ctx, "device code", slog.String("message", dc.Result.Message))
		}
		result, err = dc.AuthenticationResult(ctx)
		if err != nil {
			return fmt.Errorf("device code result: %w", err)
		}
	} else {
		result, err = ac.client.AcquireTokenInteractive(ctx, ac.config.Scopes,
			public.WithRedirectURI(fmt.Sprintf("http://localhost:%d", ac.config.Port)),
		)
		if err != nil {
			return fmt.Errorf("interactive auth: %w", err)
		}
	}
	ac.account = &result.Account
	return nil
}

// IsAuthenticated tells whether GetAccessToken would succeed without UI.
func (ac *AuthClient) IsAuthenticated(ctx context.Context, logger *slog.Logger) bool {
	if err := ac.buildClient(); err != nil {
		logger.DebugContext(ctx, "build client failed", slog.Any("error", err))
		return false
	}
	if ac.account == nil {
		accounts, err := ac.client.Accounts(ctx)
		if err != nil || len(accounts) == 0 {
			return false
		}
		ac.account = &accounts[0]
	}
	_, err := ac.client.AcquireTokenSilent(ctx, ac.config.Scopes, public.WithSilentAccount(*ac.account))
	return err == nil
}

// HasCredential reports whether Authenticate has completed successfully.
// Unlike IsAuthenticated, this never calls the token endpoint.
func (ac *AuthClient) HasCredential() bool { return ac.account != nil }

// TryNonInteractiveAuth attempts to restore an account from the persisted token
// cache and verify it can mint a token. Returns true on success. Failure paths
// are logged (at info level) so operators can diagnose startup auth problems.
func (ac *AuthClient) TryNonInteractiveAuth(ctx context.Context, logger *slog.Logger) bool {
	if ac.account != nil {
		return true
	}
	if err := ac.buildClient(); err != nil {
		logger.InfoContext(ctx, "non-interactive auth: build client failed", slog.Any("error", err))
		return false
	}
	accounts, err := ac.client.Accounts(ctx)
	if err != nil {
		logger.InfoContext(ctx, "non-interactive auth: read accounts failed", slog.Any("error", err))
		return false
	}
	if len(accounts) == 0 {
		logger.InfoContext(ctx, "non-interactive auth: no cached accounts")
		return false
	}
	if _, err := ac.client.AcquireTokenSilent(ctx, ac.config.Scopes, public.WithSilentAccount(accounts[0])); err != nil {
		logger.InfoContext(ctx, "non-interactive auth: silent acquire failed", slog.Any("error", err))
		return false
	}
	ac.account = &accounts[0]
	return true
}

// GetAccessToken returns a fresh access token, refreshing silently if the
// cached one has expired. Returns an error if no account is attached.
func (ac *AuthClient) GetAccessToken(ctx context.Context) (string, error) {
	if ac.client == nil || ac.account == nil {
		return "", fmt.Errorf("not authenticated")
	}
	result, err := ac.client.AcquireTokenSilent(ctx, ac.config.Scopes, public.WithSilentAccount(*ac.account))
	if err != nil {
		return "", fmt.Errorf("silent acquire: %w", err)
	}
	return result.AccessToken, nil
}

// Logout signs the account out of MSAL's cache and removes the cache file.
func (ac *AuthClient) Logout(ctx context.Context, logger *slog.Logger) error {
	if err := ac.buildClient(); err != nil {
		// Can't talk to MSAL; just nuke files.
		ac.account = nil
		_ = os.Remove(ac.cachePath)
		_ = os.Remove(ac.cachePath + ".lockfile")
		return nil
	}
	if ac.account != nil {
		if err := ac.client.RemoveAccount(ctx, *ac.account); err != nil {
			logger.WarnContext(ctx, "remove account failed", slog.Any("error", err))
		}
	} else if accounts, err := ac.client.Accounts(ctx); err == nil {
		for _, a := range accounts {
			_ = ac.client.RemoveAccount(ctx, a)
		}
	}
	ac.account = nil
	_ = os.Remove(ac.cachePath)
	_ = os.Remove(ac.cachePath + ".lockfile")
	return nil
}

// ---------- internal helpers ----------------------------------------------

// buildClient lazily constructs the MSAL public client with a file-backed
// unencrypted token cache.
func (ac *AuthClient) buildClient() error {
	if ac.client != nil {
		return nil
	}
	accessor, err := file.New(ac.cachePath)
	if err != nil {
		return fmt.Errorf("open cache accessor: %w", err)
	}
	cache, err := extcache.New(accessor, ac.cachePath)
	if err != nil {
		return fmt.Errorf("create cache: %w", err)
	}
	authority := fmt.Sprintf("https://login.microsoftonline.com/%s", ac.config.TenantID)
	client, err := public.New(ac.config.ClientID,
		public.WithAuthority(authority),
		public.WithCache(cache),
	)
	if err != nil {
		return fmt.Errorf("new msal public client: %w", err)
	}
	ac.client = &client
	return nil
}

// ---------- tiny validation helper ----------------------------------------

func validateConfig(cfg *Config) error {
	switch {
	case cfg.ClientID == "":
		return fmt.Errorf("client ID required")
	case cfg.TenantID == "":
		return fmt.Errorf("tenant ID required")
	case !cfg.Headless && cfg.Port <= 0:
		return fmt.Errorf("port must be > 0")
	case len(cfg.Scopes) == 0:
		return fmt.Errorf("at least one scope required")
	default:
		return nil
	}
}
