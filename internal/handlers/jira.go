package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"mcp-jira/internal/config"
	"mcp-jira/pkg/jira"
)

type JiraHandler struct {
	client *jira.Client
}

func NewJiraHandler(cfg config.JiraConfig) *JiraHandler {
	return &JiraHandler{
		client: jira.NewClient(&cfg, false),
	}
}

func (h *JiraHandler) GetTaskWithComments(params json.RawMessage) (interface{}, error) {
	var req struct {
		Key string `json:"task_key"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.Key == "" {
		return nil, fmt.Errorf("task_key is required")
	}

	issue, err := h.client.GetIssue(req.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	return map[string]interface{}{
		"titulo":      fieldAsString(issue.Fields["summary"]),
		"descricao":   fieldAsJSON(issue.Fields["description"]),
		"refinamento": h.findRefinamentoComment(issue),
	}, nil
}

func fieldAsString(field interface{}) string {
	s, _ := field.(string)
	return strings.TrimSpace(s)
}

func fieldAsJSON(field interface{}) string {
	if field == nil {
		return ""
	}

	raw, err := json.Marshal(field)
	if err != nil {
		return ""
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, raw, "", "  "); err != nil {
		return string(raw)
	}
	return pretty.String()
}

func (h *JiraHandler) CreateIssue(params json.RawMessage) (interface{}, error) {
	var req struct {
		ProjectKey  string   `json:"project_key"`
		Summary     string   `json:"summary"`
		Description string   `json:"description"`
		IssueType   string   `json:"issue_type"`
		ParentKey   string   `json:"parent_key"`
		Labels      []string `json:"labels"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.Summary == "" {
		return nil, fmt.Errorf("summary is required")
	}
	if req.ProjectKey == "" {
		req.ProjectKey = "LED"
	}
	if req.IssueType == "" {
		req.IssueType = "História"
	}

	createReq := jira.CreateIssueRequest{
		Fields: jira.CreateIssueFields{
			Project:   jira.ProjectRef{Key: req.ProjectKey},
			Summary:   req.Summary,
			IssueType: jira.IssueTypeRef{Name: req.IssueType},
		},
	}
	if req.Description != "" {
		createReq.Fields.Description = jira.TextToADF(req.Description)
	}
	if len(req.Labels) > 0 {
		createReq.Fields.Labels = req.Labels
	}

	issueTypeLower := strings.ToLower(req.IssueType)
	isSubtask := issueTypeLower == "subtarefa" || issueTypeLower == "subtask"

	if req.ParentKey != "" {
		parent, err := h.client.GetIssue(req.ParentKey)
		if err == nil {
			if comps, ok := parent.Fields["components"].([]interface{}); ok && len(comps) > 0 {
				createReq.Fields.Components = comps
			}
			if team, ok := parent.Fields["customfield_11096"].([]interface{}); ok && len(team) > 0 {
				createReq.Fields.LEDTeam = team
			}
			if sprints, ok := parent.Fields["customfield_10020"].([]interface{}); ok && len(sprints) > 0 {
				if sprint, ok := sprints[0].(map[string]interface{}); ok {
					if id, ok := sprint["id"].(float64); ok {
						createReq.Fields.Sprint = int(id)
					}
				}
			}
			if isSubtask {
				createReq.Fields.Parent = &jira.ProjectRef{Key: req.ParentKey}
			} else if grandparent, ok := parent.Fields["parent"].(map[string]interface{}); ok {
				if gpKey, ok := grandparent["key"].(string); ok && gpKey != "" {
					createReq.Fields.Parent = &jira.ProjectRef{Key: gpKey}
				}
			}
		}
	}

	result, err := h.client.CreateIssue(createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	if req.ParentKey != "" && !isSubtask {
		linkReq := jira.IssueLinkRequest{
			Type:         jira.IssueLinkType{Name: "Parent-Child"},
			OutwardIssue: jira.ProjectRef{Key: req.ParentKey},
			InwardIssue:  jira.ProjectRef{Key: result.Key},
		}
		_ = h.client.CreateIssueLink(linkReq)
	}

	return map[string]interface{}{
		"key": result.Key,
		"id":  result.ID,
		"url": result.URL,
	}, nil
}

func (h *JiraHandler) findRefinamentoComment(issue *jira.Issue) string {
	const phrase = "refinamento técnico"

	commentField, ok := issue.Fields["comment"].(map[string]interface{})
	if !ok {
		return "Nenhum comentário de refinamento técnico encontrado"
	}

	commentsList, ok := commentField["comments"].([]interface{})
	if !ok {
		return "Nenhum comentário de refinamento técnico encontrado"
	}

	for _, c := range commentsList {
		commentJSON, err := json.Marshal(c)
		if err != nil {
			continue
		}

		if !strings.Contains(strings.ToLower(string(commentJSON)), phrase) {
			continue
		}

		return fieldAsJSON(c)
	}

	return "Nenhum comentário de refinamento técnico encontrado"
}
