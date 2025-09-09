package github

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

type Client struct {
	gh      *github.Client
	token   string
	baseURL string
	isEnterprise bool
}

type ServiceInfo struct {
	Name        string
	Path        string
	Description string
}

type ResourceInfo struct {
	Name         string
	Path         string
	ResourceType string
	Namespace    string
}

type WorkflowRun struct {
	ID          int64
	Status      string
	Commit      string
	Branch      string
	StartedAt   time.Time
	CompletedAt *time.Time
}

func NewClient(token string) *Client {
	return NewClientWithBaseURL(token, "")
}

func NewClientWithBaseURL(token, baseURL string) *Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)

	var client *github.Client
	isEnterprise := false
	
	if baseURL != "" && baseURL != "https://api.github.com/" {
		// GitHub Enterprise Server
		var err error
		client, err = github.NewEnterpriseClient(baseURL, baseURL, tc)
		if err != nil {
			log.Printf("Failed to create Enterprise GitHub client: %v", err)
			// Fallback to regular client
			client = github.NewClient(tc)
		} else {
			isEnterprise = true
			log.Printf("Created GitHub Enterprise client for: %s", baseURL)
		}
	} else {
		// GitHub.com
		client = github.NewClient(tc)
	}

	return &Client{
		gh:          client,
		token:       token,
		baseURL:     baseURL,
		isEnterprise: isEnterprise,
	}
}

func (c *Client) GetRepository(ctx context.Context, owner, repo string) (*github.Repository, error) {
	repository, _, err := c.gh.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}
	return repository, nil
}

func (c *Client) DiscoverMicroservices(ctx context.Context, owner, repo string) ([]ServiceInfo, error) {
	return c.DiscoverMicroservicesInPath(ctx, owner, repo, "services")
}

func (c *Client) DiscoverMicroservicesInPath(ctx context.Context, owner, repo, servicePath string) ([]ServiceInfo, error) {
	var services []ServiceInfo

	// Clean the service path (remove trailing slash and leading ./)
	servicePath = strings.TrimSuffix(servicePath, "/")
	servicePath = strings.TrimPrefix(servicePath, "./")
	if servicePath == "" {
		servicePath = "services" // Default fallback
	}

	fmt.Printf("[GitHub Client] Discovering services in %s/%s at path: %s\n", owner, repo, servicePath)

	// Get contents of the specified directory
	_, contents, _, err := c.gh.Repositories.GetContents(ctx, owner, repo, servicePath, nil)
	if err != nil {
		if githubErr, ok := err.(*github.ErrorResponse); ok {
			fmt.Printf("[GitHub Client] Directory %s does not exist (HTTP %d): %s\n", servicePath, githubErr.Response.StatusCode, githubErr.Message)
			// Directory doesn't exist
			return services, nil
		}
		fmt.Printf("[GitHub Client] ERROR: Failed to get directory %s: %v\n", servicePath, err)
		return nil, fmt.Errorf("failed to get directory %s: %w", servicePath, err)
	}

	fmt.Printf("[GitHub Client] Found %d items in directory %s\n", len(contents), servicePath)

	for _, content := range contents {
		fmt.Printf("[GitHub Client] Processing item: %s (type: %s)\n", content.GetName(), content.GetType())
		if content.GetType() == "dir" {
			serviceName := content.GetName()
			fullServicePath := fmt.Sprintf("%s/%s", servicePath, serviceName)

			fmt.Printf("[GitHub Client] Found service directory: %s at path %s\n", serviceName, fullServicePath)

			// Try to get a description from README or package.json
			description := c.getServiceDescription(ctx, owner, repo, fullServicePath)

			service := ServiceInfo{
				Name:        serviceName,
				Path:        fullServicePath,
				Description: description,
			}

			services = append(services, service)
			fmt.Printf("[GitHub Client] Added service: %s with description: %s\n", serviceName, description)
		}
	}

	fmt.Printf("[GitHub Client] Total services discovered: %d\n", len(services))
	return services, nil
}

func (c *Client) DiscoverKubernetesResources(ctx context.Context, owner, repo string) ([]ResourceInfo, error) {
	return c.DiscoverKubernetesResourcesInPath(ctx, owner, repo, "")
}

