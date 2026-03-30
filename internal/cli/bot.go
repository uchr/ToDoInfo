package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/uchr/ToDoInfo/internal/auth"
	tgbot "github.com/uchr/ToDoInfo/internal/bot"
	"github.com/uchr/ToDoInfo/internal/service"
)

var botCmd = &cobra.Command{
	Use:   "bot",
	Short: "Run Telegram bot with periodic stats collection",
	Long: `Start a long-running Telegram bot that periodically fetches task statistics
from Microsoft Graph and provides on-demand queries + daily summaries.

Commands available in the bot:
  /login   - Authenticate via device code flow
  /stats   - Show task statistics summary
  /zombies - List all zombie tasks (14+ days old)
  /oldest  - Show the oldest task
  /chart   - Send radar chart of tasks by project
  /refresh - Force data refresh`,
	RunE: runBot,
}

func init() {
	rootCmd.AddCommand(botCmd)

	botCmd.Flags().String("telegram-token", "", "Telegram Bot API token")
	botCmd.Flags().Int64("telegram-chat-id", 0, "Allowed Telegram chat ID")
	botCmd.Flags().String("refresh-interval", "4h", "Data refresh interval (e.g. 4h, 30m)")

	_ = viper.BindPFlag("telegram-token", botCmd.Flags().Lookup("telegram-token"))
	_ = viper.BindPFlag("telegram-chat-id", botCmd.Flags().Lookup("telegram-chat-id"))
	_ = viper.BindPFlag("refresh-interval", botCmd.Flags().Lookup("refresh-interval"))
}

func runBot(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	telegramToken := viper.GetString("telegram-token")
	if telegramToken == "" {
		telegramToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	}
	if telegramToken == "" {
		return fmt.Errorf("telegram-token is required (use --telegram-token flag or TELEGRAM_BOT_TOKEN env var)")
	}

	chatID := viper.GetInt64("telegram-chat-id")
	if chatID == 0 {
		if v := os.Getenv("TELEGRAM_CHAT_ID"); v != "" {
			parsed, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid TELEGRAM_CHAT_ID: %w", err)
			}
			chatID = parsed
		}
	}
	if chatID == 0 {
		return fmt.Errorf("telegram-chat-id is required (use --telegram-chat-id flag or TELEGRAM_CHAT_ID env var)")
	}

	refreshStr := viper.GetString("refresh-interval")
	refreshInterval, err := time.ParseDuration(refreshStr)
	if err != nil {
		return fmt.Errorf("invalid refresh-interval %q: %w", refreshStr, err)
	}

	dailySummaryTime := os.Getenv("DAILY_SUMMARY_TIME")
	if dailySummaryTime == "" {
		dailySummaryTime = "09:00"
	}

	clientID := viper.GetString("client-id")
	if clientID == "" {
		return fmt.Errorf("client-id is required (use --client-id flag or AZURE_CLIENT_ID env var)")
	}

	botLogger := slog.Default()
	if verbose {
		botLogger = logger
	}

	// Declare bot pointer so the device code callback can reference it.
	// SendMessage has a nil guard on tgBot, so it's safe before Run() is called.
	var telegramBot *tgbot.Bot

	// Create auth client with headless/device code mode.
	// The callback forwards the device code prompt to Telegram.
	authCfg := auth.DefaultConfig().
		WithClientID(clientID).
		WithHeadless(func(ctx context.Context, message string) {
			telegramBot.SendMessage(ctx, message)
		})

	authClient, err := auth.NewAuthClient(authCfg)
	if err != nil {
		return fmt.Errorf("failed to create auth client: %w", err)
	}

	// Try to restore credentials from cache (no user interaction)
	if authClient.TryNonInteractiveAuth(ctx, botLogger) {
		botLogger.Info("restored authentication from cache")
	}

	collector := service.NewCollector(authClient, botLogger, refreshInterval)

	botCfg := tgbot.BotConfig{
		Token:            telegramToken,
		ChatID:           chatID,
		DailySummaryTime: dailySummaryTime,
	}
	telegramBot = tgbot.New(botCfg, collector, authClient, botLogger)

	// Start collector in background
	go func() {
		if err := collector.Run(ctx); err != nil {
			botLogger.Error("collector stopped", slog.Any("error", err))
		}
	}()

	// Start bot (blocks until ctx cancelled)
	return telegramBot.Run(ctx)
}
