package models

import (
	"database/sql"
	"fmt"
	"time"

	"dev-dashboard/pkg/types"
)

type DeploymentModel struct {
	db *sql.DB
}

func NewDeploymentModel(db *sql.DB) *DeploymentModel {
	return &DeploymentModel{db: db}
}

func (d *DeploymentModel) Create(deployment *types.Deployment) error {
	query := `
		INSERT INTO deployments (service_id, kubernetes_repo_id, commit_sha, environment, region, namespace, tag, path, discovered_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	deployment.DiscoveredAt = now
	deployment.UpdatedAt = now

	result, err := d.db.Exec(query, deployment.ServiceID, deployment.KubernetesRepoID, deployment.CommitSHA, deployment.Environment, deployment.Region, deployment.Namespace, deployment.Tag, deployment.Path, deployment.DiscoveredAt, deployment.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create deployment: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get deployment ID: %w", err)
	}

	deployment.ID = id
	return nil
}

func (d *DeploymentModel) GetByServiceID(serviceID int64) ([]*types.Deployment, error) {
	query := `
		SELECT id, service_id, kubernetes_repo_id, commit_sha, environment, region, namespace, tag, path, discovered_at, updated_at
		FROM deployments
		WHERE service_id = ?
		ORDER BY environment, region, namespace
	`
	
	rows, err := d.db.Query(query, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query deployments: %w", err)
	}
	defer rows.Close()

	var deployments []*types.Deployment
	for rows.Next() {
		deployment := &types.Deployment{}
		var namespace sql.NullString
		err := rows.Scan(
			&deployment.ID,
			&deployment.ServiceID,
			&deployment.KubernetesRepoID,
			&deployment.CommitSHA,
			&deployment.Environment,
			&deployment.Region,
			&namespace,
			&deployment.Tag,
			&deployment.Path,
			&deployment.DiscoveredAt,
			&deployment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deployment: %w", err)
		}
		
		// Handle NULL namespace
		if namespace.Valid {
			deployment.Namespace = namespace.String
		} else {
			deployment.Namespace = ""
		}
		
		deployments = append(deployments, deployment)
	}

	return deployments, nil
}

func (d *DeploymentModel) GetByID(id int64) (*types.Deployment, error) {
	query := `
		SELECT id, service_id, kubernetes_repo_id, commit_sha, environment, region, namespace, tag, path, discovered_at, updated_at
		FROM deployments
		WHERE id = ?
	`
	
	deployment := &types.Deployment{}
	var namespace sql.NullString
	err := d.db.QueryRow(query, id).Scan(
		&deployment.ID,
		&deployment.ServiceID,
		&deployment.KubernetesRepoID,
		&deployment.CommitSHA,
		&deployment.Environment,
		&deployment.Region,
		&namespace,
		&deployment.Tag,
		&deployment.Path,
		&deployment.DiscoveredAt,
		&deployment.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	// Handle NULL namespace
	if namespace.Valid {
		deployment.Namespace = namespace.String
	} else {
		deployment.Namespace = ""
	}

	return deployment, nil
}

func (d *DeploymentModel) Update(deployment *types.Deployment) error {
	query := `
		UPDATE deployments
		SET commit_sha = ?, tag = ?, path = ?, updated_at = ?
		WHERE id = ?
	`
	
	deployment.UpdatedAt = time.Now()
	_, err := d.db.Exec(query, deployment.CommitSHA, deployment.Tag, deployment.Path, deployment.UpdatedAt, deployment.ID)
	if err != nil {
		return fmt.Errorf("failed to update deployment: %w", err)
	}

	return nil
}

func (d *DeploymentModel) Upsert(deployment *types.Deployment) error {
	// Check if deployment already exists for this service, environment, and region
	existingQuery := `
		SELECT id FROM deployments
		WHERE service_id = ? AND environment = ? AND region = ? AND namespace = ?
	`
	
	var existingID int64
	err := d.db.QueryRow(existingQuery, deployment.ServiceID, deployment.Environment, deployment.Region, deployment.Namespace).Scan(&existingID)
	
	if err == sql.ErrNoRows {
		// Create new deployment
		return d.Create(deployment)
	} else if err != nil {
		return fmt.Errorf("failed to check existing deployment: %w", err)
	}
	
	// Update existing deployment
	deployment.ID = existingID
	return d.Update(deployment)
}

func (d *DeploymentModel) DeleteByServiceID(serviceID int64) error {
	query := `DELETE FROM deployments WHERE service_id = ?`
	
	_, err := d.db.Exec(query, serviceID)
	if err != nil {
		return fmt.Errorf("failed to delete deployments: %w", err)
	}

	return nil
}

func (d *DeploymentModel) GetDeploymentOverview(serviceID int64) ([]*types.DeploymentOverview, error) {
	query := `
		SELECT 
			d.commit_sha,
			d.environment,
			d.region,
			d.namespace,
			d.tag,
			d.updated_at,
			r.name as kubernetes_repo_name
		FROM deployments d
		JOIN repositories r ON d.kubernetes_repo_id = r.id
		WHERE d.service_id = ?
		ORDER BY d.environment, d.region, d.namespace
	`
	
	rows, err := d.db.Query(query, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query deployment overview: %w", err)
	}
	defer rows.Close()

	var deployments []*types.DeploymentOverview
	for rows.Next() {
		deployment := &types.DeploymentOverview{}
		var namespace sql.NullString
		err := rows.Scan(
			&deployment.CommitSHA,
			&deployment.Environment,
			&deployment.Region,
			&namespace,
			&deployment.Tag,
			&deployment.UpdatedAt,
			&deployment.KubernetesRepoName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deployment overview: %w", err)
		}
		
		// Handle NULL namespace
		if namespace.Valid {
			deployment.Namespace = namespace.String
		} else {
			deployment.Namespace = ""
		}
		
		deployments = append(deployments, deployment)
	}

	return deployments, nil
}