package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"mcp-jira/internal/config"
)

type Client struct {
	BaseURL  string
	Email    string
	APIToken string
	HTTP     *http.Client
	Debug    bool
}

type Issue struct {
	Key    string                 `json:"key"`
	Fields map[string]interface{} `json:"fields"`
}

type CreateIssueRequest struct {
	Fields CreateIssueFields `json:"fields"`
}

type CreateIssueFields struct {
	Project     ProjectRef    `json:"project"`
	Summary     string        `json:"summary"`
	Description interface{}   `json:"description,omitempty"`
	IssueType   IssueTypeRef  `json:"issuetype"`
	Parent      *ProjectRef   `json:"parent,omitempty"`
	Components  []interface{} `json:"components,omitempty"`
	LEDTeam     []interface{} `json:"customfield_11096,omitempty"` // "LED Team" — campo obrigatório no projeto LED
	Sprint      interface{}   `json:"customfield_10020,omitempty"` // Sprint — herdado do pai quando disponível
	Labels      []string      `json:"labels,omitempty"`            // Categorias da issue
}

type ProjectRef struct {
	Key string `json:"key"`
}

type IssueTypeRef struct {
	Name string `json:"name"`
}

type CreateIssueResponse struct {
	ID  string `json:"id"`
	Key string `json:"key"`
	URL string `json:"self"`
}

func NewClient(cfg *config.JiraConfig, debug bool) *Client {
	return &Client{
		BaseURL:  cfg.BaseURL,
		Email:    cfg.Email,
		APIToken: cfg.APIToken,
		HTTP: &http.Client{
			Timeout: 30 * time.Second,
		},
		Debug: debug,
	}
}

func (c *Client) doRequest(method, endpoint string, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.BaseURL, endpoint)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.Email, c.APIToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

func (c *Client) GetIssue(issueKey string) (*Issue, error) {
	endpoint := fmt.Sprintf("/rest/api/3/issue/%s", issueKey)
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var issue Issue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &issue, nil
}

type IssueLinkRequest struct {
	Type         IssueLinkType `json:"type"`
	OutwardIssue ProjectRef    `json:"outwardIssue"`
	InwardIssue  ProjectRef    `json:"inwardIssue"`
}

type IssueLinkType struct {
	Name string `json:"name"`
}

func (c *Client) CreateIssueLink(req IssueLinkRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.doRequest("POST", "/rest/api/3/issueLink", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (c *Client) CreateIssue(req CreateIssueRequest) (*CreateIssueResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.doRequest("POST", "/rest/api/3/issue", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var result CreateIssueResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func TextToADF(text string) map[string]interface{} {
	return map[string]interface{}{
		"type":    "doc",
		"version": 1,
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": text,
					},
				},
			},
		},
	}
}