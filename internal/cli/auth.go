package cli

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/uchr/ToDoInfo/internal/auth"
)

var (
	successStyleAuth = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B"))
	errorStyleAuth = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555"))
	infoStyleAuth = lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4"))
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

	fmt.Println(infoStyleAuth.Render("🔐 Logging in to Microsoft account..."))

	authClient, err := createAuthClient(clientID)
	if err != nil {
		return fmt.Errorf("failed to create auth client: %w", err)
	}

	if authClient.IsAuthenticated(ctx, logger) {
		fmt.Println(successStyleAuth.Render("✓ Already authenticated!"))
		return nil
	}

	fmt.Print(infoStyleAuth.Render("Authenticating..."))
	if err := authClient.Authenticate(ctx, logger); err != nil {
		fmt.Println(" " + errorStyleAuth.Render("✗ Authentication failed"))
		return fmt.Errorf("authentication failed: %w", err)
	}
	fmt.Println(" " + successStyleAuth.Render("✓ Login successful!"))

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

	fmt.Print(infoStyleAuth.Render("Logging out..."))
	if err := authClient.Logout(ctx, logger); err != nil {
		fmt.Println(" " + errorStyleAuth.Render("✗ Logout failed"))
		return fmt.Errorf("logout failed: %w", err)
	}
	fmt.Println(" " + successStyleAuth.Render("✓ Logged out successfully!"))

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
		fmt.Println(successStyleAuth.Render("✓ Authenticated"))
	} else {
		fmt.Println(errorStyleAuth.Render("✗ Not authenticated"))
		fmt.Println(infoStyleAuth.Render("Run 'todoinfo login' to authenticate"))
	}

	return nil
}
