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

	// Check if deployments table exists
	var deploymentsTableExists bool
	err = db.conn.QueryRow(`
		SELECT COUNT(*) > 0 
		FROM sqlite_master 
		WHERE type='table' AND name='deployments'
	`).Scan(&deploymentsTableExists)
	
	if err != nil {
		return fmt.Errorf("failed to check for deployments table: %w", err)
	}

	// Create deployments table if it doesn't exist
	if !deploymentsTableExists {
		_, err = db.conn.Exec(`
			CREATE TABLE IF NOT EXISTS deployments (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				service_id INTEGER NOT NULL,
				kubernetes_repo_id INTEGER NOT NULL,
				commit_sha TEXT NOT NULL,
				environment TEXT NOT NULL,
				region TEXT NOT NULL,
				tag TEXT NOT NULL,
				path TEXT NOT NULL,
				discovered_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (service_id) REFERENCES microservices(id) ON DELETE CASCADE,
				FOREIGN KEY (kubernetes_repo_id) REFERENCES repositories(id) ON DELETE CASCADE,
				UNIQUE(service_id, environment, region)
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create deployments table: %w", err)
		}

		// Add indexes for deployments table
		indexes := []string{
			"CREATE INDEX IF NOT EXISTS idx_deployments_service_id ON deployments(service_id)",
			"CREATE INDEX IF NOT EXISTS idx_deployments_kubernetes_repo_id ON deployments(kubernetes_repo_id)",
			"CREATE INDEX IF NOT EXISTS idx_deployments_commit_sha ON deployments(commit_sha)",
			"CREATE INDEX IF NOT EXISTS idx_deployments_environment ON deployments(environment)",
			"CREATE INDEX IF NOT EXISTS idx_deployments_region ON deployments(region)",
		}

		for _, indexSQL := range indexes {
			_, err = db.conn.Exec(indexSQL)
			if err != nil {
				return fmt.Errorf("failed to create deployments index: %w", err)
			}
		}

		// Add the deployments table trigger
		_, err = db.conn.Exec(`
			CREATE TRIGGER IF NOT EXISTS update_deployments_updated_at
				AFTER UPDATE ON deployments
			BEGIN
				UPDATE deployments SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
			END
		`)
		if err != nil {
			return fmt.Errorf("failed to create deployments trigger: %w", err)
		}
	} else {
		// Check if namespace column exists in deployments table
		var namespaceColumnExists bool
		err = db.conn.QueryRow(`
			SELECT COUNT(*) > 0 
			FROM pragma_table_info('deployments') 
			WHERE name = 'namespace'
		`).Scan(&namespaceColumnExists)
		
		if err != nil {
			return fmt.Errorf("failed to check for namespace column: %w", err)
		}

		// Add namespace column if it doesn't exist
		if !namespaceColumnExists {
			_, err = db.conn.Exec("ALTER TABLE deployments ADD COLUMN namespace TEXT")
			if err != nil {
				return fmt.Errorf("failed to add namespace column: %w", err)
			}
		}

		// Update the unique constraint to include namespace
		// Since SQLite doesn't support altering constraints, we need to check if the old constraint exists
		// and recreate the table if necessary
		var constraintExists bool
		err = db.conn.QueryRow(`
			SELECT COUNT(*) > 0 
			FROM sqlite_master 
			WHERE type = 'index' 
			AND tbl_name = 'deployments' 
			AND sql LIKE '%UNIQUE(service_id, environment, region, namespace)%'
		`).Scan(&constraintExists)
		
		if err != nil {
			return fmt.Errorf("failed to check for updated unique constraint: %w", err)
		}

		// If the constraint doesn't include namespace, we need to recreate the table
		if !constraintExists {
			// Create a temporary table with the new schema
			_, err = db.conn.Exec(`
				CREATE TABLE deployments_new (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					service_id INTEGER NOT NULL,
					kubernetes_repo_id INTEGER NOT NULL,
					commit_sha TEXT NOT NULL,
					environment TEXT NOT NULL,
					region TEXT NOT NULL,
					namespace TEXT,
					tag TEXT NOT NULL,
					path TEXT NOT NULL,
					discovered_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (service_id) REFERENCES microservices(id) ON DELETE CASCADE,
					FOREIGN KEY (kubernetes_repo_id) REFERENCES repositories(id) ON DELETE CASCADE,
					UNIQUE(service_id, environment, region, namespace)
				)
			`)
			if err != nil {
				return fmt.Errorf("failed to create new deployments table: %w", err)
			}

			// Copy data from old table to new table
			_, err = db.conn.Exec(`
				INSERT INTO deployments_new (id, service_id, kubernetes_repo_id, commit_sha, environment, region, namespace, tag, path, discovered_at, updated_at)
				SELECT id, service_id, kubernetes_repo_id, commit_sha, environment, region, namespace, tag, path, discovered_at, updated_at
				FROM deployments
			`)
			if err != nil {
				return fmt.Errorf("failed to copy data to new deployments table: %w", err)
			}

			// Drop the old table and rename the new one
			_, err = db.conn.Exec("DROP TABLE deployments")
			if err != nil {
				return fmt.Errorf("failed to drop old deployments table: %w", err)
			}

			_, err = db.conn.Exec("ALTER TABLE deployments_new RENAME TO deployments")
			if err != nil {
				return fmt.Errorf("failed to rename new deployments table: %w", err)
			}

			// Recreate indexes and triggers
			indexes := []string{
				"CREATE INDEX IF NOT EXISTS idx_deployments_service_id ON deployments(service_id)",
				"CREATE INDEX IF NOT EXISTS idx_deployments_kubernetes_repo_id ON deployments(kubernetes_repo_id)",
				"CREATE INDEX IF NOT EXISTS idx_deployments_commit_sha ON deployments(commit_sha)",
				"CREATE INDEX IF NOT EXISTS idx_deployments_environment ON deployments(environment)",
				"CREATE INDEX IF NOT EXISTS idx_deployments_region ON deployments(region)",
			}

			for _, indexSQL := range indexes {
				_, err = db.conn.Exec(indexSQL)
				if err != nil {
					return fmt.Errorf("failed to create deployments index: %w", err)
				}
			}

			// Recreate the trigger
			_, err = db.conn.Exec(`
				CREATE TRIGGER IF NOT EXISTS update_deployments_updated_at
					AFTER UPDATE ON deployments
				BEGIN
					UPDATE deployments SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
				END
			`)
			if err != nil {
				return fmt.Errorf("failed to create deployments trigger: %w", err)
			}
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