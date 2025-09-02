package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gh-dashboard/internal/database"
	"gh-dashboard/internal/models"
	"gh-dashboard/internal/sync"
	"gh-dashboard/pkg/types"
)

// App struct
type App struct {
	ctx             context.Context
	db              *database.DB
	repoModel       *models.RepositoryModel
	serviceModel    *models.MicroserviceModel
	kubernetesModel *models.KubernetesResourceModel
	actionModel     *models.ActionModel
	projectModel    *models.ProjectModel
	taskModel       *models.TaskModel
	syncService     *sync.Service
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	log.Println("GitHub Dashboard starting up...")
	
	// Initialize database
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Failed to get user home directory: %v", err)
		// Continue without database for now
		return
	}
	
	dbPath := filepath.Join(homeDir, ".gh-dashboard", "database.db")
	log.Printf("Initializing database at: %s", dbPath)
	
	db, err := database.NewDB(dbPath)
	if err != nil {
		log.Printf("Failed to initialize database: %v", err)
		log.Println("Continuing without database - some features may not work")
		// Continue without database - the UI should still load
		return
	}
	
	log.Println("Database initialized successfully")
	a.db = db
	a.repoModel = models.NewRepositoryModel(db.GetConn())
	a.serviceModel = models.NewMicroserviceModel(db.GetConn())
	a.kubernetesModel = models.NewKubernetesResourceModel(db.GetConn())
	a.actionModel = models.NewActionModel(db.GetConn())
	a.projectModel = models.NewProjectModel(db.GetConn())
	a.taskModel = models.NewTaskModel(db.GetConn())
	
	// Initialize sync service with GitHub token and SSH key from environment
	githubToken := os.Getenv("GITHUB_TOKEN")
	sshKeyPath := os.Getenv("SSH_KEY_PATH")
	
	if githubToken != "" || sshKeyPath != "" {
		syncConfig := sync.Config{
			GitHubToken:  githubToken,
			SSHKeyPath:   sshKeyPath,
			SyncInterval: 5 * time.Minute,
		}
		
		a.syncService = sync.NewService(syncConfig, a.repoModel, a.serviceModel, a.kubernetesModel, a.actionModel)
		a.syncService.Start()
		log.Println("Background sync service started")
	} else {
		log.Println("Warning: Neither GITHUB_TOKEN nor SSH_KEY_PATH is set, sync functionality disabled")
	}
	
	log.Println("GitHub Dashboard startup completed successfully")
}

// Repository Management Methods

func (a *App) GetRepositories() ([]*types.Repository, error) {
	if a.repoModel == nil {
		return []*types.Repository{}, nil
	}
	return a.repoModel.GetAll()
}

func (a *App) CreateRepository(repo types.Repository) error {
	return a.repoModel.Create(&repo)
}

func (a *App) UpdateRepository(repo types.Repository) error {
	return a.repoModel.Update(&repo)
}

func (a *App) DeleteRepository(id int64) error {
	return a.repoModel.Delete(id)
}

func (a *App) SyncRepository(id int64) error {
	if a.syncService == nil {
		return fmt.Errorf("sync service not initialized - GitHub token required")
	}
	return a.syncService.SyncRepository(id)
}

// Microservice Management Methods

func (a *App) GetMicroservices(repositoryID int64) ([]*types.Microservice, error) {
	if repositoryID == 0 {
		// Return all microservices from all repositories
		repos, err := a.repoModel.GetAll()
		if err != nil {
			return nil, err
		}
		
		var allServices []*types.Microservice
		for _, repo := range repos {
			if repo.Type == types.MonorepoType {
				services, err := a.serviceModel.GetByRepositoryID(repo.ID)
				if err != nil {
					continue
				}
				allServices = append(allServices, services...)
			}
		}
		return allServices, nil
	}
	
	return a.serviceModel.GetByRepositoryID(repositoryID)
}

func (a *App) GetMicroserviceActions(serviceID int64, limit int) ([]*types.Action, error) {
	if limit == 0 {
		limit = 50
	}
	return a.actionModel.GetByServiceID(serviceID, limit)
}

// Kubernetes Resource Management Methods

func (a *App) GetKubernetesResources(repositoryID int64) ([]*types.KubernetesResource, error) {
	if repositoryID == 0 {
		// Return all resources from all repositories
		repos, err := a.repoModel.GetAll()
		if err != nil {
			return nil, err
		}
		
		var allResources []*types.KubernetesResource
		for _, repo := range repos {
			if repo.Type == types.KubernetesType {
				resources, err := a.kubernetesModel.GetByRepositoryID(repo.ID)
				if err != nil {
					continue
				}
				allResources = append(allResources, resources...)
			}
		}
		return allResources, nil
	}
	
	return a.kubernetesModel.GetByRepositoryID(repositoryID)
}

func (a *App) GetKubernetesResourceActions(resourceID int64, limit int) ([]*types.Action, error) {
	if limit == 0 {
		limit = 50
	}
	return a.actionModel.GetByResourceID(resourceID, limit)
}

// Action Management Methods

