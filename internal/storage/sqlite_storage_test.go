package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/uchr/ToDoInfo/internal/todo"
	"github.com/uchr/ToDoInfo/internal/todometrics"
)

func newTestSQLiteStorage(t *testing.T) *SQLiteStorage {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	s, err := NewSQLiteStorage(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStorage: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestSQLiteStorage_StoreAndGetLatest(t *testing.T) {
	s := newTestSQLiteStorage(t)
	ctx := t.Context()

	snapshot := StatsSnapshot{
		Timestamp: time.Now().Truncate(time.Second),
		GlobalStats: GlobalStats{
			TotalAge:  100,
			TaskCount: 5,
		},
		ListAges: todometrics.ListAges{
			TotalAge: 100,
			Ages: []todometrics.ListAge{
				{Title: "Work", Age: 60, TaskCount: 3},
				{Title: "Personal", Age: 40, TaskCount: 2},
			},
		},
	}

	if err := s.Store(ctx, snapshot); err != nil {
		t.Fatalf("Store: %v", err)
	}

	latest, err := s.GetLatest(ctx)
	if err != nil {
		t.Fatalf("GetLatest: %v", err)
	}
	if latest == nil {
		t.Fatal("GetLatest returned nil")
	}
	if latest.GlobalStats.TotalAge != 100 {
		t.Errorf("TotalAge = %d, want 100", latest.GlobalStats.TotalAge)
	}
	if latest.GlobalStats.TaskCount != 5 {
		t.Errorf("TaskCount = %d, want 5", latest.GlobalStats.TaskCount)
	}
	if len(latest.ListAges.Ages) != 2 {
		t.Errorf("ListAges count = %d, want 2", len(latest.ListAges.Ages))
	}
}

func TestSQLiteStorage_GetLatestEmpty(t *testing.T) {
	s := newTestSQLiteStorage(t)
	ctx := t.Context()

	latest, err := s.GetLatest(ctx)
	if err != nil {
		t.Fatalf("GetLatest: %v", err)
	}
	if latest != nil {
		t.Errorf("expected nil for empty DB, got %+v", latest)
	}
}

func TestSQLiteStorage_GetLatestReturnsNewest(t *testing.T) {
	s := newTestSQLiteStorage(t)
	ctx := t.Context()

	old := StatsSnapshot{
		Timestamp:   time.Now().Add(-24 * time.Hour).Truncate(time.Second),
		GlobalStats: GlobalStats{TotalAge: 50, TaskCount: 2},
		ListAges:    todometrics.ListAges{TotalAge: 50},
	}
	newer := StatsSnapshot{
		Timestamp:   time.Now().Truncate(time.Second),
		GlobalStats: GlobalStats{TotalAge: 200, TaskCount: 10},
		ListAges:    todometrics.ListAges{TotalAge: 200},
	}

	s.Store(ctx, old)
	s.Store(ctx, newer)

	latest, _ := s.GetLatest(ctx)
	if latest.GlobalStats.TotalAge != 200 {
		t.Errorf("TotalAge = %d, want 200", latest.GlobalStats.TotalAge)
	}
}

func TestSQLiteStorage_GetHistory(t *testing.T) {
	s := newTestSQLiteStorage(t)
	ctx := t.Context()

	now := time.Now().Truncate(time.Second)
	for i := 0; i < 5; i++ {
		snap := StatsSnapshot{
			Timestamp:   now.Add(time.Duration(-i*24) * time.Hour),
			GlobalStats: GlobalStats{TotalAge: (i + 1) * 10, TaskCount: i + 1},
			ListAges:    todometrics.ListAges{TotalAge: (i + 1) * 10},
		}
		s.Store(ctx, snap)
	}

	// Query last 3 days (should get 3 snapshots: day 0, 1, 2)
	from := now.Add(-3 * 24 * time.Hour)
	to := now.Add(time.Hour)
	history, err := s.GetHistory(ctx, from, to)
	if err != nil {
		t.Fatalf("GetHistory: %v", err)
	}
	if len(history) != 3 {
		t.Errorf("history count = %d, want 3", len(history))
	}
}

func TestSQLiteStorage_GetTimeSeriesData(t *testing.T) {
	s := newTestSQLiteStorage(t)
	ctx := t.Context()

	// Pin to midday UTC so now and now-1h stay on the same UTC calendar day.
	// The query buckets by date(timestamp), which is UTC, so a test running
	// within an hour of UTC midnight would otherwise split today's pair.
	now := time.Now().UTC().Truncate(24 * time.Hour).Add(12 * time.Hour)

	// Two snapshots on the same day with different values
	s.Store(ctx, StatsSnapshot{
		Timestamp:   now.Add(-1 * time.Hour),
		GlobalStats: GlobalStats{TotalAge: 50, TaskCount: 3},
		ListAges:    todometrics.ListAges{TotalAge: 50},
	})
	s.Store(ctx, StatsSnapshot{
		Timestamp:   now,
		GlobalStats: GlobalStats{TotalAge: 100, TaskCount: 5},
		ListAges:    todometrics.ListAges{TotalAge: 100},
	})

	// One snapshot yesterday
	s.Store(ctx, StatsSnapshot{
		Timestamp:   now.Add(-24 * time.Hour),
		GlobalStats: GlobalStats{TotalAge: 30, TaskCount: 2},
		ListAges:    todometrics.ListAges{TotalAge: 30},
	})

	points, err := s.GetTimeSeriesData(ctx, 7)
	if err != nil {
		t.Fatalf("GetTimeSeriesData: %v", err)
	}
	if len(points) != 2 {
		t.Fatalf("points count = %d, want 2", len(points))
	}

	// First point is yesterday
	if points[0].MaxAge != 30 {
		t.Errorf("yesterday MaxAge = %d, want 30", points[0].MaxAge)
	}
	// Second point is today — should be MAX(50, 100) = 100
	if points[1].MaxAge != 100 {
		t.Errorf("today MaxAge = %d, want 100", points[1].MaxAge)
	}
	if points[1].TaskCount != 5 {
		t.Errorf("today TaskCount = %d, want 5", points[1].TaskCount)
	}
}

func TestSQLiteStorage_TaskListsRoundTrip(t *testing.T) {
	s := newTestSQLiteStorage(t)
	ctx := t.Context()

	taskLists := []todo.TaskList{
		{Name: "Work", Tasks: []todo.Task{{Title: "Fix bug"}}},
		{Name: "Personal", Tasks: []todo.Task{{Title: "Buy milk"}}},
	}

	snap := StatsSnapshot{
		Timestamp:   time.Now().Truncate(time.Second),
		GlobalStats: GlobalStats{TotalAge: 10, TaskCount: 2},
		ListAges:    todometrics.ListAges{TotalAge: 10},
		TaskLists:   taskLists,
	}

	s.Store(ctx, snap)

	latest, _ := s.GetLatest(ctx)
	if len(latest.TaskLists) != 2 {
		t.Fatalf("TaskLists count = %d, want 2", len(latest.TaskLists))
	}
	if latest.TaskLists[0].Name != "Work" {
		t.Errorf("TaskLists[0].Name = %q, want %q", latest.TaskLists[0].Name, "Work")
	}
	if latest.TaskLists[1].Tasks[0].Title != "Buy milk" {
		t.Errorf("TaskLists[1].Tasks[0].Title = %q, want %q", latest.TaskLists[1].Tasks[0].Title, "Buy milk")
	}
}

func TestSQLiteStorage_MigrationOnFreshDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "fresh.db")

	// Verify file doesn't exist
	if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
		t.Fatal("db file should not exist yet")
	}

	s, err := NewSQLiteStorage(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStorage on fresh path: %v", err)
	}
	defer s.Close()

	// Verify we can store and retrieve
	ctx := t.Context()
	snap := StatsSnapshot{
		Timestamp:   time.Now().Truncate(time.Second),
		GlobalStats: GlobalStats{TotalAge: 1, TaskCount: 1},
		ListAges:    todometrics.ListAges{TotalAge: 1},
	}
	if err := s.Store(ctx, snap); err != nil {
		t.Fatalf("Store on fresh DB: %v", err)
	}
	latest, err := s.GetLatest(ctx)
	if err != nil || latest == nil {
		t.Fatalf("GetLatest on fresh DB: err=%v, latest=%v", err, latest)
	}
}
