package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Jira  JiraConfig `json:"jira"`
	MCP   MCPConfig  `json:"mcp"`
	Debug bool       `json:"debug"`
}

type JiraConfig struct {
	BaseURL  string `json:"base_url"`
	Email    string `json:"email"`
	APIToken string `json:"api_token"`
}

type MCPConfig struct {
	Mode     string `json:"mode"`
	HTTPPort int    `json:"http_port"`
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Jira: JiraConfig{
			BaseURL:  getEnv("JIRA_BASE_URL", ""),
			Email:    getEnv("JIRA_EMAIL", ""),
			APIToken: getEnv("JIRA_API_TOKEN", ""),
		},
		MCP: MCPConfig{
			Mode:     getEnv("MCP_MODE", "stdio"),
			HTTPPort: getEnvAsInt("MCP_HTTP_PORT", 3001),
		},
		Debug: getEnvAsBool("DEBUG", false),
	}

	if cfg.Jira.BaseURL == "" || cfg.Jira.Email == "" || cfg.Jira.APIToken == "" {
		return nil, fmt.Errorf("missing required Jira configuration in environment variables")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
