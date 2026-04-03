package controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/models"
	"github.com/wso2/integration-control-plane/ipaas-service/services"
	"github.com/wso2/integration-control-plane/ipaas-service/utils"
)

type DeploymentController interface {
	GetComponentDeployment(w http.ResponseWriter, r *http.Request)
	GetDeploymentStatus(w http.ResponseWriter, r *http.Request)
	DeployDeploymentTrack(w http.ResponseWriter, r *http.Request)
	Promote(w http.ResponseWriter, r *http.Request)
	StopDeployment(w http.ResponseWriter, r *http.Request)
}

type deploymentController struct {
	service services.DeploymentService
}

func NewDeploymentController(service services.DeploymentService) DeploymentController {
	return &deploymentController{service: service}
}

// GetComponentDeployment handles GET /components/{componentName}/deployments?orgHandler=...&orgUuid=...&versionId=...&environmentId=...
func (c *deploymentController) GetComponentDeployment(w http.ResponseWriter, r *http.Request) {
	componentName := r.PathValue("componentName")
	orgHandler := r.URL.Query().Get("orgHandler")
	orgUUID := r.URL.Query().Get("orgUuid")
	versionID := r.URL.Query().Get("versionId")
	environmentID := r.URL.Query().Get("environmentId")

	deployment, err := c.service.GetComponentDeployment(r.Context(), orgHandler, orgUUID, componentName, versionID, environmentID)
	if err != nil {
		slog.ErrorContext(r.Context(), "get component deployment failed", "error", err, "component", componentName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to get component deployment")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, deployment)
}

// GetDeploymentStatus handles GET /components/{componentName}/deployments/status?versionId=...
func (c *deploymentController) GetDeploymentStatus(w http.ResponseWriter, r *http.Request) {
	componentName := r.PathValue("componentName")
	versionID := r.URL.Query().Get("versionId")

	statuses, err := c.service.GetDeploymentStatus(r.Context(), componentName, versionID)
	if err != nil {
		slog.ErrorContext(r.Context(), "get deployment status failed", "error", err, "component", componentName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to get deployment status")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, statuses)
}

// DeployDeploymentTrack handles POST /components/{componentName}/deployments
func (c *deploymentController) DeployDeploymentTrack(w http.ResponseWriter, r *http.Request) {
	var input models.DeployDeploymentTrackInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := c.service.DeployDeploymentTrack(r.Context(), &input)
	if err != nil {
		slog.ErrorContext(r.Context(), "deploy deployment track failed", "error", err)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to deploy deployment track")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, result)
}

// Promote handles POST /components/{componentName}/deployments/promote
func (c *deploymentController) Promote(w http.ResponseWriter, r *http.Request) {
	componentName := r.PathValue("componentName")

	var input models.PromoteInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := c.service.Promote(r.Context(), componentName, &input)
	if err != nil {
		slog.ErrorContext(r.Context(), "promote failed", "error", err, "component", componentName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to promote")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, result)
}

// StopDeployment handles DELETE /components/{componentName}/deployments
func (c *deploymentController) StopDeployment(w http.ResponseWriter, r *http.Request) {
	var input models.StopDeploymentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := c.service.StopDeployment(r.Context(), &input)
	if err != nil {
		slog.ErrorContext(r.Context(), "stop deployment failed", "error", err)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to stop deployment")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, result)
}
