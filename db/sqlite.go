package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	_ "modernc.org/sqlite"

	"deepseek_bot/loger"
)

type Message struct {
	ID        int
	SessionID string
	Content   string
	Reply     string
	Status    int
	CreatedAt string
}

type ChatMessage struct {
	Role    string
	Content string
}

const (
	StatusPending = 0
	StatusReplied = 1
)

var DB *sql.DB

func InitDB(dbPath string) error {
	dir := filepath.Dir(dbPath)
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		loger.Loger.Info("[SQLite]directory does not exist, creating: " + dir)
		err := os.MkdirAll(dir, 0775)
		if err != nil {
			loger.Loger.Fatal("[SQLite]failed to create directory", zap.Error(err))
			return err
		}
	}

	var errDB error
	DB, errDB = sql.Open("sqlite", dbPath)
	if errDB != nil {
		loger.Loger.Fatal("[SQLite]failed to open database", zap.Error(errDB))
		return errDB
	}

	DB.SetMaxOpenConns(1)
	DB.SetMaxIdleConns(1)
	DB.SetConnMaxLifetime(time.Minute * 3)

	errDB = createTable()
	if errDB != nil {
		loger.Loger.Fatal("[SQLite]failed to create table", zap.Error(errDB))
		return errDB
	}

	loger.Loger.Info("[SQLite]READY!")
	return nil
}

func createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		content TEXT NOT NULL,
		reply TEXT DEFAULT '',
		status INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := DB.Exec(query)
	return err
}

func InsertMessage(sessionID, content string) (int64, error) {
	result, err := DB.Exec("INSERT INTO messages (session_id, content, status) VALUES (?, ?, ?)", sessionID, content, StatusPending)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func GetSessionHistory(sessionID string) ([]ChatMessage, error) {
	rows, err := DB.Query("SELECT content, reply FROM messages WHERE session_id = ? AND status = ? ORDER BY created_at ASC", sessionID, StatusReplied)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []ChatMessage
	for rows.Next() {
		var content, reply string
		err := rows.Scan(&content, &reply)
		if err != nil {
			return nil, err
		}
		history = append(history, ChatMessage{Role: "user", Content: content})
		history = append(history, ChatMessage{Role: "assistant", Content: reply})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return history, nil
}

func UpdateReply(id int, reply string) error {
	_, err := DB.Exec("UPDATE messages SET reply = ?, status = ? WHERE id = ?", reply, StatusReplied, id)
	return err
}

func CloseDB() error {
	if DB != nil {
		loger.Loger.Info("[SQLite]closing database connection")
		return DB.Close()
	}
	return nil
}
