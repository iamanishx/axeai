package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	"axe-desktop/pkg/models"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	store := &Storage{db: db}
	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return store, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) migrate() error {
	migrations := []string{
		`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id),
			title TEXT NOT NULL,
			model TEXT NOT NULL,
			system_prompt TEXT,
			summary TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			archived_at DATETIME,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL REFERENCES sessions(id),
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			status TEXT DEFAULT 'completed',
			token_count INTEGER,
			metadata_json TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS tool_calls (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL REFERENCES sessions(id),
			message_id TEXT NOT NULL REFERENCES messages(id),
			tool_name TEXT NOT NULL,
			args_json TEXT,
			result_json TEXT,
			error TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE,
			FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS attachments (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL REFERENCES sessions(id),
			message_id TEXT REFERENCES messages(id),
			type TEXT NOT NULL,
			path TEXT NOT NULL,
			metadata_json TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE,
			FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value_json TEXT
		);
		`,
		`CREATE INDEX IF NOT EXISTS idx_messages_session_created ON messages(session_id, created_at);`,
		`CREATE INDEX IF NOT EXISTS idx_tool_calls_session_created ON tool_calls(session_id, created_at);`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_user_updated ON sessions(user_id, updated_at);`,
		`INSERT OR IGNORE INTO users (id, name, created_at) VALUES ('default', 'Default User', CURRENT_TIMESTAMP);`,
	}

	for _, migration := range migrations {
		if _, err := s.db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}


func (s *Storage) CreateSession(session *models.Session) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	session.CreatedAt = time.Now()
	session.UpdatedAt = session.CreatedAt

	_, err := s.db.Exec(
		`INSERT INTO sessions (id, user_id, title, model, system_prompt, summary, created_at, updated_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		session.ID, session.UserID, session.Title, session.Model, session.SystemPrompt, session.Summary,
		session.CreatedAt, session.UpdatedAt,
	)
	return err
}

func (s *Storage) GetSession(id string) (*models.Session, error) {
	var session models.Session
	err := s.db.QueryRow(
		`SELECT id, user_id, title, model, system_prompt, summary, created_at, updated_at, archived_at 
		 FROM sessions WHERE id = ?`,
		id,
	).Scan(&session.ID, &session.UserID, &session.Title, &session.Model, &session.SystemPrompt,
		&session.Summary, &session.CreatedAt, &session.UpdatedAt, &session.ArchivedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	return &session, err
}

func (s *Storage) ListSessions(userID string) ([]models.Session, error) {
	rows, err := s.db.Query(
		`SELECT id, user_id, title, model, system_prompt, summary, created_at, updated_at, archived_at 
		 FROM sessions WHERE user_id = ? AND archived_at IS NULL ORDER BY updated_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []models.Session
	for rows.Next() {
		var session models.Session
		err := rows.Scan(&session.ID, &session.UserID, &session.Title, &session.Model, &session.SystemPrompt,
			&session.Summary, &session.CreatedAt, &session.UpdatedAt, &session.ArchivedAt)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, rows.Err()
}

func (s *Storage) UpdateSession(session *models.Session) error {
	session.UpdatedAt = time.Now()
	_, err := s.db.Exec(
		`UPDATE sessions SET title = ?, model = ?, system_prompt = ?, summary = ?, updated_at = ? WHERE id = ?`,
		session.Title, session.Model, session.SystemPrompt, session.Summary, session.UpdatedAt, session.ID,
	)
	return err
}

func (s *Storage) ArchiveSession(id string) error {
	_, err := s.db.Exec(`UPDATE sessions SET archived_at = ? WHERE id = ?`, time.Now(), id)
	return err
}

func (s *Storage) DeleteSession(id string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	return err
}


func (s *Storage) CreateMessage(msg *models.Message) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	msg.CreatedAt = time.Now()

	metadataJSON, _ := json.Marshal(msg.Metadata)

	_, err := s.db.Exec(
		`INSERT INTO messages (id, session_id, role, content, status, token_count, metadata_json, created_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		msg.ID, msg.SessionID, msg.Role, msg.Content, msg.Status, msg.TokenCount, metadataJSON, msg.CreatedAt,
	)
	return err
}

func (s *Storage) GetMessage(id string) (*models.Message, error) {
	var msg models.Message
	var metadataJSON []byte
	err := s.db.QueryRow(
		`SELECT id, session_id, role, content, status, token_count, metadata_json, created_at 
		 FROM messages WHERE id = ?`,
		id,
	).Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &msg.Status, &msg.TokenCount, &metadataJSON, &msg.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("message not found: %s", id)
	}
	if err != nil {
		return nil, err
	}
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &msg.Metadata)
	}
	return &msg, nil
}

