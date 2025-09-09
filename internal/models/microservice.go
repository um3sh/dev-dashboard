package models

import (
	"database/sql"
	"fmt"
	"time"

	"dev-dashboard/pkg/types"
)

type MicroserviceModel struct {
	db *sql.DB
}

func NewMicroserviceModel(db *sql.DB) *MicroserviceModel {
	return &MicroserviceModel{db: db}
}

func (m *MicroserviceModel) Create(service *types.Microservice) error {
	query := `
		INSERT INTO microservices (repository_id, name, path, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	service.CreatedAt = now
	service.UpdatedAt = now

	result, err := m.db.Exec(query, service.RepositoryID, service.Name, service.Path, service.Description, service.CreatedAt, service.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create microservice: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get microservice ID: %w", err)
	}

	service.ID = id
	return nil
}

func (m *MicroserviceModel) GetByRepositoryID(repositoryID int64) ([]*types.Microservice, error) {
	query := `
		SELECT id, repository_id, name, path, description, created_at, updated_at
		FROM microservices
		WHERE repository_id = ?
		ORDER BY name
	`
	
	rows, err := m.db.Query(query, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to query microservices: %w", err)
	}
	defer rows.Close()

	var services []*types.Microservice
	for rows.Next() {
		service := &types.Microservice{}
		err := rows.Scan(
			&service.ID,
			&service.RepositoryID,
			&service.Name,
			&service.Path,
			&service.Description,
			&service.CreatedAt,
			&service.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan microservice: %w", err)
		}
		services = append(services, service)
	}

	return services, nil
}

func (m *MicroserviceModel) GetByID(id int64) (*types.Microservice, error) {
	query := `
		SELECT id, repository_id, name, path, description, created_at, updated_at
		FROM microservices
		WHERE id = ?
	`
	
	service := &types.Microservice{}
	err := m.db.QueryRow(query, id).Scan(
		&service.ID,
		&service.RepositoryID,
		&service.Name,
		&service.Path,
		&service.Description,
		&service.CreatedAt,
		&service.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get microservice: %w", err)
	}

	return service, nil
}

func (m *MicroserviceModel) Update(service *types.Microservice) error {
	query := `
		UPDATE microservices
		SET name = ?, path = ?, description = ?, updated_at = ?
		WHERE id = ?
	`
	
	service.UpdatedAt = time.Now()
	_, err := m.db.Exec(query, service.Name, service.Path, service.Description, service.UpdatedAt, service.ID)
	if err != nil {
		return fmt.Errorf("failed to update microservice: %w", err)
	}

	return nil
}

func (m *MicroserviceModel) Delete(id int64) error {
	query := `DELETE FROM microservices WHERE id = ?`
	
	_, err := m.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete microservice: %w", err)
	}

	return nil
}

func (m *MicroserviceModel) DeleteByRepositoryID(repositoryID int64) error {
	query := `DELETE FROM microservices WHERE repository_id = ?`
	
	_, err := m.db.Exec(query, repositoryID)
	if err != nil {
		return fmt.Errorf("failed to delete microservices: %w", err)
	}

	return nil
}

func (m *MicroserviceModel) UpsertServices(repositoryID int64, services []types.Microservice) error {
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing services for this repository
	_, err = tx.Exec("DELETE FROM microservices WHERE repository_id = ?", repositoryID)
	if err != nil {
		return fmt.Errorf("failed to delete existing services: %w", err)
	}

	// Insert new services
	if len(services) > 0 {
		query := `
			INSERT INTO microservices (repository_id, name, path, description, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`
		stmt, err := tx.Prepare(query)
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}
		defer stmt.Close()

		now := time.Now()
		for _, service := range services {
			_, err = stmt.Exec(repositoryID, service.Name, service.Path, service.Description, now, now)
			if err != nil {
				return fmt.Errorf("failed to insert service %s: %w", service.Name, err)
			}
		}
	}

	return tx.Commit()
}

func (m *MicroserviceModel) UpsertServicesPreserveID(repositoryID int64, services []types.Microservice) error {
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get existing services for this repository
	existingServices := make(map[string]*types.Microservice)
	rows, err := tx.Query("SELECT id, name, path, description, created_at, updated_at FROM microservices WHERE repository_id = ?", repositoryID)
	if err != nil {
		return fmt.Errorf("failed to query existing services: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		service := &types.Microservice{RepositoryID: repositoryID}
		err := rows.Scan(&service.ID, &service.Name, &service.Path, &service.Description, &service.CreatedAt, &service.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan existing service: %w", err)
		}
		// Use name+path as unique key
		key := service.Name + "|" + service.Path
		existingServices[key] = service
	}

	// Track which services we've processed to know which ones to delete
	processedServices := make(map[string]bool)
	now := time.Now()

	// Process new services
	for _, newService := range services {
		key := newService.Name + "|" + newService.Path
		processedServices[key] = true

		if existingService, exists := existingServices[key]; exists {
			// Update existing service
			_, err = tx.Exec(
				"UPDATE microservices SET description = ?, updated_at = ? WHERE id = ?",
				newService.Description, now, existingService.ID,
			)
			if err != nil {
				return fmt.Errorf("failed to update service %s: %w", newService.Name, err)
			}
		} else {
			// Insert new service
			_, err = tx.Exec(
				"INSERT INTO microservices (repository_id, name, path, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
				repositoryID, newService.Name, newService.Path, newService.Description, now, now,
			)
			if err != nil {
				return fmt.Errorf("failed to insert service %s: %w", newService.Name, err)
			}
		}
	}

	// Delete services that no longer exist
	for key, existingService := range existingServices {
		if !processedServices[key] {
			_, err = tx.Exec("DELETE FROM microservices WHERE id = ?", existingService.ID)
			if err != nil {
				return fmt.Errorf("failed to delete service %s: %w", existingService.Name, err)
			}
		}
	}

	return tx.Commit()
}

func (m *MicroserviceModel) GetAll() ([]*types.Microservice, error) {
	query := `
		SELECT id, repository_id, name, path, description, created_at, updated_at
		FROM microservices
		ORDER BY name
	`
	
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query microservices: %w", err)
	}
	defer rows.Close()

	var services []*types.Microservice
	for rows.Next() {
		service := &types.Microservice{}
		err := rows.Scan(
			&service.ID,
			&service.RepositoryID,
			&service.Name,
			&service.Path,
			&service.Description,
			&service.CreatedAt,
			&service.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan microservice: %w", err)
		}
		services = append(services, service)
	}

	return services, nil
}