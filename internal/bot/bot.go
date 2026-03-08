package bot

import (
	"bytes"
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/go-analyze/charts"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/uchr/ToDoInfo/internal/auth"
	"github.com/uchr/ToDoInfo/internal/service"
	"github.com/uchr/ToDoInfo/internal/storage"
	"github.com/uchr/ToDoInfo/internal/todometrics"
)

// BotConfig holds Telegram bot configuration.
type BotConfig struct {
	Token            string
	ChatID           int64
	DailySummaryTime string // "HH:MM" format, e.g. "09:00"
}

// Bot is the Telegram bot that serves task statistics.
type Bot struct {
	config    BotConfig
	collector *service.Collector
	auth      *auth.AuthClient
	logger    *slog.Logger
	tgBot     *bot.Bot
}

// New creates a new Bot instance.
func New(cfg BotConfig, collector *service.Collector, authClient *auth.AuthClient, logger *slog.Logger) *Bot {
	return &Bot{
		config:    cfg,
		collector: collector,
		auth:      authClient,
		logger:    logger,
	}
}

// Run starts the Telegram bot and blocks until ctx is cancelled.
func (b *Bot) Run(ctx context.Context) error {
	tgBot, err := bot.New(b.config.Token,
		bot.WithDefaultHandler(b.defaultHandler),
	)
	if err != nil {
		return fmt.Errorf("create telegram bot: %w", err)
	}
	b.tgBot = tgBot

	// Register command handlers
	b.tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/stats", bot.MatchTypeCommand, b.handleStats)
	b.tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/zombies", bot.MatchTypeCommand, b.handleZombies)
	b.tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/oldest", bot.MatchTypeCommand, b.handleOldest)
	b.tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/chart", bot.MatchTypeCommand, b.handleChart)
	b.tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/login", bot.MatchTypeCommand, b.handleLogin)
	b.tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/refresh", bot.MatchTypeCommand, b.handleRefresh)

	// Start daily summary scheduler
	go b.dailyScheduler(ctx)

	b.logger.Info("telegram bot starting", slog.Int64("chat_id", b.config.ChatID))

	// Start polling (blocks until ctx cancelled)
	b.tgBot.Start(ctx)
	return nil
}

// SendMessage sends a text message to the configured chat.
func (b *Bot) SendMessage(ctx context.Context, text string) {
	if b.tgBot == nil {
		return
	}
	_, err := b.tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    b.config.ChatID,
		Text:      text,
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		b.logger.Error("failed to send message", slog.Any("error", err))
	}
}

func (b *Bot) defaultHandler(ctx context.Context, tg *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	if update.Message.Chat.ID != b.config.ChatID {
		b.logger.Warn("unauthorized chat", slog.Int64("chat_id", update.Message.Chat.ID))
		return
	}
}

func (b *Bot) handleStats(ctx context.Context, tg *bot.Bot, update *models.Update) {
	if !b.authorizeChat(update) {
		return
	}

	data := b.collector.GetLatest()
	if data == nil {
		b.sendReply(ctx, tg, update, "No data yet. Use /login first, then wait for the first data collection.")
		return
	}

	text := b.formatStats(data)
	b.sendReply(ctx, tg, update, text)
}

