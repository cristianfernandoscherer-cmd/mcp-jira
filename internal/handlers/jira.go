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