func (c *Client) DiscoverKubernetesResourcesInPath(ctx context.Context, owner, repo, rootPath string) ([]ResourceInfo, error) {
	var resources []ResourceInfo

	if rootPath != "" && rootPath != "." {
		// If a specific root path is provided, scan that directory and its subdirectories
		dirResources, err := c.discoverResourcesInDir(ctx, owner, repo, strings.TrimPrefix(rootPath, "/"), "")
		if err != nil {
			return resources, fmt.Errorf("failed to scan root path %s: %w", rootPath, err)
		}
		resources = append(resources, dirResources...)
	} else {
		// No root path specified, use default behavior to check common Kubernetes directories
		kubeDirs := []string{"k8s", "kubernetes", "manifests", "deployment", "overlays"}

		for _, dir := range kubeDirs {
			dirResources, err := c.discoverResourcesInDir(ctx, owner, repo, dir, "")
			if err != nil {
				continue // Skip if directory doesn't exist
			}
			resources = append(resources, dirResources...)
		}
	}

	return resources, nil
}

func (c *Client) discoverResourcesInDir(ctx context.Context, owner, repo, path, namespace string) ([]ResourceInfo, error) {
	var resources []ResourceInfo

	_, contents, _, err := c.gh.Repositories.GetContents(ctx, owner, repo, path, nil)
	if err != nil {
		return resources, err
	}

	for _, content := range contents {
		if content.GetType() == "dir" {
			// Recursively search subdirectories
			subResources, _ := c.discoverResourcesInDir(ctx, owner, repo, content.GetPath(), namespace)
			resources = append(resources, subResources...)
		} else if content.GetType() == "file" && (strings.HasSuffix(content.GetName(), ".yaml") || strings.HasSuffix(content.GetName(), ".yml")) {
			// Parse YAML file for Kubernetes resources
			resourceInfo := c.parseKubernetesFile(ctx, owner, repo, content.GetPath())
			if resourceInfo != nil {
				resourceInfo.Namespace = namespace
				resources = append(resources, *resourceInfo)
			}
		}
	}

	return resources, nil
}

func (c *Client) parseKubernetesFile(ctx context.Context, owner, repo, path string) *ResourceInfo {
	// Get file contents
	fileContent, _, _, err := c.gh.Repositories.GetContents(ctx, owner, repo, path, nil)
	if err != nil || fileContent == nil {
		return nil
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return nil
	}

	// Simple parsing - look for kind and metadata.name
	lines := strings.Split(content, "\n")
	var kind, name string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "kind:") {
			kind = strings.TrimSpace(strings.TrimPrefix(line, "kind:"))
		}
		if strings.HasPrefix(line, "name:") && name == "" {
			name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
		}
	}

	if kind != "" && name != "" {
		return &ResourceInfo{
			Name:         name,
			Path:         path,
			ResourceType: kind,
		}
	}

	return nil
}

func (c *Client) getServiceDescription(ctx context.Context, owner, repo, servicePath string) string {
	// Try to get README.md
	readme, _, _, err := c.gh.Repositories.GetContents(ctx, owner, repo, fmt.Sprintf("%s/README.md", servicePath), nil)
	if err == nil && readme != nil {
		content, err := readme.GetContent()
		if err == nil {
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && !strings.HasPrefix(line, "#") {
					return line
				}
			}
		}
	}

	// Try to get package.json for description
	packageJSON, _, _, err := c.gh.Repositories.GetContents(ctx, owner, repo, fmt.Sprintf("%s/package.json", servicePath), nil)
	if err == nil && packageJSON != nil {
		content, err := packageJSON.GetContent()
		if err == nil && strings.Contains(content, "\"description\"") {
			// Simple JSON parsing for description field
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				if strings.Contains(line, "\"description\"") {
					parts := strings.Split(line, ":")
					if len(parts) > 1 {
						desc := strings.Trim(strings.TrimSpace(parts[1]), "\",")
						return strings.Trim(desc, "\"")
					}
				}
			}
		}
	}

	return ""
}

