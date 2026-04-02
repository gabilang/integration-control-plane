package controllers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/services"
	"github.com/wso2/integration-control-plane/ipaas-service/utils"
)

// ExecutionController handles HTTP requests for listing CronJob execution history.
type ExecutionController interface {
	ListExecutions(w http.ResponseWriter, r *http.Request)
}

type executionController struct {
	service services.ExecutionService
}

func NewExecutionController(service services.ExecutionService) ExecutionController {
	return &executionController{service: service}
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