func (b *Bot) handleZombies(ctx context.Context, tg *bot.Bot, update *models.Update) {
	if !b.authorizeChat(update) {
		return
	}

	data := b.collector.GetLatest()
	if data == nil {
		b.sendReply(ctx, tg, update, "No data yet. Use /login first.")
		return
	}

	var zombies []todometrics.TaskRottennessInfo
	for _, t := range data.SortedTasks {
		if t.Rottenness == todometrics.ZombieTaskRottenness {
			zombies = append(zombies, t)
		}
	}

	if len(zombies) == 0 {
		b.sendReply(ctx, tg, update, "No zombie tasks! Everything is fresh.")
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>%s Zombie Tasks (%d)</b>\n\n", todometrics.TaskRottenness(todometrics.ZombieTaskRottenness).String(), len(zombies)))
	for i, t := range zombies {
		sb.WriteString(fmt.Sprintf("%d. <b>%s</b>\n   %s | %d days %s\n",
			i+1,
			escapeHTML(t.TaskName),
			escapeHTML(t.TaskList),
			t.Age,
			t.Rottenness.String(),
		))
	}

	b.sendReply(ctx, tg, update, sb.String())
}

func (b *Bot) handleOldest(ctx context.Context, tg *bot.Bot, update *models.Update) {
	if !b.authorizeChat(update) {
		return
	}

	data := b.collector.GetLatest()
	if data == nil {
		b.sendReply(ctx, tg, update, "No data yet. Use /login first.")
		return
	}

	if data.Champion == nil {
		b.sendReply(ctx, tg, update, "No tasks found!")
		return
	}

	c := data.Champion
	text := fmt.Sprintf(
		"<b>Champion Procrastinator</b>\n\n"+
			"<b>%s</b>\n"+
			"List: %s\n"+
			"Age: %d days %s",
		escapeHTML(c.TaskName),
		escapeHTML(c.TaskList),
		c.Age,
		c.Rottenness.String(),
	)
	b.sendReply(ctx, tg, update, text)
}

func (b *Bot) handleChart(ctx context.Context, tg *bot.Bot, update *models.Update) {
	if !b.authorizeChat(update) {
		return
	}

	data := b.collector.GetLatest()
	if data == nil {
		b.sendReply(ctx, tg, update, "No data yet. Use /login first.")
		return
	}

	// Send radar chart
	radarPNG, err := b.renderRadarChart(data)
	if err != nil {
		b.logger.Error("radar chart render failed", slog.Any("error", err))
	} else {
		_, err = tg.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID: update.Message.Chat.ID,
			Photo: &models.InputFileUpload{
				Filename: "tasks_radar.png",
				Data:     bytes.NewReader(radarPNG),
			},
			Caption: fmt.Sprintf("Task distribution by project (%s)", data.FetchedAt.Format("2006-01-02 15:04")),
		})
		if err != nil {
			b.logger.Error("failed to send radar chart", slog.Any("error", err))
		}
	}

	// Send history line chart
	historyPNG, err := b.renderHistoryChart(ctx)
	if err != nil {
		b.logger.Warn("history chart render failed", slog.Any("error", err))
		return
	}
	_, err = tg.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID: update.Message.Chat.ID,
		Photo: &models.InputFileUpload{
			Filename: "tasks_history.png",
			Data:     bytes.NewReader(historyPNG),
		},
		Caption: "Total age over time (per project)",
	})
	if err != nil {
		b.logger.Error("failed to send history chart", slog.Any("error", err))
	}
}

func (b *Bot) handleLogin(ctx context.Context, tg *bot.Bot, update *models.Update) {
	if !b.authorizeChat(update) {
		return
	}

	if b.auth.HasCredential() {
		b.sendReply(ctx, tg, update, "Already authenticated! Use /refresh to fetch fresh data.")
		return
	}

	b.sendReply(ctx, tg, update, "Starting device code authentication...")

	if err := b.auth.Authenticate(ctx, b.logger); err != nil {
		b.sendReply(ctx, tg, update, fmt.Sprintf("Authentication failed: %v", err))
		return
	}

	b.sendReply(ctx, tg, update, "Authentication successful! Fetching initial data...")

	if err := b.collector.Refresh(ctx); err != nil {
		b.sendReply(ctx, tg, update, fmt.Sprintf("Data fetch failed: %v", err))
		return
	}

	b.sendReply(ctx, tg, update, "Ready! Use /stats, /zombies, /oldest, or /chart.")
}

func (b *Bot) handleRefresh(ctx context.Context, tg *bot.Bot, update *models.Update) {
	if !b.authorizeChat(update) {
		return
	}

	b.sendReply(ctx, tg, update, "Refreshing data...")

	if err := b.collector.Refresh(ctx); err != nil {
		b.sendReply(ctx, tg, update, fmt.Sprintf("Refresh failed: %v", err))
		return
	}

	data := b.collector.GetLatest()
	if data == nil {
		b.sendReply(ctx, tg, update, "Refresh completed but no data available.")
		return
	}

	b.sendReply(ctx, tg, update, fmt.Sprintf("Refreshed! %d tasks, total age %d days.", data.TotalTasks, data.TotalAge))
}

