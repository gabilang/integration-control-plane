package openchoreo

import (
	"context"
	"fmt"
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/clients/requests"
	"github.com/wso2/integration-control-plane/ipaas-service/middleware"
	"github.com/wso2/integration-control-plane/ipaas-service/models"
)

// ScheduleClient defines operations for managing OpenChoreo ReleaseBindings
// for scheduled-task components.
type ScheduleClient interface {
	ListReleaseBindings(ctx context.Context, orgName, projectName, componentName string) (*models.ScheduleList, error)
	GetReleaseBinding(ctx context.Context, orgName, componentName, environment string) (*models.Schedule, error)
	GetReleaseBindingNamespace(ctx context.Context, orgName, componentName, environment string) (string, error)
	CreateReleaseBinding(ctx context.Context, orgName, projectName, componentName string, req *models.UpsertScheduleRequest) (*models.Schedule, error)
	UpdateReleaseBinding(ctx context.Context, orgName, projectName, componentName string, req *models.UpsertScheduleRequest) (*models.Schedule, error)
	DeleteReleaseBinding(ctx context.Context, orgName, componentName, environment string) error
	GenerateRelease(ctx context.Context, orgName, componentName string) (string, error)
	GetResourceTree(ctx context.Context, orgName, componentName, environment string) (*models.ResourceTreeResponse, error)
	GetResourceEvents(ctx context.Context, orgName, componentName, environment, group, version, kind, name string) (*models.ResourceEventsResponse, error)
	GetResourceLogs(ctx context.Context, orgName, componentName, environment, podName string, sinceSeconds *int64) (*models.PodLogsResponse, error)
}

type scheduleClient struct {
	baseURL    string
	hostHeader string
	httpClient *http.Client
}

// NewScheduleClient creates a new OpenChoreo schedule (ReleaseBinding) client.
func NewScheduleClient(baseURL, hostHeader string) ScheduleClient {
	return &scheduleClient{
		baseURL:    baseURL,
		hostHeader: hostHeader,
		httpClient: &http.Client{},
	}
}

func (c *scheduleClient) releaseBindingsURL() string {
	return c.baseURL + "/releasebindings"
}

func (c *scheduleClient) releaseBindingURL(name string) string {
	return fmt.Sprintf("%s/releasebindings/%s", c.baseURL, name)
}

// releaseBindingName returns the conventional name for a ReleaseBinding:
// {componentName}-{environment} (e.g., "my-task-development").
func releaseBindingName(componentName, environment string) string {
	return componentName + "-" + environment
}

func (c *scheduleClient) newRequest(ctx context.Context, name, method, url string) *requests.HttpRequest {
	req := requests.NewRequest(name, method, url)
	if token := middleware.GetAuthToken(ctx); token != "" {
		req.SetHeader("Authorization", "Bearer "+token)
	}
	if c.hostHeader != "" {
		req.SetHost(c.hostHeader)
	}
	return req
}

func normalizeReleaseBinding(rb ocReleaseBinding) models.Schedule {
	labels := rb.Metadata.Labels

	projectName := rb.Spec.Owner.ProjectName
	if labels != nil && labels["openchoreo.dev/project-name"] != "" {
		projectName = labels["openchoreo.dev/project-name"]
	}

	imagePullPolicy := rb.Spec.ComponentTypeEnvironmentConfigs.ImagePullPolicy
	if imagePullPolicy == "" {
		imagePullPolicy = "IfNotPresent"
	}

	return models.Schedule{
		Environment:     rb.Spec.Environment,
		ComponentName:   rb.Spec.Owner.ComponentName,
		ProjectName:     projectName,
		CronExpression:  rb.Spec.ComponentTypeEnvironmentConfigs.Schedule,
		State:           rb.Spec.State,
		ImagePullPolicy: imagePullPolicy,
		ReleaseName:     rb.Spec.ReleaseName,
	}
}

