package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// JSONStorage implements StatsStorage using JSON files
type JSONStorage struct {
	dataDir string
}

// NewJSONStorage creates a new JSON-based storage
func NewJSONStorage(dataDir string) (*JSONStorage, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &JSONStorage{dataDir: dataDir}, nil
}

// Store saves a statistics snapshot to a JSON file
func (s *JSONStorage) Store(ctx context.Context, snapshot StatsSnapshot) error {
	filename := fmt.Sprintf("stats_%s.json", snapshot.Timestamp.Format("2006-01-02_15-04-05"))
	filepath := filepath.Join(s.dataDir, filename)

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write snapshot file: %w", err)
	}

	return nil
}

// GetLatest retrieves the most recent statistics snapshot
func (s *JSONStorage) GetLatest(ctx context.Context) (*StatsSnapshot, error) {
	files, err := s.getStatsFiles()
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, nil // No data available
	}

	// Get the most recent file
	latestFile := files[len(files)-1]
	return s.loadSnapshot(latestFile)
}

// GetHistory retrieves statistics history for a given time period
func (s *JSONStorage) GetHistory(ctx context.Context, from, to time.Time) ([]StatsSnapshot, error) {
	files, err := s.getStatsFiles()
	if err != nil {
		return nil, err
	}

	var snapshots []StatsSnapshot
	for _, file := range files {
		snapshot, err := s.loadSnapshot(file)
		if err != nil {
			continue // Skip corrupted files
		}

		if snapshot.Timestamp.After(from) && snapshot.Timestamp.Before(to) {
			snapshots = append(snapshots, *snapshot)
		}
	}

	return snapshots, nil
}

// GetTimeSeriesData retrieves time series data for graphing
func (s *JSONStorage) GetTimeSeriesData(ctx context.Context, days int) ([]TimeSeriesPoint, error) {
	cutoff := time.Now().AddDate(0, 0, -days)
	snapshots, err := s.GetHistory(ctx, cutoff, time.Now())
	if err != nil {
		return nil, err
	}

	maxAges := make(map[string]int)
	for _, snapshot := range snapshots {
		if _, exists := maxAges[snapshot.Timestamp.Format("2006-01-02")]; !exists {
			maxAges[snapshot.Timestamp.Format("2006-01-02")] = snapshot.MaxAge
		} else {
			if snapshot.MaxAge > maxAges[snapshot.Timestamp.Format("2006-01-02")] {
				maxAges[snapshot.Timestamp.Format("2006-01-02")] = snapshot.MaxAge
			}
		}
	}

	var points []TimeSeriesPoint
	for date, maxAge := range maxAges {
		d, err := time.Parse("2006-01-02", date)
		if err != nil {
			return nil, fmt.Errorf("failed to parse date: %w", err)
		}
		points = append(points, TimeSeriesPoint{
			Date:   d,
			MaxAge: maxAge,
		})
	}

	// Sort by date
	sort.Slice(points, func(i, j int) bool {
		return points[i].Date.Before(points[j].Date)
	})

	return points, nil
}

// Close releases any resources held by the storage
func (s *JSONStorage) Close() error {
	// JSON storage doesn't need cleanup
	return nil
}

// getStatsFiles returns a sorted list of stats files
func (s *JSONStorage) getStatsFiles() ([]string, error) {
	files, err := os.ReadDir(s.dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read data directory: %w", err)
	}

	var statsFiles []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			if filepath.HasPrefix(file.Name(), "stats_") {
				statsFiles = append(statsFiles, filepath.Join(s.dataDir, file.Name()))
			}
		}
	}

	// Sort files by name (which includes timestamp)
	sort.Strings(statsFiles)

	return statsFiles, nil
}

// loadSnapshot loads a snapshot from a file
func (s *JSONStorage) loadSnapshot(filename string) (*StatsSnapshot, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshot file: %w", err)
	}

	var snapshot StatsSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot: %w", err)
	}

	return &snapshot, nil
}
