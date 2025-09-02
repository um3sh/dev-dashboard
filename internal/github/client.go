package github

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

type Client struct {
	gh    *github.Client
	token string
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
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)

	client := github.NewClient(tc)

	return &Client{
		gh:    client,
		token: token,
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
	var services []ServiceInfo

	// Get contents of the services directory
	_, contents, _, err := c.gh.Repositories.GetContents(ctx, owner, repo, "services", nil)
	if err != nil {
		if _, ok := err.(*github.ErrorResponse); ok {
			// Services directory doesn't exist
			return services, nil
		}
		return nil, fmt.Errorf("failed to get services directory: %w", err)
	}

	for _, content := range contents {
		if content.GetType() == "dir" {
			serviceName := content.GetName()
			servicePath := fmt.Sprintf("services/%s", serviceName)

			// Try to get a description from README or package.json
			description := c.getServiceDescription(ctx, owner, repo, servicePath)

			services = append(services, ServiceInfo{
				Name:        serviceName,
				Path:        servicePath,
				Description: description,
			})
		}
	}

	return services, nil
}

func (c *Client) DiscoverKubernetesResources(ctx context.Context, owner, repo string) ([]ResourceInfo, error) {
	var resources []ResourceInfo

	// Common Kubernetes directories to check
	kubeDirs := []string{"k8s", "kubernetes", "manifests", "deployment", "overlays"}

	for _, dir := range kubeDirs {
		dirResources, err := c.discoverResourcesInDir(ctx, owner, repo, dir, "")
		if err != nil {
			continue // Skip if directory doesn't exist
		}
		resources = append(resources, dirResources...)
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