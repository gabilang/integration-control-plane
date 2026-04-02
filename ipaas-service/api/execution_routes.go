package api

import (
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/controllers"
)

func registerExecutionRoutes(mux *http.ServeMux, c controllers.ExecutionController) {
	mux.HandleFunc("GET /components/{componentName}/schedules/{environment}/executions", c.ListExecutions)
}
