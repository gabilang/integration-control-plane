package api

import (
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/controllers"
)

func registerComponentRoutes(mux *http.ServeMux, c controllers.ComponentController) {
	mux.HandleFunc("GET /components", c.ListComponents)
	mux.HandleFunc("POST /components", c.CreateComponent)
	mux.HandleFunc("PUT /components/{componentName}", c.UpdateComponent)
	mux.HandleFunc("DELETE /components/{componentName}", c.DeleteComponent)
	mux.HandleFunc("PUT /components/{componentName}/build-parameters", c.UpdateBuildParameters)
	mux.HandleFunc("POST /components/{componentName}/builds", c.TriggerBuild)
	mux.HandleFunc("GET /components/{componentName}/builds", c.ListBuilds)
	mux.HandleFunc("GET /components/{componentName}/builds/{buildName}", c.GetBuildStatus)
	mux.HandleFunc("GET /components/{componentName}/builds/{buildName}/logs", c.GetBuildLogs)
	mux.HandleFunc("GET /components/{componentName}/repository", c.GetComponentRepository)
	mux.HandleFunc("GET /components/{componentName}/commit-history", c.GetCommitHistory)
	mux.HandleFunc("GET /components/{componentName}/labels", c.GetComponentLabels)
	mux.HandleFunc("GET /components/{componentName}/deployment-track", c.GetDeploymentTrack)
}
