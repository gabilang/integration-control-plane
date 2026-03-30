package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wso2/integration-control-plane/ipaas-service/api"
	"github.com/wso2/integration-control-plane/ipaas-service/clients/icp"
	"github.com/wso2/integration-control-plane/ipaas-service/clients/observability"
	"github.com/wso2/integration-control-plane/ipaas-service/clients/openchoreo"
	"github.com/wso2/integration-control-plane/ipaas-service/config"
	"github.com/wso2/integration-control-plane/ipaas-service/controllers"
	"github.com/wso2/integration-control-plane/ipaas-service/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	setupLogger(cfg.LogLevel)

	// Wire dependencies
	projectClient := openchoreo.NewProjectClient(cfg.PlatformAPI.BaseURL, cfg.PlatformAPI.HostHeader)
	projectService := services.NewProjectService(projectClient)
	projectController := controllers.NewProjectController(projectService)

	environmentClient := openchoreo.NewEnvironmentClient(cfg.PlatformAPI.BaseURL, cfg.PlatformAPI.HostHeader)
	environmentService := services.NewEnvironmentService(environmentClient)
	environmentController := controllers.NewEnvironmentController(environmentService)

	componentClient := openchoreo.NewComponentClient(cfg.PlatformAPI.BaseURL, cfg.PlatformAPI.HostHeader)
	var observClient observability.Client
	if cfg.Observability.BaseURL != "" {
		observClient = observability.NewClient(cfg.Observability.BaseURL)
	}
	componentService := services.NewComponentService(componentClient, observClient)
	componentController := controllers.NewComponentController(componentService)

	graphqlProxy := icp.NewProxyClient(cfg.ICP.GraphQLURL)
	authProxy := icp.NewProxyClient(cfg.ICP.AuthBaseURL)
	observabilityProxy := icp.NewProxyClient(cfg.Observability.BaseURL)

	handler := api.NewHandler(api.AppParams{
		ProjectController:     projectController,
		ComponentController:   componentController,
		EnvironmentController: environmentController,
		GraphQLProxy:        graphqlProxy,
		AuthProxy:           authProxy,
		ObservabilityProxy:  observabilityProxy,
	})

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in background
	go func() {
		slog.Info("server started", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}

func setupLogger(level string) {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})))
}
