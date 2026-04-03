package controllers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	jwtmw "github.com/wso2/integration-control-plane/ipaas-service/middleware/jwt"
	"github.com/wso2/integration-control-plane/ipaas-service/models"
	"github.com/wso2/integration-control-plane/ipaas-service/services"
	"github.com/wso2/integration-control-plane/ipaas-service/utils"
)

// ComponentController handles HTTP requests for component operations.
type ComponentController interface {
	ListComponents(w http.ResponseWriter, r *http.Request)
	CreateComponent(w http.ResponseWriter, r *http.Request)
	UpdateBuildParameters(w http.ResponseWriter, r *http.Request)
	UpdateComponent(w http.ResponseWriter, r *http.Request)
	DeleteComponent(w http.ResponseWriter, r *http.Request)
	TriggerBuild(w http.ResponseWriter, r *http.Request)
	ListBuilds(w http.ResponseWriter, r *http.Request)
	GetBuildStatus(w http.ResponseWriter, r *http.Request)
	GetBuildLogs(w http.ResponseWriter, r *http.Request)
	GetComponentRepository(w http.ResponseWriter, r *http.Request)
	GetCommitHistory(w http.ResponseWriter, r *http.Request)
	GetComponentLabels(w http.ResponseWriter, r *http.Request)
	GetDeploymentTrack(w http.ResponseWriter, r *http.Request)
}

type componentController struct {
	service services.ComponentService
}

func NewComponentController(service services.ComponentService) ComponentController {
	return &componentController{service: service}
}

func (c *componentController) ListComponents(w http.ResponseWriter, r *http.Request) {
	claims := jwtmw.ClaimsFromContext(r.Context())
	if claims == nil || claims.OrgHandle == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	org := claims.OrgHandle
	projectName := r.URL.Query().Get("projectName")
	cursor := r.URL.Query().Get("cursor")

	list, err := c.service.ListComponents(r.Context(), org, projectName, 100, cursor)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorized) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		slog.ErrorContext(r.Context(), "list components failed",
			"error", err, "org", org, "project", projectName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to list components")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, list)
}

func (c *componentController) CreateComponent(w http.ResponseWriter, r *http.Request) {
	claims := jwtmw.ClaimsFromContext(r.Context())
	if claims == nil || claims.OrgHandle == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	org := claims.OrgHandle

	var req models.CreateComponentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Metadata.Name == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "metadata.name is required")
		return
	}

	projectName := ""
	if req.Spec.Owner != nil {
		projectName = req.Spec.Owner.ProjectName
	}

	resp, err := c.service.CreateComponent(r.Context(), org, projectName, &req)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorized) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		slog.ErrorContext(r.Context(), "create component failed",
			"error", err, "org", org, "project", projectName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to create component")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusCreated, resp)
}

func (c *componentController) UpdateBuildParameters(w http.ResponseWriter, r *http.Request) {
	claims := jwtmw.ClaimsFromContext(r.Context())
	if claims == nil || claims.OrgHandle == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	org := claims.OrgHandle
	componentName := r.PathValue("componentName")
	projectName := r.URL.Query().Get("projectName")

	var req models.UpdateBuildParametersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Workflow == nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "workflow is required")
		return
	}

	component, err := c.service.UpdateBuildParameters(r.Context(), org, projectName, componentName, &req)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorized) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		if errors.Is(err, services.ErrComponentNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, "component not found")
			return
		}
		slog.ErrorContext(r.Context(), "update build parameters failed",
			"error", err, "org", org, "project", projectName, "component", componentName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to update build parameters")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, component)
}

func (c *componentController) UpdateComponent(w http.ResponseWriter, r *http.Request) {
	claims := jwtmw.ClaimsFromContext(r.Context())
	if claims == nil || claims.OrgHandle == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	org := claims.OrgHandle
	componentName := r.PathValue("componentName")
	projectName := r.URL.Query().Get("projectName")

	var req models.UpdateComponentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	component, err := c.service.UpdateComponent(r.Context(), org, projectName, componentName, &req)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorized) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		if errors.Is(err, services.ErrComponentNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, "component not found")
			return
		}
		slog.ErrorContext(r.Context(), "update component failed",
			"error", err, "org", org, "project", projectName, "component", componentName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to update component")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, component)
}

func (c *componentController) DeleteComponent(w http.ResponseWriter, r *http.Request) {
	claims := jwtmw.ClaimsFromContext(r.Context())
	if claims == nil || claims.OrgHandle == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	org := claims.OrgHandle
	componentName := r.PathValue("componentName")
	projectName := r.URL.Query().Get("projectName")

	if err := c.service.DeleteComponent(r.Context(), org, projectName, componentName); err != nil {
		if errors.Is(err, services.ErrUnauthorized) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		if errors.Is(err, services.ErrComponentNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, "component not found")
			return
		}
		slog.ErrorContext(r.Context(), "delete component failed",
			"error", err, "org", org, "project", projectName, "component", componentName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to delete component")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *componentController) TriggerBuild(w http.ResponseWriter, r *http.Request) {
	claims := jwtmw.ClaimsFromContext(r.Context())
	if claims == nil || claims.OrgHandle == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	org := claims.OrgHandle
	componentName := r.PathValue("componentName")
	projectName := r.URL.Query().Get("projectName")

	run, err := c.service.TriggerBuild(r.Context(), org, projectName, componentName)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorized) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		if errors.Is(err, services.ErrComponentNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, "component not found")
			return
		}
		slog.ErrorContext(r.Context(), "trigger build failed",
			"error", err, "org", org, "project", projectName, "component", componentName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to trigger build")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusCreated, run)
}

