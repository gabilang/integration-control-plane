package api

import (
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/controllers"
)

func registerArtifactRoutes(mux *http.ServeMux, c controllers.ArtifactController) {
	mux.HandleFunc("GET /artifacts", c.ListArtifacts)
}