func (s *Storage) ListMessages(sessionID string, limit int, offset int) ([]models.Message, error) {
	query := `SELECT id, session_id, role, content, status, token_count, metadata_json, created_at 
		 FROM messages WHERE session_id = ? ORDER BY created_at DESC`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	rows, err := s.db.Query(query, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		var metadataJSON []byte
		err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &msg.Status, &msg.TokenCount, &metadataJSON, &msg.CreatedAt)
		if err != nil {
			return nil, err
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &msg.Metadata)
		}
		messages = append(messages, msg)
	}
	return messages, rows.Err()
}

func (s *Storage) UpdateMessage(msg *models.Message) error {
	metadataJSON, _ := json.Marshal(msg.Metadata)
	_, err := s.db.Exec(
		`UPDATE messages SET content = ?, status = ?, token_count = ?, metadata_json = ? WHERE id = ?`,
		msg.Content, msg.Status, msg.TokenCount, metadataJSON, msg.ID,
	)
	return err
}

func (s *Storage) UpdateSessionTimestamp(sessionID string) error {
	_, err := s.db.Exec(`UPDATE sessions SET updated_at = ? WHERE id = ?`, time.Now(), sessionID)
	return err
}


func (s *Storage) CreateToolCall(tc *models.ToolCall) error {
	if tc.ID == "" {
		tc.ID = uuid.New().String()
	}
	tc.CreatedAt = time.Now()

	argsJSON, _ := json.Marshal(tc.Args)
	resultJSON, _ := json.Marshal(tc.Result)

	_, err := s.db.Exec(
		`INSERT INTO tool_calls (id, session_id, message_id, tool_name, args_json, result_json, error, created_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		tc.ID, tc.SessionID, tc.MessageID, tc.ToolName, argsJSON, resultJSON, tc.Error, tc.CreatedAt,
	)
	return err
}

func (s *Storage) ListToolCalls(sessionID string) ([]models.ToolCall, error) {
	rows, err := s.db.Query(
		`SELECT id, session_id, message_id, tool_name, args_json, result_json, error, created_at 
		 FROM tool_calls WHERE session_id = ? ORDER BY created_at DESC`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var calls []models.ToolCall
	for rows.Next() {
		var tc models.ToolCall
		var argsJSON, resultJSON []byte
		err := rows.Scan(&tc.ID, &tc.SessionID, &tc.MessageID, &tc.ToolName, &argsJSON, &resultJSON, &tc.Error, &tc.CreatedAt)
		if err != nil {
			return nil, err
		}
		if len(argsJSON) > 0 {
			json.Unmarshal(argsJSON, &tc.Args)
		}
		if len(resultJSON) > 0 {
			json.Unmarshal(resultJSON, &tc.Result)
		}
		calls = append(calls, tc)
	}
	return calls, rows.Err()
}


func (s *Storage) GetSetting(key string) (any, error) {
	var valueJSON []byte
	err := s.db.QueryRow(`SELECT value_json FROM settings WHERE key = ?`, key).Scan(&valueJSON)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var value any
	if err := json.Unmarshal(valueJSON, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func (s *Storage) SetSetting(key string, value any) error {
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(
		`INSERT INTO settings (key, value_json) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value_json = excluded.value_json`,
		key, valueJSON,
	)
	return err
}
