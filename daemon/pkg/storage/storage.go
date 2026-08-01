package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"tamagops/daemon/pkg/engine"
)

type Storage struct {
	db *sql.DB
}

func Open(path string) (*Storage, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("storage: open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("storage: connect to database: %w", err)
	}

	s := &Storage{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("storage: run migrations: %w", err)
	}
	return s, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) migrate() error {
	const schema = `
	CREATE TABLE IF NOT EXISTS pet_profile (
		id                    INTEGER PRIMARY KEY CHECK (id = 1),
		name                  TEXT NOT NULL,
		level                 INTEGER NOT NULL,
		current_xp            INTEGER NOT NULL,
		next_level_xp         INTEGER NOT NULL,
		hp                    INTEGER NOT NULL,
		max_hp                INTEGER NOT NULL,
		mood                  TEXT NOT NULL,
		consecutive_high_cpu  INTEGER NOT NULL DEFAULT 0,
		consecutive_idle_cpu  INTEGER NOT NULL DEFAULT 0,
		updated_at            TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS system_logs (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		message    TEXT NOT NULL,
		mood       TEXT NOT NULL,
		created_at TEXT NOT NULL
	);
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *Storage) SavePetState(state engine.PetState) error {
	const query = `
	INSERT INTO pet_profile (
		id, name, level, current_xp, next_level_xp, hp, max_hp, mood,
		consecutive_high_cpu, consecutive_idle_cpu, updated_at
	) VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		name = excluded.name,
		level = excluded.level,
		current_xp = excluded.current_xp,
		next_level_xp = excluded.next_level_xp,
		hp = excluded.hp,
		max_hp = excluded.max_hp,
		mood = excluded.mood,
		consecutive_high_cpu = excluded.consecutive_high_cpu,
		consecutive_idle_cpu = excluded.consecutive_idle_cpu,
		updated_at = excluded.updated_at;
	`
	_, err := s.db.Exec(query,
		state.Name, state.Level, state.CurrentXP, state.NextLevelXP,
		state.HP, state.MaxHP, state.Mood,
		state.ConsecutiveHighCPU, state.ConsecutiveIdleCPU,
		time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("storage: save pet state: %w", err)
	}
	return nil
}

func (s *Storage) LoadPetState() (*engine.PetState, error) {
	const query = `
	SELECT name, level, current_xp, next_level_xp, hp, max_hp, mood,
	       consecutive_high_cpu, consecutive_idle_cpu
	FROM pet_profile WHERE id = 1;
	`
	row := s.db.QueryRow(query)

	var state engine.PetState
	err := row.Scan(
		&state.Name, &state.Level, &state.CurrentXP, &state.NextLevelXP,
		&state.HP, &state.MaxHP, &state.Mood,
		&state.ConsecutiveHighCPU, &state.ConsecutiveIdleCPU,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("storage: load pet state: %w", err)
	}
	state.Debuffs = []string{}
	return &state, nil
}

func (s *Storage) LogEvent(message, mood string) error {
	const query = `INSERT INTO system_logs (message, mood, created_at) VALUES (?, ?, ?);`
	_, err := s.db.Exec(query, message, mood, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("storage: log event: %w", err)
	}
	return nil
}

type LogEntry struct {
	Message   string
	Mood      string
	CreatedAt time.Time
}

func (s *Storage) RecentLogs(limit int) ([]LogEntry, error) {
	const query = `
	SELECT message, mood, created_at FROM system_logs
	ORDER BY id DESC LIMIT ?;
	`
	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("storage: fetch logs: %w", err)
	}
	defer rows.Close()

	var entries []LogEntry
	for rows.Next() {
		var e LogEntry
		var createdAt string
		if err := rows.Scan(&e.Message, &e.Mood, &createdAt); err != nil {
			return nil, fmt.Errorf("storage: read log: %w", err)
		}
		e.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}