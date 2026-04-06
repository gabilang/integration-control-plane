package api

import (
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/controllers"
)

func registerResourceTreeRoutes(mux *http.ServeMux, c controllers.ResourceTreeController) {
	mux.HandleFunc("GET /components/{componentName}/environments/{environment}/resource-tree/executions", c.GetExecutionsFromTree)
	mux.HandleFunc("GET /components/{componentName}/environments/{environment}/resource-tree", c.GetResourceTree)
	mux.HandleFunc("GET /components/{componentName}/environments/{environment}/resource-events", c.GetResourceEvents)
	mux.HandleFunc("GET /components/{componentName}/environments/{environment}/resource-logs", c.GetResourceLogs)
}
