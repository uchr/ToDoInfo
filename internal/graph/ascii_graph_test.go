package graph

import (
	"testing"
	"time"

	"github.com/uchr/ToDoInfo/internal/storage"
)

func TestASCIIGraph(t *testing.T) {
	// Create test data points
	points := []storage.TimeSeriesPoint{
		{Date: time.Now().AddDate(0, 0, -7), MaxAge: 10},
		{Date: time.Now().AddDate(0, 0, -6), MaxAge: 15},
		{Date: time.Now().AddDate(0, 0, -5), MaxAge: 12},
		{Date: time.Now().AddDate(0, 0, -4), MaxAge: 20},
		{Date: time.Now().AddDate(0, 0, -3), MaxAge: 18},
		{Date: time.Now().AddDate(0, 0, -2), MaxAge: 25},
		{Date: time.Now().AddDate(0, 0, -1), MaxAge: 22},
		{Date: time.Now(), MaxAge: 30},
	}

	// Create graph
	graph := NewASCIIGraph()
	if graph == nil {
		t.Fatal("Failed to create ASCII graph")
	}

	// Test findMinMax
	min, max := graph.findMinMax(points)
	if min != 10 {
		t.Errorf("Expected min age 10, got %d", min)
	}
	if max != 30 {
		t.Errorf("Expected max age 30, got %d", max)
	}

	// Test rendering (this will just ensure it doesn't panic)
	graph.RenderTimeSeriesGraph(points, "Test Graph")
}

func TestASCIIGraphEmptyData(t *testing.T) {
	graph := NewASCIIGraph()
	
	// Test with empty data (should not panic)
	graph.RenderTimeSeriesGraph([]storage.TimeSeriesPoint{}, "Empty Graph")
	
	// Test findMinMax with empty data
	min, max := graph.findMinMax([]storage.TimeSeriesPoint{})
	if min != 0 || max != 0 {
		t.Errorf("Expected min=0, max=0 for empty data, got min=%d, max=%d", min, max)
	}
}