func buildReleaseBindingBody(projectName, componentName, releaseName string, req *models.UpsertScheduleRequest) ocReleaseBinding {
	state := req.State
	if state == "" {
		state = "Active"
	}
	if releaseName == "" {
		releaseName = req.ReleaseName
	}
	return ocReleaseBinding{
		Metadata: ocObjectMeta{
			Name: releaseBindingName(componentName, req.Environment),
		},
		Spec: ocReleaseBindingSpec{
			Owner: ocReleaseBindingOwner{
				ProjectName:   projectName,
				ComponentName: componentName,
			},
			Environment: req.Environment,
			State:       state,
			ComponentTypeEnvironmentConfigs: ocReleaseBindingEnvConfigs{
				Schedule:        req.CronExpression,
				ImagePullPolicy: "IfNotPresent",
			},
			ReleaseName: releaseName,
		},
	}
}

// generateReleaseURL returns the URL for generating a release snapshot.
func (c *scheduleClient) generateReleaseURL(componentName string) string {
	return fmt.Sprintf("%s/components/%s/generate-release", c.baseURL, componentName)
}

// GenerateRelease generates a ComponentRelease snapshot from the latest Workload
// and returns the generated release name.
func (c *scheduleClient) GenerateRelease(ctx context.Context, _, componentName string) (string, error) {
	req := c.newRequest(ctx, "openchoreo.GenerateRelease", http.MethodPost, c.generateReleaseURL(componentName))
	req.SetJSON(map[string]any{})

	result := requests.SendRequest(ctx, c.httpClient, req)
	var raw ocComponentRelease
	if err := result.ScanResponse(&raw, http.StatusCreated); err != nil {
		return "", fmt.Errorf("generate release: %w", err)
	}
	return raw.Metadata.Name, nil
}

// ListReleaseBindings returns all release bindings for the given component,
// filtered by the component label.
func (c *scheduleClient) ListReleaseBindings(ctx context.Context, _, projectName, componentName string) (*models.ScheduleList, error) {
	req := c.newRequest(ctx, "openchoreo.ListReleaseBindings", http.MethodGet, c.releaseBindingsURL())
	req.SetQuery("labelSelector", fmt.Sprintf("openchoreo.dev/component=%s", componentName))

	result := requests.SendRequest(ctx, c.httpClient, req)
	var raw ocReleaseBindingList
	if err := result.ScanResponse(&raw, http.StatusOK); err != nil {
		return nil, fmt.Errorf("list release bindings: %w", err)
	}

	items := make([]models.Schedule, len(raw.Items))
	for i, rb := range raw.Items {
		items[i] = normalizeReleaseBinding(rb)
	}
	return &models.ScheduleList{Items: items}, nil
}

// GetReleaseBinding returns the release binding for the given component and environment.
func (c *scheduleClient) GetReleaseBinding(ctx context.Context, _, componentName, environment string) (*models.Schedule, error) {
	name := releaseBindingName(componentName, environment)
	req := c.newRequest(ctx, "openchoreo.GetReleaseBinding", http.MethodGet, c.releaseBindingURL(name))

	result := requests.SendRequest(ctx, c.httpClient, req)
	var raw ocReleaseBinding
	if err := result.ScanResponse(&raw, http.StatusOK); err != nil {
		return nil, fmt.Errorf("get release binding: %w", err)
	}
	s := normalizeReleaseBinding(raw)
	return &s, nil
}

// GetReleaseBindingNamespace returns the Kubernetes namespace of the ReleaseBinding,
// which is the org-scoped namespace used to label CronJob-spawned Jobs.
func (c *scheduleClient) GetReleaseBindingNamespace(ctx context.Context, _, componentName, environment string) (string, error) {
	name := releaseBindingName(componentName, environment)
	req := c.newRequest(ctx, "openchoreo.GetReleaseBindingNamespace", http.MethodGet, c.releaseBindingURL(name))

	result := requests.SendRequest(ctx, c.httpClient, req)
	var raw ocReleaseBinding
	if err := result.ScanResponse(&raw, http.StatusOK); err != nil {
		return "", fmt.Errorf("get release binding namespace: %w", err)
	}
	return raw.Metadata.Namespace, nil
}

