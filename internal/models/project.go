package models

import (
	"database/sql"
	"fmt"
	"time"

	"gh-dashboard/pkg/types"
)

type ProjectModel struct {
	db *sql.DB
}

func NewProjectModel(db *sql.DB) *ProjectModel {
	return &ProjectModel{db: db}
}

func (m *ProjectModel) Create(project *types.Project) error {
	query := `
		INSERT INTO projects (name, description, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`
	now := time.Now()
	project.CreatedAt = now
	project.UpdatedAt = now

	result, err := m.db.Exec(query, project.Name, project.Description, project.CreatedAt, project.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get project ID: %w", err)
	}

	project.ID = id
	return nil
}

func (m *ProjectModel) GetByID(id int64) (*types.Project, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM projects
		WHERE id = ?
	`
	
	project := &types.Project{}
	err := m.db.QueryRow(query, id).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return project, nil
}

func (m *ProjectModel) GetAll() ([]*types.Project, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM projects
		ORDER BY name ASC
	`
	
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}
	defer rows.Close()

	var projects []*types.Project
	for rows.Next() {
		project := &types.Project{}
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&project.CreatedAt,
			&project.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, project)
	}

	return projects, nil
}

func (m *ProjectModel) Update(project *types.Project) error {
	query := `
		UPDATE projects
		SET name = ?, description = ?, updated_at = ?
		WHERE id = ?
	`
	
	project.UpdatedAt = time.Now()
	_, err := m.db.Exec(query, project.Name, project.Description, project.UpdatedAt, project.ID)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	return nil
}

func (m *ProjectModel) Delete(id int64) error {
	query := `DELETE FROM projects WHERE id = ?`
	
	_, err := m.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}