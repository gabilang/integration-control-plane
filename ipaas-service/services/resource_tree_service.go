package services

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	ghclient "github.com/wso2/integration-control-plane/ipaas-service/clients/github"
	"github.com/wso2/integration-control-plane/ipaas-service/clients/k8s"
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
	scheduleClient  openchoreo.ScheduleClient
	jobsClient      k8s.JobsClient
	componentClient openchoreo.ComponentClient
	ghClient        *ghclient.Client
}

func NewResourceTreeService(scheduleClient openchoreo.ScheduleClient, jobsClient k8s.JobsClient, componentClient openchoreo.ComponentClient, ghClient *ghclient.Client) ResourceTreeService {
	return &resourceTreeService{scheduleClient: scheduleClient, jobsClient: jobsClient, componentClient: componentClient, ghClient: ghClient}
}

func (s *resourceTreeService) GetResourceTree(ctx context.Context, orgName, componentName, environment string) (*models.ResourceTreeResponse, error) {
	tree, err := s.scheduleClient.GetResourceTree(ctx, orgName, componentName, environment)
	if err != nil {
		return nil, translateScheduleHTTPError(err)
	}
	return tree, nil
}

// GetExecutionsFromTree fetches the resource tree and merges it with a direct
// K8s label-selector query. The resource tree only contains Jobs owned by the
// CronJob (scheduled runs); manually triggered Jobs have no ownerReference to
// the CronJob and therefore don't appear in the tree. Merging both sources
// ensures all executions are returned. Results are deduplicated by JobID and
// sorted most-recent-first.
func (s *resourceTreeService) GetExecutionsFromTree(ctx context.Context, orgName, componentName, environment string) (*models.ExecutionList, error) {
	seen := make(map[string]struct{})
	var executions []models.Execution

	// 1. Resource tree: Jobs owned by the CronJob (scheduled runs).
	tree, treeErr := s.scheduleClient.GetResourceTree(ctx, orgName, componentName, environment)
	if treeErr == nil {
		for _, release := range tree.RenderedReleases {
			for _, node := range release.Nodes {
				if node.Kind != "Job" {
					continue
				}
				seen[node.Name] = struct{}{}
				executions = append(executions, extractExecutionFromNode(node))
			}
		}
	}

	// 2. Direct K8s query: picks up manually triggered Jobs that are not in the tree.
	if fallback, fbErr := s.fallbackListExecutions(ctx, orgName, componentName, environment); fbErr == nil {
		for _, exec := range fallback.Items {
			if _, dup := seen[exec.JobID]; dup {
				continue
			}
			executions = append(executions, exec)
		}
	}

	if executions == nil {
		executions = []models.Execution{}
	}
	s.resolvePathSpecificRevisions(ctx, componentName, executions)
	sortExecutionsDesc(executions)
	return &models.ExecutionList{Items: executions}, nil
}

// sortExecutionsDesc sorts executions by StartTime descending (most recent first).
func sortExecutionsDesc(items []models.Execution) {
	sort.Slice(items, func(i, j int) bool {
		ti, _ := time.Parse(time.RFC3339, items[i].StartTime)
		tj, _ := time.Parse(time.RFC3339, items[j].StartTime)
		return ti.After(tj)
	})
}

// fallbackListExecutions queries K8s Jobs directly using label selectors,
// bypassing the resource tree. Returns empty (not an error) when unavailable.
func (s *resourceTreeService) fallbackListExecutions(ctx context.Context, orgName, componentName, environment string) (*models.ExecutionList, error) {
	if s.jobsClient == nil {
		return &models.ExecutionList{Items: []models.Execution{}}, nil
	}
	orgNs, err := s.scheduleClient.GetReleaseBindingNamespace(ctx, orgName, componentName, environment)
	if err != nil {
		return nil, err
	}
	labelSelector := fmt.Sprintf(
		"openchoreo.dev/component=%s,openchoreo.dev/environment=%s,openchoreo.dev/namespace=%s",
		componentName, environment, orgNs,
	)
	return s.jobsClient.ListJobs(ctx, labelSelector)
}

// resolvePathSpecificRevisions replaces each execution's revisionId (which is the
// repo-wide HEAD at build time) with the last commit that actually touched the
// component's subdirectory.
func (s *resourceTreeService) resolvePathSpecificRevisions(ctx context.Context, componentName string, executions []models.Execution) {
	if s.ghClient == nil || !s.ghClient.IsAvailable() || s.componentClient == nil {
		return
	}

	track, err := s.componentClient.GetDeploymentTrack(ctx, componentName)
	if err != nil || track.URL == "" || track.AppPath == "" || track.AppPath == "/" || track.AppPath == "." {
		return
	}

	// Collect unique revisionIds to avoid duplicate GitHub API calls.
	unique := make(map[string]string) // repoCommit → pathCommit
	for _, e := range executions {
		if e.RevisionID != "" {
			unique[e.RevisionID] = ""
		}
	}

	for repoCommit := range unique {
		commits, ghErr := s.ghClient.GetCommits(ctx, track.URL, repoCommit, track.AppPath, 1)
		if ghErr != nil {
			slog.WarnContext(ctx, "failed to resolve path-specific commit",
				"component", componentName, "commit", repoCommit, "error", ghErr)
			continue
		}
		if len(commits) > 0 {
			unique[repoCommit] = commits[0].SHA
		}
	}

	for i := range executions {
		if mapped, ok := unique[executions[i].RevisionID]; ok && mapped != "" {
			executions[i].RevisionID = mapped
		}
	}
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
