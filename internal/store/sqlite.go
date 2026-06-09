package store

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"scoreador/internal/sim"
)

type SingleMatchRun struct {
	ID          int64
	CreatedAt   time.Time
	Seed        int64
	LambdaPath  string
	TeamA       string
	TeamB       string
	ShotsA      int
	ShotsB      int
	MotivationA string
	MotivationB string
	Tiebreaker  string
	Simulations int
	WinsA       int
	WinsB       int
	GoalsA      int
	GoalsB      int
	Regulation  int
	Penalties   int
	RandomTie   int
}

func (r SingleMatchRun) DisplayCreatedAt() string {
	if r.CreatedAt.IsZero() {
		return ""
	}
	return r.CreatedAt.Local().Format("2006-01-02 15:04:05")
}

type SQLiteStore struct {
	db *sql.DB
}

func OpenSQLite(path string) (*SQLiteStore, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		path = filepath.Join("out", "single_matches.db")
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	store := &SQLiteStore{db: db}
	if err := store.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *SQLiteStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *SQLiteStore) migrate() error {
	if s == nil || s.db == nil {
		return errors.New("sqlite store no inicializado")
	}
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS single_match_runs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	created_at TEXT NOT NULL,
	seed INTEGER NOT NULL,
	lambda_path TEXT NOT NULL,
	team_a TEXT NOT NULL,
	team_b TEXT NOT NULL,
	shots_a INTEGER NOT NULL,
	shots_b INTEGER NOT NULL,
	motivation_a TEXT NOT NULL,
	motivation_b TEXT NOT NULL,
	tiebreaker TEXT NOT NULL,
	simulations INTEGER NOT NULL,
	wins_a INTEGER NOT NULL,
	wins_b INTEGER NOT NULL,
	goals_a INTEGER NOT NULL,
	goals_b INTEGER NOT NULL,
	regulation INTEGER NOT NULL,
	penalties INTEGER NOT NULL,
	random_tie INTEGER NOT NULL
);`)
	return err
}

func (s *SQLiteStore) SaveSingleMatchRun(seed int64, lambdaPath string, input sim.SingleMatchInput, summary sim.SingleMatchSeries) (SingleMatchRun, error) {
	if s == nil || s.db == nil {
		return SingleMatchRun{}, errors.New("sqlite store no inicializado")
	}
	now := time.Now().UTC()
	record := SingleMatchRun{
		CreatedAt:   now,
		Seed:        seed,
		LambdaPath:  strings.TrimSpace(lambdaPath),
		TeamA:       strings.TrimSpace(input.TeamA),
		TeamB:       strings.TrimSpace(input.TeamB),
		ShotsA:      input.ShotsA,
		ShotsB:      input.ShotsB,
		MotivationA: input.MotivationA.String(),
		MotivationB: input.MotivationB.String(),
		Tiebreaker:  strings.TrimSpace(input.Tiebreaker),
		Simulations: summary.Simulations,
		WinsA:       summary.WinsA,
		WinsB:       summary.WinsB,
		GoalsA:      summary.GoalsA,
		GoalsB:      summary.GoalsB,
		Regulation:  summary.Regulation,
		Penalties:   summary.Penalties,
		RandomTie:   summary.RandomTie,
	}

	res, err := s.db.Exec(`
INSERT INTO single_match_runs (
	created_at, seed, lambda_path, team_a, team_b, shots_a, shots_b, motivation_a, motivation_b, tiebreaker,
	simulations, wins_a, wins_b, goals_a, goals_b, regulation, penalties, random_tie
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		now.Format(time.RFC3339Nano),
		record.Seed,
		record.LambdaPath,
		record.TeamA,
		record.TeamB,
		record.ShotsA,
		record.ShotsB,
		record.MotivationA,
		record.MotivationB,
		record.Tiebreaker,
		record.Simulations,
		record.WinsA,
		record.WinsB,
		record.GoalsA,
		record.GoalsB,
		record.Regulation,
		record.Penalties,
		record.RandomTie,
	)
	if err != nil {
		return SingleMatchRun{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return SingleMatchRun{}, err
	}
	record.ID = id
	return record, nil
}

func (s *SQLiteStore) ListSingleMatchRuns(limit int) ([]SingleMatchRun, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("sqlite store no inicializado")
	}
	if limit <= 0 {
		limit = 10
	}

	rows, err := s.db.Query(`
SELECT id, created_at, seed, lambda_path, team_a, team_b, shots_a, shots_b, motivation_a, motivation_b, tiebreaker,
       simulations, wins_a, wins_b, goals_a, goals_b, regulation, penalties, random_tie
FROM single_match_runs
ORDER BY id DESC
LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	runs := make([]SingleMatchRun, 0, limit)
	for rows.Next() {
		var (
			record       SingleMatchRun
			createdAtRaw string
		)
		if err := rows.Scan(
			&record.ID,
			&createdAtRaw,
			&record.Seed,
			&record.LambdaPath,
			&record.TeamA,
			&record.TeamB,
			&record.ShotsA,
			&record.ShotsB,
			&record.MotivationA,
			&record.MotivationB,
			&record.Tiebreaker,
			&record.Simulations,
			&record.WinsA,
			&record.WinsB,
			&record.GoalsA,
			&record.GoalsB,
			&record.Regulation,
			&record.Penalties,
			&record.RandomTie,
		); err != nil {
			return nil, err
		}
		parsed, err := time.Parse(time.RFC3339Nano, createdAtRaw)
		if err != nil {
			return nil, fmt.Errorf("parse created_at: %w", err)
		}
		record.CreatedAt = parsed
		runs = append(runs, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return runs, nil
}
