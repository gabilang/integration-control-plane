package api

import (
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/controllers"
)

func registerDeploymentRoutes(mux *http.ServeMux, c controllers.DeploymentController) {
	// More specific patterns must come before less specific ones.
	mux.HandleFunc("GET /components/{componentName}/deployments/status", c.GetDeploymentStatus)
	mux.HandleFunc("GET /components/{componentName}/deployments", c.GetComponentDeployment)
	mux.HandleFunc("POST /components/{componentName}/deployments/promote", c.Promote)
	mux.HandleFunc("POST /components/{componentName}/deployments", c.DeployDeploymentTrack)
	mux.HandleFunc("DELETE /components/{componentName}/deployments", c.StopDeployment)
}
