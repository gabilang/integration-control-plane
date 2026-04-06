package controllers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/wso2/integration-control-plane/ipaas-service/services"
	"github.com/wso2/integration-control-plane/ipaas-service/utils"
)

// ResourceTreeController handles HTTP requests for OpenChoreo resource tree APIs.
type ResourceTreeController interface {
	GetResourceTree(w http.ResponseWriter, r *http.Request)
	GetExecutionsFromTree(w http.ResponseWriter, r *http.Request)
	GetResourceEvents(w http.ResponseWriter, r *http.Request)
	GetResourceLogs(w http.ResponseWriter, r *http.Request)
}

type resourceTreeController struct {
	service services.ResourceTreeService
}

func NewResourceTreeController(service services.ResourceTreeService) ResourceTreeController {
	return &resourceTreeController{service: service}
}

// GetResourceTree handles GET /components/{componentName}/environments/{environment}/resource-tree
func (c *resourceTreeController) GetResourceTree(w http.ResponseWriter, r *http.Request) {
	org, ok := orgHandle(r)
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	componentName := r.PathValue("componentName")
	environment := r.PathValue("environment")

	tree, err := c.service.GetResourceTree(r.Context(), org, componentName, environment)
	if err != nil {
		slog.ErrorContext(r.Context(), "get resource tree failed",
			"error", err, "component", componentName, "environment", environment)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to get resource tree")
		return
	}
	utils.WriteSuccessResponse(w, http.StatusOK, tree)
}

// GetExecutionsFromTree handles GET /components/{componentName}/environments/{environment}/resource-tree/executions
func (c *resourceTreeController) GetExecutionsFromTree(w http.ResponseWriter, r *http.Request) {
	org, ok := orgHandle(r)
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	componentName := r.PathValue("componentName")
	environment := r.PathValue("environment")

	list, err := c.service.GetExecutionsFromTree(r.Context(), org, componentName, environment)
	if err != nil {
		slog.ErrorContext(r.Context(), "get executions from tree failed",
			"error", err, "component", componentName, "environment", environment)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to get executions from resource tree")
		return
	}
	utils.WriteSuccessResponse(w, http.StatusOK, list)
}

// GetResourceEvents handles GET /components/{componentName}/environments/{environment}/resource-events?version=&kind=&name=&group=
func (c *resourceTreeController) GetResourceEvents(w http.ResponseWriter, r *http.Request) {
	org, ok := orgHandle(r)
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	componentName := r.PathValue("componentName")
	environment := r.PathValue("environment")
	group := r.URL.Query().Get("group")
	version := r.URL.Query().Get("version")
	kind := r.URL.Query().Get("kind")
	name := r.URL.Query().Get("name")

	if version == "" || kind == "" || name == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "version, kind, and name query parameters are required")
		return
	}

	events, err := c.service.GetResourceEvents(r.Context(), org, componentName, environment, group, version, kind, name)
	if err != nil {
		slog.ErrorContext(r.Context(), "get resource events failed",
			"error", err, "component", componentName, "environment", environment, "kind", kind, "name", name)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to get resource events")
		return
	}
	utils.WriteSuccessResponse(w, http.StatusOK, events)
}

// GetResourceLogs handles GET /components/{componentName}/environments/{environment}/resource-logs?podName=&sinceSeconds=
func (c *resourceTreeController) GetResourceLogs(w http.ResponseWriter, r *http.Request) {
	org, ok := orgHandle(r)
	if !ok {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "missing org context")
		return
	}
	componentName := r.PathValue("componentName")
	environment := r.PathValue("environment")
	podName := r.URL.Query().Get("podName")

	if podName == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "podName query parameter is required")
		return
	}

	var sinceSeconds *int64
	if ss := r.URL.Query().Get("sinceSeconds"); ss != "" {
		v, err := strconv.ParseInt(ss, 10, 64)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, "sinceSeconds must be a valid integer")
			return
		}
		sinceSeconds = &v
	}

	logs, err := c.service.GetResourceLogs(r.Context(), org, componentName, environment, podName, sinceSeconds)
	if err != nil {
		slog.ErrorContext(r.Context(), "get resource logs failed",
			"error", err, "component", componentName, "environment", environment, "pod", podName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to get resource logs")
		return
	}
	utils.WriteSuccessResponse(w, http.StatusOK, logs)
}