func (c *Client) GetWorkflowRuns(ctx context.Context, owner, repo string, workflowID int64, limit int) ([]WorkflowRun, error) {
	opts := &github.ListWorkflowRunsOptions{
		ListOptions: github.ListOptions{PerPage: limit},
	}

	runs, _, err := c.gh.Actions.ListWorkflowRunsByID(ctx, owner, repo, workflowID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow runs: %w", err)
	}

	var workflowRuns []WorkflowRun
	for _, run := range runs.WorkflowRuns {
		workflowRun := WorkflowRun{
			ID:        run.GetID(),
			Status:    run.GetStatus(),
			Commit:    run.GetHeadSHA(),
			Branch:    run.GetHeadBranch(),
			StartedAt: run.GetCreatedAt().Time,
		}

		if run.UpdatedAt != nil {
			workflowRun.CompletedAt = &run.UpdatedAt.Time
		}

		workflowRuns = append(workflowRuns, workflowRun)
	}

	return workflowRuns, nil
}

func (c *Client) ListWorkflows(ctx context.Context, owner, repo string) ([]*github.Workflow, error) {
	workflows, _, err := c.gh.Actions.ListWorkflows(ctx, owner, repo, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflows: %w", err)
	}

	return workflows.Workflows, nil
}

// KustomizationDeployment represents a deployment found in kustomization.yaml
type KustomizationDeployment struct {
	ServiceName  string
	Environment  string
	Region       string
	Namespace    string
	Tag          string
	Path         string
	CommitSHA    string
}

// ScanKustomizationFiles scans the Kubernetes repository for kustomization.yaml files
func (c *Client) ScanKustomizationFiles(ctx context.Context, owner, repo string) ([]KustomizationDeployment, error) {
	var deployments []KustomizationDeployment

	// Use Contents API to traverse repository structure instead of Search API
	// This is more reliable for private repositories and newly created files
	kustomizationPaths, err := c.findKustomizationFiles(ctx, owner, repo, "services", make([]string, 0))
	if err != nil {
		return nil, fmt.Errorf("failed to find kustomization files: %w", err)
	}

	for _, path := range kustomizationPaths {
		
		// Parse service name, environment, region, and namespace from path
		// Expected: services/service-b/overlays/prd/us-west-2/ns-a/kustomization.yaml
		pathParts := strings.Split(path, "/")
		if len(pathParts) < 7 || pathParts[0] != "services" || pathParts[2] != "overlays" {
			continue
		}

		serviceName := pathParts[1]
		// Skip overlays directory at pathParts[2]
		environment := pathParts[3] 
		region := pathParts[4]
		namespace := pathParts[5]

		// Get the content of the kustomization.yaml file
		fileContent, _, _, err := c.gh.Repositories.GetContents(ctx, owner, repo, path, nil)
		if err != nil {
			log.Printf("Failed to get kustomization file %s: %v", path, err)
			continue
		}

		if fileContent == nil {
			continue
		}

		content, err := fileContent.GetContent()
		if err != nil {
			log.Printf("Failed to decode kustomization file %s: %v", path, err)
			continue
		}

		// Parse YAML to extract image tag
		tag := c.extractImageTagFromKustomization(content, serviceName)
		if tag == "" {
			log.Printf("No tag found for service %s in %s", serviceName, path)
			continue
		}

		// Get the commit SHA for this file
		commitSHA := ""
		commits, _, err := c.gh.Repositories.ListCommits(ctx, owner, repo, &github.CommitsListOptions{
			Path: path,
			ListOptions: github.ListOptions{PerPage: 1},
		})
		if err == nil && len(commits) > 0 && commits[0].SHA != nil {
			commitSHA = *commits[0].SHA
		}

		deployment := KustomizationDeployment{
			ServiceName: serviceName,
			Environment: environment,
			Region:      region,
			Namespace:   namespace,
			Tag:         tag,
			Path:        path,
			CommitSHA:   commitSHA,
		}

		deployments = append(deployments, deployment)
	}

	return deployments, nil
}

