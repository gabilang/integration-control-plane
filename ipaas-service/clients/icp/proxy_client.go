package icp

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/wso2/integration-control-plane/ipaas-service/middleware"
)

// ProxyClient forwards authenticated requests to an upstream service, injecting
// the caller's Bearer token from the request context.
type ProxyClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewProxyClient creates a ProxyClient targeting baseURL.
// baseURL must not end with a slash.
func NewProxyClient(baseURL string) *ProxyClient {
	return &ProxyClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{},
	}
}

// IsAvailable reports whether a baseURL was configured.
func (c *ProxyClient) IsAvailable() bool {
	return c.baseURL != ""
}

// ProxyHTTP forwards the incoming request to the upstream at the same path,
// copying headers, body, and the status code back to w.
// The caller's Bearer token is forwarded as-is.
func (c *ProxyClient) ProxyHTTP(w http.ResponseWriter, r *http.Request, upstreamPath string) {
	upstreamURL := c.baseURL + upstreamPath

	// Build the outbound request
	outReq, err := http.NewRequestWithContext(r.Context(), r.Method, upstreamURL, r.Body)
	if err != nil {
		slog.ErrorContext(r.Context(), "proxy: failed to create upstream request", "url", upstreamURL, "error", err)
		http.Error(w, `{"error":"Internal Server Error","message":"failed to create upstream request"}`, http.StatusInternalServerError)
		return
	}

	// Copy original request headers
	for key, values := range r.Header {
		// Skip hop-by-hop headers
		switch strings.ToLower(key) {
		case "connection", "keep-alive", "proxy-authenticate", "proxy-authorization",
			"te", "trailers", "transfer-encoding", "upgrade":
			continue
		}
		for _, v := range values {
			outReq.Header.Add(key, v)
		}
	}

	// Inject Bearer token from context (gateway already validated it)
	if token := middleware.GetAuthToken(r.Context()); token != "" {
		outReq.Header.Set("Authorization", "Bearer "+token)
	}

	// Forward query string
	if r.URL.RawQuery != "" {
		outReq.URL.RawQuery = r.URL.RawQuery
	}

	resp, err := c.httpClient.Do(outReq)
	if err != nil {
		slog.ErrorContext(r.Context(), "proxy: upstream request failed", "url", upstreamURL, "error", err)
		http.Error(w, `{"error":"Bad Gateway","message":"upstream request failed"}`, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, v := range values {
			w.Header().Add(key, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		slog.ErrorContext(r.Context(), "proxy: failed to copy response body", "error", err)
	}
}

// ServiceUnavailableHandler returns a handler that responds 503 when the
// upstream service is not configured.
func ServiceUnavailableHandler(service string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w,
			fmt.Sprintf(`{"error":"Service Unavailable","message":"%s not configured"}`, service),
			http.StatusServiceUnavailable,
		)
	}
}
