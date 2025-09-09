package types

import "time"

type RepositoryType string

const (
	MonorepoType    RepositoryType = "monorepo"
	KubernetesType  RepositoryType = "kubernetes"
)

type Repository struct {
	ID              int64          `json:"id" db:"id"`
	Name            string         `json:"name" db:"name"`
	URL             string         `json:"url" db:"url"`
	Type            RepositoryType `json:"type" db:"type"`
	Description     string         `json:"description" db:"description"`
	ServiceName     string         `json:"service_name,omitempty" db:"service_name"`
	ServiceLocation string         `json:"service_location,omitempty" db:"service_location"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
	LastSyncAt      *time.Time     `json:"last_sync_at" db:"last_sync_at"`
}

type Microservice struct {
	ID           int64     `json:"id" db:"id"`
	RepositoryID int64     `json:"repository_id" db:"repository_id"`
	Name         string    `json:"name" db:"name"`
	Path         string    `json:"path" db:"path"`
	Description  string    `json:"description" db:"description"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type KubernetesResource struct {
	ID           int64     `json:"id" db:"id"`
	RepositoryID int64     `json:"repository_id" db:"repository_id"`
	Name         string    `json:"name" db:"name"`
	Path         string    `json:"path" db:"path"`
	ResourceType string    `json:"resource_type" db:"resource_type"`
	Namespace    string    `json:"namespace" db:"namespace"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type ActionType string

const (
	BuildAction      ActionType = "build"
	DeploymentAction ActionType = "deployment"
)

type Action struct {
	ID            int64      `json:"id" db:"id"`
	RepositoryID  int64      `json:"repository_id" db:"repository_id"`
	ServiceID     *int64     `json:"service_id" db:"service_id"`
	ResourceID    *int64     `json:"resource_id" db:"resource_id"`
	Type          ActionType `json:"type" db:"type"`
	Status        string     `json:"status" db:"status"`
	WorkflowRunID int64      `json:"workflow_run_id" db:"workflow_run_id"`
	Commit        string     `json:"commit" db:"commit_sha"`
	Branch        string     `json:"branch" db:"branch"`
	BuildHash     string     `json:"build_hash" db:"build_hash"`
	StartedAt     time.Time  `json:"started_at" db:"started_at"`
	CompletedAt   *time.Time `json:"completed_at" db:"completed_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type ActionWithDetails struct {
	Action
	ServiceName  *string `json:"service_name,omitempty"`
	ResourceName *string `json:"resource_name,omitempty"`
}

type TaskStatus string

const (
	TaskPending    TaskStatus = "pending"
	TaskInProgress TaskStatus = "in_progress"
	TaskCompleted  TaskStatus = "completed"
)

type Project struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Task struct {
	ID            int64      `json:"id" db:"id"`
	ProjectID     int64      `json:"project_id" db:"project_id"`
	JiraTicketID  string     `json:"jira_ticket_id" db:"jira_ticket_id"`
	JiraTitle     string     `json:"jira_title" db:"jira_title"`
	Title         string     `json:"title" db:"title"`
	Description   string     `json:"description" db:"description"`
	ScheduledDate *time.Time `json:"scheduled_date" db:"scheduled_date"`
	Deadline      *time.Time `json:"deadline" db:"deadline"`
	Status        TaskStatus `json:"status" db:"status"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type TaskWithProject struct {
	Task
	ProjectName string `json:"project_name"`
}

type PullRequest struct {
	ID        int64     `json:"id"`
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	Author    string    `json:"author"`
	Branch    string    `json:"branch"`
	CreatedAt time.Time `json:"created_at"`
}

type Commit struct {
	Hash    string    `json:"hash"`
	Message string    `json:"message"`
	Author  string    `json:"author"`
	Date    time.Time `json:"date"`
}

type Deployment struct {
	ID                int64     `json:"id" db:"id"`
	ServiceID         int64     `json:"service_id" db:"service_id"`
	KubernetesRepoID  int64     `json:"kubernetes_repo_id" db:"kubernetes_repo_id"`
	CommitSHA         string    `json:"commit_sha" db:"commit_sha"`
	Environment       string    `json:"environment" db:"environment"`
	Region            string    `json:"region" db:"region"`
	Namespace         string    `json:"namespace" db:"namespace"`
	Tag               string    `json:"tag" db:"tag"`
	Path              string    `json:"path" db:"path"`
	DiscoveredAt      time.Time `json:"discovered_at" db:"discovered_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

type DeploymentOverview struct {
	CommitSHA            string    `json:"commit_sha"`
	Environment          string    `json:"environment"`
	Region               string    `json:"region"`
	Namespace            string    `json:"namespace"`
	Tag                  string    `json:"tag"`
	UpdatedAt            time.Time `json:"updated_at"`
	KubernetesRepoName   string    `json:"kubernetes_repo_name"`
}

type DeploymentStatus struct {
	Environment  string    `json:"environment"`
	Region       string    `json:"region"`
	Namespace    string    `json:"namespace"`
	Tag          string    `json:"tag"`
	IsDeployed   bool      `json:"is_deployed"`
	DeployedAt   time.Time `json:"deployed_at"`
}

type CommitDeploymentStatus struct {
	Commit        Commit             `json:"commit"`
	Deployments   []DeploymentStatus `json:"deployments"`
}