package api

import (
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/controllers"
)

func registerProjectRoutes(mux *http.ServeMux, c controllers.ProjectController) {
	mux.HandleFunc("GET /projects", c.ListProjects)
	mux.HandleFunc("POST /projects", c.CreateProject)
	mux.HandleFunc("GET /projects/{projectName}", c.GetProject)
	mux.HandleFunc("PUT /projects/{projectName}", c.UpdateProject)
	mux.HandleFunc("DELETE /projects/{projectName}", c.DeleteProject)
	mux.HandleFunc("GET /projects/{projectName}/contributors", c.GetProjectContributors)
}
