package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dev-dashboard/internal/database"
	"dev-dashboard/internal/github"
	"dev-dashboard/internal/jira"
	"dev-dashboard/internal/models"
	"dev-dashboard/internal/sync"
	"dev-dashboard/pkg/types"
	
	goGithub "github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// App struct
type App struct {
	ctx             context.Context
	db              *database.DB
	repoModel       *models.RepositoryModel
	serviceModel    *models.MicroserviceModel
	kubernetesModel *models.KubernetesResourceModel
	actionModel     *models.ActionModel
	deploymentModel *models.DeploymentModel
	projectModel    *models.ProjectModel
	taskModel       *models.TaskModel
	configModel     *models.ConfigModel
	jiraClient      *jira.Client
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
	log.Println("Dev Dashboard starting up...")
	
	// Initialize database
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Failed to get user home directory: %v", err)
		// Continue without database for now
		return
	}
	
	dbPath := filepath.Join(homeDir, ".dev-dashboard", "database.db")
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
	a.deploymentModel = models.NewDeploymentModel(db.GetConn())
	a.projectModel = models.NewProjectModel(db.GetConn())
	a.taskModel = models.NewTaskModel(db.GetConn())
	a.configModel = models.NewConfigModel(db.GetConn())
	
	// Initialize JIRA client if configured
	a.initJiraClient()
	
	// Initialize sync service with GitHub token from config
	githubToken := a.getGitHubToken()
	
	if githubToken != "" {
		syncConfig := sync.Config{
			GitHubToken:         githubToken,
			GitHubEnterpriseURL: a.getGitHubEnterpriseURL(),
			SyncInterval:        5 * time.Minute,
		}
		
		a.syncService = sync.NewService(syncConfig, a.repoModel, a.serviceModel, a.kubernetesModel, a.actionModel, a.deploymentModel)
		a.syncService.Start()
		log.Println("Background sync service started")
	} else {
		log.Println("Warning: GITHUB_TOKEN not configured, sync functionality disabled")
	}
	
	log.Println("Dev Dashboard startup completed successfully")
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

func (a *App) CreateRepositoryWithAuth(repoData map[string]interface{}) error {
	repo := types.Repository{
		Name:            repoData["name"].(string),
		URL:             repoData["url"].(string),
		Type:            types.RepositoryType(repoData["type"].(string)),
		Description:     repoData["description"].(string),
		ServiceLocation: repoData["service_location"].(string),
	}

	// Create repository first
	err := a.repoModel.Create(&repo)
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	// If it's a monorepo, discover and create services
	if repo.Type == types.MonorepoType {
		log.Printf("Repository is monorepo type, starting service discovery for %s", repo.Name)
		authMethod := repoData["auth_method"].(string)
		credentials := repoData["credentials"].(map[string]interface{})
		
		log.Printf("Auth method: %s, Service location: %s", authMethod, repo.ServiceLocation)
		
		services, err := a.discoverServices(repo.URL, repo.ServiceLocation, authMethod, credentials)
		if err != nil {
			log.Printf("ERROR: Failed to discover services for repository %s: %v", repo.Name, err)
		} else {
			log.Printf("Successfully discovered %d services for repository %s", len(services), repo.Name)
			// Create discovered services
			for _, service := range services {
				log.Printf("Creating microservice: %s at path %s", service.Name, service.Path)
				microservice := types.Microservice{
					RepositoryID: repo.ID,
					Name:         service.Name,
					Path:         service.Path,
					Description:  service.Description,
				}
				err := a.serviceModel.Create(&microservice)
				if err != nil {
					log.Printf("ERROR: Failed to create microservice %s: %v", service.Name, err)
				} else {
					log.Printf("Successfully created microservice %s", service.Name)
				}
			}
		}
	} else {
		log.Printf("Repository %s is type %s, skipping service discovery", repo.Name, repo.Type)
	}

	return nil
}

func (a *App) ValidateRepositoryAccess(url, authMethod string, credentials map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{
		"success": false,
		"error":   "",
	}

	ctx := context.Background()

	if authMethod == "pat" {
		token, ok := credentials["githubToken"].(string)
		if !ok || token == "" {
			// Use globally configured GitHub token
			token = a.getGitHubToken()
			if token == "" {
				result["error"] = "GitHub token is required - please configure it in Settings"
				return result
			}
		}

		// Extract owner and repo from URL
		owner, repoName, err := a.parseGitHubURL(url)
		if err != nil {
			result["error"] = fmt.Sprintf("Invalid GitHub URL: %v", err)
			return result
		}

		// Test GitHub API access
		client := a.createGitHubClient(token)
		_, _, err = client.Repositories.Get(ctx, owner, repoName)
		if err != nil {
			result["error"] = fmt.Sprintf("Cannot access repository: %v", err)
			return result
		}

		result["success"] = true
	} else {
		result["error"] = "Only GitHub Personal Access Token authentication is supported"
	}

	return result
}

