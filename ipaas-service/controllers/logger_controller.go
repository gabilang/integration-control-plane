package controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/models"
	"github.com/wso2/integration-control-plane/ipaas-service/services"
	"github.com/wso2/integration-control-plane/ipaas-service/utils"
)

type LoggerController interface {
	ListLoggers(w http.ResponseWriter, r *http.Request)
	UpdateLogLevel(w http.ResponseWriter, r *http.Request)
}

type loggerController struct {
	service services.LoggerService
}

func NewLoggerController(service services.LoggerService) LoggerController {
	return &loggerController{service: service}
}

// ListLoggers handles GET /components/{componentName}/loggers?environmentId=...
func (c *loggerController) ListLoggers(w http.ResponseWriter, r *http.Request) {
	componentName := r.PathValue("componentName")
	environmentID := r.URL.Query().Get("environmentId")

	if environmentID == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "environmentId is required")
		return
	}

	list, err := c.service.ListLoggers(r.Context(), environmentID, componentName)
	if err != nil {
		slog.ErrorContext(r.Context(), "list loggers failed", "error", err, "component", componentName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to list loggers")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, list)
}

// UpdateLogLevel handles PUT /components/{componentName}/loggers
func (c *loggerController) UpdateLogLevel(w http.ResponseWriter, r *http.Request) {
	var input models.UpdateLogLevelInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := c.service.UpdateLogLevel(r.Context(), &input)
	if err != nil {
		slog.ErrorContext(r.Context(), "update log level failed", "error", err)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to update log level")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, result)
}
