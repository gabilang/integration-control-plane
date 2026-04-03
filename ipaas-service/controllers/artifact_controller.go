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

// ArtifactController handles HTTP requests for MI runtime artifact queries.
type ArtifactController interface {
	ListArtifacts(w http.ResponseWriter, r *http.Request)
	GetArtifactTypes(w http.ResponseWriter, r *http.Request)
	GetArtifactSource(w http.ResponseWriter, r *http.Request)
	GetLocalEntryValue(w http.ResponseWriter, r *http.Request)
	GetArtifactParams(w http.ResponseWriter, r *http.Request)
	GetArtifactWsdl(w http.ResponseWriter, r *http.Request)
	UpdateArtifactStatus(w http.ResponseWriter, r *http.Request)
	UpdateListenerState(w http.ResponseWriter, r *http.Request)
	TriggerArtifact(w http.ResponseWriter, r *http.Request)
	UpdateArtifactTracingStatus(w http.ResponseWriter, r *http.Request)
	UpdateArtifactStatisticsStatus(w http.ResponseWriter, r *http.Request)
}

type artifactController struct {
	service services.ArtifactService
}

func NewArtifactController(service services.ArtifactService) ArtifactController {
	return &artifactController{service: service}
}

// ListArtifacts handles GET /artifacts?artifactType=X&environmentId=Y&componentId=Z
func (c *artifactController) ListArtifacts(w http.ResponseWriter, r *http.Request) {
	artifactType := r.URL.Query().Get("artifactType")
	environmentID := r.URL.Query().Get("environmentId")
	componentID := r.URL.Query().Get("componentId")

	if artifactType == "" || environmentID == "" || componentID == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "artifactType, environmentId and componentId are required")
		return
	}

	list, err := c.service.ListArtifacts(r.Context(), artifactType, environmentID, componentID)
	if err != nil {
		if errors.Is(err, services.ErrUnknownArtifactType) {
			utils.WriteErrorResponse(w, http.StatusBadRequest, "unknown artifact type: "+artifactType)
			return
		}
		slog.ErrorContext(r.Context(), "list artifacts failed",
			"error", err, "artifactType", artifactType,
			"environmentId", environmentID, "componentId", componentID)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to list artifacts")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, list)
}

