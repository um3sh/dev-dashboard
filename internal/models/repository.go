package models

import (
	"database/sql"
	"fmt"
	"time"

	"dev-dashboard/pkg/types"
)

type RepositoryModel struct {
	db *sql.DB
}

func NewRepositoryModel(db *sql.DB) *RepositoryModel {
	return &RepositoryModel{db: db}
}

func (m *RepositoryModel) Create(repo *types.Repository) error {
	query := `
		INSERT INTO repositories (name, url, type, description, service_name, service_location, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	repo.CreatedAt = now
	repo.UpdatedAt = now

	result, err := m.db.Exec(query, repo.Name, repo.URL, repo.Type, repo.Description, repo.ServiceName, repo.ServiceLocation, repo.CreatedAt, repo.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get repository ID: %w", err)
	}

	repo.ID = id
	return nil
}

func (m *RepositoryModel) GetByID(id int64) (*types.Repository, error) {
	query := `
		SELECT id, name, url, type, description, service_name, service_location, created_at, updated_at, last_sync_at
		FROM repositories
		WHERE id = ?
	`
	
	repo := &types.Repository{}
	err := m.db.QueryRow(query, id).Scan(
		&repo.ID,
		&repo.Name,
		&repo.URL,
		&repo.Type,
		&repo.Description,
		&repo.ServiceName,
		&repo.ServiceLocation,
		&repo.CreatedAt,
		&repo.UpdatedAt,
		&repo.LastSyncAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	return repo, nil
}

func (m *RepositoryModel) GetAll() ([]*types.Repository, error) {
	query := `
		SELECT id, name, url, type, description, service_name, service_location, created_at, updated_at, last_sync_at
		FROM repositories
		ORDER BY created_at DESC
	`
	
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query repositories: %w", err)
	}
	defer rows.Close()

	var repositories []*types.Repository
	for rows.Next() {
		repo := &types.Repository{}
		err := rows.Scan(
			&repo.ID,
			&repo.Name,
			&repo.URL,
			&repo.Type,
			&repo.Description,
			&repo.ServiceName,
			&repo.ServiceLocation,
			&repo.CreatedAt,
			&repo.UpdatedAt,
			&repo.LastSyncAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan repository: %w", err)
		}
		repositories = append(repositories, repo)
	}

	return repositories, nil
}

func (m *RepositoryModel) Update(repo *types.Repository) error {
	query := `
		UPDATE repositories
		SET name = ?, url = ?, type = ?, description = ?, service_name = ?, service_location = ?, updated_at = ?
		WHERE id = ?
	`
	
	repo.UpdatedAt = time.Now()
	_, err := m.db.Exec(query, repo.Name, repo.URL, repo.Type, repo.Description, repo.ServiceName, repo.ServiceLocation, repo.UpdatedAt, repo.ID)
	if err != nil {
		return fmt.Errorf("failed to update repository: %w", err)
	}

	return nil
}

func (m *RepositoryModel) UpdateLastSync(id int64) error {
	query := `
		UPDATE repositories
		SET last_sync_at = ?, updated_at = ?
		WHERE id = ?
	`
	
	now := time.Now()
	_, err := m.db.Exec(query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to update last sync: %w", err)
	}

	return nil
}

func (m *RepositoryModel) Delete(id int64) error {
	// Start a transaction to ensure atomic deletion
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Enable foreign keys for this transaction to ensure CASCADE deletes work
	if _, err := tx.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}
	
	query := `DELETE FROM repositories WHERE id = ?`
	
	result, err := tx.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete repository: %w", err)
	}

	// Check if a row was actually deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("repository with ID %d not found", id)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}