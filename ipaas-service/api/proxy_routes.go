package api

import (
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/clients/icp"
)

// registerGraphQLRoute registers a transparent proxy for ICP GraphQL requests.
// POST /graphql → forwards to ICP_GRAPHQL_URL
func registerGraphQLRoute(mux *http.ServeMux, client *icp.ProxyClient) {
	if !client.IsAvailable() {
		mux.HandleFunc("POST /graphql", icp.ServiceUnavailableHandler("ICP GraphQL"))
		return
	}
	mux.HandleFunc("POST /graphql", func(w http.ResponseWriter, r *http.Request) {
		client.ProxyHTTP(w, r, "/graphql")
	})
}

// registerAuthProxyRoutes registers transparent proxy routes for the ICP auth service.
// All methods on /auth/{path...} → forwarded to ICP_AUTH_BASE_URL/{path...}
func registerAuthProxyRoutes(mux *http.ServeMux, client *icp.ProxyClient) {
	unavailable := icp.ServiceUnavailableHandler("ICP Auth")

	handleAuth := func(w http.ResponseWriter, r *http.Request) {
		if !client.IsAvailable() {
			unavailable(w, r)
			return
		}
		// Strip the /auth prefix and proxy the rest
		path := r.URL.Path[len("/auth"):]
		if path == "" {
			path = "/"
		}
		client.ProxyHTTP(w, r, path)
	}

	mux.HandleFunc("/auth/", handleAuth)
}

// registerObservabilityProxyRoutes registers transparent proxy routes for the
// ICP observability service.
// All methods on /observability/{path...} → forwarded to OBSERVABILITY_SERVICE_BASE_URL/{path...}
func registerObservabilityProxyRoutes(mux *http.ServeMux, client *icp.ProxyClient) {
	unavailable := icp.ServiceUnavailableHandler("Observability")

	handleObs := func(w http.ResponseWriter, r *http.Request) {
		if !client.IsAvailable() {
			unavailable(w, r)
			return
		}
		path := r.URL.Path[len("/observability"):]
		if path == "" {
			path = "/"
		}
		client.ProxyHTTP(w, r, path)
	}

	mux.HandleFunc("/observability/", handleObs)
}
