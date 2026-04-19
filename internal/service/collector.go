package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/uchr/ToDoInfo/internal/auth"
	"github.com/uchr/ToDoInfo/internal/storage"
	"github.com/uchr/ToDoInfo/internal/todoclient"
	"github.com/uchr/ToDoInfo/internal/todometrics"
)

// StatsData holds the latest computed statistics.
type StatsData struct {
	FetchedAt   time.Time
	TotalTasks  int
	TotalAge    int
	ListAges    todometrics.ListAges
	SortedTasks []todometrics.TaskRottennessInfo
	Champion    *todometrics.TaskRottennessInfo
	TimeSeries  []storage.TimeSeriesPoint
}

// Collector periodically fetches tasks from Microsoft Graph and caches stats.
type Collector struct {
	authClient *auth.AuthClient
	logger     *slog.Logger
	interval   time.Duration

	mu             sync.RWMutex
	cached         *StatsData
	lastRefreshAt  time.Time
	lastRefreshErr error
}

// NewCollector creates a new Collector.
func NewCollector(authClient *auth.AuthClient, logger *slog.Logger, interval time.Duration) *Collector {
	return &Collector{
		authClient: authClient,
		logger:     logger,
		interval:   interval,
	}
}

// Run performs an immediate fetch, then refreshes on a ticker until ctx is cancelled.
// Skips refresh attempts when the auth client is not yet authenticated.
func (c *Collector) Run(ctx context.Context) error {
	if c.authClient.HasCredential() {
		if err := c.Refresh(ctx); err != nil {
			c.logger.Error("initial refresh failed", slog.Any("error", err))
		}
	} else {
		c.logger.Info("skipping initial refresh — not yet authenticated")
	}

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if !c.authClient.HasCredential() {
				c.logger.Debug("skipping refresh — not yet authenticated")
				continue
			}
			if err := c.Refresh(ctx); err != nil {
				c.logger.Error("periodic refresh failed", slog.Any("error", err))
			}
		}
	}
}

// Refresh fetches tasks, computes metrics, stores a snapshot, and updates the cache.
func (c *Collector) Refresh(ctx context.Context) (retErr error) {
	c.logger.Info("refreshing task data")

	var fresh *StatsData
	defer func() {
		c.mu.Lock()
		c.lastRefreshAt = time.Now()
		c.lastRefreshErr = retErr
		if fresh != nil {
			c.cached = fresh
		}
		c.mu.Unlock()
	}()

	token, err := c.authClient.GetAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("get access token: %w", err)
	}

	parser := todoclient.New()
	taskLists, err := parser.GetTasks(ctx, c.logger, token)
	if err != nil {
		return fmt.Errorf("get tasks: %w", err)
	}

	metrics := todometrics.New(taskLists)
	sortedTasks := metrics.GetSortedTasks()
	listAges := metrics.GetListAges()

	totalAge := 0
	for _, t := range sortedTasks {
		totalAge += t.Age
	}

	var champion *todometrics.TaskRottennessInfo
	if top := metrics.GetTopTasksByAge(1); len(top) > 0 {
		champion = &top[0]
	}

	// Store snapshot
	dbPath, err := storage.DefaultDBPath()
	if err != nil {
		return fmt.Errorf("get db path: %w", err)
	}
	store, err := storage.NewSQLiteStorage(dbPath)
	if err != nil {
		return fmt.Errorf("create storage: %w", err)
	}
	defer store.Close()

	snapshot := storage.StatsSnapshot{
		Timestamp: time.Now(),
		GlobalStats: storage.GlobalStats{
			TotalAge:  totalAge,
			TaskCount: len(sortedTasks),
		},
		ListAges:  listAges,
		TaskLists: taskLists,
	}
	if err := store.Store(ctx, snapshot); err != nil {
		c.logger.Warn("failed to store snapshot", slog.Any("error", err))
	}

	// Fetch time-series for chart data
	timeSeries, err := store.GetTimeSeriesData(ctx, 90)
	if err != nil {
		c.logger.Warn("failed to load time series", slog.Any("error", err))
	}

	fresh = &StatsData{
		FetchedAt:   time.Now(),
		TotalTasks:  len(sortedTasks),
		TotalAge:    totalAge,
		ListAges:    listAges,
		SortedTasks: sortedTasks,
		Champion:    champion,
		TimeSeries:  timeSeries,
	}

	c.logger.Info("refresh complete", slog.Int("tasks", len(sortedTasks)), slog.Int("totalAge", totalAge))
	return nil
}

// EnsureFresh runs Refresh if the last successful refresh is older than maxAge,
// otherwise returns nil immediately. Errors from Refresh propagate to the caller.
func (c *Collector) EnsureFresh(ctx context.Context, maxAge time.Duration) error {
	c.mu.RLock()
	age := time.Since(c.lastRefreshAt)
	hadSuccess := !c.lastRefreshAt.IsZero() && c.lastRefreshErr == nil
	c.mu.RUnlock()

	if hadSuccess && age < maxAge {
		return nil
	}
	return c.Refresh(ctx)
}

// LastRefreshErr returns the error from the most recent Refresh attempt, or nil if it succeeded.
func (c *Collector) LastRefreshErr() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastRefreshErr
}

// GetLatest returns the most recently cached stats. May be nil if no fetch has succeeded yet.
func (c *Collector) GetLatest() *StatsData {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cached
}
