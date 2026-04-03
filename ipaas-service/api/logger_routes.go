package api

import (
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/controllers"
)

func registerLoggerRoutes(mux *http.ServeMux, c controllers.LoggerController) {
	mux.HandleFunc("GET /components/{componentName}/loggers", c.ListLoggers)
	mux.HandleFunc("PUT /components/{componentName}/loggers", c.UpdateLogLevel)
}
