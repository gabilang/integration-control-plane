package icp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/wso2/integration-control-plane/ipaas-service/middleware"
)

// Client executes queries against the ICP service.
type Client struct {
	url        string
	httpClient *http.Client
}

// NewClient creates a Client targeting the ICP service. baseURL is the ICP
// service base URL (without /graphql); the path is appended automatically —
// consistent with how the ProxyClient routes POST /graphql requests.
func NewClient(baseURL string) *Client {
	return &Client{
		url:        baseURL + "/graphql",
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// IsAvailable reports whether a baseURL was configured.
func (c *Client) IsAvailable() bool {
	return c.url != "/graphql"
}

type queryRequest struct {
	Query     string            `json:"query"`
	Variables map[string]string `json:"variables,omitempty"`
}

type queryResponse struct {
	Data   map[string]json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

// Query executes a query against the ICP service and returns the raw data
// fields keyed by field name. The Bearer token is read from the request context.
func (c *Client) Query(ctx context.Context, query string, variables map[string]string) (map[string]json.RawMessage, error) {
	body, err := json.Marshal(queryRequest{Query: query, Variables: variables})
	if err != nil {
		return nil, fmt.Errorf("icp: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("icp: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token := middleware.GetAuthToken(ctx); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("icp: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("icp: unexpected status %d", resp.StatusCode)
	}

	var result queryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("icp: decode response: %w", err)
	}
	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("icp: %s", result.Errors[0].Message)
	}

	return result.Data, nil
}
