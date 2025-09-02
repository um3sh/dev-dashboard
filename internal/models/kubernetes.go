package models

import (
	"database/sql"
	"fmt"
	"time"

	"gh-dashboard/pkg/types"
)

type KubernetesResourceModel struct {
	db *sql.DB
}

func NewKubernetesResourceModel(db *sql.DB) *KubernetesResourceModel {
	return &KubernetesResourceModel{db: db}
}

func (m *KubernetesResourceModel) Create(resource *types.KubernetesResource) error {
	query := `
		INSERT INTO kubernetes_resources (repository_id, name, path, resource_type, namespace, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	resource.CreatedAt = now
	resource.UpdatedAt = now

	result, err := m.db.Exec(query, resource.RepositoryID, resource.Name, resource.Path, resource.ResourceType, resource.Namespace, resource.CreatedAt, resource.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes resource: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get kubernetes resource ID: %w", err)
	}

	resource.ID = id
	return nil
}

func (m *KubernetesResourceModel) GetByRepositoryID(repositoryID int64) ([]*types.KubernetesResource, error) {
	query := `
		SELECT id, repository_id, name, path, resource_type, namespace, created_at, updated_at
		FROM kubernetes_resources
		WHERE repository_id = ?
		ORDER BY namespace, name
	`
	
	rows, err := m.db.Query(query, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to query kubernetes resources: %w", err)
	}
	defer rows.Close()

	var resources []*types.KubernetesResource
	for rows.Next() {
		resource := &types.KubernetesResource{}
		err := rows.Scan(
			&resource.ID,
			&resource.RepositoryID,
			&resource.Name,
			&resource.Path,
			&resource.ResourceType,
			&resource.Namespace,
			&resource.CreatedAt,
			&resource.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan kubernetes resource: %w", err)
		}
		resources = append(resources, resource)
	}

	return resources, nil
}

func (m *KubernetesResourceModel) GetByID(id int64) (*types.KubernetesResource, error) {
	query := `
		SELECT id, repository_id, name, path, resource_type, namespace, created_at, updated_at
		FROM kubernetes_resources
		WHERE id = ?
	`
	
	resource := &types.KubernetesResource{}
	err := m.db.QueryRow(query, id).Scan(
		&resource.ID,
		&resource.RepositoryID,
		&resource.Name,
		&resource.Path,
		&resource.ResourceType,
		&resource.Namespace,
		&resource.CreatedAt,
		&resource.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get kubernetes resource: %w", err)
	}

	return resource, nil
}

func (m *KubernetesResourceModel) Update(resource *types.KubernetesResource) error {
	query := `
		UPDATE kubernetes_resources
		SET name = ?, path = ?, resource_type = ?, namespace = ?, updated_at = ?
		WHERE id = ?
	`
	
	resource.UpdatedAt = time.Now()
	_, err := m.db.Exec(query, resource.Name, resource.Path, resource.ResourceType, resource.Namespace, resource.UpdatedAt, resource.ID)
	if err != nil {
		return fmt.Errorf("failed to update kubernetes resource: %w", err)
	}

	return nil
}

func (m *KubernetesResourceModel) Delete(id int64) error {
	query := `DELETE FROM kubernetes_resources WHERE id = ?`
	
	_, err := m.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete kubernetes resource: %w", err)
	}

	return nil
}

func (m *KubernetesResourceModel) DeleteByRepositoryID(repositoryID int64) error {
	query := `DELETE FROM kubernetes_resources WHERE repository_id = ?`
	
	_, err := m.db.Exec(query, repositoryID)
	if err != nil {
		return fmt.Errorf("failed to delete kubernetes resources: %w", err)
	}

	return nil
}

func (m *KubernetesResourceModel) UpsertResources(repositoryID int64, resources []types.KubernetesResource) error {
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing resources for this repository
	_, err = tx.Exec("DELETE FROM kubernetes_resources WHERE repository_id = ?", repositoryID)
	if err != nil {
		return fmt.Errorf("failed to delete existing resources: %w", err)
	}

	// Insert new resources
	if len(resources) > 0 {
		query := `
			INSERT INTO kubernetes_resources (repository_id, name, path, resource_type, namespace, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`
		stmt, err := tx.Prepare(query)
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}
		defer stmt.Close()

		now := time.Now()
		for _, resource := range resources {
			_, err = stmt.Exec(repositoryID, resource.Name, resource.Path, resource.ResourceType, resource.Namespace, now, now)
			if err != nil {
				return fmt.Errorf("failed to insert resource %s: %w", resource.Name, err)
			}
		}
	}

	return tx.Commit()
}