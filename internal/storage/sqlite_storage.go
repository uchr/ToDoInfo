package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/uchr/ToDoInfo/internal/todo"
	"github.com/uchr/ToDoInfo/internal/todometrics"

	_ "modernc.org/sqlite"
)

// SQLiteStorage implements StatsStorage using a SQLite database.
type SQLiteStorage struct {
	db *sql.DB
}

// NewSQLiteStorage opens (or creates) a SQLite database at dbPath and runs migrations.
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}

	// Enable WAL mode for better concurrent read performance.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}

	s := &SQLiteStorage{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return s, nil
}

func (s *SQLiteStorage) migrate() error {
	const ddl = `
CREATE TABLE IF NOT EXISTS snapshots (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp       DATETIME NOT NULL,
    total_age       INTEGER NOT NULL,
    task_count      INTEGER NOT NULL,
    task_lists_json TEXT
);

CREATE TABLE IF NOT EXISTS list_ages (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    snapshot_id INTEGER NOT NULL REFERENCES snapshots(id),
    title       TEXT NOT NULL,
    age         INTEGER NOT NULL,
    task_count  INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_snapshots_timestamp ON snapshots(timestamp);
CREATE INDEX IF NOT EXISTS idx_list_ages_snapshot   ON list_ages(snapshot_id);
`
	_, err := s.db.Exec(ddl)
	return err
}

// Store saves a statistics snapshot.
func (s *SQLiteStorage) Store(ctx context.Context, snapshot StatsSnapshot) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Serialize task lists to JSON.
	var taskListsJSON []byte
	if len(snapshot.TaskLists) > 0 {
		taskListsJSON, err = json.Marshal(snapshot.TaskLists)
		if err != nil {
			return fmt.Errorf("marshal task lists: %w", err)
		}
	}

	res, err := tx.ExecContext(ctx,
		`INSERT INTO snapshots (timestamp, total_age, task_count, task_lists_json)
		 VALUES (?, ?, ?, ?)`,
		snapshot.Timestamp.UTC().Format(time.RFC3339),
		snapshot.GlobalStats.TotalAge,
		snapshot.GlobalStats.TaskCount,
		taskListsJSON,
	)
	if err != nil {
		return fmt.Errorf("insert snapshot: %w", err)
	}

	snapshotID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}

	// Insert list ages.
	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO list_ages (snapshot_id, title, age, task_count) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare list_ages insert: %w", err)
	}
	defer stmt.Close()

	for _, la := range snapshot.ListAges.Ages {
		if _, err := stmt.ExecContext(ctx, snapshotID, la.Title, la.Age, la.TaskCount); err != nil {
			return fmt.Errorf("insert list age: %w", err)
		}
	}

	return tx.Commit()
}

// GetLatest retrieves the most recent statistics snapshot.
func (s *SQLiteStorage) GetLatest(ctx context.Context) (*StatsSnapshot, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, timestamp, total_age, task_count, task_lists_json
		 FROM snapshots ORDER BY timestamp DESC LIMIT 1`)

	snap, err := s.scanSnapshot(ctx, row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return snap, nil
}

// GetHistory retrieves statistics history for a given time period.
func (s *SQLiteStorage) GetHistory(ctx context.Context, from, to time.Time) ([]StatsSnapshot, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, timestamp, total_age, task_count, task_lists_json
		 FROM snapshots
		 WHERE timestamp > ? AND timestamp < ?
		 ORDER BY timestamp ASC`,
		from.UTC().Format(time.RFC3339),
		to.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return nil, fmt.Errorf("query history: %w", err)
	}
	defer rows.Close()

	var snapshots []StatsSnapshot
	for rows.Next() {
		snap, err := s.scanSnapshotFromRows(ctx, rows)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, *snap)
	}
	return snapshots, rows.Err()
}

