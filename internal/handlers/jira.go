package handlers

import (
	"encoding/json"
	"fmt"
	"strings"

	"mcp-jira/internal/config"
	"mcp-jira/pkg/jira"
)

type Comment struct {
	Body CommentBody `json:"body"`
}

type CommentBody struct {
	Content []ContentBlock `json:"content"`
}

type ContentBlock struct {
	Content []TextNode `json:"content"`
}

type TextNode struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

type JiraHandler struct {
	client *jira.Client
}

func NewJiraHandler(cfg config.JiraConfig) *JiraHandler {
	return &JiraHandler{
		client: jira.NewClient(&cfg, false),
	}
}

func (h *JiraHandler) GetIssueWithComments(params json.RawMessage) (interface{}, error) {
	var req struct {
		Key string `json:"issue_key"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.Key == "" {
		return nil, fmt.Errorf("issue_key is required")
	}

	issue, err := h.client.GetIssue(req.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	refinamento := h.findRefinamentoComment(issue)

	return map[string]interface{}{
		"refinamento": refinamento,
	}, nil
}

func (h *JiraHandler) findRefinamentoComment(issue *jira.Issue) string {
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

		var comment Comment
		if err := json.Unmarshal(commentJSON, &comment); err != nil {
			continue
		}

		text := h.extractTextFromComment(comment)
		if strings.Contains(strings.ToLower(text), "refinamento técnico") {
			return text
		}
	}

	return "Nenhum comentário de refinamento técnico encontrado"
}

func (h *JiraHandler) extractTextFromComment(comment Comment) string {
	var result strings.Builder

	for _, block := range comment.Body.Content {
		for _, node := range block.Content {
			if node.Type == "text" {
				result.WriteString(node.Text)
			}
		}
		result.WriteString("\n")
	}

	return strings.TrimSpace(result.String())
}
