package openchoreo

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/wso2/integration-control-plane/ipaas-service/clients/requests"
	"github.com/wso2/integration-control-plane/ipaas-service/middleware"
	"github.com/wso2/integration-control-plane/ipaas-service/models"
)

// ComponentClient defines operations for managing OpenChoreo components.
type ComponentClient interface {
	ListComponents(ctx context.Context, orgName, projectName string, limit int, cursor string) (*models.ComponentList, error)
	GetComponent(ctx context.Context, componentName string) (*models.Component, error)
	CreateComponent(ctx context.Context, orgName, projectName string, req *models.CreateComponentRequest) (*models.Component, error)
	UpdateBuildParameters(ctx context.Context, orgName, projectName, componentName string, req *models.UpdateBuildParametersRequest) (*models.Component, error)
	TriggerBuild(ctx context.Context, orgName, projectName, componentName string) (*models.WorkflowRun, error)
	ListWorkflowRuns(ctx context.Context, orgName, projectName, componentName string, limit int, cursor string) (*models.WorkflowRunList, error)
	GetWorkflowRun(ctx context.Context, orgName, projectName, componentName, runName string) (*models.WorkflowRun, error)
	GetComponentParameters(ctx context.Context, componentName string) (*ComponentParameters, error)
	PatchComponentParameters(ctx context.Context, componentName string, params *ComponentParameters) error
	DeleteComponent(ctx context.Context, orgName, projectName, componentName string) error
	UpdateComponent(ctx context.Context, orgName, projectName, componentName string, req *models.UpdateComponentRequest) (*models.Component, error)
	GetDeploymentTrack(ctx context.Context, componentName string) (*models.DeploymentTrack, error)
}

type componentClient struct {
	baseURL    string
	hostHeader string
	httpClient *http.Client
}

// NewComponentClient creates a new OpenChoreo component client.
func NewComponentClient(baseURL, hostHeader string) ComponentClient {
	return &componentClient{
		baseURL:    baseURL,
		hostHeader: hostHeader,
		httpClient: &http.Client{},
	}
}

func (c *componentClient) componentsURL() string {
	return c.baseURL + "/components"
}

func (c *componentClient) componentURL(componentName string) string {
	return fmt.Sprintf("%s/components/%s", c.baseURL, componentName)
}

func (c *componentClient) workflowRunsURL() string {
	return c.baseURL + "/workflowruns"
}

func (c *componentClient) workflowRunURL(runName string) string {
	return fmt.Sprintf("%s/workflowruns/%s", c.baseURL, runName)
}

func (c *componentClient) newRequest(ctx context.Context, name, method, url string) *requests.HttpRequest {
	req := requests.NewRequest(name, method, url)
	if token := middleware.GetAuthToken(ctx); token != "" {
		req.SetHeader("Authorization", "Bearer "+token)
	}
	if c.hostHeader != "" {
		req.SetHost(c.hostHeader)
	}
	return req
}

// normalizeComponent converts a K8s-style OpenChoreo component into the flat
// model returned to callers.
func normalizeComponent(comp ocComponent) models.Component {
	ann := comp.Metadata.Annotations
	labels := comp.Metadata.Labels

	var displayName, description string
	if ann != nil {
		displayName = ann["openchoreo.dev/display-name"]
		description = ann["openchoreo.dev/description"]
	}

	var projectName string
	if comp.Spec.Owner != nil {
		projectName = comp.Spec.Owner.ProjectName
	}
	if projectName == "" && labels != nil {
		projectName = labels["openchoreo.dev/project-name"]
	}

	var componentType string
	if comp.Spec.ComponentType != nil {
		componentType = comp.Spec.ComponentType.Name
	}
	if componentType == "" && labels != nil {
		componentType = labels["openchoreo.dev/component-type"]
	}

	c := models.Component{
		UID:         comp.Metadata.UID,
		Name:        comp.Metadata.Name,
		ProjectName: projectName,
		DisplayName: displayName,
		Description: description,
		Type:        componentType,
		AutoDeploy:  comp.Spec.AutoDeploy,
		AutoBuild:   comp.Spec.AutoBuild,
		CreatedAt:   comp.Metadata.CreationTimestamp,
		Status:      latestConditionReason(comp.Status.Conditions),
	}

	// Synthesize a single deployment track from the component's workflow spec.
	if track := buildDeploymentTrack(comp); track != nil {
		c.DeploymentTracks = []models.DeploymentTrack{*track}
	}

	return c
}

