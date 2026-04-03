package api

import (
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/controllers"
)

func registerExecutionRoutes(mux *http.ServeMux, c controllers.ExecutionController) {
	mux.HandleFunc("GET /components/{componentName}/schedules/{environment}/executions", c.ListExecutions)
	mux.HandleFunc("POST /components/{componentName}/schedules/{environment}/executions", c.TriggerExecution)
	mux.HandleFunc("GET /components/{componentName}/execution-configs", c.GetExecutionConfigs)
	mux.HandleFunc("GET /components/{componentName}/executions/{runId}/arguments", c.GetExecutionArguments)
	mux.HandleFunc("PUT /components/{componentName}/job-configs", c.UpdateJobConfigs)
}
