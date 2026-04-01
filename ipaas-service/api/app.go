package api

import (
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/clients/icp"
	"github.com/wso2/integration-control-plane/ipaas-service/controllers"
	"github.com/wso2/integration-control-plane/ipaas-service/middleware"
	jwtmw "github.com/wso2/integration-control-plane/ipaas-service/middleware/jwt"
	"github.com/wso2/integration-control-plane/ipaas-service/middleware/logger"
)

// AppParams holds all dependencies needed to build the HTTP handler.
type AppParams struct {
	ProjectController     controllers.ProjectController
	ComponentController   controllers.ComponentController
	EnvironmentController controllers.EnvironmentController
	ArtifactController    controllers.ArtifactController
	ScheduleController    controllers.ScheduleController
	GraphQLProxy          *icp.ProxyClient
	AuthProxy             *icp.ProxyClient
	ObservabilityProxy    *icp.ProxyClient
}

// NewHandler assembles the full HTTP handler with middleware and routes.
// The internal API gateway strips the context path (/integration-platform-api/v1.0)
// before forwarding, so routes are registered at root level.
func NewHandler(params AppParams) http.Handler {
	mux := http.NewServeMux()

	// Health check — unauthenticated
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`)) //nolint:errcheck
	})

	// API routes — protected by JWT middleware
	apiMux := http.NewServeMux()
	registerProjectRoutes(apiMux, params.ProjectController)
	registerComponentRoutes(apiMux, params.ComponentController)
	registerScheduleRoutes(apiMux, params.ScheduleController)
	registerEnvironmentRoutes(apiMux, params.EnvironmentController)
	registerArtifactRoutes(apiMux, params.ArtifactController)
	registerGraphQLRoute(apiMux, params.GraphQLProxy)
	registerAuthProxyRoutes(apiMux, params.AuthProxy)
	registerObservabilityProxyRoutes(apiMux, params.ObservabilityProxy)

	mux.Handle("/", jwtmw.Middleware(apiMux))

	// Global middleware stack (outermost applied last)
	var handler http.Handler = mux
	handler = middleware.ExtractAuthToken()(handler)
	handler = logger.RequestLogger()(handler)
	handler = middleware.AddCorrelationID()(handler)
	handler = middleware.RecovererOnPanic()(handler)

	return handler
}
