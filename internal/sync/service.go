package sync

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"gh-dashboard/internal/github"
	"gh-dashboard/internal/models"
	"gh-dashboard/pkg/types"
)

type Service struct {
	githubClient        *github.Client
	sshClient          *github.SSHClient
	repoModel          *models.RepositoryModel
	microserviceModel  *models.MicroserviceModel
	kubernetesModel    *models.KubernetesResourceModel
	actionModel        *models.ActionModel
	syncInterval       time.Duration
	ctx                context.Context
	cancelFunc         context.CancelFunc
}

type Config struct {
	GitHubToken  string
	SSHKeyPath   string
	SyncInterval time.Duration
}

func NewService(config Config, repoModel *models.RepositoryModel, microserviceModel *models.MicroserviceModel, kubernetesModel *models.KubernetesResourceModel, actionModel *models.ActionModel) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Initialize SSH client
	sshClient, err := github.NewSSHClient(config.GitHubToken, config.SSHKeyPath)
	if err != nil {
		log.Printf("Warning: Failed to initialize SSH client: %v", err)
	}
	
	return &Service{
		githubClient:       github.NewClient(config.GitHubToken),
		sshClient:         sshClient,
		repoModel:         repoModel,
		microserviceModel: microserviceModel,
		kubernetesModel:   kubernetesModel,
		actionModel:       actionModel,
		syncInterval:      config.SyncInterval,
		ctx:               ctx,
		cancelFunc:        cancel,
	}
}

func (s *Service) Start() {
	go func() {
		ticker := time.NewTicker(s.syncInterval)
		defer ticker.Stop()

		// Initial sync
		s.syncAll()

		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				s.syncAll()
			}
		}
	}()
}

func (s *Service) Stop() {
	s.cancelFunc()
}

func (s *Service) SyncRepository(repositoryID int64) error {
	repo, err := s.repoModel.GetByID(repositoryID)
	if err != nil {
		return fmt.Errorf("failed to get repository: %w", err)
	}

	owner, repoName, err := parseGitHubURL(repo.URL)
	if err != nil {
		return fmt.Errorf("invalid repository URL: %w", err)
	}

	switch repo.Type {
	case types.MonorepoType:
		return s.syncMonorepo(repo, owner, repoName)
	case types.KubernetesType:
		return s.syncKubernetesRepo(repo, owner, repoName)
	default:
		return fmt.Errorf("unknown repository type: %s", repo.Type)
	}
}

func (s *Service) syncAll() {
	repositories, err := s.repoModel.GetAll()
	if err != nil {
		log.Printf("Failed to get repositories for sync: %v", err)
		return
	}

	for _, repo := range repositories {
		if err := s.SyncRepository(repo.ID); err != nil {
			log.Printf("Failed to sync repository %s: %v", repo.Name, err)
			continue
		}

		if err := s.repoModel.UpdateLastSync(repo.ID); err != nil {
			log.Printf("Failed to update last sync time for repository %s: %v", repo.Name, err)
		}
	}
}

func (s *Service) syncMonorepo(repo *types.Repository, owner, repoName string) error {
	var services []github.ServiceInfo
	var err error

	// Use SSH client if available, otherwise fall back to GitHub API client
	if s.sshClient != nil {
		services, err = s.sshClient.DiscoverMicroservices(s.ctx, repo.URL, repo.ServiceName, repo.ServiceLocation)
	} else if s.githubClient != nil {
		services, err = s.githubClient.DiscoverMicroservices(s.ctx, owner, repoName)
	} else {
		return fmt.Errorf("no GitHub client available")
	}

	if err != nil {
		return fmt.Errorf("failed to discover microservices: %w", err)
	}

	// If no services discovered but we have specific service info, create one
	if len(services) == 0 && repo.ServiceName != "" && repo.ServiceLocation != "" {
		services = append(services, github.ServiceInfo{
			Name:        repo.ServiceName,
			Path:        repo.ServiceLocation,
			Description: fmt.Sprintf("Service %s located at %s", repo.ServiceName, repo.ServiceLocation),
		})
	}

	// Convert to types
	var microservices []types.Microservice
	for _, service := range services {
		microservices = append(microservices, types.Microservice{
			RepositoryID: repo.ID,
			Name:         service.Name,
			Path:         service.Path,
			Description:  service.Description,
		})
	}

	// Upsert microservices
	if err := s.microserviceModel.UpsertServices(repo.ID, microservices); err != nil {
		return fmt.Errorf("failed to upsert microservices: %w", err)
	}

	// Sync workflow runs for build and deployment actions
	if err := s.syncWorkflowRuns(repo, owner, repoName); err != nil {
		log.Printf("Failed to sync workflow runs for %s: %v", repo.Name, err)
	}

	return nil
}