func (c *componentController) ListBuilds(w http.ResponseWriter, r *http.Request) {
	claims := jwtmw.ClaimsFromContext(r.Context())
	if claims == nil || claims.OrgHandle == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	org := claims.OrgHandle
	componentName := r.PathValue("componentName")
	projectName := r.URL.Query().Get("projectName")
	cursor := r.URL.Query().Get("cursor")

	list, err := c.service.ListBuilds(r.Context(), org, projectName, componentName, 20, cursor)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorized) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		slog.ErrorContext(r.Context(), "list builds failed",
			"error", err, "org", org, "project", projectName, "component", componentName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to list builds")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, list)
}

func (c *componentController) GetBuildStatus(w http.ResponseWriter, r *http.Request) {
	claims := jwtmw.ClaimsFromContext(r.Context())
	if claims == nil || claims.OrgHandle == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	org := claims.OrgHandle
	componentName := r.PathValue("componentName")
	buildName := r.PathValue("buildName")
	projectName := r.URL.Query().Get("projectName")

	run, err := c.service.GetBuildStatus(r.Context(), org, projectName, componentName, buildName)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorized) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		if errors.Is(err, services.ErrBuildNotFound) {
			utils.WriteErrorResponse(w, http.StatusNotFound, "build not found")
			return
		}
		slog.ErrorContext(r.Context(), "get build status failed",
			"error", err, "org", org, "project", projectName, "component", componentName, "build", buildName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to get build status")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, run)
}

func (c *componentController) GetBuildLogs(w http.ResponseWriter, r *http.Request) {
	claims := jwtmw.ClaimsFromContext(r.Context())
	if claims == nil || claims.OrgHandle == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	org := claims.OrgHandle
	componentName := r.PathValue("componentName")
	buildName := r.PathValue("buildName")
	projectName := r.URL.Query().Get("projectName")

	logs, err := c.service.GetBuildLogs(r.Context(), org, projectName, componentName, buildName)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorized) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		if errors.Is(err, services.ErrLogsUnavailable) {
			utils.WriteErrorResponse(w, http.StatusServiceUnavailable, "build logs service not available")
			return
		}
		slog.ErrorContext(r.Context(), "get build logs failed",
			"error", err, "org", org, "project", projectName, "component", componentName, "build", buildName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to get build logs")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, logs)
}

func (c *componentController) GetComponentRepository(w http.ResponseWriter, r *http.Request) {
	claims := jwtmw.ClaimsFromContext(r.Context())
	if claims == nil || claims.OrgHandle == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	componentName := r.PathValue("componentName")
	projectName := r.URL.Query().Get("projectName")

	repo, err := c.service.GetComponentRepository(r.Context(), projectName, componentName)
	if err != nil {
		slog.ErrorContext(r.Context(), "get component repository failed", "error", err, "component", componentName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to get component repository")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, repo)
}

func (c *componentController) GetCommitHistory(w http.ResponseWriter, r *http.Request) {
	claims := jwtmw.ClaimsFromContext(r.Context())
	if claims == nil || claims.OrgHandle == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	componentName := r.PathValue("componentName")
	branch := r.URL.Query().Get("branch")

	list, err := c.service.GetCommitHistory(r.Context(), componentName, branch)
	if err != nil {
		slog.ErrorContext(r.Context(), "get commit history failed", "error", err, "component", componentName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to get commit history")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, list)
}

func (c *componentController) GetComponentLabels(w http.ResponseWriter, r *http.Request) {
	claims := jwtmw.ClaimsFromContext(r.Context())
	if claims == nil || claims.OrgHandle == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	projectName := r.URL.Query().Get("projectName")
	orgIDStr := r.URL.Query().Get("orgId")

	var orgID int
	if orgIDStr != "" {
		var parseErr error
		orgID, parseErr = strconv.Atoi(orgIDStr)
		if parseErr != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, "orgId must be an integer")
			return
		}
	}

	list, err := c.service.GetComponentLabels(r.Context(), projectName, orgID)
	if err != nil {
		slog.ErrorContext(r.Context(), "get component labels failed", "error", err)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to get component labels")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, list)
}

func (c *componentController) GetDeploymentTrack(w http.ResponseWriter, r *http.Request) {
	claims := jwtmw.ClaimsFromContext(r.Context())
	if claims == nil || claims.OrgHandle == "" {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	componentName := r.PathValue("componentName")

	track, err := c.service.GetDeploymentTrack(r.Context(), componentName)
	if err != nil {
		slog.ErrorContext(r.Context(), "get deployment track failed", "error", err, "component", componentName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to get deployment track")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, track)
}
