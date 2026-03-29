package openchoreo

import (
	"context"
	"fmt"
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/clients/requests"
	"github.com/wso2/integration-control-plane/ipaas-service/middleware"
	"github.com/wso2/integration-control-plane/ipaas-service/models"
)

// ProjectClient defines operations for managing OpenChoreo projects.
type ProjectClient interface {
	ListProjects(ctx context.Context, orgName string, limit int, cursor string) (*models.ProjectList, error)
	GetProject(ctx context.Context, orgName, projectName string) (*models.Project, error)
	CreateProject(ctx context.Context, orgName string, req *models.CreateProjectRequest) (*models.Project, error)
	UpdateProject(ctx context.Context, orgName, projectName string, req *models.UpdateProjectRequest) (*models.Project, error)
	DeleteProject(ctx context.Context, orgName, projectName string) error
}

type projectClient struct {
	baseURL    string
	hostHeader string
	httpClient *http.Client
}

// NewProjectClient creates a new OpenChoreo project client.
// baseURL is the platform-api-service gateway URL.
// hostHeader is the Host header required by the data-plane gateway for routing.
// The Bearer token is read from the request context on each call.
func NewProjectClient(baseURL, hostHeader string) ProjectClient {
	return &projectClient{
		baseURL:    baseURL,
		hostHeader: hostHeader,
		httpClient: &http.Client{},
	}
}

func (c *projectClient) projectsURL() string {
	return c.baseURL + "/projects"
}

func (c *projectClient) projectURL(projectName string) string {
	return fmt.Sprintf("%s/projects/%s", c.baseURL, projectName)
}

func (c *projectClient) newRequest(ctx context.Context, name, method, url string) *requests.HttpRequest {
	req := requests.NewRequest(name, method, url)
	if token := middleware.GetAuthToken(ctx); token != "" {
		req.SetHeader("Authorization", "Bearer "+token)
	}
	if c.hostHeader != "" {
		req.SetHost(c.hostHeader)
	}
	return req
}

// normalizeProject converts a K8s-style OpenChoreo project into the flat model
// returned to callers.
func normalizeProject(p ocProject) models.Project {
	ann := p.Metadata.Annotations
	var displayName, description string
	if ann != nil {
		displayName = ann["openchoreo.dev/display-name"]
		description = ann["openchoreo.dev/description"]
	}

	var deploymentPipeline string
	if p.Spec.DeploymentPipelineRef != nil {
		deploymentPipeline = p.Spec.DeploymentPipelineRef.Name
	}

	return models.Project{
		UID:                p.Metadata.UID,
		Name:               p.Metadata.Name,
		NamespaceName:      p.Metadata.Namespace,
		DisplayName:        displayName,
		Description:        description,
		DeploymentPipeline: deploymentPipeline,
		CreatedAt:          p.Metadata.CreationTimestamp,
		Status:             latestConditionReason(p.Status.Conditions),
	}
}

// buildCreateProjectBody converts a flat CreateProjectRequest into the K8s-style
// body expected by the OpenChoreo API.
func buildCreateProjectBody(req *models.CreateProjectRequest) ocProject {
	body := ocProject{
		Metadata: ocObjectMeta{
			Name: req.Name,
		},
	}
	if req.DisplayName != "" || req.Description != "" {
		body.Metadata.Annotations = map[string]string{}
		if req.DisplayName != "" {
			body.Metadata.Annotations["openchoreo.dev/display-name"] = req.DisplayName
		}
		if req.Description != "" {
			body.Metadata.Annotations["openchoreo.dev/description"] = req.Description
		}
	}
	if req.DeploymentPipeline != "" {
		body.Spec.DeploymentPipelineRef = &ocRef{Name: req.DeploymentPipeline}
	}
	return body
}

// buildUpdateProjectBody converts a flat UpdateProjectRequest into the K8s-style
// body expected by the OpenChoreo API.
func buildUpdateProjectBody(name string, req *models.UpdateProjectRequest) ocProject {
	body := ocProject{
		Metadata: ocObjectMeta{
			Name: name,
		},
	}
	if req.DisplayName != "" || req.Description != "" {
		body.Metadata.Annotations = map[string]string{}
		if req.DisplayName != "" {
			body.Metadata.Annotations["openchoreo.dev/display-name"] = req.DisplayName
		}
		if req.Description != "" {
			body.Metadata.Annotations["openchoreo.dev/description"] = req.Description
		}
	}
	if req.DeploymentPipeline != "" {
		body.Spec.DeploymentPipelineRef = &ocRef{Name: req.DeploymentPipeline}
	}
	return body
}

func (c *projectClient) ListProjects(ctx context.Context, _ string, limit int, cursor string) (*models.ProjectList, error) {
	req := c.newRequest(ctx, "openchoreo.ListProjects", http.MethodGet, c.projectsURL())
	if limit > 0 {
		req.SetQuery("limit", fmt.Sprintf("%d", limit))
	}
	if cursor != "" {
		req.SetQuery("cursor", cursor)
	}

	result := requests.SendRequest(ctx, c.httpClient, req)
	var raw ocProjectList
	if err := result.ScanResponse(&raw, http.StatusOK); err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}

	items := make([]models.Project, len(raw.Items))
	for i, p := range raw.Items {
		items[i] = normalizeProject(p)
	}
	return &models.ProjectList{Items: items}, nil
}

func (c *projectClient) GetProject(ctx context.Context, _, projectName string) (*models.Project, error) {
	req := c.newRequest(ctx, "openchoreo.GetProject", http.MethodGet, c.projectURL(projectName))

	result := requests.SendRequest(ctx, c.httpClient, req)
	var raw ocProject
	if err := result.ScanResponse(&raw, http.StatusOK); err != nil {
		return nil, fmt.Errorf("get project: %w", err)
	}
	p := normalizeProject(raw)
	return &p, nil
}

func (c *projectClient) CreateProject(ctx context.Context, _ string, body *models.CreateProjectRequest) (*models.Project, error) {
	req := c.newRequest(ctx, "openchoreo.CreateProject", http.MethodPost, c.projectsURL())
	req.SetJSON(buildCreateProjectBody(body))

	result := requests.SendRequest(ctx, c.httpClient, req)
	var raw ocProject
	if err := result.ScanResponse(&raw, http.StatusCreated); err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	p := normalizeProject(raw)
	return &p, nil
}

func (c *projectClient) UpdateProject(ctx context.Context, _, projectName string, body *models.UpdateProjectRequest) (*models.Project, error) {
	req := c.newRequest(ctx, "openchoreo.UpdateProject", http.MethodPut, c.projectURL(projectName))
	req.SetJSON(buildUpdateProjectBody(projectName, body))

	result := requests.SendRequest(ctx, c.httpClient, req)
	var raw ocProject
	if err := result.ScanResponse(&raw, http.StatusOK); err != nil {
		return nil, fmt.Errorf("update project: %w", err)
	}
	p := normalizeProject(raw)
	return &p, nil
}

func (c *projectClient) DeleteProject(ctx context.Context, _, projectName string) error {
	req := c.newRequest(ctx, "openchoreo.DeleteProject", http.MethodDelete, c.projectURL(projectName))

	result := requests.SendRequest(ctx, c.httpClient, req)
	if err := result.ScanResponse(nil, http.StatusNoContent); err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	return nil
}
