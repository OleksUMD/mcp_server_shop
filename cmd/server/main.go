package main

import (
	"os"

	"github.com/OleksUMD/mcp_server_shop/configs"
	"github.com/OleksUMD/mcp_server_shop/internal/mcp"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.SetFormatter(&logrus.JSONFormatter{})

	config, err := configs.Load()
	if err != nil {
		logger.WithError(err).Fatal("Failed to load configuration")
	}

	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		logger.WithError(err).Warn("Invalid log level, using info")
		level = logrus.InfoLevel
	}

	logger.SetLevel(level)

	logger.WithFields(logrus.Fields{
		"api_url":               config.ApiURL,
		"log_level":             config.LogLevel,
		"auth_token_configures": config.AuthToken != "",
	}).Info("Starting MCP E-commerce Server")

	//restClient := client.NewRestClient(config.ApiURL, config.AuthToken, logger)
	toolRegistry := mcp.NewRegistry(logger)
	mcpServer := mcp.NewServer(toolRegistry, logger)
	logger.Info("Starting server on stdio transport")

	if err := mcpServer.Start(); err != nil {
		logger.WithError(err).Fatal("Server error")
	}
}