// CreateReleaseBinding creates a new release binding for the given component.
func (c *scheduleClient) CreateReleaseBinding(ctx context.Context, orgName, projectName, componentName string, req *models.UpsertScheduleRequest) (*models.Schedule, error) {
	releaseName := req.ReleaseName
	if releaseName == "" {
		var err error
		releaseName, err = c.GenerateRelease(ctx, orgName, componentName)
		if err != nil {
			return nil, fmt.Errorf("auto generate release: %w", err)
		}
	}
	body := buildReleaseBindingBody(projectName, componentName, releaseName, req)
	httpReq := c.newRequest(ctx, "openchoreo.CreateReleaseBinding", http.MethodPost, c.releaseBindingsURL())
	httpReq.SetJSON(body)

	result := requests.SendRequest(ctx, c.httpClient, httpReq)
	var raw ocReleaseBinding
	if err := result.ScanResponse(&raw, http.StatusCreated); err != nil {
		return nil, fmt.Errorf("create release binding: %w", err)
	}
	s := normalizeReleaseBinding(raw)
	return &s, nil
}

// UpdateReleaseBinding patches the release binding for the given component and environment.
func (c *scheduleClient) UpdateReleaseBinding(ctx context.Context, _, projectName, componentName string, req *models.UpsertScheduleRequest) (*models.Schedule, error) {
	name := releaseBindingName(componentName, req.Environment)
	body := buildReleaseBindingBody(projectName, componentName, "", req)
	httpReq := c.newRequest(ctx, "openchoreo.UpdateReleaseBinding", http.MethodPut, c.releaseBindingURL(name))
	httpReq.SetJSON(body)

	result := requests.SendRequest(ctx, c.httpClient, httpReq)
	var raw ocReleaseBinding
	if err := result.ScanResponse(&raw, http.StatusOK); err != nil {
		return nil, fmt.Errorf("update release binding: %w", err)
	}
	s := normalizeReleaseBinding(raw)
	return &s, nil
}

// DeleteReleaseBinding deletes the release binding for the given component and environment.
func (c *scheduleClient) DeleteReleaseBinding(ctx context.Context, _, componentName, environment string) error {
	name := releaseBindingName(componentName, environment)
	req := c.newRequest(ctx, "openchoreo.DeleteReleaseBinding", http.MethodDelete, c.releaseBindingURL(name))

	result := requests.SendRequest(ctx, c.httpClient, req)
	if err := result.ScanResponse(nil, http.StatusNoContent); err != nil {
		return fmt.Errorf("delete release binding: %w", err)
	}
	return nil
}

// GetResourceTree returns the K8s resource tree for the given component's release binding.
func (c *scheduleClient) GetResourceTree(ctx context.Context, _, componentName, environment string) (*models.ResourceTreeResponse, error) {
	name := releaseBindingName(componentName, environment)
	url := fmt.Sprintf("%s/releasebindings/%s/k8sresources/tree", c.baseURL, name)
	req := c.newRequest(ctx, "openchoreo.GetResourceTree", http.MethodGet, url)

	result := requests.SendRequest(ctx, c.httpClient, req)
	var raw ocK8sResourceTreeResponse
	if err := result.ScanResponse(&raw, http.StatusOK); err != nil {
		return nil, fmt.Errorf("get resource tree: %w", err)
	}
	return normalizeResourceTree(raw), nil
}

// GetResourceEvents returns K8s events for a specific resource in the release binding's resource tree.
func (c *scheduleClient) GetResourceEvents(ctx context.Context, _, componentName, environment, group, version, kind, name string) (*models.ResourceEventsResponse, error) {
	rbName := releaseBindingName(componentName, environment)
	url := fmt.Sprintf("%s/releasebindings/%s/k8sresources/events", c.baseURL, rbName)
	req := c.newRequest(ctx, "openchoreo.GetResourceEvents", http.MethodGet, url)
	if group != "" {
		req.SetQuery("group", group)
	}
	req.SetQuery("version", version)
	req.SetQuery("kind", kind)
	req.SetQuery("name", name)

	result := requests.SendRequest(ctx, c.httpClient, req)
	var raw ocResourceEventsResponse
	if err := result.ScanResponse(&raw, http.StatusOK); err != nil {
		return nil, fmt.Errorf("get resource events: %w", err)
	}
	return normalizeResourceEvents(raw), nil
}

