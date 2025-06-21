package cli

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/uchr/ToDoInfo/internal/auth"
	"github.com/uchr/ToDoInfo/internal/todo"
	"github.com/uchr/ToDoInfo/internal/todoclient"
	"github.com/uchr/ToDoInfo/internal/todometrics"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Display beautiful statistics about your ToDo tasks",
	Long: `Analyze your Microsoft ToDo tasks and display comprehensive statistics including:
- Task rottenness levels (Fresh, Ripe, Tired, Zombie)
- Age distribution across lists
- Complete list of all tasks sorted by age
- Completion rates and trends`,
	RunE: runStats,
}

func init() {
	rootCmd.AddCommand(statsCmd)

	// Mark client-id as required for this command
	statsCmd.MarkPersistentFlagRequired("client-id")
}

func runStats(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Show banner
	showBanner()

	// Get client ID from viper (handles both flags and env vars)
	clientID := viper.GetString("client-id")
	if clientID == "" {
		return fmt.Errorf("client-id is required (use --client-id flag or AZURE_CLIENT_ID env var)")
	}

	// Create auth client
	authClient, err := createAuthClient(clientID)
	if err != nil {
		return fmt.Errorf("failed to create auth client: %w", err)
	}

	// Authenticate with loading spinner
	spinner, _ := pterm.DefaultSpinner.Start("🔐 Authenticating with Microsoft...")
	if err := authClient.Authenticate(ctx); err != nil {
		spinner.Fail("Authentication failed")
		return fmt.Errorf("authentication failed: %w", err)
	}
	spinner.Success("✅ Authentication successful!")

	// Fetch tasks with progress
	tasks, err := fetchTasks(ctx, logger, authClient)
	if err != nil {
		return fmt.Errorf("failed to fetch tasks: %w", err)
	}

	// Calculate metrics
	metrics := todometrics.New(tasks)

	// Display statistics
	displayStatistics(metrics, tasks)

	return nil
}

func showBanner() {
	banner := pterm.DefaultBigText.WithLetters(
		pterm.NewLettersFromStringWithStyle("TODO", pterm.NewStyle(pterm.FgCyan)),
		pterm.NewLettersFromStringWithStyle("INFO", pterm.NewStyle(pterm.FgLightMagenta)),
	)
	banner.Render()

	pterm.DefaultCenter.WithCenterEachLineSeparately().Println(
		pterm.Gray("Microsoft ToDo Task Analysis Tool"),
	)
	pterm.Println()
}

func fetchTasks(ctx context.Context, logger *slog.Logger, authClient *auth.AuthClient) ([]todo.TaskList, error) {
	// Extract access token from the auth client for use with old HTTP client
	token, err := extractAccessToken(ctx, authClient)
	if err != nil {
		return nil, fmt.Errorf("failed to extract access token: %w", err)
	}

	// Use the old TodoParser that was working
	parser := todoclient.New()

	taskLists, err := parser.GetTasks(ctx, logger, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get task lists: %w", err)
	}

	return taskLists, nil
}

// extractAccessToken gets the access token from the new auth client
func extractAccessToken(ctx context.Context, authClient *auth.AuthClient) (string, error) {
	return authClient.GetAccessToken(ctx)
}

func displayStatistics(metrics *todometrics.Metrics, allTasks []todo.TaskList) {
	// Global stats
	displayGlobalStats(metrics, allTasks)

	// Task Age by List (second section)
	displayListAges(metrics)

	// Top 10 Oldest Tasks
	displayTopOldestTasks(metrics)

	// Champion Procrastinator box at the very top
	displayChampionProcrastinator(metrics)
}

func displayChampionProcrastinator(metrics *todometrics.Metrics) {
	topTasks := metrics.GetTopTasksByAge(1)
	if len(topTasks) > 0 {
		oldest := topTasks[0]
		pterm.DefaultBox.WithTitle("🏅 Champion Procrastinator").WithTitleTopCenter().Println(
			fmt.Sprintf("The oldest task is \"%s\" from list \"%s\"\nAge: %d days %s",
				truncateString(oldest.TaskName, 50),
				truncateString(oldest.TaskList, 30),
				oldest.Age,
				oldest.Rottenness.String(),
			),
		)
	}
	pterm.Println()
}

func displayGlobalStats(metrics *todometrics.Metrics, allTasks []todo.TaskList) {
	pterm.DefaultSection.Println("📊 Global Stats")

	// Calculate total age of all tasks
	allTasksInfo := metrics.GetSortedTasks()
	totalAge := 0
	for _, task := range allTasksInfo {
		totalAge += task.Age
	}

	globalData := pterm.TableData{
		{"Age", fmt.Sprintf("%d days", totalAge)},
	}

	pterm.DefaultTable.WithData(globalData).Render()
}

func displayListAges(metrics *todometrics.Metrics) {
	pterm.DefaultSection.Println("📋 Task Age by List")

	listAges := metrics.GetListAges()

	listData := pterm.TableData{{"List Name", "Total Age (days)", "Share"}}

	for _, listAge := range listAges.Ages {
		percentage := 0.0
		if listAges.TotalAge > 0 {
			percentage = float64(listAge.Age) / float64(listAges.TotalAge) * 100
		}

		listData = append(listData, []string{
			listAge.Title,
			strconv.Itoa(listAge.Age),
			fmt.Sprintf("%.1f%%", percentage),
		})
	}

	pterm.DefaultTable.WithHasHeader().WithData(listData).WithSeparator(" | ").Render()
}

func displayTopOldestTasks(metrics *todometrics.Metrics) {
	pterm.DefaultSection.Println("🏆 All Tasks (Oldest First)")

	// Get all tasks instead of limiting to topN
	allTasks := metrics.GetSortedTasks()

	if len(allTasks) == 0 {
		pterm.Info.Println("No tasks found!")
		return
	}

	taskData := pterm.TableData{{"Rank", "Task", "List", "Age (days)", "Status"}}

	for i, task := range allTasks {
		taskData = append(taskData, []string{
			strconv.Itoa(i + 1),
			truncateString(task.TaskName, 40),
			truncateString(task.TaskList, 20),
			strconv.Itoa(task.Age),
			task.Rottenness.String(),
		})
	}

	pterm.DefaultTable.WithHasHeader().WithData(taskData).WithSeparator(" | ").Render()
}

func truncateString(s string, maxLen int) string {
	return s
}
