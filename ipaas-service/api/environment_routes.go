package api

import (
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/controllers"
)

func registerEnvironmentRoutes(mux *http.ServeMux, c controllers.EnvironmentController) {
	mux.HandleFunc("GET /environments", c.ListEnvironments)
	mux.HandleFunc("POST /environments", c.CreateEnvironment)
	mux.HandleFunc("GET /environments/{environmentName}", c.GetEnvironment)
	mux.HandleFunc("PUT /environments/{environmentName}", c.UpdateEnvironment)
	mux.HandleFunc("DELETE /environments/{environmentName}", c.DeleteEnvironment)
}
