package controllers

import (
	"log/slog"
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/models"
	"github.com/wso2/integration-control-plane/ipaas-service/services"
	"github.com/wso2/integration-control-plane/ipaas-service/utils"
)

type SecretController interface {
	GenerateJwtSecret(w http.ResponseWriter, r *http.Request)
	RotateJwtSecret(w http.ResponseWriter, r *http.Request)
}

type secretController struct {
	service services.SecretService
}

func NewSecretController(service services.SecretService) SecretController {
	return &secretController{service: service}
}

// GenerateJwtSecret handles POST /components/{componentName}/environments/{environmentName}/jwt-secret
func (c *secretController) GenerateJwtSecret(w http.ResponseWriter, r *http.Request) {
	componentName := r.PathValue("componentName")
	environmentName := r.PathValue("environmentName")

	secret, err := c.service.GenerateJwtSecret(r.Context(), componentName, environmentName)
	if err != nil {
		slog.ErrorContext(r.Context(), "generate jwt secret failed", "error", err, "component", componentName, "environment", environmentName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to generate jwt secret")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusCreated, models.JwtSecret{Secret: secret})
}

// RotateJwtSecret handles PUT /components/{componentName}/environments/{environmentName}/jwt-secret/rotate
func (c *secretController) RotateJwtSecret(w http.ResponseWriter, r *http.Request) {
	componentName := r.PathValue("componentName")
	environmentName := r.PathValue("environmentName")

	secret, err := c.service.RotateJwtSecret(r.Context(), componentName, environmentName)
	if err != nil {
		slog.ErrorContext(r.Context(), "rotate jwt secret failed", "error", err, "component", componentName, "environment", environmentName)
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to rotate jwt secret")
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, models.JwtSecret{Secret: secret})
}
