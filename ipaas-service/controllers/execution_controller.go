package controllers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/models"
	"github.com/wso2/integration-control-plane/ipaas-service/services"
	"github.com/wso2/integration-control-plane/ipaas-service/utils"
)

// ExecutionController handles HTTP requests for CronJob execution history and manual triggers.
type ExecutionController interface {
	ListExecutions(w http.ResponseWriter, r *http.Request)
	TriggerExecution(w http.ResponseWriter, r *http.Request)
	GetExecutionConfigs(w http.ResponseWriter, r *http.Request)
	GetExecutionArguments(w http.ResponseWriter, r *http.Request)
	UpdateJobConfigs(w http.ResponseWriter, r *http.Request)
}

type executionController struct {
	service       services.ExecutionService
	configService services.ExecutionConfigService
}

func NewExecutionController(service services.ExecutionService, configService services.ExecutionConfigService) ExecutionController {
	return &executionController{service: service, configService: configService}
}

func (c *executionController) TriggerExecution(w http.ResponseWriter, r *http.Request) {
	org, ok := orgHandle(r)
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	componentName := r.PathValue("componentName")
	environment := r.PathValue("environment")
	projectName := r.URL.Query().Get("projectName")

	execution, err := c.service.TriggerExecution(r.Context(), org, projectName, componentName, environment)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorized) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		if errors.Is(err, services.ErrScheduleNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, "schedule not found")
			return
		}
		slog.ErrorContext(r.Context(), "trigger execution failed",
			"error", err, "org", org, "project", projectName, "component", componentName, "environment", environment)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to trigger execution")
		return
	}
	utils.WriteSuccessResponse(w, http.StatusCreated, execution)
}

func (c *executionController) ListExecutions(w http.ResponseWriter, r *http.Request) {
	org, ok := orgHandle(r)
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	componentName := r.PathValue("componentName")
	environment := r.PathValue("environment")
	projectName := r.URL.Query().Get("projectName")

	list, err := c.service.ListExecutions(r.Context(), org, projectName, componentName, environment)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorized) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		if errors.Is(err, services.ErrScheduleNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, "schedule not found")
			return
		}
		slog.ErrorContext(r.Context(), "list executions failed",
			"error", err, "org", org, "project", projectName, "component", componentName, "environment", environment)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to list executions")
		return
	}
	utils.WriteSuccessResponse(w, http.StatusOK, list)
}

// GetExecutionConfigs handles GET /components/{componentName}/execution-configs?releaseId=...
func (c *executionController) GetExecutionConfigs(w http.ResponseWriter, r *http.Request) {
	componentName := r.PathValue("componentName")
	releaseID := r.URL.Query().Get("releaseId")

	configs, err := c.configService.GetExecutionConfigs(r.Context(), componentName, releaseID)
	if err != nil {
		slog.ErrorContext(r.Context(), "get execution configs failed", "error", err, "component", componentName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to get execution configs")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, configs)
}

// GetExecutionArguments handles GET /components/{componentName}/executions/{runId}/arguments?releaseId=...
func (c *executionController) GetExecutionArguments(w http.ResponseWriter, r *http.Request) {
	componentName := r.PathValue("componentName")
	runID := r.PathValue("runId")
	releaseID := r.URL.Query().Get("releaseId")

	args, err := c.configService.GetExecutionArguments(r.Context(), runID, componentName, releaseID)
	if err != nil {
		slog.ErrorContext(r.Context(), "get execution arguments failed", "error", err, "component", componentName, "runId", runID)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to get execution arguments")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, args)
}

// UpdateJobConfigs handles PUT /components/{componentName}/job-configs
func (c *executionController) UpdateJobConfigs(w http.ResponseWriter, r *http.Request) {
	var input models.UpdateJobConfigsInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ok2, err := c.configService.UpdateJobConfigs(r.Context(), &input)
	if err != nil {
		slog.ErrorContext(r.Context(), "update job configs failed", "error", err)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to update job configs")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, map[string]bool{"success": ok2})
}