// GetResourceLogs returns logs for a specific pod in the release binding's resource tree.
func (c *scheduleClient) GetResourceLogs(ctx context.Context, _, componentName, environment, podName string, sinceSeconds *int64) (*models.PodLogsResponse, error) {
	rbName := releaseBindingName(componentName, environment)
	url := fmt.Sprintf("%s/releasebindings/%s/k8sresources/logs", c.baseURL, rbName)
	req := c.newRequest(ctx, "openchoreo.GetResourceLogs", http.MethodGet, url)
	req.SetQuery("podName", podName)
	if sinceSeconds != nil {
		req.SetQuery("sinceSeconds", fmt.Sprintf("%d", *sinceSeconds))
	}

	result := requests.SendRequest(ctx, c.httpClient, req)
	var raw ocResourcePodLogsResponse
	if err := result.ScanResponse(&raw, http.StatusOK); err != nil {
		return nil, fmt.Errorf("get resource logs: %w", err)
	}
	return normalizeResourceLogs(raw), nil
}

func normalizeResourceTree(raw ocK8sResourceTreeResponse) *models.ResourceTreeResponse {
	releases := make([]models.ReleaseResourceTree, len(raw.RenderedReleases))
	for i, rr := range raw.RenderedReleases {
		nodes := make([]models.ResourceNode, len(rr.Nodes))
		for j, n := range rr.Nodes {
			parentRefs := make([]models.ResourceRef, len(n.ParentRefs))
			for k, pr := range n.ParentRefs {
				parentRefs[k] = models.ResourceRef{
					Group:     pr.Group,
					Version:   pr.Version,
					Kind:      pr.Kind,
					Namespace: pr.Namespace,
					Name:      pr.Name,
					UID:       pr.UID,
				}
			}
			var health *models.HealthInfo
			if n.Health != nil {
				health = &models.HealthInfo{
					Status:  n.Health.Status,
					Message: n.Health.Message,
				}
			}
			nodes[j] = models.ResourceNode{
				Group:           n.Group,
				Version:         n.Version,
				Kind:            n.Kind,
				Namespace:       n.Namespace,
				Name:            n.Name,
				UID:             n.UID,
				ResourceVersion: n.ResourceVersion,
				CreatedAt:       n.CreatedAt,
				ParentRefs:      parentRefs,
				Object:          n.Object,
				Health:          health,
			}
		}
		releases[i] = models.ReleaseResourceTree{
			Name:        rr.Name,
			TargetPlane: rr.TargetPlane,
			Nodes:       nodes,
		}
	}
	return &models.ResourceTreeResponse{RenderedReleases: releases}
}

func normalizeResourceEvents(raw ocResourceEventsResponse) *models.ResourceEventsResponse {
	events := make([]models.ResourceEvent, len(raw.Events))
	for i, e := range raw.Events {
		events[i] = models.ResourceEvent{
			Type:           e.Type,
			Reason:         e.Reason,
			Message:        e.Message,
			Count:          e.Count,
			FirstTimestamp: e.FirstTimestamp,
			LastTimestamp:   e.LastTimestamp,
			Source:         e.Source,
		}
	}
	return &models.ResourceEventsResponse{Events: events}
}

func normalizeResourceLogs(raw ocResourcePodLogsResponse) *models.PodLogsResponse {
	entries := make([]models.PodLogEntry, len(raw.LogEntries))
	for i, e := range raw.LogEntries {
		entries[i] = models.PodLogEntry{
			Timestamp: e.Timestamp,
			Log:       e.Log,
		}
	}
	return &models.PodLogsResponse{LogEntries: entries}
}
