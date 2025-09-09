package sync

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"

	"dev-dashboard/internal/github"
	"dev-dashboard/internal/kubernetes"
	"dev-dashboard/internal/models"
	"dev-dashboard/pkg/types"
	
	goGithub "github.com/google/go-github/v57/github"
)

type Service struct {
	githubClient        *github.Client
	repoModel          *models.RepositoryModel
	microserviceModel  *models.MicroserviceModel
	kubernetesModel    *models.KubernetesResourceModel
	actionModel        *models.ActionModel
	deploymentModel    *models.DeploymentModel
	kubernetesScanner  *kubernetes.Scanner
	syncInterval       time.Duration
	ctx                context.Context
	cancelFunc         context.CancelFunc
}

type Config struct {
	GitHubToken       string
	GitHubEnterpriseURL string
	SyncInterval      time.Duration
}

func NewService(config Config, repoModel *models.RepositoryModel, microserviceModel *models.MicroserviceModel, kubernetesModel *models.KubernetesResourceModel, actionModel *models.ActionModel, deploymentModel *models.DeploymentModel) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Service{
		githubClient:       github.NewClientWithBaseURL(config.GitHubToken, config.GitHubEnterpriseURL),
		repoModel:         repoModel,
		microserviceModel: microserviceModel,
		kubernetesModel:   kubernetesModel,
		actionModel:       actionModel,
		deploymentModel:   deploymentModel,
		kubernetesScanner: kubernetes.NewScanner(),
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

	// Use GitHub API client for service discovery
	if s.githubClient != nil {
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

	// Upsert microservices preserving existing IDs
	if err := s.microserviceModel.UpsertServicesPreserveID(repo.ID, microservices); err != nil {
		return fmt.Errorf("failed to upsert microservices: %w", err)
	}

	// Sync workflow runs for build and deployment actions
	if err := s.syncWorkflowRuns(repo, owner, repoName); err != nil {
		log.Printf("Failed to sync workflow runs for %s: %v", repo.Name, err)
	}

	return nil
}

func (s *Service) syncKubernetesRepo(repo *types.Repository, owner, repoName string) error {
	// Scan for real deployment data using GitHub API
	if s.githubClient != nil {
		log.Printf("Scanning kustomization files for Kubernetes repo: %s", repo.Name)
		
		// Use GitHub API to scan for kustomization.yaml files
		kustomizationDeployments, err := s.githubClient.ScanKustomizationFiles(s.ctx, owner, repoName)
		if err != nil {
			log.Printf("Failed to scan kustomization files in %s: %v", repo.Name, err)
		} else {
			log.Printf("Found %d kustomization deployments in %s", len(kustomizationDeployments), repo.Name)
			
			// Get all microservices to match with deployments
			allServices, err := s.microserviceModel.GetAll()
			if err != nil {
				log.Printf("Failed to get services for deployment matching: %v", err)
			} else {
				// Convert GitHub API results to deployment records
				for _, kustomDeploy := range kustomizationDeployments {
					// Find matching service by name
					var serviceID int64
					for _, service := range allServices {
						if strings.Contains(strings.ToLower(service.Name), strings.ToLower(kustomDeploy.ServiceName)) ||
						   strings.Contains(strings.ToLower(kustomDeploy.ServiceName), strings.ToLower(service.Name)) {
							serviceID = service.ID
							break
						}
					}
					
					if serviceID == 0 {
						log.Printf("No matching service found for %s, skipping", kustomDeploy.ServiceName)
						continue
					}
					
					// Try to correlate tag with actual monorepo commit
					var commitSHA string
					// Check if tag is already a commit SHA (40 hex characters)
					if len(kustomDeploy.Tag) == 40 && isHexString(kustomDeploy.Tag) {
						// Tag is likely a commit SHA, use it directly
						commitSHA = kustomDeploy.Tag
						log.Printf("Using tag as commit SHA for service %s: %s", kustomDeploy.ServiceName, kustomDeploy.Tag)
					} else {
						// Try to correlate tag with actual monorepo commit
						commitSHA = s.correlateTagWithCommit(serviceID, kustomDeploy.Tag)
						if commitSHA == "" {
							commitSHA = kustomDeploy.CommitSHA // Fallback to k8s repo commit
						}
					}

					deployment := &types.Deployment{
						ServiceID:        serviceID,
						KubernetesRepoID: repo.ID,
						CommitSHA:        commitSHA,
						Environment:      kustomDeploy.Environment,
						Region:           kustomDeploy.Region,
						Namespace:        kustomDeploy.Namespace,
						Tag:              kustomDeploy.Tag,
						Path:             kustomDeploy.Path,
					}
					
					if err := s.deploymentModel.Upsert(deployment); err != nil {
						log.Printf("Failed to upsert deployment: %v", err)
					} else {
						log.Printf("Upserted deployment for service %s (%d) in %s/%s with tag %s", 
							kustomDeploy.ServiceName, serviceID, kustomDeploy.Environment, kustomDeploy.Region, kustomDeploy.Tag)
					}
				}
			}
		}
	} else {
		log.Printf("No GitHub client available for scanning %s", repo.Name)
	}

	// Discover Kubernetes resources
	rootPath := repo.ServiceLocation // Use service_location as root path for Kubernetes repos too
	if rootPath == "" {
		log.Printf("No root path specified for Kubernetes repository %s, using default discovery", repo.Name)
	} else {
		log.Printf("Using root path '%s' for Kubernetes repository %s", rootPath, repo.Name)
	}
	
	resources, err := s.githubClient.DiscoverKubernetesResourcesInPath(s.ctx, owner, repoName, rootPath)
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

	// Handle HTTPS URLs only
	var pathStr string
	if u.Host == "github.com" {
		pathStr = u.Path
	} else {
		return "", "", fmt.Errorf("only HTTPS GitHub URLs are supported")
	}

	pathStr = strings.TrimPrefix(pathStr, "/")
	pathStr = strings.TrimSuffix(pathStr, ".git")
	
	parts := strings.Split(pathStr, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository path")
	}

	return parts[0], parts[1], nil
}

// correlateTagWithCommit attempts to find the monorepo commit that corresponds to a deployment tag
func (s *Service) correlateTagWithCommit(serviceID int64, tag string) string {
	// Get the service to find its monorepo
	service, err := s.microserviceModel.GetByID(serviceID)
	if err != nil {
		log.Printf("Failed to get service %d: %v", serviceID, err)
		return ""
	}

	// Get the monorepo details
	repo, err := s.repoModel.GetByID(service.RepositoryID)
	if err != nil {
		log.Printf("Failed to get repository %d: %v", service.RepositoryID, err)
		return ""
	}

	// Only process monorepo type repositories
	if repo.Type != types.MonorepoType {
		return ""
	}

	// Parse GitHub URL to get owner and repo name
	owner, repoName, err := parseGitHubURL(repo.URL)
	if err != nil {
		log.Printf("Failed to parse repo URL %s: %v", repo.URL, err)
		return ""
	}

	// Search for commits that might match this tag
	// This is a simple heuristic - in production you might want more sophisticated matching
	if s.githubClient != nil {
		// Try to find a commit message or tag that references this release
		// Look for commits in the service path that might correspond to the tag
		commitOpts := &goGithub.CommitsListOptions{
			Path: service.Path,
			ListOptions: goGithub.ListOptions{PerPage: 50},
		}

		commits, _, err := s.githubClient.GetGitHubClient().Repositories.ListCommits(s.ctx, owner, repoName, commitOpts)
		if err != nil {
			log.Printf("Failed to get commits for service %s: %v", service.Name, err)
			return ""
		}

		// Look for commits that might match the tag
		for _, commit := range commits {
			if commit.SHA == nil || commit.Commit == nil || commit.Commit.Message == nil {
				continue
			}

			message := *commit.Commit.Message
			sha := *commit.SHA

			// Simple matching logic - look for tag reference in commit message
			if strings.Contains(strings.ToLower(message), strings.ToLower(tag)) {
				log.Printf("Found matching commit %s for tag %s: %s", sha[:7], tag, message)
				return sha
			}

			// Also check if the tag format matches common patterns
			if strings.Contains(tag, "release-") {
				version := strings.TrimPrefix(tag, "release-")
				if strings.Contains(strings.ToLower(message), version) {
					log.Printf("Found version matching commit %s for tag %s: %s", sha[:7], tag, message)
					return sha
				}
			}
		}

		// Try to find Git tags in the repository that match
		tags, _, err := s.githubClient.GetGitHubClient().Repositories.ListTags(s.ctx, owner, repoName, nil)
		if err == nil {
			for _, gitTag := range tags {
				if gitTag.Name != nil && gitTag.Commit != nil && gitTag.Commit.SHA != nil {
					if strings.EqualFold(*gitTag.Name, tag) {
						log.Printf("Found exact git tag match for %s: %s", tag, *gitTag.Commit.SHA)
						return *gitTag.Commit.SHA
					}
				}
			}
		}
	}

	log.Printf("No commit correlation found for tag %s in service %s", tag, service.Name)
	return ""
}

// isHexString checks if a string contains only hexadecimal characters
func isHexString(s string) bool {
	hexPattern := regexp.MustCompile(`^[a-fA-F0-9]+$`)
	return hexPattern.MatchString(s)
}