// dailyScheduler sends a daily summary at the configured time.
func (b *Bot) dailyScheduler(ctx context.Context) {
	for {
		next := b.nextDailySummary()
		b.logger.Info("next daily summary", slog.Time("at", next))

		timer := time.NewTimer(time.Until(next))
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			b.sendDailySummary(ctx)
		}
	}
}

func (b *Bot) nextDailySummary() time.Time {
	hour, minute := 9, 0 // default 09:00
	if b.config.DailySummaryTime != "" {
		if _, err := fmt.Sscanf(b.config.DailySummaryTime, "%d:%d", &hour, &minute); err != nil {
			b.logger.Warn("invalid daily summary time, using 09:00", slog.String("value", b.config.DailySummaryTime))
			hour, minute = 9, 0
		}
	}

	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next
}

func (b *Bot) sendDailySummary(ctx context.Context) {
	data := b.collector.GetLatest()
	if data == nil {
		return
	}

	text := "<b>Daily Summary</b>\n\n" + b.formatStats(data)
	b.SendMessage(ctx, text)
}

// formatStats produces a text summary from StatsData.
// Format:
//
//	Total: (X days, Y tasks)
//	# Group 1 (X days, Y tasks)
//	  top 5 oldest tasks
//	# Group 2 ...
func (b *Bot) formatStats(data *service.StatsData) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>Total: %d days, %d tasks</b>\n", data.TotalAge, data.TotalTasks))

	// Build a map: list name → top 5 oldest tasks
	tasksByList := make(map[string][]todometrics.TaskRottennessInfo)
	for _, t := range data.SortedTasks {
		if len(tasksByList[t.TaskList]) < 5 {
			tasksByList[t.TaskList] = append(tasksByList[t.TaskList], t)
		}
	}

	for _, la := range data.ListAges.Ages {
		sb.WriteString(fmt.Sprintf("\n<b>%s</b> (%d days, %d tasks)\n",
			escapeHTML(la.Title), la.Age, la.TaskCount))

		tasks := tasksByList[la.Title]
		for _, t := range tasks {
			sb.WriteString(fmt.Sprintf("  %s %s — %d days\n",
				t.Rottenness.String(),
				escapeHTML(t.TaskName),
				t.Age))
		}
	}

	return sb.String()
}

// renderRadarChart generates a radar chart PNG with two series: age and count per list.
// Both series are normalized to 0–100% of their respective maximums so they're
// visually comparable on the same radar axes.
func (b *Bot) renderRadarChart(data *service.StatsData) ([]byte, error) {
	ages := data.ListAges.Ages
	if len(ages) == 0 {
		return nil, fmt.Errorf("no list data for chart")
	}

	// Find global max for each series to normalize
	var maxAge, maxCount float64
	for _, la := range ages {
		if float64(la.Age) > maxAge {
			maxAge = float64(la.Age)
		}
		if float64(la.TaskCount) > maxCount {
			maxCount = float64(la.TaskCount)
		}
	}
	if maxAge < 1 {
		maxAge = 1
	}
	if maxCount < 1 {
		maxCount = 1
	}

	names := make([]string, len(ages))
	maxValues := make([]float64, len(ages))
	ageValues := make([]float64, len(ages))
	countValues := make([]float64, len(ages))

	for i, la := range ages {
		names[i] = stripEmojis(la.Title)
		// Normalize to 0–100 scale
		ageValues[i] = float64(la.Age) / maxAge * 100
		countValues[i] = float64(la.TaskCount) / maxCount * 100
		maxValues[i] = 100
	}

	seriesData := [][]float64{ageValues, countValues}

	opt := charts.NewRadarChartOptionWithData(seriesData, names, maxValues)
	opt.Legend = charts.LegendOption{
		SeriesNames: []string{
			fmt.Sprintf("Age (max %d days)", int(maxAge)),
			fmt.Sprintf("Count (max %d)", int(maxCount)),
		},
	}

	p := charts.NewPainter(charts.PainterOptions{
		Width:  600,
		Height: 400,
	}, charts.PainterThemeOption(charts.GetTheme(charts.ThemeGrafana)))

	if err := p.RadarChart(opt); err != nil {
		return nil, fmt.Errorf("render radar chart: %w", err)
	}

	buf, err := p.Bytes()
	if err != nil {
		return nil, fmt.Errorf("encode chart png: %w", err)
	}

	return buf, nil
}

