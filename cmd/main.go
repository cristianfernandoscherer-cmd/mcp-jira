package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"mcp-jira/internal/config"
	"mcp-jira/internal/handlers"
	"mcp-jira/pkg/mcp"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	server := mcp.NewServer(cfg.MCP.Mode)

	jiraHandler := handlers.NewJiraHandler(cfg.Jira)

	server.RegisterHandler("get_task_with_comments", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"task_key": map[string]interface{}{
				"type":        "string",
				"description": "Jira issue key (e.g. LED-53292)",
			},
		},
		"required": []string{"task_key"},
	}, jiraHandler.GetTaskWithComments)

	server.RegisterHandler("create_issue", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_key": map[string]interface{}{
				"type":        "string",
				"description": "Chave do projeto Jira (ex: LED). Padrão: LED",
			},
			"summary": map[string]interface{}{
				"type":        "string",
				"description": "Título da issue (obrigatório)",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Descrição da issue em texto simples",
			},
			"issue_type": map[string]interface{}{
				"type":        "string",
				"description": "Tipo da issue (Story, Bug, Task). Padrão: Story",
			},
			"parent_key": map[string]interface{}{
				"type":        "string",
				"description": "Chave da issue pai (ex: LED-53292)",
			},
		},
		"required": []string{"summary"},
	}, jiraHandler.CreateIssue)

	log.Printf("Starting MCP server in %s mode", cfg.MCP.Mode)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.Start(); err != nil {
			log.Printf("Server error: %v", err)
			done <- syscall.SIGTERM
		}
	}()

	<-done
	log.Println("Shutting down server...")
}
