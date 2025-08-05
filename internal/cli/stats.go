package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/NimbleMarkets/ntcharts/canvas"
	"github.com/NimbleMarkets/ntcharts/linechart"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/uchr/ToDoInfo/internal/auth"
	"github.com/uchr/ToDoInfo/internal/storage"
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
- Historical trends and charts

Use --offline flag to view previously stored data without authentication.`,
	RunE: runStats,
}

func init() {
	rootCmd.AddCommand(statsCmd)

	// Add offline flag
	statsCmd.Flags().Bool("offline", false, "Use existing stored data without fetching new statistics (no authentication required)")

	// Don't mark client-id as required since we have offline mode
	// The command will check for it when not in offline mode
}

func runStats(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Show banner
	showBanner()

	// Check if offline mode is enabled
	offlineMode, _ := cmd.Flags().GetBool("offline")

	var metrics *todometrics.Metrics

	if !offlineMode {
		// Fetch new data
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

		// Authenticate
		fmt.Print(infoStyle.Render("🔐 Authenticating with Microsoft..."))
		if err := authClient.Authenticate(ctx, logger); err != nil {
			fmt.Println(" " + errorStyle.Render("✗ Authentication failed"))
			return fmt.Errorf("authentication failed: %w", err)
		}
		fmt.Println(" " + successStyle.Render("✓ Authentication successful!"))

		// Fetch tasks with progress
		tasks, err := fetchTasks(ctx, logger, authClient)
		if err != nil {
			return fmt.Errorf("failed to fetch tasks: %w", err)
		}

		// Calculate metrics
		metrics = todometrics.New(tasks)

		// Store statistics for historical tracking
		if err := storeStatistics(ctx, metrics, tasks); err != nil {
			// Don't fail the command if storage fails, just log a warning
			fmt.Println(warningStyle.Render(fmt.Sprintf("⚠ Failed to store statistics: %v", err)))
		}
	} else {
		// Use existing stored data
		fmt.Println(infoStyle.Render("📊 Offline Mode: Using existing stored data"))
		fmt.Println()

		snapshot, err := loadLatestSnapshot(ctx)
		if err != nil {
			return fmt.Errorf("failed to load stored data: %w", err)
		}

		fmt.Println(infoStyle.Render(fmt.Sprintf("📅 Data from: %s", snapshot.Timestamp.Format("2006-01-02 15:04"))))
		fmt.Println()

		// Create metrics from stored snapshot
		metrics = createMetricsFromSnapshot(snapshot)

		// If no tasks available (due to parsing issues or old format), show limited info
		if len(metrics.GetSortedTasks()) == 0 && len(snapshot.TaskLists) > 0 {
			fmt.Println(warningStyle.Render("⚠ Full task data unavailable due to format changes. Showing summary only."))
			fmt.Println()
		}
	}

	// Always use the same render function
	displayStatistics(metrics)

	// Display historical graphs at the bottom
	displayHistoricalGraphs(ctx)

	return nil
}

// loadLatestSnapshot loads the most recent stored statistics snapshot
func loadLatestSnapshot(ctx context.Context) (*storage.StatsSnapshot, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot get home directory: %w", err)
	}

	dataDir := filepath.Join(homeDir, ".todoinfo", "data")
	store, err := storage.NewJSONStorage(dataDir)
	if err != nil {
		return nil, fmt.Errorf("cannot access data directory: %w", err)
	}
	defer store.Close()

	snapshot, err := store.GetLatest(ctx)
	if err != nil {
		// Check if it's a JSON parsing error due to task data format incompatibility
		if strings.Contains(err.Error(), "unmarshal") || strings.Contains(err.Error(), "cannot unmarshal") {
			return nil, fmt.Errorf("stored data format is incompatible - run 'todoinfo stats' with authentication to generate new data")
		}
		return nil, fmt.Errorf("no stored data available - run 'todoinfo stats' with authentication first to generate data: %w", err)
	}

	return snapshot, nil
}

// createMetricsFromSnapshot creates a metrics object from stored snapshot data
func createMetricsFromSnapshot(snapshot *storage.StatsSnapshot) *todometrics.Metrics {
	// If we have full task data stored, try to use it to create proper metrics
	if len(snapshot.TaskLists) > 0 {
		// Try to create metrics from stored task data
		// If there are issues with the data format, we'll fall back to empty metrics
		return todometrics.New(snapshot.TaskLists)
	}

	// Fallback for backward compatibility with old snapshots that don't have task data
	// Return empty metrics - the display functions will show limited info from snapshot data
	return todometrics.New([]todo.TaskList{})
}

var (
	// Define styles
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D7FF")).
			Bold(true).
			MarginLeft(2).
			MarginRight(2)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF79C6")).
			Bold(true)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8BE9FD")).
			Bold(true).
			MarginTop(1).
			MarginBottom(1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FFB86C")).
			Padding(1, 2).
			Margin(1, 0)

	tableHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#50FA7B")).
				Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B"))

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F1FA8C"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4"))
)

func showBanner() {
	// ASCII art banner like the original
	banner := `