// renderHistoryChart generates a line chart PNG showing total age and per-group age over time.
func (b *Bot) renderHistoryChart(ctx context.Context) ([]byte, error) {
	dbPath, err := storage.DefaultDBPath()
	if err != nil {
		return nil, fmt.Errorf("get db path: %w", err)
	}
	store, err := storage.NewSQLiteStorage(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open storage: %w", err)
	}
	defer store.Close()

	cutoff := time.Now().AddDate(0, 0, -90)
	snapshots, err := store.GetHistory(ctx, cutoff, time.Now())
	if err != nil {
		return nil, fmt.Errorf("query history: %w", err)
	}
	if len(snapshots) == 0 {
		return nil, fmt.Errorf("no history data available")
	}

	// Group by day, keep latest snapshot per day
	type dayData struct {
		date     string
		snapshot storage.StatsSnapshot
	}
	dailyMap := make(map[string]storage.StatsSnapshot)
	for _, s := range snapshots {
		key := s.Timestamp.Format("2006-01-02")
		existing, ok := dailyMap[key]
		if !ok || s.Timestamp.After(existing.Timestamp) {
			dailyMap[key] = s
		}
	}

	// Sort days chronologically
	days := make([]dayData, 0, len(dailyMap))
	for k, v := range dailyMap {
		days = append(days, dayData{date: k, snapshot: v})
	}
	slices.SortFunc(days, func(a, b dayData) int { return cmp.Compare(a.date, b.date) })

	if len(days) < 2 {
		return nil, fmt.Errorf("need at least 2 data points for history chart")
	}

	// Collect all group names across all snapshots
	groupSet := make(map[string]bool)
	for _, d := range days {
		for _, la := range d.snapshot.ListAges.Ages {
			groupSet[la.Title] = true
		}
	}
	groups := slices.Sorted(maps.Keys(groupSet))

	// Build series: first is "Total", then one per group
	labels := make([]string, len(days))
	totalSeries := make([]float64, len(days))
	groupSeries := make([][]float64, len(groups))
	for i := range groups {
		groupSeries[i] = make([]float64, len(days))
	}

	for i, d := range days {
		labels[i] = d.date[5:] // "MM-DD"
		totalSeries[i] = float64(d.snapshot.GlobalStats.TotalAge)

		ageByTitle := make(map[string]int)
		for _, la := range d.snapshot.ListAges.Ages {
			ageByTitle[la.Title] = la.Age
		}
		for gi, g := range groups {
			groupSeries[gi][i] = float64(ageByTitle[g])
		}
	}

	// Build data matrix: total + groups
	seriesNames := make([]string, 0, 1+len(groups))
	seriesNames = append(seriesNames, "Total")
	allData := make([][]float64, 0, 1+len(groups))
	allData = append(allData, totalSeries)
	for i, g := range groups {
		seriesNames = append(seriesNames, stripEmojis(g))
		allData = append(allData, groupSeries[i])
	}

	opt := charts.NewLineChartOptionWithData(allData)
	opt.Legend = charts.LegendOption{
		SeriesNames: seriesNames,
	}
	opt.XAxis = charts.XAxisOption{
		Labels: labels,
	}

	p := charts.NewPainter(charts.PainterOptions{
		Width:  800,
		Height: 400,
	}, charts.PainterThemeOption(charts.GetTheme(charts.ThemeGrafana)))

	if err := p.LineChart(opt); err != nil {
		return nil, fmt.Errorf("render line chart: %w", err)
	}

	buf, err := p.Bytes()
	if err != nil {
		return nil, fmt.Errorf("encode chart: %w", err)
	}

	return buf, nil
}

// stripEmojis removes emoji characters from a string, keeping only letters, digits, punctuation, and spaces.
func stripEmojis(s string) string {
	return strings.Map(func(r rune) rune {
		if r <= unicode.MaxASCII || unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsPunct(r) {
			return r
		}
		return -1
	}, strings.TrimSpace(s))
}

func (b *Bot) authorizeChat(update *models.Update) bool {
	if update.Message == nil {
		return false
	}
	return update.Message.Chat.ID == b.config.ChatID
}

func (b *Bot) sendReply(ctx context.Context, tg *bot.Bot, update *models.Update, text string) {
	_, err := tg.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      text,
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		b.logger.Error("failed to send reply", slog.Any("error", err))
	}
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
