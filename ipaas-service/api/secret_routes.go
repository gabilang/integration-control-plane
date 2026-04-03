package api

import (
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/controllers"
)

func registerSecretRoutes(mux *http.ServeMux, c controllers.SecretController) {
	mux.HandleFunc("POST /components/{componentName}/environments/{environmentName}/jwt-secret", c.GenerateJwtSecret)
	mux.HandleFunc("PUT /components/{componentName}/environments/{environmentName}/jwt-secret/rotate", c.RotateJwtSecret)
}
