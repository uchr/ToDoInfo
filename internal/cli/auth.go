package cli

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/uchr/ToDoInfo/internal/auth"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Microsoft account",
	Long:  "Authenticate with Microsoft Graph API to access your ToDo data",
	RunE:  runLogin,
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout and clear cached credentials",
	Long:  "Clear cached authentication tokens and logout from Microsoft account",
	RunE:  runLogout,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	Long:  "Check if you are currently authenticated with Microsoft Graph API",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(statusCmd)

	// Mark client-id as required for auth commands
	loginCmd.MarkPersistentFlagRequired("client-id")
	logoutCmd.MarkPersistentFlagRequired("client-id")
	statusCmd.MarkPersistentFlagRequired("client-id")
}

func createAuthClient(clientID string) (*auth.AuthClient, error) {
	config := auth.DefaultConfig().WithClientID(clientID)
	return auth.NewAuthClient(config)
}

func runLogin(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	clientID := viper.GetString("client-id")
	if clientID == "" {
		return fmt.Errorf("client-id is required (use --client-id flag or AZURE_CLIENT_ID env var)")
	}

	pterm.Info.Println("🔐 Logging in to Microsoft account...")

	authClient, err := createAuthClient(clientID)
	if err != nil {
		return fmt.Errorf("failed to create auth client: %w", err)
	}

	if authClient.IsAuthenticated(ctx, logger) {
		pterm.Success.Println("✅ Already authenticated!")
		return nil
	}

	spinner, _ := pterm.DefaultSpinner.Start("Authenticating...")
	if err := authClient.Authenticate(ctx, logger); err != nil {
		spinner.Fail("Authentication failed")
		return fmt.Errorf("authentication failed: %w", err)
	}
	spinner.Success("✅ Login successful!")

	return nil
}

func runLogout(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	clientID := viper.GetString("client-id")
	if clientID == "" {
		return fmt.Errorf("client-id is required (use --client-id flag or AZURE_CLIENT_ID env var)")
	}

	authClient, err := createAuthClient(clientID)
	if err != nil {
		return fmt.Errorf("failed to create auth client: %w", err)
	}

	spinner, _ := pterm.DefaultSpinner.Start("Logging out...")
	if err := authClient.Logout(ctx, logger); err != nil {
		spinner.Fail("Logout failed")
		return fmt.Errorf("logout failed: %w", err)
	}
	spinner.Success("✅ Logged out successfully!")

	return nil
}

func runStatus(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	clientID := viper.GetString("client-id")
	if clientID == "" {
		return fmt.Errorf("client-id is required (use --client-id flag or AZURE_CLIENT_ID env var)")
	}

	authClient, err := createAuthClient(clientID)
	if err != nil {
		return fmt.Errorf("failed to create auth client: %w", err)
	}

	if authClient.IsAuthenticated(ctx, logger) {
		pterm.Success.Println("✅ Authenticated")
	} else {
		pterm.Error.Println("❌ Not authenticated")
		pterm.Info.Println("Run 'todoinfo login' to authenticate")
	}

	return nil
}
