package jira

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	username   string
	authMethod string // "bearer", "basic", or "token"
	client     *http.Client
}

type Issue struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Fields struct {
		Summary     string `json:"summary"`
		Description string `json:"description"`
		Status      struct {
			Name string `json:"name"`
		} `json:"status"`
		IssueType struct {
			Name string `json:"name"`
		} `json:"issuetype"`
		Priority struct {
			Name string `json:"name"`
		} `json:"priority"`
	} `json:"fields"`
}

func NewClient(baseURL, token string) *Client {
	return NewClientWithAuth(baseURL, "", token, "")
}

func NewClientWithAuth(baseURL, username, token, authMethod string) *Client {
	// Clean up baseURL
	baseURL = strings.TrimSuffix(baseURL, "/")
	
	// Remove any existing API path
	baseURL = strings.TrimSuffix(baseURL, "/rest/api/3")
	baseURL = strings.TrimSuffix(baseURL, "/rest/api/2")
	baseURL = strings.TrimSuffix(baseURL, "/rest/api")
	
	// Auto-detect authentication method if not specified
	if authMethod == "" {
		if username != "" && token != "" {
			authMethod = "basic"
		} else if token != "" {
			// Try bearer first, fallback to basic if needed
			authMethod = "bearer"
		} else {
			authMethod = "basic"
		}
	}

	return &Client{
		baseURL:    baseURL,
		username:   username,
		token:      token,
		authMethod: authMethod,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) getAPIURL(apiVersion string) string {
	if apiVersion == "" {
		apiVersion = "2" // Default to API v2 for enterprise compatibility
	}
	return fmt.Sprintf("%s/rest/api/%s", c.baseURL, apiVersion)
}

func (c *Client) setAuthHeaders(req *http.Request) {
	switch c.authMethod {
	case "basic":
		if c.username != "" && c.token != "" {
			// Basic auth with username:token
			auth := base64.StdEncoding.EncodeToString([]byte(c.username + ":" + c.token))
			req.Header.Set("Authorization", "Basic "+auth)
		}
	case "bearer":
		if c.token != "" {
			req.Header.Set("Authorization", "Bearer "+c.token)
		}
	case "token":
		if c.token != "" {
			// Some enterprise setups use X-Atlassian-Token
			req.Header.Set("X-Atlassian-Token", c.token)
		}
	}
}

func (c *Client) GetIssue(issueKey string) (*Issue, error) {
	if c.token == "" && c.username == "" {
		return nil, fmt.Errorf("JIRA authentication not configured")
	}

	// Try API v2 first (enterprise), then v3 (cloud)
	apiVersions := []string{"2", "3"}
	
	for _, apiVersion := range apiVersions {
		issue, err := c.getIssueWithAPI(issueKey, apiVersion)
		if err == nil {
			return issue, nil
		}
		
		// If it's an auth error, don't try other versions
		if strings.Contains(err.Error(), "unauthorized") || strings.Contains(err.Error(), "401") {
			return nil, err
		}
	}
	
	return nil, fmt.Errorf("failed to fetch issue %s with both API v2 and v3", issueKey)
}

func (c *Client) getIssueWithAPI(issueKey, apiVersion string) (*Issue, error) {
	url := fmt.Sprintf("%s/issue/%s", c.getAPIURL(apiVersion), issueKey)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set auth headers
	c.setAuthHeaders(req)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to %s: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("issue %s not found", issueKey)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("unauthorized (401) - check your JIRA credentials and permissions")
	}

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("forbidden (403) - check your JIRA permissions for issue %s", issueKey)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JIRA API error %d: %s", resp.StatusCode, string(body))
	}

	var issue Issue
	if err := json.Unmarshal(body, &issue); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w (body: %s)", err, string(body))
	}

	return &issue, nil
}

func (c *Client) TestConnection() error {
	if c.token == "" && c.username == "" {
		return fmt.Errorf("JIRA authentication not configured")
	}

	// Try both API versions
	apiVersions := []string{"2", "3"}
	
	for _, apiVersion := range apiVersions {
		err := c.testConnectionWithAPI(apiVersion)
		if err == nil {
			return nil
		}
		
		// If it's an auth error, don't try other versions
		if strings.Contains(err.Error(), "unauthorized") || strings.Contains(err.Error(), "401") {
			return err
		}
	}
	
	return fmt.Errorf("failed to connect to JIRA with both API v2 and v3")
}

func (c *Client) testConnectionWithAPI(apiVersion string) error {
	url := fmt.Sprintf("%s/myself", c.getAPIURL(apiVersion))
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.setAuthHeaders(req)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request to %s: %w", url, err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("unauthorized (401) - check your JIRA credentials and URL")
	}

	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("forbidden (403) - check your JIRA permissions")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JIRA API error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}