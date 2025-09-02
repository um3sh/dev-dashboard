package models

import (
	"database/sql"
	"fmt"
	"time"
)

type ConfigModel struct {
	db *sql.DB
}

type Config struct {
	Key       string    `json:"key" db:"key"`
	Value     string    `json:"value" db:"value"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func NewConfigModel(db *sql.DB) *ConfigModel {
	return &ConfigModel{db: db}
}

func (m *ConfigModel) Get(key string) (*Config, error) {
	query := `SELECT key, value, updated_at FROM config WHERE key = ?`
	
	config := &Config{}
	err := m.db.QueryRow(query, key).Scan(
		&config.Key,
		&config.Value,
		&config.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No config found, not an error
		}
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	return config, nil
}

func (m *ConfigModel) Set(key, value string) error {
	query := `
		INSERT INTO config (key, value, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			updated_at = excluded.updated_at
	`
	
	now := time.Now()
	_, err := m.db.Exec(query, key, value, now)
	if err != nil {
		return fmt.Errorf("failed to set config: %w", err)
	}

	return nil
}

func (m *ConfigModel) GetAll() (map[string]string, error) {
	query := `SELECT key, value FROM config`
	
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query config: %w", err)
	}
	defer rows.Close()

	configs := make(map[string]string)
	for rows.Next() {
		var key, value string
		err := rows.Scan(&key, &value)
		if err != nil {
			return nil, fmt.Errorf("failed to scan config: %w", err)
		}
		configs[key] = value
	}

	return configs, nil
}

func (m *ConfigModel) Delete(key string) error {
	query := `DELETE FROM config WHERE key = ?`
	
	_, err := m.db.Exec(query, key)
	if err != nil {
		return fmt.Errorf("failed to delete config: %w", err)
	}

	return nil
}