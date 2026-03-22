package history

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

type DB struct {
	db *sql.DB
}

type TestRun struct {
	ID        string    `json:"id"`
	Package   string    `json:"package"`
	TestName  string    `json:"testName"`
	Status    string    `json:"status"`
	Duration  int64     `json:"duration"`
	Timestamp time.Time `json:"timestamp"`
	Output    string    `json:"output,omitempty"`
}

func New(dbPath string) (*DB, error) {
	if dbPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = "."
		}
		dbPath = filepath.Join(homeDir, ".alout", "history.db")
	}

	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	h := &DB{db: db}
	if err := h.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate: %w", err)
	}

	return h, nil
}

func (h *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS test_runs (
		id TEXT PRIMARY KEY,
		package TEXT NOT NULL,
		test_name TEXT NOT NULL,
		status TEXT NOT NULL,
		duration_ms INTEGER NOT NULL,
		timestamp DATETIME NOT NULL,
		output TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_test_runs_package ON test_runs(package);
	CREATE INDEX IF NOT EXISTS idx_test_runs_test_name ON test_runs(test_name);
	CREATE INDEX IF NOT EXISTS idx_test_runs_timestamp ON test_runs(timestamp);
	`

	_, err := h.db.Exec(schema)
	return err
}

func (h *DB) SaveTestRun(run *TestRun) error {
	if run.ID == "" {
		run.ID = uuid.New().String()
	}

	query := `
	INSERT INTO test_runs (id, package, test_name, status, duration_ms, timestamp, output)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := h.db.Exec(query,
		run.ID,
		run.Package,
		run.TestName,
		run.Status,
		run.Duration,
		run.Timestamp,
		run.Output,
	)

	return err
}

func (h *DB) GetTestRuns(limit int) ([]TestRun, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
	SELECT id, package, test_name, status, duration_ms, timestamp, output
	FROM test_runs
	ORDER BY timestamp DESC
	LIMIT ?
	`

	rows, err := h.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []TestRun
	for rows.Next() {
		var run TestRun
		var durationMs int64
		var output sql.NullString

		err := rows.Scan(
			&run.ID,
			&run.Package,
			&run.TestName,
			&run.Status,
			&durationMs,
			&run.Timestamp,
			&output,
		)
		if err != nil {
			return nil, err
		}

		run.Duration = durationMs
		if output.Valid {
			run.Output = output.String
		}

		runs = append(runs, run)
	}

	return runs, nil
}

func (h *DB) GetTestRunsByPackage(pkg string, limit int) ([]TestRun, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
	SELECT id, package, test_name, status, duration_ms, timestamp, output
	FROM test_runs
	WHERE package = ?
	ORDER BY timestamp DESC
	LIMIT ?
	`

	rows, err := h.db.Query(query, pkg, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []TestRun
	for rows.Next() {
		var run TestRun
		var durationMs int64
		var output sql.NullString

		err := rows.Scan(
			&run.ID,
			&run.Package,
			&run.TestName,
			&run.Status,
			&durationMs,
			&run.Timestamp,
			&output,
		)
		if err != nil {
			return nil, err
		}

		run.Duration = durationMs
		if output.Valid {
			run.Output = output.String
		}

		runs = append(runs, run)
	}

	return runs, nil
}

func (h *DB) GetLastRunResult(pkg, testName string) (*TestRun, error) {
	query := `
	SELECT id, package, test_name, status, duration_ms, timestamp, output
	FROM test_runs
	WHERE package = ? AND test_name = ?
	ORDER BY timestamp DESC
	LIMIT 1
	`

	var run TestRun
	var durationMs int64
	var output sql.NullString

	err := h.db.QueryRow(query, pkg, testName).Scan(
		&run.ID,
		&run.Package,
		&run.TestName,
		&run.Status,
		&durationMs,
		&run.Timestamp,
		&output,
	)
	if err != nil {
		return nil, err
	}

	run.Duration = durationMs
	if output.Valid {
		run.Output = output.String
	}

	return &run, nil
}

func (h *DB) DeleteOldRuns(keepDays int) error {
	if keepDays <= 0 {
		keepDays = 30
	}

	cutoff := time.Now().AddDate(0, 0, -keepDays)

	query := `DELETE FROM test_runs WHERE timestamp < ?`
	_, err := h.db.Exec(query, cutoff)

	return err
}

func (h *DB) Close() error {
	return h.db.Close()
}