func (a *App) DiscoverRepositoryServices(url, serviceLocation, authMethod string, credentials map[string]interface{}) []map[string]interface{} {
	services := []map[string]interface{}{}

	ctx := context.Background()

	if authMethod == "pat" {
		token, ok := credentials["githubToken"].(string)
		if !ok || token == "" {
			// Use globally configured GitHub token
			token = a.getGitHubToken()
			if token == "" {
				return services // Return empty services if no token configured
			}
		}

		// Create GitHub client with Enterprise support
		enterpriseURL := a.getGitHubEnterpriseURL()
		githubClient := github.NewClientWithBaseURL(token, enterpriseURL)
		
		owner, repo, err := githubClient.ParseRepositoryURL(url)
		if err != nil {
			return services
		}

		discoveredServices, err := githubClient.DiscoverMicroservicesInPath(ctx, owner, repo, serviceLocation)
		if err != nil {
			log.Printf("Failed to discover services: %v", err)
			return services
		}

		for _, service := range discoveredServices {
			services = append(services, map[string]interface{}{
				"name":        service.Name,
				"path":        service.Path,
				"description": service.Description,
			})
		}
	} else {
		log.Printf("Only GitHub PAT authentication is supported, got: %s", authMethod)
	}

	return services
}

// Helper methods for repository operations
func (a *App) parseGitHubURL(url string) (owner, repo string, err error) {
	// Create a GitHub client to use its URL parsing capabilities
	githubToken := a.getGitHubToken()
	enterpriseURL := a.getGitHubEnterpriseURL()
	
	githubClient := github.NewClientWithBaseURL(githubToken, enterpriseURL)
	return githubClient.ParseRepositoryURL(url)
}

func (a *App) createGitHubClient(token string) *goGithub.Client {
	// Get Enterprise configuration
	enterpriseURL := a.getGitHubEnterpriseURL()
	
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	
	if enterpriseURL != "" {
		client, err := goGithub.NewEnterpriseClient(enterpriseURL, enterpriseURL, tc)
		if err != nil {
			log.Printf("Failed to create Enterprise GitHub client: %v", err)
			return goGithub.NewClient(tc)
		}
		return client
	}
	
	return goGithub.NewClient(tc)
}


