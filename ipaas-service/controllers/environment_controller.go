package controllers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/services"
	"github.com/wso2/integration-control-plane/ipaas-service/utils"
)

// EnvironmentController handles HTTP requests for environment operations.
type EnvironmentController interface {
	ListEnvironments(w http.ResponseWriter, r *http.Request)
	GetEnvironment(w http.ResponseWriter, r *http.Request)
}

type environmentController struct {
	service services.EnvironmentService
}

func NewEnvironmentController(service services.EnvironmentService) EnvironmentController {
	return &environmentController{service: service}
}

func (c *environmentController) ListEnvironments(w http.ResponseWriter, r *http.Request) {
	org, ok := orgHandle(r)
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}

	list, err := c.service.ListEnvironments(r.Context(), org)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorized) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		slog.ErrorContext(r.Context(), "list environments failed", "error", err, "org", org)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to list environments")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, list)
}

func (c *environmentController) GetEnvironment(w http.ResponseWriter, r *http.Request) {
	org, ok := orgHandle(r)
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	environmentName := r.PathValue("environmentName")

	env, err := c.service.GetEnvironment(r.Context(), org, environmentName)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorized) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		if errors.Is(err, services.ErrEnvironmentNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, "environment not found")
			return
		}
		slog.ErrorContext(r.Context(), "get environment failed", "error", err, "org", org, "environment", environmentName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to get environment")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, env)
}
