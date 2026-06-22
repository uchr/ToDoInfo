package bot

import (
	"testing"
	"time"

	"github.com/uchr/ToDoInfo/internal/storage"
)

func snap(ts time.Time, totalAge int) storage.StatsSnapshot {
	return storage.StatsSnapshot{
		Timestamp:   ts,
		GlobalStats: storage.GlobalStats{TotalAge: totalAge, TaskCount: 1},
	}
}

// Regression: a day with several snapshots must be represented by its latest
// snapshot, not the one with the lowest total age. Otherwise the chart ignores
// the data a /chart or /summary request just fetched.
func TestCollapseHistoryByDay_KeepsLatestPerDay(t *testing.T) {
	day := time.Date(2026, 6, 22, 0, 0, 0, 0, time.UTC)
	earlyLowAge := snap(day.Add(8*time.Hour), 100)  // earlier, lower total age
	lateHighAge := snap(day.Add(20*time.Hour), 150) // latest snapshot of the day

	// Order shuffled to confirm selection is by timestamp, not input order.
	got := collapseHistoryByDay([]storage.StatsSnapshot{lateHighAge, earlyLowAge})

	if len(got) != 1 {
		t.Fatalf("expected 1 day, got %d", len(got))
	}
	if got[0].snapshot.GlobalStats.TotalAge != 150 {
		t.Fatalf("expected latest snapshot (totalAge 150), got %d",
			got[0].snapshot.GlobalStats.TotalAge)
	}
}

func TestCollapseHistoryByDay_SortsAscendingByDate(t *testing.T) {
	d1 := time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)
	d2 := time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC)
	d3 := time.Date(2026, 6, 22, 12, 0, 0, 0, time.UTC)

	got := collapseHistoryByDay([]storage.StatsSnapshot{snap(d3, 3), snap(d1, 1), snap(d2, 2)})

	want := []string{"2026-06-20", "2026-06-21", "2026-06-22"}
	if len(got) != len(want) {
		t.Fatalf("expected %d days, got %d", len(want), len(got))
	}
	for i, w := range want {
		if got[i].date != w {
			t.Fatalf("day %d: expected %s, got %s", i, w, got[i].date)
		}
	}
}
