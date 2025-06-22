package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/uchr/ToDoInfo/internal/todometrics"
)

func TestJSONStorage(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "todoinfo_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create storage
	storage, err := NewJSONStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	ctx := context.Background()

	// Create test snapshot
	snapshot := StatsSnapshot{
		Timestamp: time.Now(),
		GlobalStats: GlobalStats{
			TotalAge:  100,
			TaskCount: 5,
		},
		ListAges: todometrics.ListAges{
			TotalAge: 100,
			Ages: []todometrics.ListAge{
				{Title: "Work", Age: 60},
				{Title: "Personal", Age: 40},
			},
		},
		MaxAge:    30,
		TaskCount: 5,
	}

	// Test Store
	err = storage.Store(ctx, snapshot)
	if err != nil {
		t.Fatalf("Failed to store snapshot: %v", err)
	}

	// Test GetLatest
	latest, err := storage.GetLatest(ctx)
	if err != nil {
		t.Fatalf("Failed to get latest snapshot: %v", err)
	}

	if latest == nil {
		t.Fatal("Latest snapshot is nil")
	}

	if latest.GlobalStats.TotalAge != 100 {
		t.Errorf("Expected TotalAge 100, got %d", latest.GlobalStats.TotalAge)
	}

	if latest.MaxAge != 30 {
		t.Errorf("Expected MaxAge 30, got %d", latest.MaxAge)
	}

	// Test GetTimeSeriesData
	points, err := storage.GetTimeSeriesData(ctx, 7)
	if err != nil {
		t.Fatalf("Failed to get time series data: %v", err)
	}

	if len(points) != 1 {
		t.Errorf("Expected 1 data point, got %d", len(points))
	}

	if len(points) > 0 && points[0].MaxAge != 30 {
		t.Errorf("Expected MaxAge 30 in time series, got %d", points[0].MaxAge)
	}
}