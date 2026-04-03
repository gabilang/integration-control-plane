package api

import (
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/controllers"
)

func registerRuntimeRoutes(mux *http.ServeMux, c controllers.RuntimeController) {
	mux.HandleFunc("GET /components/{componentName}/runtimes", c.ListRuntimes)
	mux.HandleFunc("GET /projects/{projectName}/runtimes", c.ListProjectRuntimes)
	mux.HandleFunc("DELETE /runtimes/{runtimeId}", c.DeleteRuntime)
}