// extractImageTagFromKustomization parses kustomization.yaml content to find the newTag for a service
func (c *Client) extractImageTagFromKustomization(content, serviceName string) string {
	// Simple YAML parsing to find images section and extract newTag
	lines := strings.Split(content, "\n")
	inImagesSection := false
	inServiceImage := false

	for _, line := range lines {
		originalLine := line
		line = strings.TrimSpace(line)
		
		// Look for images: section
		if line == "images:" {
			inImagesSection = true
			continue
		}

		if inImagesSection {
			// Check if we're out of images section (non-indented line that's not part of list)
			if len(line) > 0 && !strings.HasPrefix(originalLine, " ") && !strings.HasPrefix(line, "-") && line != "---" {
				inImagesSection = false
				inServiceImage = false
				continue
			}

			// Look for service name in image name or newName
			if strings.Contains(line, "name:") && strings.Contains(line, serviceName) {
				inServiceImage = true
				continue
			}
			if strings.Contains(line, "newName:") && strings.Contains(line, serviceName) {
				inServiceImage = true
				continue
			}

			// Extract newTag if we're in the correct service image
			if inServiceImage && strings.Contains(line, "newTag:") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					tag := strings.TrimSpace(parts[1])
					tag = strings.Trim(tag, "\"'")
					return tag
				}
			}

			// Reset service image flag when we hit a new image entry (new list item)
			if strings.HasPrefix(line, "-") {
				inServiceImage = false
			}
		}
	}

	return ""
}

// findKustomizationFiles recursively searches for kustomization.yaml files using Contents API
func (c *Client) findKustomizationFiles(ctx context.Context, owner, repo, path string, foundFiles []string) ([]string, error) {
	// Get contents of the directory
	_, contents, _, err := c.gh.Repositories.GetContents(ctx, owner, repo, path, nil)
	if err != nil {
		// Directory doesn't exist, skip silently
		return foundFiles, nil
	}

	for _, content := range contents {
		if content.GetType() == "dir" {
			// Recursively search subdirectories
			subPath := content.GetPath()
			foundFiles, err = c.findKustomizationFiles(ctx, owner, repo, subPath, foundFiles)
			if err != nil {
				continue // Skip directories we can't access
			}
		} else if content.GetType() == "file" && content.GetName() == "kustomization.yaml" {
			// Found a kustomization.yaml file
			foundFiles = append(foundFiles, content.GetPath())
		}
	}

	return foundFiles, nil
}

// GetGitHubClient returns the underlying GitHub client for advanced operations
func (c *Client) GetGitHubClient() *github.Client {
	return c.gh
}

// IsEnterprise returns true if this client is configured for GitHub Enterprise
func (c *Client) IsEnterprise() bool {
	return c.isEnterprise
}

// GetBaseURL returns the base URL for this GitHub client
func (c *Client) GetBaseURL() string {
	return c.baseURL
}

// ParseRepositoryURL extracts owner and repo name from various GitHub URL formats
// Supports both GitHub.com and GitHub Enterprise URLs
func (c *Client) ParseRepositoryURL(repoURL string) (owner, repo string, err error) {
	if repoURL == "" {
		return "", "", fmt.Errorf("repository URL is empty")
	}

	// Remove .git suffix
	repoURL = strings.TrimSuffix(repoURL, ".git")
	
	// Handle HTTPS URLs
	if strings.HasPrefix(repoURL, "https://") {
		return c.parseHTTPSURL(repoURL)
	}
	
	return "", "", fmt.Errorf("only HTTPS URLs are supported")
}

// parseHTTPSURL handles HTTPS GitHub URLs for both github.com and Enterprise
func (c *Client) parseHTTPSURL(repoURL string) (owner, repo string, err error) {
	// Remove https:// prefix
	urlPath := strings.TrimPrefix(repoURL, "https://")
	
	// Split by /
	parts := strings.Split(urlPath, "/")
	if len(parts) < 3 {
		return "", "", fmt.Errorf("invalid repository URL format")
	}
	
	// For GitHub.com: github.com/owner/repo
	// For Enterprise: enterprise.example.com/owner/repo
	if len(parts) >= 3 {
		return parts[len(parts)-2], parts[len(parts)-1], nil
	}
	
	return "", "", fmt.Errorf("unable to parse repository URL")
}

// IsValidGitHubURL checks if the provided URL matches this client's configuration
func (c *Client) IsValidGitHubURL(repoURL string) bool {
	if repoURL == "" {
		return false
	}
	
	if c.isEnterprise && c.baseURL != "" {
		// Extract domain from base URL
		// baseURL format: https://enterprise.example.com/api/v3/
		baseURLParts := strings.Split(strings.TrimPrefix(c.baseURL, "https://"), "/")
		if len(baseURLParts) > 0 {
			expectedDomain := baseURLParts[0]
			return strings.Contains(repoURL, expectedDomain)
		}
	} else {
		// GitHub.com
		return strings.Contains(repoURL, "github.com")
	}
	
	return false
}
