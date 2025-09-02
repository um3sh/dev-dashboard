package github

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
	gossh "golang.org/x/crypto/ssh"
)

type SSHClient struct {
	gh       *github.Client
	sshAuth  transport.AuthMethod
	token    string
	sshKey   string
}

func NewSSHClient(token string, sshKeyPath string) (*SSHClient, error) {
	client := &SSHClient{
		token:  token,
		sshKey: sshKeyPath,
	}

	// Set up GitHub API client if token is provided
	if token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(context.Background(), ts)
		client.gh = github.NewClient(tc)
	}

	// Set up SSH authentication
	if sshKeyPath != "" {
		pubKeys, err := ssh.NewPublicKeysFromFile("git", sshKeyPath, "")
		if err != nil {
			// Try with passphrase from environment
			passphrase := os.Getenv("SSH_PASSPHRASE")
			pubKeys, err = ssh.NewPublicKeysFromFile("git", sshKeyPath, passphrase)
			if err != nil {
				return nil, fmt.Errorf("failed to load SSH key: %w", err)
			}
		}
		
		// Configure host key callback to accept known hosts
		pubKeys.HostKeyCallback = gossh.InsecureIgnoreHostKey()
		client.sshAuth = pubKeys
	} else {
		// Try default SSH key locations
		homeDir, _ := os.UserHomeDir()
		defaultKeys := []string{
			filepath.Join(homeDir, ".ssh", "id_rsa"),
			filepath.Join(homeDir, ".ssh", "id_ed25519"),
			filepath.Join(homeDir, ".ssh", "id_ecdsa"),
		}

		for _, keyPath := range defaultKeys {
			if _, err := os.Stat(keyPath); err == nil {
				pubKeys, err := ssh.NewPublicKeysFromFile("git", keyPath, "")
				if err != nil {
					// Try with passphrase from environment
					passphrase := os.Getenv("SSH_PASSPHRASE")
					pubKeys, err = ssh.NewPublicKeysFromFile("git", keyPath, passphrase)
					if err != nil {
						continue // Try next key
					}
				}
				pubKeys.HostKeyCallback = gossh.InsecureIgnoreHostKey()
				client.sshAuth = pubKeys
				client.sshKey = keyPath
				break
			}
		}
	}

	return client, nil
}

func (c *SSHClient) CloneRepository(ctx context.Context, repoURL, targetDir string) error {
	if c.sshAuth == nil {
		return fmt.Errorf("SSH authentication not configured")
	}

	// Convert HTTPS URLs to SSH format
	sshURL := c.convertToSSHURL(repoURL)

	_, err := git.PlainClone(targetDir, false, &git.CloneOptions{
		URL:  sshURL,
		Auth: c.sshAuth,
	})

	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	return nil
}

func (c *SSHClient) CloneToMemory(ctx context.Context, repoURL string) (*git.Repository, error) {
	if c.sshAuth == nil {
		return nil, fmt.Errorf("SSH authentication not configured")
	}

	// Convert HTTPS URLs to SSH format
	sshURL := c.convertToSSHURL(repoURL)

	repo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:  sshURL,
		Auth: c.sshAuth,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to clone repository to memory: %w", err)
	}

	return repo, nil
}

func (c *SSHClient) DiscoverMicroservices(ctx context.Context, repoURL, serviceName, serviceLocation string) ([]ServiceInfo, error) {
	var services []ServiceInfo

	// If specific service name and location are provided, use them
	if serviceName != "" && serviceLocation != "" {
		services = append(services, ServiceInfo{
			Name: serviceName,
			Path: serviceLocation,
			Description: fmt.Sprintf("Service %s at %s", serviceName, serviceLocation),
		})
		return services, nil
	}

	// Otherwise, try to clone and discover services
	repo, err := c.CloneToMemory(ctx, repoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository for discovery: %w", err)
	}

	// Get the repository's working tree
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	// Try common service directories
	serviceDirs := []string{"services", "apps", "packages", "microservices"}
	
	for _, dir := range serviceDirs {
		dirServices, err := c.discoverServicesInDir(worktree, dir)
		if err == nil {
			services = append(services, dirServices...)
		}
	}

	return services, nil
}

func (c *SSHClient) discoverServicesInDir(worktree *git.Worktree, dirPath string) ([]ServiceInfo, error) {
	var services []ServiceInfo

	// This is a simplified implementation
	// In a real scenario, you would walk the filesystem and look for service indicators
	// like package.json, Dockerfile, go.mod, etc.
	
	return services, nil
}

func (c *SSHClient) convertToSSHURL(httpsURL string) string {
	// Convert HTTPS GitHub URLs to SSH format
	if strings.HasPrefix(httpsURL, "https://github.com/") {
		// Extract owner/repo from URL
		path := strings.TrimPrefix(httpsURL, "https://github.com/")
		path = strings.TrimSuffix(path, ".git")
		return fmt.Sprintf("git@github.com:%s.git", path)
	}
	
	// If it's already an SSH URL, return as is
	if strings.HasPrefix(httpsURL, "git@") {
		return httpsURL
	}
	
	// For other URLs, assume they're SSH-compatible
	return httpsURL
}

func (c *SSHClient) GetRepository(ctx context.Context, owner, repo string) (*github.Repository, error) {
	if c.gh == nil {
		return nil, fmt.Errorf("GitHub API client not configured")
	}
	
	repository, _, err := c.gh.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}
	return repository, nil
}

func (c *SSHClient) GetWorkflowRuns(ctx context.Context, owner, repo string, workflowID int64, limit int) ([]WorkflowRun, error) {
	if c.gh == nil {
		return nil, fmt.Errorf("GitHub API client not configured")
	}

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

func (c *SSHClient) ListWorkflows(ctx context.Context, owner, repo string) ([]*github.Workflow, error) {
	if c.gh == nil {
		return nil, fmt.Errorf("GitHub API client not configured")
	}

	workflows, _, err := c.gh.Actions.ListWorkflows(ctx, owner, repo, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflows: %w", err)
	}

	return workflows.Workflows, nil
}