// buildDeploymentTrack creates a DeploymentTrack from the component's workflow
// spec. Returns nil if the component has no workflow/repository configured.
func buildDeploymentTrack(comp ocComponent) *models.DeploymentTrack {
	if comp.Spec.Workflow == nil || comp.Spec.Workflow.Parameters == nil ||
		comp.Spec.Workflow.Parameters.Repository == nil {
		return nil
	}

	repo := comp.Spec.Workflow.Parameters.Repository
	track := &models.DeploymentTrack{
		ID:                comp.Metadata.UID,
		ComponentID:       comp.Metadata.Name,
		Latest:            true,
		AutoDeployEnabled: comp.Spec.AutoDeploy,
		CreatedAt:         comp.Metadata.CreationTimestamp,
		UpdatedAt:         comp.Metadata.CreationTimestamp,
		URL:               repo.URL,
		AppPath:           repo.AppPath,
	}

	if repo.Revision != nil {
		track.Branch = repo.Revision.Branch
		track.CommitSHA = repo.Revision.Commit
	}

	return track
}

// normalizeWorkflowRun converts a K8s-style OpenChoreo workflow run into the
// flat model returned to callers.
func normalizeWorkflowRun(run ocWorkflowRun) models.WorkflowRun {
	labels := run.Metadata.Labels
	var componentName, projectName string
	if labels != nil {
		componentName = labels["openchoreo.dev/component"]
		projectName = labels["openchoreo.dev/project"]
	}

	// Determine status from conditions:
	// - WorkflowCompleted reason if present
	// - "Running" if WorkflowRunning condition is True
	// - else "Pending"
	status := "Pending"
	for _, c := range run.Status.Conditions {
		if c.Type == "WorkflowCompleted" && c.Reason != "" {
			status = c.Reason
			break
		}
		if c.Type == "WorkflowRunning" && c.Status == "True" {
			status = "Running"
		}
	}

	// Extract image from publish-image task output and commit from checkout-source
	var image, commit string
	for _, task := range run.Status.Tasks {
		if task.Outputs == nil {
			continue
		}
		switch task.Name {
		case "publish-image":
			for _, p := range task.Outputs.Parameters {
				if p.Name == "image" {
					image = p.Value
				}
			}
		case "checkout-source":
			for _, p := range task.Outputs.Parameters {
				if p.Name == "git-revision" {
					commit = p.Value
				}
			}
		}
	}
	// Fallback: read commit from the run spec (set when the build is triggered).
	if commit == "" && run.Spec.Workflow != nil &&
		run.Spec.Workflow.Parameters != nil &&
		run.Spec.Workflow.Parameters.Repository != nil &&
		run.Spec.Workflow.Parameters.Repository.Revision != nil {
		commit = run.Spec.Workflow.Parameters.Repository.Revision.Commit
	}

	return models.WorkflowRun{
		Name:          run.Metadata.Name,
		Status:        status,
		StartedAt:     run.Metadata.CreationTimestamp,
		ComponentName: componentName,
		ProjectName:   projectName,
		Image:         image,
		Commit:        commit,
	}
}

