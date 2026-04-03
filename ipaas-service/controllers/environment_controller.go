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

// EnvironmentController handles HTTP requests for environment operations.
type EnvironmentController interface {
	ListEnvironments(w http.ResponseWriter, r *http.Request)
	GetEnvironment(w http.ResponseWriter, r *http.Request)
	CreateEnvironment(w http.ResponseWriter, r *http.Request)
	UpdateEnvironment(w http.ResponseWriter, r *http.Request)
	DeleteEnvironment(w http.ResponseWriter, r *http.Request)
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

func (c *environmentController) CreateEnvironment(w http.ResponseWriter, r *http.Request) {
	org, ok := orgHandle(r)
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}

	var req models.CreateEnvironmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "name is required")
		return
	}

	env, err := c.service.CreateEnvironment(r.Context(), org, &req)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorized) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		slog.ErrorContext(r.Context(), "create environment failed", "error", err, "org", org)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to create environment")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusCreated, env)
}

func (c *environmentController) UpdateEnvironment(w http.ResponseWriter, r *http.Request) {
	org, ok := orgHandle(r)
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	environmentName := r.PathValue("environmentName")

	var req models.UpdateEnvironmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	env, err := c.service.UpdateEnvironment(r.Context(), org, environmentName, &req)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorized) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		if errors.Is(err, services.ErrEnvironmentNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, "environment not found")
			return
		}
		slog.ErrorContext(r.Context(), "update environment failed", "error", err, "org", org, "environment", environmentName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to update environment")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, env)
}

func (c *environmentController) DeleteEnvironment(w http.ResponseWriter, r *http.Request) {
	org, ok := orgHandle(r)
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	environmentName := r.PathValue("environmentName")

	if err := c.service.DeleteEnvironment(r.Context(), org, environmentName); err != nil {
		if errors.Is(err, services.ErrUnauthorized) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		if errors.Is(err, services.ErrEnvironmentNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, "environment not found")
			return
		}
		slog.ErrorContext(r.Context(), "delete environment failed", "error", err, "org", org, "environment", environmentName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to delete environment")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
