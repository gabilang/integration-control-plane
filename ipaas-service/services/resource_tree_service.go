package services

import (
	"context"
	"strings"

	"github.com/wso2/integration-control-plane/ipaas-service/clients/openchoreo"
	"github.com/wso2/integration-control-plane/ipaas-service/models"
)

// ResourceTreeService retrieves K8s resource trees, executions, events, and
// pod logs via the OpenChoreo ReleaseBinding resource tree API.
type ResourceTreeService interface {
	GetResourceTree(ctx context.Context, orgName, componentName, environment string) (*models.ResourceTreeResponse, error)
	GetExecutionsFromTree(ctx context.Context, orgName, componentName, environment string) (*models.ExecutionList, error)
	GetResourceEvents(ctx context.Context, orgName, componentName, environment, group, version, kind, name string) (*models.ResourceEventsResponse, error)
	GetResourceLogs(ctx context.Context, orgName, componentName, environment, podName string, sinceSeconds *int64) (*models.PodLogsResponse, error)
}

type resourceTreeService struct {
	scheduleClient openchoreo.ScheduleClient
}

func NewResourceTreeService(scheduleClient openchoreo.ScheduleClient) ResourceTreeService {
	return &resourceTreeService{scheduleClient: scheduleClient}
}

func (s *resourceTreeService) GetResourceTree(ctx context.Context, orgName, componentName, environment string) (*models.ResourceTreeResponse, error) {
	tree, err := s.scheduleClient.GetResourceTree(ctx, orgName, componentName, environment)
	if err != nil {
		return nil, translateScheduleHTTPError(err)
	}
	return tree, nil
}

// GetExecutionsFromTree fetches the resource tree and extracts Job nodes as executions.
func (s *resourceTreeService) GetExecutionsFromTree(ctx context.Context, orgName, componentName, environment string) (*models.ExecutionList, error) {
	tree, err := s.scheduleClient.GetResourceTree(ctx, orgName, componentName, environment)
	if err != nil {
		return nil, translateScheduleHTTPError(err)
	}

	var executions []models.Execution
	for _, release := range tree.RenderedReleases {
		for _, node := range release.Nodes {
			if node.Kind != "Job" {
				continue
			}
			executions = append(executions, extractExecutionFromNode(node))
		}
	}
	if executions == nil {
		executions = []models.Execution{}
	}
	return &models.ExecutionList{Items: executions}, nil
}

func (s *resourceTreeService) GetResourceEvents(ctx context.Context, orgName, componentName, environment, group, version, kind, name string) (*models.ResourceEventsResponse, error) {
	resp, err := s.scheduleClient.GetResourceEvents(ctx, orgName, componentName, environment, group, version, kind, name)
	if err != nil {
		return nil, translateScheduleHTTPError(err)
	}
	return resp, nil
}

func (s *resourceTreeService) GetResourceLogs(ctx context.Context, orgName, componentName, environment, podName string, sinceSeconds *int64) (*models.PodLogsResponse, error) {
	resp, err := s.scheduleClient.GetResourceLogs(ctx, orgName, componentName, environment, podName, sinceSeconds)
	if err != nil {
		return nil, translateScheduleHTTPError(err)
	}
	return resp, nil
}

// extractExecutionFromNode converts a Job ResourceNode into an Execution model
// by reading the full K8s Job object embedded in the node.
func extractExecutionFromNode(node models.ResourceNode) models.Execution {
	obj := node.Object

	status := "Unknown"
	if node.Health != nil {
		switch node.Health.Status {
		case "Healthy":
			status = "Succeeded"
		case "Degraded":
			status = "Failed"
		case "Progressing":
			status = "Running"
		}
	}

	// Try to extract more precise status from the embedded Job object.
	if jobStatus, ok := getNestedMap(obj, "status"); ok {
		active, _ := getNestedFloat(jobStatus, "active")
		succeeded, _ := getNestedFloat(jobStatus, "succeeded")
		failed, _ := getNestedFloat(jobStatus, "failed")
		switch {
		case active > 0:
			status = "Running"
		case succeeded > 0:
			status = "Succeeded"
		case failed > 0:
			status = "Failed"
		}
	}

	var startTime, completionTime string
	if jobStatus, ok := getNestedMap(obj, "status"); ok {
		startTime, _ = getNestedString(jobStatus, "startTime")
		completionTime, _ = getNestedString(jobStatus, "completionTime")
	}

	var revisionID string
	if spec, ok := getNestedMap(obj, "spec"); ok {
		if tmpl, ok := getNestedMap(spec, "template"); ok {
			if podSpec, ok := getNestedMap(tmpl, "spec"); ok {
				if containers, ok := podSpec["containers"].([]any); ok && len(containers) > 0 {
					if c, ok := containers[0].(map[string]any); ok {
						if image, ok := c["image"].(string); ok {
							revisionID = extractRevisionFromImage(image)
						}
					}
				}
			}
		}
	}

	return models.Execution{
		JobID:          node.Name,
		Status:         status,
		StartTime:      startTime,
		CompletionTime: completionTime,
		RevisionID:     revisionID,
	}
}

// extractRevisionFromImage extracts the git revision from a container image tag.
// Image format: "registry/org/component:v1-3c574505" → "3c574505"
func extractRevisionFromImage(image string) string {
	tag := image
	if idx := strings.LastIndex(image, ":"); idx >= 0 {
		tag = image[idx+1:]
	}
	parts := strings.Split(tag, "-")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return ""
}

func getNestedMap(obj map[string]any, key string) (map[string]any, bool) {
	v, ok := obj[key]
	if !ok {
		return nil, false
	}
	m, ok := v.(map[string]any)
	return m, ok
}

func getNestedString(obj map[string]any, key string) (string, bool) {
	v, ok := obj[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func getNestedFloat(obj map[string]any, key string) (float64, bool) {
	v, ok := obj[key]
	if !ok {
		return 0, false
	}
	f, ok := v.(float64)
	return f, ok
}

// Ensure the service satisfies the interface at compile time.
var _ ResourceTreeService = (*resourceTreeService)(nil)