// ListComponents returns all components belonging to the given project,
// using a label selector to filter by project name.
func (c *componentClient) ListComponents(ctx context.Context, _, projectName string, limit int, cursor string) (*models.ComponentList, error) {
	req := c.newRequest(ctx, "openchoreo.ListComponents", http.MethodGet, c.componentsURL())
	req.SetQuery("labelSelector", fmt.Sprintf("openchoreo.dev/project=%s", projectName))
	if limit > 0 {
		req.SetQuery("limit", fmt.Sprintf("%d", limit))
	}
	if cursor != "" {
		req.SetQuery("cursor", cursor)
	}

	result := requests.SendRequest(ctx, c.httpClient, req)
	var raw ocComponentList
	if err := result.ScanResponse(&raw, http.StatusOK); err != nil {
		return nil, fmt.Errorf("list components: %w", err)
	}

	items := make([]models.Component, len(raw.Items))
	for i, comp := range raw.Items {
		items[i] = normalizeComponent(comp)
	}
	return &models.ComponentList{Items: items}, nil
}

// CreateComponent creates a component in the OpenChoreo API.
// The request body is forwarded as-is since it already matches the K8s-style
// format expected by the OpenChoreo API.
func (c *componentClient) CreateComponent(ctx context.Context, _, projectName string, req *models.CreateComponentRequest) (*models.Component, error) {
	httpReq := c.newRequest(ctx, "openchoreo.CreateComponent", http.MethodPost, c.componentsURL())
	httpReq.SetJSON(req)

	result := requests.SendRequest(ctx, c.httpClient, httpReq)
	var raw ocComponent
	if err := result.ScanResponse(&raw, http.StatusCreated); err != nil {
		return nil, fmt.Errorf("create component: %w", err)
	}
	comp := normalizeComponent(raw)
	return &comp, nil
}

// UpdateBuildParameters patches the component's workflow configuration.
func (c *componentClient) UpdateBuildParameters(ctx context.Context, _, projectName, componentName string, req *models.UpdateBuildParametersRequest) (*models.Component, error) {
	body := ocComponent{
		Metadata: ocObjectMeta{Name: componentName},
		Spec: ocComponentSpec{
			Owner: &ocOwner{ProjectName: projectName},
		},
	}
	if req.Workflow != nil {
		body.Spec.Workflow = &ocWorkflow{
			Kind: req.Workflow.Kind,
			Name: req.Workflow.Name,
		}
		if req.Workflow.Parameters != nil && req.Workflow.Parameters.Repository != nil {
			repo := req.Workflow.Parameters.Repository
			body.Spec.Workflow.Parameters = &ocWorkflowParameters{
				Repository: &ocWorkflowRepository{
					URL:     repo.URL,
					AppPath: repo.AppPath,
				},
			}
			if repo.Revision != nil {
				body.Spec.Workflow.Parameters.Repository.Revision = &ocWorkflowRevision{
					Branch: repo.Revision.Branch,
					Commit: repo.Revision.Commit,
				}
			}
		}
	}

	httpReq := c.newRequest(ctx, "openchoreo.UpdateBuildParameters", http.MethodPatch, c.componentURL(componentName))
	httpReq.SetJSON(body)

	result := requests.SendRequest(ctx, c.httpClient, httpReq)
	var raw ocComponent
	if err := result.ScanResponse(&raw, http.StatusOK); err != nil {
		return nil, fmt.Errorf("update build parameters: %w", err)
	}
	comp := normalizeComponent(raw)
	return &comp, nil
}

// TriggerBuild triggers a build workflow run for the component.
// It first fetches the component to obtain its workflow configuration, then
// creates a WorkflowRun resource in the OpenChoreo API.
func (c *componentClient) TriggerBuild(ctx context.Context, _, projectName, componentName string) (*models.WorkflowRun, error) {
	// Fetch the component to get its workflow config.
	getReq := c.newRequest(ctx, "openchoreo.GetComponentForBuild", http.MethodGet, c.componentURL(componentName))
	getResult := requests.SendRequest(ctx, c.httpClient, getReq)
	var rawComp ocComponent
	if err := getResult.ScanResponse(&rawComp, http.StatusOK); err != nil {
		return nil, fmt.Errorf("get component for build trigger: %w", err)
	}

	runName := fmt.Sprintf("%s-%d", componentName, time.Now().UnixMilli())
	body := ocWorkflowRun{
		Metadata: ocObjectMeta{
			Name: runName,
			Labels: map[string]string{
				"openchoreo.dev/component": componentName,
				"openchoreo.dev/project":   projectName,
			},
		},
		Spec: ocWorkflowRunSpec{
			Workflow: rawComp.Spec.Workflow,
		},
	}

	httpReq := c.newRequest(ctx, "openchoreo.TriggerBuild", http.MethodPost, c.workflowRunsURL())
	httpReq.SetJSON(body)

	result := requests.SendRequest(ctx, c.httpClient, httpReq)
	var raw ocWorkflowRun
	if err := result.ScanResponse(&raw, http.StatusCreated); err != nil {
		return nil, fmt.Errorf("trigger build: %w", err)
	}
	run := normalizeWorkflowRun(raw)
	return &run, nil
}