████████╗ ██████╗ ██████╗  ██████╗     ██╗███╗   ██╗███████╗ ██████╗ 
╚══██╔══╝██╔═══██╗██╔══██╗██╔═══██╗    ██║████╗  ██║██╔════╝██╔═══██╗
   ██║   ██║   ██║██║  ██║██║   ██║    ██║██╔██╗ ██║█████╗  ██║   ██║
   ██║   ██║   ██║██║  ██║██║   ██║    ██║██║╚██╗██║██╔══╝  ██║   ██║
   ██║   ╚██████╔╝██████╔╝╚██████╔╝    ██║██║ ╚████║██║     ╚██████╔╝
   ╚═╝    ╚═════╝ ╚═════╝  ╚═════╝     ╚═╝╚═╝  ╚═══╝╚═╝      ╚═════╝ `

	subtitle := infoStyle.Render("Microsoft ToDo Task Analysis Tool")

	// Use the ASCII art with colors
	lines := strings.Split(banner, "\n")
	for i, line := range lines {
		if i == 0 {
			continue // Skip first empty line
		}
		if i <= 3 {
			fmt.Println(titleStyle.Render(line))
		} else {
			fmt.Println(subtitleStyle.Render(line))
		}
	}

	fmt.Println()
	fmt.Println(lipgloss.Place(80, 1, lipgloss.Center, lipgloss.Center, subtitle))
	fmt.Println()
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

func displayStatistics(metrics *todometrics.Metrics) {
	// Global stats
	displayGlobalStats(metrics)

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
		content := fmt.Sprintf("The oldest task is \"%s\" from list \"%s\"\nAge: %d days %s",
			truncateString(oldest.TaskName, 50),
			truncateString(oldest.TaskList, 30),
			oldest.Age,
			oldest.Rottenness.String(),
		)

		box := boxStyle.
			BorderForeground(lipgloss.Color("#FFD700")).
			Render(headerStyle.Render("🏅 Champion Procrastinator") + "\n" + content)

		fmt.Println(box)
	}
}

func displayGlobalStats(metrics *todometrics.Metrics) {
	fmt.Println(headerStyle.Render("📊 Global Stats"))

	// Calculate total age of all tasks
	allTasksInfo := metrics.GetSortedTasks()
	totalAge := 0
	for _, task := range allTasksInfo {
		totalAge += task.Age
	}

	// Create lipgloss table
	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB86C"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if col == 0 {
				return tableHeaderStyle
			}
			return lipgloss.NewStyle()
		}).
		Headers("Metric", "Value").
		Row("Tasks", fmt.Sprintf("%d", len(allTasksInfo))).
		Row("Total Age", fmt.Sprintf("%d days", totalAge))

	fmt.Println(t.Render())
}

func displayListAges(metrics *todometrics.Metrics) {
	fmt.Println(headerStyle.Render("📋 Task Age by List"))

	listAges := metrics.GetListAges()

	// Create lipgloss table
	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB86C"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return tableHeaderStyle
			}
			return lipgloss.NewStyle()
		}).
		Headers("List Name", "Total Age (days)", "Task Count", "Share")

	for _, listAge := range listAges.Ages {
		percentage := 0.0
		if listAges.TotalAge > 0 {
			percentage = float64(listAge.Age) / float64(listAges.TotalAge) * 100
		}

		t.Row(
			listAge.Title,
			strconv.Itoa(listAge.Age),
			strconv.Itoa(listAge.TaskCount),
			fmt.Sprintf("%.1f%%", percentage),
		)
	}

	fmt.Println(t.Render())
}

func displayTopOldestTasks(metrics *todometrics.Metrics) {
	fmt.Println(headerStyle.Render("🏆 All Tasks (Oldest First)"))

	// Get all tasks instead of limiting to topN
	allTasks := metrics.GetSortedTasks()

	if len(allTasks) == 0 {
		fmt.Println(infoStyle.Render("No tasks found!"))
		return
	}

	// Create lipgloss table
	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB86C"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return tableHeaderStyle
			}
			return lipgloss.NewStyle()
		}).
		Headers("Rank", "Task", "List", "Age (days)", "Status")

	for i, task := range allTasks {
		t.Row(
			strconv.Itoa(i+1),
			task.TaskName,
			task.TaskList,
			strconv.Itoa(task.Age),
			task.Rottenness.String(),
		)
	}

	fmt.Println(t.Render())
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// storeStatistics stores current statistics to persistent storage
func storeStatistics(ctx context.Context, metrics *todometrics.Metrics, tasks []todo.TaskList) error {
	// Create data directory in user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	dataDir := filepath.Join(homeDir, ".todoinfo", "data")

	// Create storage
	store, err := storage.NewJSONStorage(dataDir)
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}
	defer store.Close()

	// Calculate max age and task count
	allTasks := metrics.GetSortedTasks()
	totalAge := 0
	if len(allTasks) > 0 {
		for _, task := range allTasks {
			totalAge += task.Age
		}
	}

	// Create snapshot with full task data for better offline support
	snapshot := storage.StatsSnapshot{
		Timestamp: time.Now(),
		GlobalStats: storage.GlobalStats{
			TotalAge:  totalAge,
			TaskCount: len(allTasks),
		},
		ListAges:  metrics.GetListAges(),
		TaskCount: len(allTasks),
		TaskLists: tasks, // Store full task data
	}

	return store.Store(ctx, snapshot)
}

// displayHistoricalGraphs displays the historical graphs at the bottom
func displayHistoricalGraphs(ctx context.Context) {
	// Create data directory path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(warningStyle.Render("⚠ Cannot get home directory for historical data"))
		return
	}

	dataDir := filepath.Join(homeDir, ".todoinfo", "data")

	// Create storage
	store, err := storage.NewJSONStorage(dataDir)
	if err != nil {
		fmt.Println(warningStyle.Render("⚠ Cannot access historical data storage"))
		return
	}
	defer store.Close()

	// Get time series data for the last 90 days
	points, err := store.GetTimeSeriesData(ctx, 90)
	if err != nil {
		fmt.Println(warningStyle.Render("⚠ No historical data available yet - run stats a few times to build history"))
		return
	}

	if len(points) == 0 {
		fmt.Println(infoStyle.Render("📊 No historical data points found - run stats regularly to build trends"))
		return
	}

	// Create and render the graphs with ntcharts
	renderHistoricalCharts(points)
}

// renderHistoricalCharts creates bar charts using ntcharts
func renderHistoricalCharts(points []storage.TimeSeriesPoint) {
	if len(points) == 0 {
		return
	}

	fmt.Println()
	fmt.Println(headerStyle.Render("📈 Historical Task Age Trend (Last 90 Days)"))

	// Prepare data for age trend chart
	ageData := make([]float64, len(points))
	labels := make([]string, len(points))

	for i, point := range points {
		ageData[i] = float64(point.MaxAge)
		labels[i] = point.Date.Format("01-02")
	}

	renderLineChart("Task Age (Days)", ageData, labels)

	fmt.Println()
	fmt.Println(headerStyle.Render("📈 Historical Task Count Trend (Last 90 Days)"))

	// Prepare data for task count chart
	countData := make([]float64, len(points))

	for i, point := range points {
		countData[i] = float64(point.TaskCount)
	}

	renderLineChart("Task Count", countData, labels)
}

// renderLineChart creates a line chart using ntcharts
func renderLineChart(_ string, data []float64, labels []string) {
	if len(data) == 0 {
		emptyChart := "No data points available"
		fmt.Println(boxStyle.Render(emptyChart))
		return
	}

	// Show data range for context
	minVal, maxVal := findMinMax(data)
	rangeInfo := fmt.Sprintf("Range: %.0f - %.0f | Data points: %d", minVal, maxVal, len(data))
	fmt.Println(infoStyle.Render(rangeInfo))
	fmt.Println()

	// Create line chart with appropriate dimensions
	chartWidth := 80
	chartHeight := 12

	// Set up coordinate ranges
	minX := 0.0
	maxX := float64(len(data) - 1)

	lc := linechart.New(chartWidth, chartHeight, minX, maxX, minVal, maxVal,
		linechart.WithXLabelFormatter(func(i int, x float64) string {
			idx := int(x)
			if idx >= 0 && idx < len(labels) {
				return labels[idx]
			}
			return ""
		}),
		linechart.WithYLabelFormatter(func(i int, y float64) string {
			return fmt.Sprintf("%.0f", y)
		}),
	)

	// Draw the axes and labels first
	lc.DrawXYAxisAndLabel()

	// Draw line connecting all data points
	for i := 0; i < len(data)-1; i++ {
		point1 := canvas.Float64Point{X: float64(i), Y: data[i]}
		point2 := canvas.Float64Point{X: float64(i + 1), Y: data[i+1]}
		lc.DrawBrailleLine(point1, point2)
	}

	// Render the chart
	chartView := lc.View()

	if chartView == "" || len(strings.TrimSpace(chartView)) == 0 {
		// Fallback to simple text chart
		fallbackChart := createSimpleLineChart(data, labels)
		fmt.Println(boxStyle.Render(fallbackChart))
	} else {
		fmt.Println(boxStyle.Render(chartView))
	}
}

// findMinMax finds the minimum and maximum values in a slice
func findMinMax(data []float64) (float64, float64) {
	if len(data) == 0 {
		return 0, 0
	}

	min := data[0]
	max := data[0]

	for _, val := range data {
		if val < min {
			min = val
		}
		if val > max {
			max = val
		}
	}

	return min, max
}

// createSimpleLineChart creates a simple ASCII line chart when ntcharts fails
func createSimpleLineChart(data []float64, labels []string) string {
	if len(data) == 0 {
		return "No data"
	}

	// Find max value for scaling
	maxVal := 0.0
	for _, val := range data {
		if val > maxVal {
			maxVal = val
		}
	}

	if maxVal == 0 {
		return "All values are zero"
	}

	chart := strings.Builder{}
	maxBarLength := 40

	// Add header
	chart.WriteString(fmt.Sprintf("%-6s │ %-40s │ %s\n", "Date", "Line Chart", "Value"))
	chart.WriteString(strings.Repeat("-", 55) + "\n")

	for i, val := range data {
		label := labels[i]
		if len(label) > 6 {
			label = label[:6]
		}

		barLength := int((val / maxVal) * float64(maxBarLength))
		if barLength < 1 && val > 0 {
			barLength = 1 // Show at least 1 character for non-zero values
		}

		// Create a simple line representation
		bar := strings.Repeat(" ", barLength-1) + "●"
		if barLength == 0 {
			bar = "○"
		}

		// Pad bar to consistent width for alignment
		paddedBar := fmt.Sprintf("%-40s", bar)

		chart.WriteString(fmt.Sprintf("%-6s │ %s │ %.0f\n", label, paddedBar, val))
	}

	return chart.String()
}