func (a *App) discoverServices(url, serviceLocation, authMethod string, credentials map[string]interface{}) ([]github.ServiceInfo, error) {
	ctx := context.Background()

	log.Printf("Starting service discovery for %s using %s auth", url, authMethod)

	if authMethod == "pat" {
		token, ok := credentials["githubToken"].(string)
		if !ok || token == "" {
			// Use globally configured GitHub token
			token = a.getGitHubToken()
			if token == "" {
				log.Printf("ERROR: GitHub token not configured globally or provided in credentials")
				return nil, fmt.Errorf("GitHub token is required - please configure it in Settings")
			}
		}

		log.Printf("Using GitHub PAT authentication (token length: %d)", len(token))

		// Create GitHub client with Enterprise support
		enterpriseURL := a.getGitHubEnterpriseURL()
		githubClient := github.NewClientWithBaseURL(token, enterpriseURL)
		
		owner, repo, err := githubClient.ParseRepositoryURL(url)
		if err != nil {
			log.Printf("ERROR: Failed to parse GitHub URL %s: %v", url, err)
			return nil, err
		}

		log.Printf("Parsed GitHub URL - Owner: %s, Repo: %s, Service location: %s", owner, repo, serviceLocation)

		log.Printf("Created GitHub client, calling DiscoverMicroservicesInPath...")
		
		services, err := githubClient.DiscoverMicroservicesInPath(ctx, owner, repo, serviceLocation)
		if err != nil {
			log.Printf("ERROR: DiscoverMicroservicesInPath failed: %v", err)
			return nil, err
		}
		
		log.Printf("DiscoverMicroservicesInPath returned %d services", len(services))
		for i, service := range services {
			log.Printf("  Service %d: Name=%s, Path=%s, Description=%s", i+1, service.Name, service.Path, service.Description)
		}
		
		return services, nil
	}

	return nil, fmt.Errorf("only GitHub PAT authentication is supported, got: %s", authMethod)
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

func (a *App) RediscoverRepositoryServices(id int64, authMethod string, credentials map[string]interface{}) error {
	// Get the repository
	repo, err := a.repoModel.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get repository: %w", err)
	}

	if repo.Type != types.MonorepoType {
		return fmt.Errorf("repository is not a monorepo")
	}

	log.Printf("Rediscovering services for repository %s (%s)", repo.Name, repo.URL)

	// Only support PAT authentication
	if authMethod != "pat" {
		return fmt.Errorf("only GitHub PAT authentication is supported")
	}

	// Discover services using the provided credentials
	discoveredServices, err := a.discoverServices(repo.URL, repo.ServiceLocation, authMethod, credentials)
	if err != nil {
		return fmt.Errorf("failed to discover services: %w", err)
	}

	log.Printf("Discovered %d services for repository %s", len(discoveredServices), repo.Name)

	// Convert to microservice types
	var microservices []types.Microservice
	for _, service := range discoveredServices {
		microservices = append(microservices, types.Microservice{
			RepositoryID: repo.ID,
			Name:         service.Name,
			Path:         service.Path,
			Description:  service.Description,
		})
	}

	// Upsert services preserving existing IDs
	err = a.serviceModel.UpsertServicesPreserveID(id, microservices)
	if err != nil {
		return fmt.Errorf("failed to upsert services: %w", err)
	}

	log.Printf("Successfully updated services for repository %s", repo.Name)

	return nil
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
			// Only include services from actual monorepo repositories (exclude kubernetes repositories)
			if repo.Type == types.MonorepoType && !a.isKubernetesRepository(repo) {
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

// GetServicePullRequests returns service-specific pull requests from GitHub
func (a *App) GetServicePullRequests(serviceID int64) ([]*types.PullRequest, error) {
	// Get service details
	service, err := a.serviceModel.GetByID(serviceID)
	if err != nil {
		return nil, err
	}
	
	// Get repository details
	repo, err := a.repoModel.GetByID(service.RepositoryID)
	if err != nil {
		return nil, err
	}
	
	// Create GitHub client if we have a token
	githubToken := a.getGitHubToken()
	if githubToken == "" {
		return []*types.PullRequest{}, nil // Return empty list if no token
	}
	
	ctx := context.Background()
	client := a.createGitHubClient(githubToken)
	
	// Parse repository URL to get owner and repo name
	owner, repoName := parseRepositoryURL(repo.URL)
	if owner == "" || repoName == "" {
		return []*types.PullRequest{}, nil
	}
	
	// Get pull requests
	log.Printf("Fetching PRs for %s/%s, service path: %s", owner, repoName, service.Path)
	prs, _, err := client.PullRequests.List(ctx, owner, repoName, &goGithub.PullRequestListOptions{
		State: "all",
		ListOptions: goGithub.ListOptions{PerPage: 50},
	})
	if err != nil {
		log.Printf("Failed to fetch pull requests for %s/%s: %v", owner, repoName, err)
		return []*types.PullRequest{}, nil
	}
	
	log.Printf("Found %d total PRs for repository %s/%s", len(prs), owner, repoName)
	
	// Filter PRs that affect the service directory
	var servicePRs []*types.PullRequest
	for _, pr := range prs {
		if pr == nil || pr.Number == nil {
			continue
		}
		
		// Get files changed in this PR
		files, _, err := client.PullRequests.ListFiles(ctx, owner, repoName, *pr.Number, nil)
		if err != nil {
			continue
		}
		
		// Check if any files in the service directory were changed
		serviceAffected := false
		for _, file := range files {
			if file.Filename != nil && strings.HasPrefix(*file.Filename, service.Path) {
				serviceAffected = true
				break
			}
		}
		
		if serviceAffected {
			status := "open"
			if pr.State != nil {
				status = *pr.State
			}
			if pr.Merged != nil && *pr.Merged {
				status = "merged"
			}
			
			author := ""
			if pr.User != nil && pr.User.Login != nil {
				author = *pr.User.Login
			}
			
			title := ""
			if pr.Title != nil {
				title = *pr.Title
			}
			
			branch := ""
			if pr.Head != nil && pr.Head.Ref != nil {
				branch = *pr.Head.Ref
			}
			
			createdAt := time.Now()
			if pr.CreatedAt != nil {
				createdAt = pr.CreatedAt.Time
			}
			
			servicePRs = append(servicePRs, &types.PullRequest{
				ID:        int64(*pr.Number),
				Number:    *pr.Number,
				Title:     title,
				Status:    status,
				Author:    author,
				Branch:    branch,
				CreatedAt: createdAt,
			})
		}
	}
	
	return servicePRs, nil
}

// GetServiceCommits returns service-specific commit history from GitHub
func (a *App) GetServiceCommits(serviceID int64) ([]*types.Commit, error) {
	// Get service details
	service, err := a.serviceModel.GetByID(serviceID)
	if err != nil {
		return nil, err
	}
	
	// Get repository details
	repo, err := a.repoModel.GetByID(service.RepositoryID)
	if err != nil {
		return nil, err
	}
	
	// Create GitHub client if we have a token
	githubToken := a.getGitHubToken()
	if githubToken == "" {
		return []*types.Commit{}, nil // Return empty list if no token
	}
	
	ctx := context.Background()
	client := a.createGitHubClient(githubToken)
	
	// Parse repository URL to get owner and repo name
	owner, repoName := parseRepositoryURL(repo.URL)
	if owner == "" || repoName == "" {
		return []*types.Commit{}, nil
	}
	
	// Get commits for the service directory
	log.Printf("Fetching commits for %s/%s path: %s", owner, repoName, service.Path)
	commits, _, err := client.Repositories.ListCommits(ctx, owner, repoName, &goGithub.CommitsListOptions{
		Path: service.Path,
		ListOptions: goGithub.ListOptions{PerPage: 50},
	})
	if err != nil {
		log.Printf("Failed to fetch commits for %s/%s path %s: %v", owner, repoName, service.Path, err)
		return []*types.Commit{}, nil
	}
	
	// Also get deployment commits that might not have touched the service path
	// but are specifically for this service
	deployments, err := a.deploymentModel.GetByServiceID(serviceID)
	if err == nil && len(deployments) > 0 {
		commitSHASet := make(map[string]bool)
		for _, commit := range commits {
			if commit.SHA != nil {
				commitSHASet[*commit.SHA] = true
			}
		}
		
		// Add deployment commits for this specific service that aren't already in the list
		for _, deployment := range deployments {
			if deployment.CommitSHA != "" && !commitSHASet[deployment.CommitSHA] {
				// Fetch this specific commit
				commit, _, err := client.Repositories.GetCommit(ctx, owner, repoName, deployment.CommitSHA, nil)
				if err != nil {
					log.Printf("Failed to fetch deployment commit %s: %v", deployment.CommitSHA, err)
					continue
				}
				commits = append(commits, commit)
				log.Printf("Added deployment commit %s to service %s commits", deployment.CommitSHA[:7], service.Name)
			}
		}
	}
	
	log.Printf("Found %d total commits for service %s", len(commits), service.Name)
	
	// Log all commit SHAs for debugging
	for i, commit := range commits {
		if commit != nil && commit.SHA != nil {
			log.Printf("Commit %d: %s", i, (*commit.SHA)[:7])
		}
	}
	
	// Convert to our types
	var serviceCommits []*types.Commit
	for _, commit := range commits {
		if commit == nil || commit.SHA == nil {
			continue
		}
		
		message := ""
		author := ""
		date := time.Now()
		
		if commit.Commit != nil {
			if commit.Commit.Message != nil {
				message = *commit.Commit.Message
			}
			if commit.Commit.Author != nil {
				if commit.Commit.Author.Name != nil {
					author = *commit.Commit.Author.Name
				}
				if commit.Commit.Author.Date != nil {
					date = commit.Commit.Author.Date.Time
				}
			}
		}
		
		serviceCommits = append(serviceCommits, &types.Commit{
			Hash:    *commit.SHA,
			Message: message,
			Author:  author,
			Date:    date,
		})
	}
	
	return serviceCommits, nil
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

// Deployment Management Methods

func (a *App) GetServiceDeployments(serviceID int64) ([]*types.DeploymentOverview, error) {
	log.Printf("GetServiceDeployments called with serviceID: %d", serviceID)
	if a.deploymentModel == nil {
		log.Printf("ERROR: deployment model not initialized")
		return nil, fmt.Errorf("deployment model not initialized")
	}
	deployments, err := a.deploymentModel.GetDeploymentOverview(serviceID)
	if err != nil {
		log.Printf("ERROR: Failed to get deployments for service %d: %v", serviceID, err)
		return nil, err
	}
	log.Printf("Successfully retrieved %d deployments for service %d", len(deployments), serviceID)
	return deployments, nil
}

func (a *App) GetServiceCommitDeployments(serviceID int64) ([]*types.CommitDeploymentStatus, error) {
	log.Printf("GetServiceCommitDeployments called with serviceID: %d", serviceID)
	
	// Get service commits first
	commits, err := a.GetServiceCommits(serviceID)
	if err != nil {
		log.Printf("ERROR: Failed to get service commits: %v", err)
		return nil, err
	}
	
	// Get all deployments for this service
	deployments, err := a.deploymentModel.GetByServiceID(serviceID)
	if err != nil {
		log.Printf("ERROR: Failed to get deployments: %v", err)
		return nil, err
	}
	log.Printf("Found %d deployments for service %d", len(deployments), serviceID)
	
	// Create a map of commit SHA to deployments
	commitDeploymentMap := make(map[string][]*types.Deployment)
	for _, deployment := range deployments {
		if deployment.CommitSHA != "" {
			commitDeploymentMap[deployment.CommitSHA] = append(commitDeploymentMap[deployment.CommitSHA], deployment)
			log.Printf("Added deployment for commit %s in %s/%s/%s", deployment.CommitSHA[:7], deployment.Environment, deployment.Region, deployment.Namespace)
		}
	}
	log.Printf("Built commitDeploymentMap with %d unique commits", len(commitDeploymentMap))
	
	// Get unique environment/region/namespace combinations
	envRegionNamespaceSet := make(map[string]bool)
	for _, deployment := range deployments {
		key := deployment.Environment + "/" + deployment.Region + "/" + deployment.Namespace
		envRegionNamespaceSet[key] = true
	}
	
	// Build commit deployment status
	var result []*types.CommitDeploymentStatus
	for _, commit := range commits {
		commitStatus := &types.CommitDeploymentStatus{
			Commit:      *commit,
			Deployments: []types.DeploymentStatus{},
		}
		
		log.Printf("Processing commit %s", commit.Hash[:7])
		// Check deployments for this commit
		if commitDeployments, exists := commitDeploymentMap[commit.Hash]; exists {
			log.Printf("Found %d deployments for commit %s", len(commitDeployments), commit.Hash[:7])
			for _, deployment := range commitDeployments {
				deploymentStatus := types.DeploymentStatus{
					Environment: deployment.Environment,
					Region:      deployment.Region,
					Namespace:   deployment.Namespace,
					Tag:         deployment.Tag,
					IsDeployed:  true,
					DeployedAt:  deployment.UpdatedAt,
				}
				commitStatus.Deployments = append(commitStatus.Deployments, deploymentStatus)
			}
		} else {
			log.Printf("No deployments found for commit %s", commit.Hash[:7])
			// Add empty deployment statuses for all env/region/namespace combinations to show "not deployed"
			for envRegionNamespace := range envRegionNamespaceSet {
				parts := strings.Split(envRegionNamespace, "/")
				if len(parts) == 3 {
					deploymentStatus := types.DeploymentStatus{
						Environment: parts[0],
						Region:      parts[1],
						Namespace:   parts[2],
						Tag:         "",
						IsDeployed:  false,
						DeployedAt:  time.Time{},
					}
					commitStatus.Deployments = append(commitStatus.Deployments, deploymentStatus)
				}
			}
		}
		
		result = append(result, commitStatus)
	}
	
	log.Printf("Successfully retrieved %d commit deployment statuses for service %d", len(result), serviceID)
	return result, nil
}

// TestServiceCommitsFetch is a debug method to test GetServiceCommits specifically
func (a *App) TestServiceCommitsFetch(serviceID int64) string {
	log.Printf("TestServiceCommitsFetch called with serviceID: %d", serviceID)
	
	// Get service details
	service, err := a.serviceModel.GetByID(serviceID)
	if err != nil {
		return fmt.Sprintf("ERROR getting service: %v", err)
	}
	
	// Get repository details
	repo, err := a.repoModel.GetByID(service.RepositoryID)
	if err != nil {
		return fmt.Sprintf("ERROR getting repository: %v", err)
	}
	
	// Check GitHub token
	githubToken := a.getGitHubToken()
	tokenStatus := "configured"
	if githubToken == "" {
		tokenStatus = "missing"
	}
	
	result := fmt.Sprintf("Service: %s (ID: %d)\n", service.Name, serviceID)
	result += fmt.Sprintf("Path: %s\n", service.Path)
	result += fmt.Sprintf("Repository: %s (ID: %d)\n", repo.Name, repo.ID)
	result += fmt.Sprintf("Repository URL: %s\n", repo.URL)
	result += fmt.Sprintf("GitHub token: %s\n", tokenStatus)
	
	// Get commits
	commits, err := a.GetServiceCommits(serviceID)
	if err != nil {
		result += fmt.Sprintf("ERROR getting commits: %v\n", err)
	} else {
		result += fmt.Sprintf("Found %d commits:\n", len(commits))
		for i, commit := range commits {
			if len(commit.Message) > 50 {
				result += fmt.Sprintf("  %d: %s - %s...\n", i, commit.Hash[:7], commit.Message[:47])
			} else {
				result += fmt.Sprintf("  %d: %s - %s\n", i, commit.Hash[:7], commit.Message)
			}
		}
	}
	
	return result
}

// TestCommitDeploymentCorrelation is a debug method to test the correlation logic
func (a *App) TestCommitDeploymentCorrelation(serviceID int64) string {
	log.Printf("TestCommitDeploymentCorrelation called with serviceID: %d", serviceID)
	
	// Get service commits
	commits, err := a.GetServiceCommits(serviceID)
	if err != nil {
		return fmt.Sprintf("ERROR getting commits: %v", err)
	}
	
	// Get deployments
	deployments, err := a.deploymentModel.GetByServiceID(serviceID)
	if err != nil {
		return fmt.Sprintf("ERROR getting deployments: %v", err)
	}
	
	result := fmt.Sprintf("Found %d commits and %d deployments\n", len(commits), len(deployments))
	result += "Commits:\n"
	for i, commit := range commits {
		result += fmt.Sprintf("  %d: %s - %s\n", i, commit.Hash[:7], commit.Message[:50])
	}
	result += "Deployments:\n"
	for i, deployment := range deployments {
		result += fmt.Sprintf("  %d: %s in %s/%s/%s\n", i, deployment.CommitSHA[:7], deployment.Environment, deployment.Region, deployment.Namespace)
	}
	
	return result
}

func (a *App) GetServiceDeploymentHistory(serviceID int64) ([]*types.Commit, error) {
	// Get the service to find its repository
	service, err := a.serviceModel.GetByID(serviceID)
	if err != nil {
		return nil, fmt.Errorf("service not found: %w", err)
	}

	// Get repository details
	repo, err := a.repoModel.GetByID(service.RepositoryID)
	if err != nil {
		return nil, fmt.Errorf("repository not found: %w", err)
	}

	// Parse GitHub URL to get owner and repo name
	owner, repoName, err := a.parseGitHubURL(repo.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid repository URL: %w", err)
	}

	// Get GitHub token
	githubToken := a.getGitHubToken()
	if githubToken == "" {
		return nil, fmt.Errorf("GitHub token not configured")
	}

	// Create GitHub client
	client := a.createGitHubClient(githubToken)

	// Get commits for the service path
	opts := &goGithub.CommitsListOptions{
		Path: service.Path,
		ListOptions: goGithub.ListOptions{
			PerPage: 100,
		},
	}

	commits, _, err := client.Repositories.ListCommits(a.ctx, owner, repoName, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get service commits: %w", err)
	}

	var serviceCommits []*types.Commit
	for _, commit := range commits {
		if commit.Commit == nil {
			continue
		}

		author := "Unknown"
		if commit.Commit.Author != nil && commit.Commit.Author.Name != nil {
			author = *commit.Commit.Author.Name
		}

		message := ""
		if commit.Commit.Message != nil {
			message = *commit.Commit.Message
		}

		date := time.Now()
		if commit.Commit.Author != nil && commit.Commit.Author.Date != nil {
			date = commit.Commit.Author.Date.Time
		}

		serviceCommits = append(serviceCommits, &types.Commit{
			Hash:    *commit.SHA,
			Message: message,
			Author:  author,
			Date:    date,
		})
	}

	return serviceCommits, nil
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

// Configuration Management Methods

func (a *App) GetConfig(key string) (string, error) {
	if a.configModel == nil {
		return "", fmt.Errorf("config model not initialized")
	}
	
	config, err := a.configModel.Get(key)
	if err != nil {
		return "", err
	}
	
	if config == nil {
		return "", nil // No config found
	}
	
	return config.Value, nil
}

func (a *App) SetConfig(key, value string) error {
	if a.configModel == nil {
		return fmt.Errorf("config model not initialized")
	}
	
	err := a.configModel.Set(key, value)
	if err != nil {
		return err
	}
	
	// Reinitialize JIRA client if JIRA config was changed
	if strings.HasPrefix(key, "jira_") {
		a.initJiraClient()
	}
	
	return nil
}

func (a *App) GetAllConfig() (map[string]string, error) {
	if a.configModel == nil {
		return map[string]string{}, nil
	}
	return a.configModel.GetAll()
}

// JIRA Integration Methods

func (a *App) initJiraClient() {
	if a.configModel == nil {
		return
	}
	
	jiraURL, _ := a.configModel.Get("jira_url")
	jiraToken, _ := a.configModel.Get("jira_token")
	jiraUsername, _ := a.configModel.Get("jira_username")
	jiraAuthMethod, _ := a.configModel.Get("jira_auth_method")
	
	if jiraURL != nil && jiraURL.Value != "" && jiraToken != nil && jiraToken.Value != "" {
		var username, authMethod string
		if jiraUsername != nil {
			username = jiraUsername.Value
		}
		if jiraAuthMethod != nil {
			authMethod = jiraAuthMethod.Value
		}
		
		a.jiraClient = jira.NewClientWithAuth(jiraURL.Value, username, jiraToken.Value, authMethod)
		log.Printf("JIRA client initialized with auth method: %s", authMethod)
	}
}

func (a *App) TestJiraConnection() error {
	if a.jiraClient == nil {
		return fmt.Errorf("JIRA client not configured")
	}
	return a.jiraClient.TestConnection()
}

func (a *App) FetchJiraTicketTitle(ticketID string) (string, error) {
	if a.jiraClient == nil {
		return "", fmt.Errorf("JIRA client not configured")
	}
	
	issue, err := a.jiraClient.GetIssue(ticketID)
	if err != nil {
		return "", err
	}
	
	return issue.Fields.Summary, nil
}

func (a *App) UpdateTaskJiraTitle(taskID int64, ticketID string) error {
	if a.taskModel == nil {
		return fmt.Errorf("task model not initialized")
	}
	
	if a.jiraClient == nil {
		return fmt.Errorf("JIRA client not configured")
	}
	
	title, err := a.FetchJiraTicketTitle(ticketID)
	if err != nil {
		log.Printf("Failed to fetch JIRA ticket title for %s: %v", ticketID, err)
		return err
	}
	
	return a.taskModel.UpdateJiraTitle(taskID, title)
}

func (a *App) RefreshAllJiraTitles() error {
	if a.taskModel == nil {
		return fmt.Errorf("task model not initialized")
	}
	
	if a.jiraClient == nil {
		return fmt.Errorf("JIRA client not configured")
	}
	
	// Get all tasks
	tasks, err := a.taskModel.GetAllWithProjects()
	if err != nil {
		return err
	}
	
	var errors []string
	successCount := 0
	
	for _, task := range tasks {
		if task.JiraTicketID != "" {
			title, err := a.FetchJiraTicketTitle(task.JiraTicketID)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Failed to fetch title for %s: %v", task.JiraTicketID, err))
				continue
			}
			
			err = a.taskModel.UpdateJiraTitle(task.ID, title)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Failed to update title for task %d: %v", task.ID, err))
				continue
			}
			
			successCount++
		}
	}
	
	log.Printf("Refreshed %d JIRA titles, %d errors", successCount, len(errors))
	
	if len(errors) > 0 {
		return fmt.Errorf("some titles failed to refresh: %v", errors)
	}
	
	return nil
}

// Enhanced Task Methods

func (a *App) CreateTaskWithJiraTitle(task types.Task) error {
	log.Printf("CreateTaskWithJiraTitle called with task: %+v", task)
	
	if a.taskModel == nil {
		log.Printf("Error: task model not initialized")
		return fmt.Errorf("task model not initialized")
	}
	
	// If JIRA ticket ID is provided and JIRA client is configured, fetch the title
	if task.JiraTicketID != "" && a.jiraClient != nil {
		log.Printf("Fetching JIRA title for ticket: %s", task.JiraTicketID)
		title, err := a.FetchJiraTicketTitle(task.JiraTicketID)
		if err != nil {
			log.Printf("Warning: Failed to fetch JIRA title for %s: %v", task.JiraTicketID, err)
		} else {
			task.JiraTitle = title
			log.Printf("Successfully fetched JIRA title: %s", title)
		}
	} else {
		log.Printf("Skipping JIRA title fetch - ticketID: %s, jiraClient: %v", task.JiraTicketID, a.jiraClient != nil)
	}
	
	log.Printf("Creating task with data: %+v", task)
	err := a.taskModel.Create(&task)
	if err != nil {
		log.Printf("Error creating task: %v", err)
		return fmt.Errorf("failed to create task: %w", err)
	}
	
	log.Printf("Task created successfully with ID: %d", task.ID)
	return nil
}

// Greet returns a greeting for the given name (keeping original method for compatibility)
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// TestDeploymentData is a test method to verify deployment functionality
func (a *App) TestDeploymentData() map[string]interface{} {
	result := make(map[string]interface{})
	
	// Test if we have services
	services, err := a.serviceModel.GetAll()
	if err != nil {
		result["error"] = fmt.Sprintf("Failed to get services: %v", err)
		return result
	}
	result["services_count"] = len(services)
	result["services"] = services
	
	// Test deployment data for service-a (ID: 3)
	if len(services) > 0 {
		serviceID := int64(3) // service-a
		deployments, err := a.deploymentModel.GetDeploymentOverview(serviceID)
		if err != nil {
			result["deployment_error"] = fmt.Sprintf("Failed to get deployments: %v", err)
		} else {
			result["deployments_count"] = len(deployments)
			result["deployments"] = deployments
		}
	}
	
	return result
}

// isKubernetesRepository checks if a repository is actually a kubernetes repository
// based on name patterns and URL content, even if incorrectly typed as monorepo
func (a *App) isKubernetesRepository(repo *types.Repository) bool {
	// Check if repository name suggests it's a kubernetes repository
	name := strings.ToLower(repo.Name)
	if strings.Contains(name, "k8s") || strings.Contains(name, "kubernetes") {
		return true
	}
	
	// Check if repository URL suggests it's a kubernetes repository
	url := strings.ToLower(repo.URL)
	if strings.Contains(url, "k8s") || strings.Contains(url, "kubernetes") {
		return true
	}
	
	return false
}

// parseRepositoryURL extracts owner and repo name from GitHub repository URL
func parseRepositoryURL(url string) (string, string) {
	// Handle both SSH and HTTPS URLs
	// SSH: git@github.com:owner/repo.git
	// HTTPS: https://github.com/owner/repo or https://github.com/owner/repo.git
	
	url = strings.TrimSuffix(url, ".git")
	
	if strings.HasPrefix(url, "git@github.com:") {
		// SSH format
		parts := strings.Split(strings.TrimPrefix(url, "git@github.com:"), "/")
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
	} else if strings.Contains(url, "github.com/") {
		// HTTPS format
		idx := strings.Index(url, "github.com/")
		if idx >= 0 {
			remaining := url[idx+len("github.com/"):]
			parts := strings.Split(remaining, "/")
			if len(parts) >= 2 {
				return parts[0], parts[1]
			}
		}
	}
	
	return "", ""
}

// getGitHubToken retrieves the GitHub token from config, falling back to environment variable
func (a *App) getGitHubToken() string {
	// Try to get from database config first
	if a.configModel != nil {
		if config, err := a.configModel.Get("github_token"); err == nil && config != nil && config.Value != "" {
			return config.Value
		}
	}
	
	// Fall back to environment variable for backward compatibility
	return os.Getenv("GITHUB_TOKEN")
}

// getGitHubEnterpriseURL retrieves the GitHub Enterprise URL from config
func (a *App) getGitHubEnterpriseURL() string {
	if a.configModel != nil {
		if config, err := a.configModel.Get("github_enterprise_url"); err == nil && config != nil && config.Value != "" {
			return config.Value
		}
	}
	return ""
}

// TestGitHubConnection tests the GitHub connection using the stored token
func (a *App) TestGitHubConnection() error {
	githubToken := a.getGitHubToken()
	if githubToken == "" {
		return fmt.Errorf("no GitHub token configured")
	}
	
	ctx := context.Background()
	client := a.createGitHubClient(githubToken)
	
	// Test the token by making a simple API call to get the authenticated user
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return fmt.Errorf("GitHub API test failed: %w", err)
	}
	
	log.Printf("GitHub connection test successful. Authenticated as: %s", user.GetLogin())
	return nil
}

// TestScanKubernetesDeployments manually triggers a scan of kubernetes deployments for testing
func (a *App) TestScanKubernetesDeployments() error {
	log.Printf("TestScanKubernetesDeployments called")
	
	// Get kubernetes repository
	repos, err := a.repoModel.GetAll()
	if err != nil {
		return fmt.Errorf("failed to get repositories: %w", err)
	}
	
	var kubernetesRepo *types.Repository
	for _, repo := range repos {
		if repo.Type == types.KubernetesType {
			kubernetesRepo = repo
			break
		}
	}
	
	if kubernetesRepo == nil {
		return fmt.Errorf("no kubernetes repository found")
	}
	
	log.Printf("Found kubernetes repository: %s (%s)", kubernetesRepo.Name, kubernetesRepo.URL)
	
	// Clear existing deployments
	if err := a.clearAllDeployments(); err != nil {
		log.Printf("Warning: failed to clear existing deployments: %v", err)
	}
	
	// Trigger sync for kubernetes repository
	if a.syncService != nil {
		return a.syncService.SyncRepository(kubernetesRepo.ID)
	}
	
	return fmt.Errorf("sync service not initialized - GitHub token required")
}

// Helper method to clear all deployments for testing
func (a *App) clearAllDeployments() error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}
	
	_, err := a.db.GetConn().Exec("DELETE FROM deployments")
	return err
}