// ListWorkflowRuns returns all workflow runs for the given component,
// filtering by the component label.
func (c *componentClient) ListWorkflowRuns(ctx context.Context, _, projectName, componentName string, limit int, cursor string) (*models.WorkflowRunList, error) {
	httpReq := c.newRequest(ctx, "openchoreo.ListWorkflowRuns", http.MethodGet, c.workflowRunsURL())
	httpReq.SetQuery("labelSelector", fmt.Sprintf("openchoreo.dev/component=%s", componentName))
	if limit > 0 {
		httpReq.SetQuery("limit", fmt.Sprintf("%d", limit))
	}
	if cursor != "" {
		httpReq.SetQuery("cursor", cursor)
	}

	result := requests.SendRequest(ctx, c.httpClient, httpReq)
	var raw ocWorkflowRunList
	if err := result.ScanResponse(&raw, http.StatusOK); err != nil {
		return nil, fmt.Errorf("list workflow runs: %w", err)
	}

	items := make([]models.WorkflowRun, len(raw.Items))
	for i, run := range raw.Items {
		items[i] = normalizeWorkflowRun(run)
	}
	return &models.WorkflowRunList{Items: items}, nil
}

// GetWorkflowRun returns a specific workflow run by name.
func (c *componentClient) GetWorkflowRun(ctx context.Context, _, projectName, componentName, runName string) (*models.WorkflowRun, error) {
	httpReq := c.newRequest(ctx, "openchoreo.GetWorkflowRun", http.MethodGet, c.workflowRunURL(runName))

	result := requests.SendRequest(ctx, c.httpClient, httpReq)
	var raw ocWorkflowRun
	if err := result.ScanResponse(&raw, http.StatusOK); err != nil {
		return nil, fmt.Errorf("get workflow run: %w", err)
	}
	run := normalizeWorkflowRun(raw)
	return &run, nil
}

// GetComponentParameters reads the scheduled-task component parameters (backoffLimit, activeDeadlineSeconds).
func (c *componentClient) GetComponentParameters(ctx context.Context, componentName string) (*ComponentParameters, error) {
	httpReq := c.newRequest(ctx, "openchoreo.GetComponentParameters", http.MethodGet, c.componentURL(componentName))

	result := requests.SendRequest(ctx, c.httpClient, httpReq)
	var raw ocComponent
	if err := result.ScanResponse(&raw, http.StatusOK); err != nil {
		return nil, fmt.Errorf("get component parameters: %w", err)
	}
	return raw.Spec.Parameters, nil
}

// PatchComponentParameters updates the scheduled-task component parameters (backoffLimit, activeDeadlineSeconds).
// The Platform API does not support PATCH on components, so this does a GET-then-PUT to preserve
// all existing fields (workflow, autoBuild, etc.) while updating only spec.parameters.
func (c *componentClient) PatchComponentParameters(ctx context.Context, componentName string, params *ComponentParameters) error {
	getReq := c.newRequest(ctx, "openchoreo.GetComponentForPatch", http.MethodGet, c.componentURL(componentName))
	getResult := requests.SendRequest(ctx, c.httpClient, getReq)
	var existing ocComponent
	if err := getResult.ScanResponse(&existing, http.StatusOK); err != nil {
		return fmt.Errorf("get component for parameter update: %w", err)
	}

	existing.Spec.Parameters = params
	putReq := c.newRequest(ctx, "openchoreo.PatchComponentParameters", http.MethodPut, c.componentURL(componentName))
	putReq.SetJSON(existing)

	result := requests.SendRequest(ctx, c.httpClient, putReq)
	if err := result.ScanResponse(nil, http.StatusOK); err != nil {
		return fmt.Errorf("put component parameters: %w", err)
	}
	return nil
}