func (s *Service) syncKubernetesRepo(repo *types.Repository, owner, repoName string) error {
	// Discover Kubernetes resources
	resources, err := s.githubClient.DiscoverKubernetesResources(s.ctx, owner, repoName)
	if err != nil {
		return fmt.Errorf("failed to discover kubernetes resources: %w", err)
	}

	// Convert to types
	var kubernetesResources []types.KubernetesResource
	for _, resource := range resources {
		kubernetesResources = append(kubernetesResources, types.KubernetesResource{
			RepositoryID: repo.ID,
			Name:         resource.Name,
			Path:         resource.Path,
			ResourceType: resource.ResourceType,
			Namespace:    resource.Namespace,
		})
	}

	// Upsert Kubernetes resources
	if err := s.kubernetesModel.UpsertResources(repo.ID, kubernetesResources); err != nil {
		return fmt.Errorf("failed to upsert kubernetes resources: %w", err)
	}

	// Sync workflow runs for deployment actions
	if err := s.syncWorkflowRuns(repo, owner, repoName); err != nil {
		log.Printf("Failed to sync workflow runs for %s: %v", repo.Name, err)
	}

	return nil
}

func (s *Service) syncWorkflowRuns(repo *types.Repository, owner, repoName string) error {
	// Get all workflows
	workflows, err := s.githubClient.ListWorkflows(s.ctx, owner, repoName)
	if err != nil {
		return fmt.Errorf("failed to list workflows: %w", err)
	}

	var actions []types.Action
	
	for _, workflow := range workflows {
		// Get recent workflow runs
		runs, err := s.githubClient.GetWorkflowRuns(s.ctx, owner, repoName, workflow.GetID(), 50)
		if err != nil {
			log.Printf("Failed to get workflow runs for %s: %v", workflow.GetName(), err)
			continue
		}

		for _, run := range runs {
			actionType := s.determineActionType(workflow.GetName())
			if actionType == "" {
				continue // Skip non-build/deploy workflows
			}

			action := types.Action{
				RepositoryID:  repo.ID,
				Type:          types.ActionType(actionType),
				Status:        run.Status,
				WorkflowRunID: run.ID,
				Commit:        run.Commit,
				Branch:        run.Branch,
				StartedAt:     run.StartedAt,
				CompletedAt:   run.CompletedAt,
			}

			// Try to match with services or resources based on workflow name or path
			if repo.Type == types.MonorepoType {
				serviceID := s.matchWorkflowToService(repo.ID, workflow.GetName(), run.Branch)
				if serviceID != 0 {
					action.ServiceID = &serviceID
				}
			} else if repo.Type == types.KubernetesType {
				resourceID := s.matchWorkflowToResource(repo.ID, workflow.GetName())
				if resourceID != 0 {
					action.ResourceID = &resourceID
				}
			}

			actions = append(actions, action)
		}
	}

	if len(actions) > 0 {
		if err := s.actionModel.UpsertActions(actions); err != nil {
			return fmt.Errorf("failed to upsert actions: %w", err)
		}
	}

	return nil
}

func (s *Service) determineActionType(workflowName string) string {
	workflowName = strings.ToLower(workflowName)
	
	if strings.Contains(workflowName, "build") || strings.Contains(workflowName, "ci") {
		return "build"
	}
	
	if strings.Contains(workflowName, "deploy") || strings.Contains(workflowName, "cd") {
		return "deployment"
	}
	
	return ""
}

func (s *Service) matchWorkflowToService(repositoryID int64, workflowName, branch string) int64 {
	services, err := s.microserviceModel.GetByRepositoryID(repositoryID)
	if err != nil {
		return 0
	}

	workflowName = strings.ToLower(workflowName)
	
	for _, service := range services {
		serviceName := strings.ToLower(service.Name)
		if strings.Contains(workflowName, serviceName) {
			return service.ID
		}
		
		// Check if branch contains service name (for feature branches)
		if strings.Contains(strings.ToLower(branch), serviceName) {
			return service.ID
		}
	}
	
	return 0
}

func (s *Service) matchWorkflowToResource(repositoryID int64, workflowName string) int64 {
	resources, err := s.kubernetesModel.GetByRepositoryID(repositoryID)
	if err != nil {
		return 0
	}

	workflowName = strings.ToLower(workflowName)
	
	for _, resource := range resources {
		resourceName := strings.ToLower(resource.Name)
		if strings.Contains(workflowName, resourceName) {
			return resource.ID
		}
	}
	
	return 0
}

func parseGitHubURL(repoURL string) (owner, repo string, err error) {
	u, err := url.Parse(repoURL)
	if err != nil {
		return "", "", err
	}

	// Handle both HTTPS and SSH URLs
	var pathStr string
	if u.Host == "github.com" {
		pathStr = u.Path
	} else if strings.Contains(repoURL, "git@github.com:") {
		// SSH URL format: git@github.com:owner/repo.git
		parts := strings.Split(repoURL, ":")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid SSH URL format")
		}
		pathStr = "/" + parts[1]
	} else {
		return "", "", fmt.Errorf("not a GitHub URL")
	}

	pathStr = strings.TrimPrefix(pathStr, "/")
	pathStr = strings.TrimSuffix(pathStr, ".git")
	
	parts := strings.Split(pathStr, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository path")
	}

	return parts[0], parts[1], nil
}