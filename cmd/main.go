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

	server.RegisterHandler("get_issue_with_comments", jiraHandler.GetIssueWithComments)

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