// GetTimeSeriesData retrieves time series data for graphing.
func (s *SQLiteStorage) GetTimeSeriesData(ctx context.Context, days int) ([]TimeSeriesPoint, error) {
	cutoff := time.Now().AddDate(0, 0, -days).UTC().Format(time.RFC3339)

	rows, err := s.db.QueryContext(ctx,
		`SELECT date(timestamp) AS d, MAX(total_age), MAX(task_count)
		 FROM snapshots
		 WHERE timestamp > ?
		 GROUP BY d
		 ORDER BY d ASC`,
		cutoff,
	)
	if err != nil {
		return nil, fmt.Errorf("query time series: %w", err)
	}
	defer rows.Close()

	var points []TimeSeriesPoint
	for rows.Next() {
		var dateStr string
		var maxAge, taskCount int
		if err := rows.Scan(&dateStr, &maxAge, &taskCount); err != nil {
			return nil, fmt.Errorf("scan time series row: %w", err)
		}
		d, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("parse date %q: %w", dateStr, err)
		}
		points = append(points, TimeSeriesPoint{
			Date:      d,
			MaxAge:    maxAge,
			TaskCount: taskCount,
		})
	}
	return points, rows.Err()
}

// Close releases the database connection.
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

// scanSnapshot scans a single snapshot row (from QueryRow).
func (s *SQLiteStorage) scanSnapshot(ctx context.Context, row *sql.Row) (*StatsSnapshot, error) {
	var (
		id           int64
		tsStr        string
		totalAge     int
		taskCount    int
		taskListJSON sql.NullString
	)
	if err := row.Scan(&id, &tsStr, &totalAge, &taskCount, &taskListJSON); err != nil {
		return nil, err
	}
	return s.buildSnapshot(ctx, id, tsStr, totalAge, taskCount, taskListJSON)
}

// scanSnapshotFromRows scans a single snapshot from an active Rows cursor.
func (s *SQLiteStorage) scanSnapshotFromRows(ctx context.Context, rows *sql.Rows) (*StatsSnapshot, error) {
	var (
		id           int64
		tsStr        string
		totalAge     int
		taskCount    int
		taskListJSON sql.NullString
	)
	if err := rows.Scan(&id, &tsStr, &totalAge, &taskCount, &taskListJSON); err != nil {
		return nil, fmt.Errorf("scan snapshot row: %w", err)
	}
	return s.buildSnapshot(ctx, id, tsStr, totalAge, taskCount, taskListJSON)
}

func (s *SQLiteStorage) buildSnapshot(ctx context.Context, id int64, tsStr string, totalAge, taskCount int, taskListJSON sql.NullString) (*StatsSnapshot, error) {
	ts, err := time.Parse(time.RFC3339, tsStr)
	if err != nil {
		return nil, fmt.Errorf("parse timestamp %q: %w", tsStr, err)
	}

	// Load list ages for this snapshot.
	laRows, err := s.db.QueryContext(ctx,
		`SELECT title, age, task_count FROM list_ages WHERE snapshot_id = ?`, id)
	if err != nil {
		return nil, fmt.Errorf("query list ages: %w", err)
	}
	defer laRows.Close()

	var listAges todometrics.ListAges
	for laRows.Next() {
		var la todometrics.ListAge
		if err := laRows.Scan(&la.Title, &la.Age, &la.TaskCount); err != nil {
			return nil, fmt.Errorf("scan list age: %w", err)
		}
		listAges.Ages = append(listAges.Ages, la)
	}
	if err := laRows.Err(); err != nil {
		return nil, err
	}
	listAges.TotalAge = totalAge

	// Deserialize task lists.
	var taskLists []todo.TaskList
	if taskListJSON.Valid && taskListJSON.String != "" {
		if err := json.Unmarshal([]byte(taskListJSON.String), &taskLists); err != nil {
			return nil, fmt.Errorf("unmarshal task lists: %w", err)
		}
	}

	return &StatsSnapshot{
		Timestamp: ts,
		GlobalStats: GlobalStats{
			TotalAge:  totalAge,
			TaskCount: taskCount,
		},
		ListAges:  listAges,
		TaskLists: taskLists,
	}, nil
}
