package database

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaFS embed.FS

type DB struct {
	conn *sql.DB
}

func NewDB(dbPath string) (*DB, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := conn.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	db := &DB{conn: conn}

	if err := db.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

func (db *DB) initSchema() error {
	// Check if tables already exist
	var tableCount int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('repositories', 'projects', 'tasks', 'config')").Scan(&tableCount)
	
	if err == nil && tableCount >= 4 {
		// Tables exist, check if we need migrations
		return db.runMigrations()
	}

	// Create fresh schema
	schemaSQL, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	// Execute the schema as a single batch to avoid transaction issues
	_, err = db.conn.Exec(string(schemaSQL))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}

func (db *DB) runMigrations() error {
	// Check if jira_title column exists in tasks table
	var columnExists bool
	err := db.conn.QueryRow(`
		SELECT COUNT(*) > 0 
		FROM pragma_table_info('tasks') 
		WHERE name = 'jira_title'
	`).Scan(&columnExists)
	
	if err != nil {
		return fmt.Errorf("failed to check for jira_title column: %w", err)
	}

	// Add jira_title column if it doesn't exist
	if !columnExists {
		_, err = db.conn.Exec("ALTER TABLE tasks ADD COLUMN jira_title TEXT")
		if err != nil {
			return fmt.Errorf("failed to add jira_title column: %w", err)
		}
	}

	// Check if config table exists
	var configTableExists bool
	err = db.conn.QueryRow(`
		SELECT COUNT(*) > 0 
		FROM sqlite_master 
		WHERE type='table' AND name='config'
	`).Scan(&configTableExists)
	
	if err != nil {
		return fmt.Errorf("failed to check for config table: %w", err)
	}

	// Create config table if it doesn't exist
	if !configTableExists {
		_, err = db.conn.Exec(`
			CREATE TABLE IF NOT EXISTS config (
				key TEXT PRIMARY KEY,
				value TEXT NOT NULL,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create config table: %w", err)
		}

		// Add the config table trigger
		_, err = db.conn.Exec(`
			CREATE TRIGGER IF NOT EXISTS update_config_updated_at
				AFTER UPDATE ON config
			BEGIN
				UPDATE config SET updated_at = CURRENT_TIMESTAMP WHERE key = NEW.key;
			END
		`)
		if err != nil {
			return fmt.Errorf("failed to create config trigger: %w", err)
		}

		// Add the config table index
		_, err = db.conn.Exec("CREATE INDEX IF NOT EXISTS idx_config_key ON config(key)")
		if err != nil {
			return fmt.Errorf("failed to create config index: %w", err)
		}
	}

	return nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) GetConn() *sql.DB {
	return db.conn
}