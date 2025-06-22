package storage

import (
	"context"
	"time"

	"github.com/uchr/ToDoInfo/internal/todometrics"
)

// StatsSnapshot represents a point-in-time statistics snapshot
type StatsSnapshot struct {
	Timestamp   time.Time            `json:"timestamp"`
	GlobalStats GlobalStats          `json:"global_stats"`
	ListAges    todometrics.ListAges `json:"list_ages"`
	TaskCount   int                  `json:"task_count"`
}

// GlobalStats represents global statistics
type GlobalStats struct {
	TotalAge  int `json:"total_age"`
	TaskCount int `json:"task_count"`
}

// TimeSeriesPoint represents a data point in time series
type TimeSeriesPoint struct {
	Date      time.Time `json:"date"`
	MaxAge    int       `json:"max_age"`
	TaskCount int       `json:"task_count"`
}

// StatsStorage defines the interface for storing and retrieving statistics
type StatsStorage interface {
	// Store saves a statistics snapshot
	Store(ctx context.Context, snapshot StatsSnapshot) error

	// GetLatest retrieves the most recent statistics snapshot
	GetLatest(ctx context.Context) (*StatsSnapshot, error)

	// GetHistory retrieves statistics history for a given time period
	GetHistory(ctx context.Context, from, to time.Time) ([]StatsSnapshot, error)

	// GetTimeSeriesData retrieves time series data for graphing
	GetTimeSeriesData(ctx context.Context, days int) ([]TimeSeriesPoint, error)

	// Close releases any resources held by the storage
	Close() error
}
