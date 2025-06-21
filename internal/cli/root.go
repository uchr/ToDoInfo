package cli

import (
	"context"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile  string
	clientID string
	verbose  bool

	logger *slog.Logger
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "todoinfo",
	Short: "Analyze and visualize your Microsoft ToDo tasks",
	Long: `ToDo Info is a beautiful CLI tool that helps you analyze old tasks from Microsoft ToDo.
It connects to Microsoft Graph API to fetch tasks, calculates task ages, and categorizes 
them by "rottenness" levels with beautiful visualizations.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose {
			pterm.EnableDebugMessages()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute(ctx context.Context) error {
	rootCmd.SetContext(ctx)
	return rootCmd.Execute()
}

func init() {
	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.todoinfo.yaml)")
	rootCmd.PersistentFlags().StringVar(&clientID, "client-id", "", "Azure AD Client ID (required)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Bind flags to viper
	viper.BindPFlag("client-id", rootCmd.PersistentFlags().Lookup("client-id"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	// Note: client-id is marked as required per command, not globally
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Load .env file if it exists
	_ = godotenv.Load()

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".todoinfo" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".todoinfo")
	}

	// Environment variables
	viper.SetEnvPrefix("TODOINFO")
	viper.AutomaticEnv()

	// Also check for AZURE_CLIENT_ID
	if clientID == "" && viper.GetString("client-id") == "" {
		if azureClientID := os.Getenv("AZURE_CLIENT_ID"); azureClientID != "" {
			viper.Set("client-id", azureClientID)
		}
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			pterm.Info.Printf("Using config file: %s", viper.ConfigFileUsed())
		}
	}
}
