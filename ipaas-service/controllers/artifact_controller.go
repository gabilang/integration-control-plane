package controllers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/services"
	"github.com/wso2/integration-control-plane/ipaas-service/utils"
)

// ArtifactController handles HTTP requests for MI runtime artifact queries.
type ArtifactController interface {
	ListArtifacts(w http.ResponseWriter, r *http.Request)
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
