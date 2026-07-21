package database

import (
	"database/sql"
	"shared/golog"
	"strings"
	"sync"

	_ "modernc.org/sqlite"
)

var log = golog.New("SQLITE")
var db *sql.DB

type FlagRecord struct {
	Value     string
	Status    int
	TeamId    int
	Submitter string
	Timestamp int64
}

type Manager struct {
	mu sync.Mutex
}

func (*Manager) LoadDatabase() {
	var err error
	db, err = sql.Open("sqlite", "database.db")

	if err != nil {
		log.Errorf("Failed to open database: %s", err.Error())
		return
	}

	if err := db.Ping(); err != nil {
		log.Errorf("Failed to ping database: %s", err.Error())
		return
	}

	log.Info("Connected to SQLITE database")
}

func (*Manager) CreateTable() {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS flags(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			value TEXT NOT NULL,
			submitter TEXT NOT NULL,
			team_id INTEGER NOT NULL,
			status INT NOT NULL,
			timestamp LONG NOT NULL
		);
	`)

	if err != nil {
		log.Errorf("Failed to create table: %s", err.Error())
		return
	}
}

func (m *Manager) InsertAllFlags(flags []FlagRecord) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(flags) == 0 {
		return
	}

	var (
		placeholders []string
		args         []any
	)

	for _, f := range flags {
		if f.Value == "" {
			continue
		}
		placeholders = append(placeholders, "(?, ?, ?, ?, ?)")
		args = append(args,
			f.Value,
			f.Submitter,
			f.TeamId,
			f.Status,
			f.Timestamp,
		)
	}

	query := `INSERT INTO flags(value, submitter, team_id, status, timestamp) VALUES ` + strings.Join(placeholders, ",")

	_, err := db.Exec(query, args...)

	if err != nil {
		log.Errorf("Failed to insert flags: %s", err.Error())
		return
	}

	log.Info("Saved %d flags", len(flags))
}

func (m *Manager) LoadAllFlags() ([]FlagRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	rows, err := db.Query("SELECT value, submitter, team_id, status, timestamp FROM flags ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []FlagRecord
	for rows.Next() {
		var r FlagRecord
		if err := rows.Scan(&r.Value, &r.Submitter, &r.TeamId, &r.Status, &r.Timestamp); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}
