package kubernetes

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"dev-dashboard/pkg/types"

	"gopkg.in/yaml.v3"
)

type KustomizationConfig struct {
	Images []struct {
		Name    string `yaml:"name"`
		NewName string `yaml:"newName"`
		NewTag  string `yaml:"newTag"`
	} `yaml:"images"`
}

type Scanner struct{}

func NewScanner() *Scanner {
	return &Scanner{}
}

func (s *Scanner) ScanRepository(repoPath string, repositoryID int64) ([]*types.Deployment, error) {
	var deployments []*types.Deployment

	servicesPath := filepath.Join(repoPath, "services")
	if _, err := fs.Stat(os.DirFS(repoPath), "services"); err != nil {
		return deployments, nil
	}

	err := filepath.WalkDir(servicesPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(d.Name(), "kustomization.yaml") {
			return nil
		}

		deployment, err := s.parseKustomizationFile(path, repositoryID)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		if deployment != nil {
			deployments = append(deployments, deployment)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan kubernetes repository: %w", err)
	}

	return deployments, nil
}

func (s *Scanner) parseKustomizationFile(filePath string, repositoryID int64) (*types.Deployment, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var config KustomizationConfig
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if len(config.Images) == 0 {
		return nil, nil
	}

	// Extract service, environment, region, and namespace from path
	// Expected path: kubernetes-resources/services/service-b/overlays/prd/us-west-2/ns-a/kustomization.yaml
	pathParts := strings.Split(filepath.Dir(filePath), string(filepath.Separator))
	if len(pathParts) < 6 {
		return nil, fmt.Errorf("invalid path structure: %s", filePath)
	}

	var serviceName, environment, region, namespace string
	for i, part := range pathParts {
		if part == "services" && i+5 < len(pathParts) {
			serviceName = pathParts[i+1]
			// Skip the "overlays" directory at pathParts[i+2]
			environment = pathParts[i+3]
			region = pathParts[i+4]
			namespace = pathParts[i+5]
			break
		}
	}

	if serviceName == "" || environment == "" || region == "" || namespace == "" {
		return nil, fmt.Errorf("could not extract service info from path: %s", filePath)
	}

	// Find the image for this service
	var imageTag string
	for _, image := range config.Images {
		if strings.Contains(image.Name, serviceName) || strings.Contains(image.NewName, serviceName) {
			imageTag = image.NewTag
			break
		}
	}

	if imageTag == "" {
		return nil, nil
	}

	deployment := &types.Deployment{
		KubernetesRepoID: repositoryID,
		Environment:      environment,
		Region:           region,
		Namespace:        namespace,
		Tag:              imageTag,
		Path:             filePath,
		CommitSHA:        "", // Will be populated when matching with monorepo commits
	}

	return deployment, nil
}