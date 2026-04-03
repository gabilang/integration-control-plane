package openchoreo

import (
	"context"
	"fmt"
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/clients/requests"
	"github.com/wso2/integration-control-plane/ipaas-service/middleware"
	"github.com/wso2/integration-control-plane/ipaas-service/models"
)

// EnvironmentClient defines operations for reading OpenChoreo environments.
type EnvironmentClient interface {
	ListEnvironments(ctx context.Context, orgName string) (*models.EnvironmentList, error)
	GetEnvironment(ctx context.Context, orgName, environmentName string) (*models.Environment, error)
	CreateEnvironment(ctx context.Context, orgName string, req *models.CreateEnvironmentRequest) (*models.Environment, error)
	UpdateEnvironment(ctx context.Context, orgName, environmentName string, req *models.UpdateEnvironmentRequest) (*models.Environment, error)
	DeleteEnvironment(ctx context.Context, orgName, environmentName string) error
}

type environmentClient struct {
	baseURL    string
	hostHeader string
	httpClient *http.Client
}

// NewEnvironmentClient creates a new OpenChoreo environment client.
func NewEnvironmentClient(baseURL, hostHeader string) EnvironmentClient {
	return &environmentClient{
		baseURL:    baseURL,
		hostHeader: hostHeader,
		httpClient: &http.Client{},
	}
}

func (c *environmentClient) environmentsURL() string {
	return c.baseURL + "/environments"
}

func (c *environmentClient) environmentURL(environmentName string) string {
	return fmt.Sprintf("%s/environments/%s", c.baseURL, environmentName)
}

func (c *environmentClient) newRequest(ctx context.Context, name, method, url string) *requests.HttpRequest {
	req := requests.NewRequest(name, method, url)
	if token := middleware.GetAuthToken(ctx); token != "" {
		req.SetHeader("Authorization", "Bearer "+token)
	}
	if c.hostHeader != "" {
		req.SetHost(c.hostHeader)
	}
	return req
}

func normalizeEnvironment(e ocEnvironment) models.Environment {
	ann := e.Metadata.Annotations
	var displayName string
	if ann != nil {
		displayName = ann["openchoreo.dev/display-name"]
	}

	var dataPlaneRef string
	var isProduction bool
	if e.Spec.DataPlaneRef != nil {
		dataPlaneRef = e.Spec.DataPlaneRef.Name
	}
	if e.Spec.IsProduction != nil {
		isProduction = *e.Spec.IsProduction
	}

	return models.Environment{
		UID:          e.Metadata.UID,
		Name:         e.Metadata.Name,
		DisplayName:  displayName,
		DataPlaneRef: dataPlaneRef,
		IsProduction: isProduction,
		CreatedAt:    e.Metadata.CreationTimestamp,
	}
}

func (c *environmentClient) ListEnvironments(ctx context.Context, _ string) (*models.EnvironmentList, error) {
	req := c.newRequest(ctx, "openchoreo.ListEnvironments", http.MethodGet, c.environmentsURL())

	result := requests.SendRequest(ctx, c.httpClient, req)
	var raw ocEnvironmentList
	if err := result.ScanResponse(&raw, http.StatusOK); err != nil {
		return nil, fmt.Errorf("list environments: %w", err)
	}

	items := make([]models.Environment, len(raw.Items))
	for i, e := range raw.Items {
		items[i] = normalizeEnvironment(e)
	}
	return &models.EnvironmentList{Items: items}, nil
}

func (c *environmentClient) GetEnvironment(ctx context.Context, _, environmentName string) (*models.Environment, error) {
	req := c.newRequest(ctx, "openchoreo.GetEnvironment", http.MethodGet, c.environmentURL(environmentName))

	result := requests.SendRequest(ctx, c.httpClient, req)
	var raw ocEnvironment
	if err := result.ScanResponse(&raw, http.StatusOK); err != nil {
		return nil, fmt.Errorf("get environment: %w", err)
	}
	e := normalizeEnvironment(raw)
	return &e, nil
}

func (c *environmentClient) CreateEnvironment(ctx context.Context, _ string, reqBody *models.CreateEnvironmentRequest) (*models.Environment, error) {
	body := ocEnvironment{
		Metadata: ocObjectMeta{
			Name: reqBody.Name,
			Annotations: map[string]string{
				"openchoreo.dev/display-name": reqBody.DisplayName,
			},
		},
		Spec: ocEnvironmentSpec{
			IsProduction: &reqBody.IsProduction,
		},
	}

	req := c.newRequest(ctx, "openchoreo.CreateEnvironment", http.MethodPost, c.environmentsURL())
	req.SetJSON(body)

	result := requests.SendRequest(ctx, c.httpClient, req)
	var raw ocEnvironment
	if err := result.ScanResponse(&raw, http.StatusCreated); err != nil {
		return nil, fmt.Errorf("create environment: %w", err)
	}
	e := normalizeEnvironment(raw)
	return &e, nil
}

func (c *environmentClient) UpdateEnvironment(ctx context.Context, _, environmentName string, reqBody *models.UpdateEnvironmentRequest) (*models.Environment, error) {
	getReq := c.newRequest(ctx, "openchoreo.GetEnvironmentForUpdate", http.MethodGet, c.environmentURL(environmentName))
	getResult := requests.SendRequest(ctx, c.httpClient, getReq)
	var existing ocEnvironment
	if err := getResult.ScanResponse(&existing, http.StatusOK); err != nil {
		return nil, fmt.Errorf("get environment for update: %w", err)
	}

	if existing.Metadata.Annotations == nil {
		existing.Metadata.Annotations = make(map[string]string)
	}
	if reqBody.DisplayName != "" {
		existing.Metadata.Annotations["openchoreo.dev/display-name"] = reqBody.DisplayName
	}
	if reqBody.IsProduction != nil {
		existing.Spec.IsProduction = reqBody.IsProduction
	}

	putReq := c.newRequest(ctx, "openchoreo.UpdateEnvironment", http.MethodPut, c.environmentURL(environmentName))
	putReq.SetJSON(existing)

	result := requests.SendRequest(ctx, c.httpClient, putReq)
	var raw ocEnvironment
	if err := result.ScanResponse(&raw, http.StatusOK); err != nil {
		return nil, fmt.Errorf("update environment: %w", err)
	}
	e := normalizeEnvironment(raw)
	return &e, nil
}

func (c *environmentClient) DeleteEnvironment(ctx context.Context, _, environmentName string) error {
	req := c.newRequest(ctx, "openchoreo.DeleteEnvironment", http.MethodDelete, c.environmentURL(environmentName))
	result := requests.SendRequest(ctx, c.httpClient, req)
	if err := result.ScanResponse(nil, http.StatusNoContent); err != nil {
		return fmt.Errorf("delete environment: %w", err)
	}
	return nil
}
