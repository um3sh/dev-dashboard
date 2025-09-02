package models

import (
	"database/sql"
	"fmt"
	"time"

	"gh-dashboard/pkg/types"
)

type TaskModel struct {
	db *sql.DB
}

func NewTaskModel(db *sql.DB) *TaskModel {
	return &TaskModel{db: db}
}

func (m *TaskModel) Create(task *types.Task) error {
	query := `
		INSERT INTO tasks (project_id, jira_ticket_id, jira_title, title, description, scheduled_date, deadline, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now
	if task.Status == "" {
		task.Status = types.TaskPending
	}

	result, err := m.db.Exec(query, task.ProjectID, task.JiraTicketID, task.JiraTitle, task.Title, task.Description, task.ScheduledDate, task.Deadline, task.Status, task.CreatedAt, task.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get task ID: %w", err)
	}

	task.ID = id
	return nil
}

func (m *TaskModel) GetByID(id int64) (*types.Task, error) {
	query := `
		SELECT id, project_id, jira_ticket_id, jira_title, title, description, scheduled_date, deadline, status, created_at, updated_at
		FROM tasks
		WHERE id = ?
	`
	
	task := &types.Task{}
	err := m.db.QueryRow(query, id).Scan(
		&task.ID,
		&task.ProjectID,
		&task.JiraTicketID,
		&task.JiraTitle,
		&task.Title,
		&task.Description,
		&task.ScheduledDate,
		&task.Deadline,
		&task.Status,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return task, nil
}

func (m *TaskModel) GetByProjectID(projectID int64) ([]*types.Task, error) {
	query := `
		SELECT id, project_id, jira_ticket_id, jira_title, title, description, scheduled_date, deadline, status, created_at, updated_at
		FROM tasks
		WHERE project_id = ?
		ORDER BY 
			CASE WHEN scheduled_date IS NULL THEN 1 ELSE 0 END,
			scheduled_date ASC,
			deadline ASC
	`
	
	rows, err := m.db.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*types.Task
	for rows.Next() {
		task := &types.Task{}
		err := rows.Scan(
			&task.ID,
			&task.ProjectID,
			&task.JiraTicketID,
			&task.JiraTitle,
			&task.Title,
			&task.Description,
			&task.ScheduledDate,
			&task.Deadline,
			&task.Status,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (m *TaskModel) GetAllWithProjects() ([]*types.TaskWithProject, error) {
	query := `
		SELECT t.id, t.project_id, t.jira_ticket_id, t.jira_title, t.title, t.description, t.scheduled_date, t.deadline, t.status, t.created_at, t.updated_at, p.name
		FROM tasks t
		JOIN projects p ON t.project_id = p.id
		ORDER BY t.deadline ASC
	`
	
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks with projects: %w", err)
	}
	defer rows.Close()

	var tasks []*types.TaskWithProject
	for rows.Next() {
		task := &types.TaskWithProject{}
		err := rows.Scan(
			&task.ID,
			&task.ProjectID,
			&task.JiraTicketID,
			&task.JiraTitle,
			&task.Title,
			&task.Description,
			&task.ScheduledDate,
			&task.Deadline,
			&task.Status,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.ProjectName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task with project: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (m *TaskModel) GetTasksInDateRange(startDate, endDate time.Time) ([]*types.TaskWithProject, error) {
	query := `
		SELECT t.id, t.project_id, t.jira_ticket_id, t.jira_title, t.title, t.description, t.scheduled_date, t.deadline, t.status, t.created_at, t.updated_at, p.name
		FROM tasks t
		JOIN projects p ON t.project_id = p.id
		WHERE t.deadline BETWEEN ? AND ?
		ORDER BY t.deadline ASC
	`
	
	rows, err := m.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks in date range: %w", err)
	}
	defer rows.Close()

	var tasks []*types.TaskWithProject
	for rows.Next() {
		task := &types.TaskWithProject{}
		err := rows.Scan(
			&task.ID,
			&task.ProjectID,
			&task.JiraTicketID,
			&task.JiraTitle,
			&task.Title,
			&task.Description,
			&task.ScheduledDate,
			&task.Deadline,
			&task.Status,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.ProjectName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task with project: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (m *TaskModel) Update(task *types.Task) error {
	query := `
		UPDATE tasks
		SET project_id = ?, jira_ticket_id = ?, jira_title = ?, title = ?, description = ?, scheduled_date = ?, deadline = ?, status = ?, updated_at = ?
		WHERE id = ?
	`
	
	task.UpdatedAt = time.Now()
	_, err := m.db.Exec(query, task.ProjectID, task.JiraTicketID, task.JiraTitle, task.Title, task.Description, task.ScheduledDate, task.Deadline, task.Status, task.UpdatedAt, task.ID)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

func (m *TaskModel) UpdateStatus(id int64, status types.TaskStatus) error {
	query := `
		UPDATE tasks
		SET status = ?, updated_at = ?
		WHERE id = ?
	`
	
	now := time.Now()
	_, err := m.db.Exec(query, status, now, id)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	return nil
}

func (m *TaskModel) UpdateJiraTitle(id int64, jiraTitle string) error {
	query := `
		UPDATE tasks
		SET jira_title = ?, updated_at = ?
		WHERE id = ?
	`
	
	now := time.Now()
	_, err := m.db.Exec(query, jiraTitle, now, id)
	if err != nil {
		return fmt.Errorf("failed to update JIRA title: %w", err)
	}

	return nil
}

func (m *TaskModel) GetTasksGroupedByScheduledDate() ([]*types.TaskWithProject, error) {
	query := `
		SELECT t.id, t.project_id, t.jira_ticket_id, t.jira_title, t.title, t.description, t.scheduled_date, t.deadline, t.status, t.created_at, t.updated_at, p.name
		FROM tasks t
		JOIN projects p ON t.project_id = p.id
		ORDER BY 
			CASE WHEN t.scheduled_date IS NULL THEN 1 ELSE 0 END,
			t.scheduled_date ASC,
			t.created_at DESC
	`
	
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks grouped by scheduled date: %w", err)
	}
	defer rows.Close()

	var tasks []*types.TaskWithProject
	for rows.Next() {
		task := &types.TaskWithProject{}
		err := rows.Scan(
			&task.ID,
			&task.ProjectID,
			&task.JiraTicketID,
			&task.JiraTitle,
			&task.Title,
			&task.Description,
			&task.ScheduledDate,
			&task.Deadline,
			&task.Status,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.ProjectName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task with project: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (m *TaskModel) Delete(id int64) error {
	query := `DELETE FROM tasks WHERE id = ?`
	
	_, err := m.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	return nil
}