func (a *App) GetRecentActions(repositoryID int64, limit int) ([]*types.ActionWithDetails, error) {
	if limit == 0 {
		limit = 50
	}
	return a.actionModel.GetByRepositoryID(repositoryID, limit)
}

// Dashboard Statistics

func (a *App) GetDashboardStats() (map[string]interface{}, error) {
	if a.repoModel == nil {
		return map[string]interface{}{
			"repositories":       0,
			"microservices":      0,
			"kubernetesResources": 0,
			"recentActions":      []*types.ActionWithDetails{},
		}, nil
	}
	
	repos, err := a.repoModel.GetAll()
	if err != nil {
		return nil, err
	}
	
	var totalServices, totalResources int
	var recentActions []*types.ActionWithDetails
	
	for _, repo := range repos {
		if repo.Type == types.MonorepoType {
			services, err := a.serviceModel.GetByRepositoryID(repo.ID)
			if err == nil {
				totalServices += len(services)
			}
		} else if repo.Type == types.KubernetesType {
			resources, err := a.kubernetesModel.GetByRepositoryID(repo.ID)
			if err == nil {
				totalResources += len(resources)
			}
		}
		
		// Get recent actions for this repo
		actions, err := a.actionModel.GetByRepositoryID(repo.ID, 10)
		if err == nil {
			recentActions = append(recentActions, actions...)
		}
	}
	
	// Sort recent actions by timestamp (most recent first)
	// This is a simple bubble sort for demonstration
	for i := 0; i < len(recentActions)-1; i++ {
		for j := 0; j < len(recentActions)-i-1; j++ {
			if recentActions[j].StartedAt.Before(recentActions[j+1].StartedAt) {
				recentActions[j], recentActions[j+1] = recentActions[j+1], recentActions[j]
			}
		}
	}
	
	// Limit to 10 most recent
	if len(recentActions) > 10 {
		recentActions = recentActions[:10]
	}
	
	return map[string]interface{}{
		"repositories":       len(repos),
		"microservices":      totalServices,
		"kubernetesResources": totalResources,
		"recentActions":      recentActions,
	}, nil
}

// Project Management Methods

func (a *App) GetProjects() ([]*types.Project, error) {
	if a.projectModel == nil {
		return []*types.Project{}, nil
	}
	return a.projectModel.GetAll()
}

func (a *App) GetProject(id int64) (*types.Project, error) {
	if a.projectModel == nil {
		return nil, fmt.Errorf("project model not initialized")
	}
	return a.projectModel.GetByID(id)
}

func (a *App) CreateProject(project types.Project) error {
	if a.projectModel == nil {
		return fmt.Errorf("project model not initialized")
	}
	return a.projectModel.Create(&project)
}

func (a *App) UpdateProject(project types.Project) error {
	if a.projectModel == nil {
		return fmt.Errorf("project model not initialized")
	}
	return a.projectModel.Update(&project)
}

func (a *App) DeleteProject(id int64) error {
	if a.projectModel == nil {
		return fmt.Errorf("project model not initialized")
	}
	return a.projectModel.Delete(id)
}

// Task Management Methods

func (a *App) GetTasks() ([]*types.TaskWithProject, error) {
	if a.taskModel == nil {
		return []*types.TaskWithProject{}, nil
	}
	return a.taskModel.GetAllWithProjects()
}

func (a *App) GetTasksByProject(projectID int64) ([]*types.Task, error) {
	if a.taskModel == nil {
		return []*types.Task{}, nil
	}
	return a.taskModel.GetByProjectID(projectID)
}

func (a *App) GetTask(id int64) (*types.Task, error) {
	if a.taskModel == nil {
		return nil, fmt.Errorf("task model not initialized")
	}
	return a.taskModel.GetByID(id)
}

func (a *App) CreateTask(task types.Task) error {
	if a.taskModel == nil {
		return fmt.Errorf("task model not initialized")
	}
	return a.taskModel.Create(&task)
}

func (a *App) UpdateTask(task types.Task) error {
	if a.taskModel == nil {
		return fmt.Errorf("task model not initialized")
	}
	return a.taskModel.Update(&task)
}

func (a *App) UpdateTaskStatus(id int64, status types.TaskStatus) error {
	if a.taskModel == nil {
		return fmt.Errorf("task model not initialized")
	}
	return a.taskModel.UpdateStatus(id, status)
}

func (a *App) DeleteTask(id int64) error {
	if a.taskModel == nil {
		return fmt.Errorf("task model not initialized")
	}
	return a.taskModel.Delete(id)
}

func (a *App) GetTasksInDateRange(startDate, endDate time.Time) ([]*types.TaskWithProject, error) {
	if a.taskModel == nil {
		return []*types.TaskWithProject{}, nil
	}
	return a.taskModel.GetTasksInDateRange(startDate, endDate)
}

func (a *App) GetTasksGroupedByScheduledDate() ([]*types.TaskWithProject, error) {
	if a.taskModel == nil {
		return []*types.TaskWithProject{}, nil
	}
	return a.taskModel.GetTasksGroupedByScheduledDate()
}

// Greet returns a greeting for the given name (keeping original method for compatibility)
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}