// GetArtifactTypes handles GET /artifacts/types?componentId=...&environmentId=...
func (c *artifactController) GetArtifactTypes(w http.ResponseWriter, r *http.Request) {
	componentID := r.URL.Query().Get("componentId")
	environmentID := r.URL.Query().Get("environmentId")

	list, err := c.service.GetArtifactTypes(r.Context(), componentID, environmentID)
	if err != nil {
		slog.ErrorContext(r.Context(), "get artifact types failed", "error", err, "componentId", componentID)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to get artifact types")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, list)
}

// GetArtifactSource handles GET /artifacts/source?environmentId=...&componentId=...&artifactType=...&artifactName=...
func (c *artifactController) GetArtifactSource(w http.ResponseWriter, r *http.Request) {
	environmentID := r.URL.Query().Get("environmentId")
	componentID := r.URL.Query().Get("componentId")
	artifactType := r.URL.Query().Get("artifactType")
	artifactName := r.URL.Query().Get("artifactName")

	source, err := c.service.GetArtifactSource(r.Context(), environmentID, componentID, artifactType, artifactName)
	if err != nil {
		slog.ErrorContext(r.Context(), "get artifact source failed", "error", err)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to get artifact source")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, source)
}

// GetLocalEntryValue handles GET /artifacts/local-entry?componentId=...&entryName=...&environmentId=...
func (c *artifactController) GetLocalEntryValue(w http.ResponseWriter, r *http.Request) {
	componentID := r.URL.Query().Get("componentId")
	entryName := r.URL.Query().Get("entryName")
	environmentID := r.URL.Query().Get("environmentId")

	value, err := c.service.GetLocalEntryValue(r.Context(), componentID, entryName, environmentID)
	if err != nil {
		slog.ErrorContext(r.Context(), "get local entry value failed", "error", err)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to get local entry value")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, value)
}

// GetArtifactParams handles GET /artifacts/params?componentId=...&artifactType=...&artifactName=...&environmentId=...&runtimeId=...
func (c *artifactController) GetArtifactParams(w http.ResponseWriter, r *http.Request) {
	componentID := r.URL.Query().Get("componentId")
	artifactType := r.URL.Query().Get("artifactType")
	artifactName := r.URL.Query().Get("artifactName")
	environmentID := r.URL.Query().Get("environmentId")
	runtimeID := r.URL.Query().Get("runtimeId")

	params, err := c.service.GetArtifactParams(r.Context(), componentID, artifactType, artifactName, environmentID, runtimeID)
	if err != nil {
		slog.ErrorContext(r.Context(), "get artifact params failed", "error", err)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to get artifact params")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, params)
}

// GetArtifactWsdl handles GET /artifacts/wsdl?componentId=...&artifactType=...&artifactName=...&environmentId=...&runtimeId=...
func (c *artifactController) GetArtifactWsdl(w http.ResponseWriter, r *http.Request) {
	componentID := r.URL.Query().Get("componentId")
	artifactType := r.URL.Query().Get("artifactType")
	artifactName := r.URL.Query().Get("artifactName")
	environmentID := r.URL.Query().Get("environmentId")
	runtimeID := r.URL.Query().Get("runtimeId")

	wsdl, err := c.service.GetArtifactWsdl(r.Context(), componentID, artifactType, artifactName, environmentID, runtimeID)
	if err != nil {
		slog.ErrorContext(r.Context(), "get artifact wsdl failed", "error", err)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to get artifact wsdl")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, wsdl)
}

// UpdateArtifactStatus handles PUT /artifacts/status
func (c *artifactController) UpdateArtifactStatus(w http.ResponseWriter, r *http.Request) {
	var input models.ArtifactStatusInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := c.service.UpdateArtifactStatus(r.Context(), &input)
	if err != nil {
		slog.ErrorContext(r.Context(), "update artifact status failed", "error", err)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to update artifact status")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, result)
}

// UpdateListenerState handles PUT /artifacts/listener-state
func (c *artifactController) UpdateListenerState(w http.ResponseWriter, r *http.Request) {
	var input models.ListenerStateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := c.service.UpdateListenerState(r.Context(), &input)
	if err != nil {
		slog.ErrorContext(r.Context(), "update listener state failed", "error", err)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to update listener state")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, result)
}

// TriggerArtifact handles POST /artifacts/trigger
func (c *artifactController) TriggerArtifact(w http.ResponseWriter, r *http.Request) {
	var input models.TriggerArtifactInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := c.service.TriggerArtifact(r.Context(), &input)
	if err != nil {
		slog.ErrorContext(r.Context(), "trigger artifact failed", "error", err)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to trigger artifact")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusCreated, result)
}

// UpdateArtifactTracingStatus handles PUT /artifacts/tracing
func (c *artifactController) UpdateArtifactTracingStatus(w http.ResponseWriter, r *http.Request) {
	var input models.ArtifactTracingInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := c.service.UpdateArtifactTracingStatus(r.Context(), &input)
	if err != nil {
		slog.ErrorContext(r.Context(), "update artifact tracing failed", "error", err)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to update artifact tracing")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, result)
}

// UpdateArtifactStatisticsStatus handles PUT /artifacts/statistics
func (c *artifactController) UpdateArtifactStatisticsStatus(w http.ResponseWriter, r *http.Request) {
	var input models.ArtifactTracingInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := c.service.UpdateArtifactStatisticsStatus(r.Context(), &input)
	if err != nil {
		slog.ErrorContext(r.Context(), "update artifact statistics failed", "error", err)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to update artifact statistics")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, result)
}
