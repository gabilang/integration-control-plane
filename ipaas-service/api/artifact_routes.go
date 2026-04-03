package api

import (
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/controllers"
)

func registerArtifactRoutes(mux *http.ServeMux, c controllers.ArtifactController) {
	mux.HandleFunc("GET /artifacts", c.ListArtifacts)
	mux.HandleFunc("GET /artifacts/types", c.GetArtifactTypes)
	mux.HandleFunc("GET /artifacts/source", c.GetArtifactSource)
	mux.HandleFunc("GET /artifacts/local-entry", c.GetLocalEntryValue)
	mux.HandleFunc("GET /artifacts/params", c.GetArtifactParams)
	mux.HandleFunc("GET /artifacts/wsdl", c.GetArtifactWsdl)
	mux.HandleFunc("PUT /artifacts/status", c.UpdateArtifactStatus)
	mux.HandleFunc("PUT /artifacts/listener-state", c.UpdateListenerState)
	mux.HandleFunc("POST /artifacts/trigger", c.TriggerArtifact)
	mux.HandleFunc("PUT /artifacts/tracing", c.UpdateArtifactTracingStatus)
	mux.HandleFunc("PUT /artifacts/statistics", c.UpdateArtifactStatisticsStatus)
}