// DeleteComponent deletes a component from the OpenChoreo API.
func (c *componentClient) DeleteComponent(ctx context.Context, _, _, componentName string) error {
	httpReq := c.newRequest(ctx, "openchoreo.DeleteComponent", http.MethodDelete, c.componentURL(componentName))
	result := requests.SendRequest(ctx, c.httpClient, httpReq)
	if err := result.ScanResponse(nil, http.StatusNoContent); err != nil {
		return fmt.Errorf("delete component: %w", err)
	}
	return nil
}

// UpdateComponent updates the mutable metadata (displayName, description) of a component
// using a GET-then-PUT strategy to preserve all existing fields.
func (c *componentClient) UpdateComponent(ctx context.Context, _, _, componentName string, req *models.UpdateComponentRequest) (*models.Component, error) {
	getReq := c.newRequest(ctx, "openchoreo.GetComponentForUpdate", http.MethodGet, c.componentURL(componentName))
	getResult := requests.SendRequest(ctx, c.httpClient, getReq)
	var existing ocComponent
	if err := getResult.ScanResponse(&existing, http.StatusOK); err != nil {
		return nil, fmt.Errorf("get component for update: %w", err)
	}

	if existing.Metadata.Annotations == nil {
		existing.Metadata.Annotations = make(map[string]string)
	}
	if req.DisplayName != "" {
		existing.Metadata.Annotations["openchoreo.dev/display-name"] = req.DisplayName
	}
	if req.Description != "" {
		existing.Metadata.Annotations["openchoreo.dev/description"] = req.Description
	}

	putReq := c.newRequest(ctx, "openchoreo.UpdateComponent", http.MethodPut, c.componentURL(componentName))
	putReq.SetJSON(existing)

	result := requests.SendRequest(ctx, c.httpClient, putReq)
	var raw ocComponent
	if err := result.ScanResponse(&raw, http.StatusOK); err != nil {
		return nil, fmt.Errorf("update component: %w", err)
	}
	comp := normalizeComponent(raw)
	return &comp, nil
}

// GetComponent fetches a single component from OpenChoreo.
func (c *componentClient) GetComponent(ctx context.Context, componentName string) (*models.Component, error) {
	httpReq := c.newRequest(ctx, "openchoreo.GetComponent", http.MethodGet, c.componentURL(componentName))
	result := requests.SendRequest(ctx, c.httpClient, httpReq)
	var raw ocComponent
	if err := result.ScanResponse(&raw, http.StatusOK); err != nil {
		return nil, fmt.Errorf("get component: %w", err)
	}
	comp := normalizeComponent(raw)
	return &comp, nil
}

// GetDeploymentTrack returns the deployment track synthesized from a
// component's workflow spec in OpenChoreo.
func (c *componentClient) GetDeploymentTrack(ctx context.Context, componentName string) (*models.DeploymentTrack, error) {
	httpReq := c.newRequest(ctx, "openchoreo.GetDeploymentTrack", http.MethodGet, c.componentURL(componentName))
	result := requests.SendRequest(ctx, c.httpClient, httpReq)
	var raw ocComponent
	if err := result.ScanResponse(&raw, http.StatusOK); err != nil {
		return nil, fmt.Errorf("get deployment track: %w", err)
	}

	track := buildDeploymentTrack(raw)
	if track == nil {
		return &models.DeploymentTrack{}, nil
	}
	return track, nil
}
