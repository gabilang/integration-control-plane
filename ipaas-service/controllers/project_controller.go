package controllers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	jwtmw "github.com/wso2/integration-control-plane/ipaas-service/middleware/jwt"
	"github.com/wso2/integration-control-plane/ipaas-service/models"
	"github.com/wso2/integration-control-plane/ipaas-service/services"
	"github.com/wso2/integration-control-plane/ipaas-service/utils"
)

// ProjectController handles HTTP requests for project operations.
type ProjectController interface {
	ListProjects(w http.ResponseWriter, r *http.Request)
	GetProject(w http.ResponseWriter, r *http.Request)
	CreateProject(w http.ResponseWriter, r *http.Request)
	UpdateProject(w http.ResponseWriter, r *http.Request)
	DeleteProject(w http.ResponseWriter, r *http.Request)
}

type projectController struct {
	service services.ProjectService
}

func NewProjectController(service services.ProjectService) ProjectController {
	return &projectController{service: service}
}

// orgHandle extracts the organization handle from the JWT claims stored in context.
func orgHandle(r *http.Request) (string, bool) {
	claims := jwtmw.ClaimsFromContext(r.Context())
	if claims == nil || claims.OrgHandle == "" {
		return "", false
	}
	return claims.OrgHandle, true
}

func (c *projectController) ListProjects(w http.ResponseWriter, r *http.Request) {
	org, ok := orgHandle(r)
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	cursor := r.URL.Query().Get("cursor")

	list, err := c.service.ListProjects(r.Context(), org, 20, cursor)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorized) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		slog.ErrorContext(r.Context(), "list projects failed", "error", err, "org", org)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to list projects")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, list)
}

func (c *projectController) GetProject(w http.ResponseWriter, r *http.Request) {
	org, ok := orgHandle(r)
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	projectName := r.PathValue("projectName")

	project, err := c.service.GetProject(r.Context(), org, projectName)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorized) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		if errors.Is(err, services.ErrProjectNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, "project not found")
			return
		}
		slog.ErrorContext(r.Context(), "get project failed", "error", err, "org", org, "project", projectName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to get project")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, project)
}

func (c *projectController) CreateProject(w http.ResponseWriter, r *http.Request) {
	org, ok := orgHandle(r)
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}

	var req models.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "name is required")
		return
	}

	project, err := c.service.CreateProject(r.Context(), org, &req)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorized) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		slog.ErrorContext(r.Context(), "create project failed", "error", err, "org", org)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to create project")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusCreated, project)
}

func (c *projectController) UpdateProject(w http.ResponseWriter, r *http.Request) {
	org, ok := orgHandle(r)
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	projectName := r.PathValue("projectName")

	var req models.UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	project, err := c.service.UpdateProject(r.Context(), org, projectName, &req)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorized) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		if errors.Is(err, services.ErrProjectNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, "project not found")
			return
		}
		slog.ErrorContext(r.Context(), "update project failed", "error", err, "org", org, "project", projectName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to update project")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, project)
}

func (c *projectController) DeleteProject(w http.ResponseWriter, r *http.Request) {
	org, ok := orgHandle(r)
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	projectName := r.PathValue("projectName")

	err := c.service.DeleteProject(r.Context(), org, projectName)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorized) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		if errors.Is(err, services.ErrProjectNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, "project not found")
			return
		}
		slog.ErrorContext(r.Context(), "delete project failed", "error", err, "org", org, "project", projectName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to delete project")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
