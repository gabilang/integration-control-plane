package api

import (
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/controllers"
)

func registerComponentRoutes(mux *http.ServeMux, c controllers.ComponentController) {
	mux.HandleFunc("GET /components", c.ListComponents)
	mux.HandleFunc("POST /components", c.CreateComponent)
	mux.HandleFunc("PUT /components/{componentName}/build-parameters", c.UpdateBuildParameters)
	mux.HandleFunc("POST /components/{componentName}/builds", c.TriggerBuild)
	mux.HandleFunc("GET /components/{componentName}/builds", c.ListBuilds)
	mux.HandleFunc("GET /components/{componentName}/builds/{buildName}", c.GetBuildStatus)
	mux.HandleFunc("GET /components/{componentName}/builds/{buildName}/logs", c.GetBuildLogs)
}