// TestKustomizationFileAccess tests if we can access the kustomization.yaml file directly
func (a *App) TestKustomizationFileAccess() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	githubToken := a.getGitHubToken()
	if githubToken == "" {
		result["error"] = "No GitHub token configured"
		result["github_token_configured"] = false
		return result, nil
	}
	
	result["github_token_configured"] = true
	
	// Get kubernetes repository
	repos, err := a.repoModel.GetAll()
	if err != nil {
		result["error"] = fmt.Sprintf("Failed to get repositories: %v", err)
		return result, err
	}
	
	var kubernetesRepo *types.Repository
	for _, repo := range repos {
		if repo.Type == types.KubernetesType {
			kubernetesRepo = repo
			break
		}
	}
	
	if kubernetesRepo == nil {
		result["error"] = "No kubernetes repository found"
		return result, nil
	}
	
	result["kubernetes_repo"] = map[string]interface{}{
		"name": kubernetesRepo.Name,
		"url":  kubernetesRepo.URL,
		"type": kubernetesRepo.Type,
	}
	
	// Parse GitHub URL
	owner, repoName, err := a.parseGitHubURL(kubernetesRepo.URL)
	if err != nil {
		result["error"] = fmt.Sprintf("Invalid repository URL: %v", err)
		return result, err
	}
	
	result["parsed_url"] = map[string]interface{}{
		"owner": owner,
		"repo":  repoName,
	}
	
	// Test GitHub client
	ctx := context.Background()
	client := a.createGitHubClient(githubToken)
	
	// Test repository access
	repo, _, err := client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		result["error"] = fmt.Sprintf("Cannot access repository: %v", err)
		return result, err
	}
	
	result["repo_access"] = "success"
	result["default_branch"] = repo.GetDefaultBranch()
	
	// Search for kustomization.yaml files
	searchQuery := fmt.Sprintf("filename:kustomization.yaml repo:%s/%s", owner, repoName)
	searchResult, _, err := client.Search.Code(ctx, searchQuery, &goGithub.SearchOptions{
		ListOptions: goGithub.ListOptions{PerPage: 10},
	})
	
	if err != nil {
		result["search_error"] = fmt.Sprintf("Search failed: %v", err)
	} else {
		var files []map[string]interface{}
		for _, codeResult := range searchResult.CodeResults {
			if codeResult.Path != nil {
				files = append(files, map[string]interface{}{
					"path": *codeResult.Path,
					"url":  *codeResult.HTMLURL,
				})
			}
		}
		result["kustomization_files"] = files
		result["files_found"] = len(files)
	}
	
	// Try to get specific file content
	testPath := "services/service-a/overlays/stg/us-west-2/kustomization.yaml"
	fileContent, _, _, err := client.Repositories.GetContents(ctx, owner, repoName, testPath, nil)
	if err != nil {
		result["file_access_error"] = fmt.Sprintf("Cannot access %s: %v", testPath, err)
	} else {
		content, err := fileContent.GetContent()
		if err != nil {
			result["content_decode_error"] = fmt.Sprintf("Cannot decode content: %v", err)
		} else {
			result["file_content"] = content
			result["file_access"] = "success"
		}
	}
	
	return result, nil
}