package models

import (
	"database/sql"
	"fmt"
	"time"

	"dev-dashboard/pkg/types"
)

type ActionModel struct {
	db *sql.DB
}

func NewActionModel(db *sql.DB) *ActionModel {
	return &ActionModel{db: db}
}

func (m *ActionModel) Create(action *types.Action) error {
	query := `
		INSERT INTO actions (repository_id, service_id, resource_id, type, status, workflow_run_id, commit_sha, branch, build_hash, started_at, completed_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	action.CreatedAt = now
	action.UpdatedAt = now

	result, err := m.db.Exec(query, action.RepositoryID, action.ServiceID, action.ResourceID, action.Type, action.Status, action.WorkflowRunID, action.Commit, action.Branch, action.BuildHash, action.StartedAt, action.CompletedAt, action.CreatedAt, action.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create action: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get action ID: %w", err)
	}

	action.ID = id
	return nil
}

func (m *ActionModel) GetByRepositoryID(repositoryID int64, limit int) ([]*types.ActionWithDetails, error) {
	query := `
		SELECT 
			a.id, a.repository_id, a.service_id, a.resource_id, a.type, a.status, 
			a.workflow_run_id, a.commit_sha, a.branch, a.build_hash, a.started_at, 
			a.completed_at, a.created_at, a.updated_at,
			ms.name as service_name,
			kr.name as resource_name
		FROM actions a
		LEFT JOIN microservices ms ON a.service_id = ms.id
		LEFT JOIN kubernetes_resources kr ON a.resource_id = kr.id
		WHERE a.repository_id = ?
		ORDER BY a.started_at DESC
		LIMIT ?
	`
	
	rows, err := m.db.Query(query, repositoryID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query actions: %w", err)
	}
	defer rows.Close()

	var actions []*types.ActionWithDetails
	for rows.Next() {
		action := &types.ActionWithDetails{}
		err := rows.Scan(
			&action.ID,
			&action.RepositoryID,
			&action.ServiceID,
			&action.ResourceID,
			&action.Type,
			&action.Status,
			&action.WorkflowRunID,
			&action.Commit,
			&action.Branch,
			&action.BuildHash,
			&action.StartedAt,
			&action.CompletedAt,
			&action.CreatedAt,
			&action.UpdatedAt,
			&action.ServiceName,
			&action.ResourceName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan action: %w", err)
		}
		actions = append(actions, action)
	}

	return actions, nil
}

func (m *ActionModel) GetByServiceID(serviceID int64, limit int) ([]*types.Action, error) {
	query := `
		SELECT id, repository_id, service_id, resource_id, type, status, workflow_run_id, commit_sha, branch, build_hash, started_at, completed_at, created_at, updated_at
		FROM actions
		WHERE service_id = ?
		ORDER BY started_at DESC
		LIMIT ?
	`
	
	rows, err := m.db.Query(query, serviceID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query actions by service: %w", err)
	}
	defer rows.Close()

	var actions []*types.Action
	for rows.Next() {
		action := &types.Action{}
		err := rows.Scan(
			&action.ID,
			&action.RepositoryID,
			&action.ServiceID,
			&action.ResourceID,
			&action.Type,
			&action.Status,
			&action.WorkflowRunID,
			&action.Commit,
			&action.Branch,
			&action.BuildHash,
			&action.StartedAt,
			&action.CompletedAt,
			&action.CreatedAt,
			&action.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan action: %w", err)
		}
		actions = append(actions, action)
	}

	return actions, nil
}

func (m *ActionModel) GetByResourceID(resourceID int64, limit int) ([]*types.Action, error) {
	query := `
		SELECT id, repository_id, service_id, resource_id, type, status, workflow_run_id, commit_sha, branch, build_hash, started_at, completed_at, created_at, updated_at
		FROM actions
		WHERE resource_id = ?
		ORDER BY started_at DESC
		LIMIT ?
	`
	
	rows, err := m.db.Query(query, resourceID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query actions by resource: %w", err)
	}
	defer rows.Close()

	var actions []*types.Action
	for rows.Next() {
		action := &types.Action{}
		err := rows.Scan(
			&action.ID,
			&action.RepositoryID,
			&action.ServiceID,
			&action.ResourceID,
			&action.Type,
			&action.Status,
			&action.WorkflowRunID,
			&action.Commit,
			&action.Branch,
			&action.BuildHash,
			&action.StartedAt,
			&action.CompletedAt,
			&action.CreatedAt,
			&action.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan action: %w", err)
		}
		actions = append(actions, action)
	}

	return actions, nil
}

func (m *ActionModel) Update(action *types.Action) error {
	query := `
		UPDATE actions
		SET status = ?, build_hash = ?, completed_at = ?, updated_at = ?
		WHERE id = ?
	`
	
	action.UpdatedAt = time.Now()
	_, err := m.db.Exec(query, action.Status, action.BuildHash, action.CompletedAt, action.UpdatedAt, action.ID)
	if err != nil {
		return fmt.Errorf("failed to update action: %w", err)
	}

	return nil
}

func (m *ActionModel) UpsertActions(actions []types.Action) error {
	if len(actions) == 0 {
		return nil
	}

	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT OR REPLACE INTO actions 
		(repository_id, service_id, resource_id, type, status, workflow_run_id, commit_sha, branch, build_hash, started_at, completed_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for _, action := range actions {
		if action.CreatedAt.IsZero() {
			action.CreatedAt = now
		}
		action.UpdatedAt = now

		_, err = stmt.Exec(
			action.RepositoryID,
			action.ServiceID,
			action.ResourceID,
			action.Type,
			action.Status,
			action.WorkflowRunID,
			action.Commit,
			action.Branch,
			action.BuildHash,
			action.StartedAt,
			action.CompletedAt,
			action.CreatedAt,
			action.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to upsert action: %w", err)
		}
	}

	return tx.Commit()
}