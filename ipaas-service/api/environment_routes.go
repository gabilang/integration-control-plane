package api

import (
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/controllers"
)

func registerEnvironmentRoutes(mux *http.ServeMux, c controllers.EnvironmentController) {
	mux.HandleFunc("GET /environments", c.ListEnvironments)
	mux.HandleFunc("GET /environments/{environmentName}", c.GetEnvironment)
}
