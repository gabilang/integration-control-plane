package controllers

import (
	"log/slog"
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/services"
	"github.com/wso2/integration-control-plane/ipaas-service/utils"
)

type RuntimeController interface {
	ListRuntimes(w http.ResponseWriter, r *http.Request)
	ListProjectRuntimes(w http.ResponseWriter, r *http.Request)
	DeleteRuntime(w http.ResponseWriter, r *http.Request)
}

type runtimeController struct {
	service services.RuntimeService
}

func NewRuntimeController(service services.RuntimeService) RuntimeController {
	return &runtimeController{service: service}
}

// ListRuntimes handles GET /components/{componentName}/runtimes?environmentId=...&projectName=...
func (c *runtimeController) ListRuntimes(w http.ResponseWriter, r *http.Request) {
	componentName := r.PathValue("componentName")
	environmentID := r.URL.Query().Get("environmentId")
	projectName := r.URL.Query().Get("projectName")

	if environmentID == "" || projectName == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "environmentId and projectName are required")
		return
	}

	list, err := c.service.ListRuntimes(r.Context(), environmentID, projectName, componentName)
	if err != nil {
		slog.ErrorContext(r.Context(), "list runtimes failed", "error", err, "component", componentName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to list runtimes")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, list)
}

// ListProjectRuntimes handles GET /projects/{projectName}/runtimes?environmentId=...
func (c *runtimeController) ListProjectRuntimes(w http.ResponseWriter, r *http.Request) {
	projectName := r.PathValue("projectName")
	environmentID := r.URL.Query().Get("environmentId")

	if environmentID == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "environmentId is required")
		return
	}

	list, err := c.service.ListProjectRuntimes(r.Context(), environmentID, projectName)
	if err != nil {
		slog.ErrorContext(r.Context(), "list project runtimes failed", "error", err, "project", projectName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to list project runtimes")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, list)
}

// DeleteRuntime handles DELETE /runtimes/{runtimeId}
func (c *runtimeController) DeleteRuntime(w http.ResponseWriter, r *http.Request) {
	runtimeID := r.PathValue("runtimeId")

	if err := c.service.DeleteRuntime(r.Context(), runtimeID); err != nil {
		slog.ErrorContext(r.Context(), "delete runtime failed", "error", err, "runtimeId", runtimeID)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to delete runtime")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
