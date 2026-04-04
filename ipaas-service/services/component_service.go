package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	ghclient "github.com/wso2/integration-control-plane/ipaas-service/clients/github"
	"github.com/wso2/integration-control-plane/ipaas-service/clients/icp"
	"github.com/wso2/integration-control-plane/ipaas-service/clients/observability"
	"github.com/wso2/integration-control-plane/ipaas-service/clients/openchoreo"
	"github.com/wso2/integration-control-plane/ipaas-service/clients/requests"
	"github.com/wso2/integration-control-plane/ipaas-service/models"
)

var (
	ErrComponentNotFound = errors.New("component not found")
	ErrBuildNotFound     = errors.New("build not found")
	ErrLogsUnavailable   = errors.New("observability service not configured")
)

// ComponentService handles business logic for component operations.
type ComponentService interface {
	ListComponents(ctx context.Context, orgName, projectName string, limit int, cursor string) (*models.ComponentList, error)
	CreateComponent(ctx context.Context, orgName, projectName string, req *models.CreateComponentRequest) (*models.CreateComponentResponse, error)
	UpdateBuildParameters(ctx context.Context, orgName, projectName, componentName string, req *models.UpdateBuildParametersRequest) (*models.Component, error)
	UpdateComponent(ctx context.Context, orgName, projectName, componentName string, req *models.UpdateComponentRequest) (*models.Component, error)
	DeleteComponent(ctx context.Context, orgName, projectName, componentName string) error
	TriggerBuild(ctx context.Context, orgName, projectName, componentName string) (*models.WorkflowRun, error)
	ListBuilds(ctx context.Context, orgName, projectName, componentName string, limit int, cursor string) (*models.WorkflowRunList, error)
	GetBuildStatus(ctx context.Context, orgName, projectName, componentName, buildName string) (*models.WorkflowRun, error)
	GetBuildLogs(ctx context.Context, orgName, projectName, componentName, buildName string) (*models.BuildLogs, error)
	GetComponentRepository(ctx context.Context, projectID, componentHandler string) (*models.ComponentRepository, error)
	GetCommitHistory(ctx context.Context, orgName, projectName, componentName, branch string) (*models.CommitList, error)
	GetComponentLabels(ctx context.Context, projectID string, orgID int) (*models.LabelList, error)
	GetDeploymentTrack(ctx context.Context, componentName string) (*models.DeploymentTrack, error)
}

type componentService struct {
	client       openchoreo.ComponentClient
	observClient observability.Client
	icpClient    *icp.Client
	ghClient     *ghclient.Client
}

func NewComponentService(client openchoreo.ComponentClient, observClient observability.Client, icpClient *icp.Client, ghClient *ghclient.Client) ComponentService {
	return &componentService{client: client, observClient: observClient, icpClient: icpClient, ghClient: ghClient}
}

func (s *componentService) ListComponents(ctx context.Context, orgName, projectName string, limit int, cursor string) (*models.ComponentList, error) {
	list, err := s.client.ListComponents(ctx, orgName, projectName, limit, cursor)
	if err != nil {
		return nil, translateComponentHTTPError(err)
	}
	return list, nil
}

// CreateComponent creates the component. If autoBuild is enabled in the request,
// it also triggers the initial build. If the build trigger fails the component
// is still returned — the caller can retry the build later.
func (s *componentService) CreateComponent(ctx context.Context, orgName, projectName string, req *models.CreateComponentRequest) (*models.CreateComponentResponse, error) {
	component, err := s.client.CreateComponent(ctx, orgName, projectName, req)
	if err != nil {
		return nil, translateComponentHTTPError(err)
	}

	var buildRun *models.WorkflowRun
	if req.Spec.AutoBuild {
		buildRun, err = s.client.TriggerBuild(ctx, orgName, projectName, component.Name)
		if err != nil {
			slog.WarnContext(ctx, "initial build trigger failed",
				"error", err,
				"org", orgName,
				"project", projectName,
				"component", component.Name,
			)
		}
	}

	return &models.CreateComponentResponse{
		Component: component,
		BuildRun:  buildRun,
	}, nil
}

func (s *componentService) UpdateBuildParameters(ctx context.Context, orgName, projectName, componentName string, req *models.UpdateBuildParametersRequest) (*models.Component, error) {
	component, err := s.client.UpdateBuildParameters(ctx, orgName, projectName, componentName, req)
	if err != nil {
		return nil, translateComponentHTTPError(err)
	}
	return component, nil
}

func (s *componentService) UpdateComponent(ctx context.Context, orgName, projectName, componentName string, req *models.UpdateComponentRequest) (*models.Component, error) {
	component, err := s.client.UpdateComponent(ctx, orgName, projectName, componentName, req)
	if err != nil {
		return nil, translateComponentHTTPError(err)
	}
	return component, nil
}

func (s *componentService) DeleteComponent(ctx context.Context, orgName, projectName, componentName string) error {
	if err := s.client.DeleteComponent(ctx, orgName, projectName, componentName); err != nil {
		return translateComponentHTTPError(err)
	}
	return nil
}
func (s *componentService) TriggerBuild(ctx context.Context, orgName, projectName, componentName string) (*models.WorkflowRun, error) {
	run, err := s.client.TriggerBuild(ctx, orgName, projectName, componentName)
	if err != nil {
		return nil, translateComponentHTTPError(err)
	}
	return run, nil
}

func (s *componentService) ListBuilds(ctx context.Context, orgName, projectName, componentName string, limit int, cursor string) (*models.WorkflowRunList, error) {
	list, err := s.client.ListWorkflowRuns(ctx, orgName, projectName, componentName, limit, cursor)
	if err != nil {
		return nil, translateComponentHTTPError(err)
	}
	return list, nil
}

func (s *componentService) GetBuildStatus(ctx context.Context, orgName, projectName, componentName, buildName string) (*models.WorkflowRun, error) {
	run, err := s.client.GetWorkflowRun(ctx, orgName, projectName, componentName, buildName)
	if err != nil {
		return nil, translateComponentHTTPError(err)
	}
	return run, nil
}

func (s *componentService) GetBuildLogs(ctx context.Context, orgName, projectName, componentName, buildName string) (*models.BuildLogs, error) {
	if s.observClient == nil {
		return nil, ErrLogsUnavailable
	}
	logs, err := s.observClient.GetBuildLogs(ctx, orgName, projectName, componentName, buildName)
	if err != nil {
		return nil, fmt.Errorf("get build logs: %w", err)
	}
	return logs, nil
}

func translateComponentHTTPError(err error) error {
	if err == nil {
		return nil
	}
	var httpErr *requests.HttpError
	if errors.As(err, &httpErr) {
		switch httpErr.StatusCode {
		case http.StatusNotFound:
			return fmt.Errorf("%w: %s", ErrComponentNotFound, httpErr.Body)
		case http.StatusUnauthorized:
			return ErrUnauthorized
		}
	}
	return err
}

func translateBuildHTTPError(err error) error {
	if err == nil {
		return nil
	}
	var httpErr *requests.HttpError
	if errors.As(err, &httpErr) {
		switch httpErr.StatusCode {
		case http.StatusNotFound:
			return fmt.Errorf("%w: %s", ErrBuildNotFound, httpErr.Body)
		case http.StatusUnauthorized:
			return ErrUnauthorized
		}
	}
	return err
}

const (
	commitHistoryQuery = `query GetCommitHistory($componentId: String!, $branch: String!) { commitHistory(componentId: $componentId, branch: $branch) { sha, message, isLatest, author { name, date, email, avatarUrl } } }`
	projectLabelsQuery = `query GetProjectComponentLabels($projectId: String!, $orgId: Int!) { projectComponentLabels(projectId: $projectId, orgId: $orgId) }`
)

// GetComponentRepository fetches repository config from the OpenChoreo component
// spec (spec.workflow.parameters.repository). The projectID parameter is kept for
// interface compatibility but is not used — the componentName path param is sufficient.
func (s *componentService) GetComponentRepository(ctx context.Context, _, componentName string) (*models.ComponentRepository, error) {
	track, err := s.client.GetDeploymentTrack(ctx, componentName)
	if err != nil {
		return nil, fmt.Errorf("get component repository: %w", err)
	}

	repo := &models.ComponentRepository{
		Branch:     track.Branch,
		AppSubPath: track.AppPath,
	}

	// Derive gitProvider, organizationApp, and nameApp from the repository URL.
	// URL format: https://{host}/{org}/{repo}[.git]
	if track.URL != "" {
		repo.ServerURL = track.URL
		parts := splitRepoURL(track.URL)
		if len(parts) >= 3 {
			repo.GitProvider = parts[0]
			repo.OrganizationApp = parts[1]
			repo.NameApp = parts[2]
		}
	}

	repo.TreeURL = buildTreeURL(repo.GitProvider, repo.OrganizationApp, repo.NameApp, repo.Branch, repo.AppSubPath, repo.ServerURL)
	return repo, nil
}

// buildTreeURL constructs a browser-navigable URL to the repository tree at the given branch.
// For GitHub: https://github.com/{org}/{repo}/tree/{branch}[/{subPath}]
// For Bitbucket: https://bitbucket.org/{org}/{repo}/src/HEAD/{subPath}?at={branch}
// For GitLab: https://gitlab.com/{org}/{repo}/-/tree/{branch}[/{subPath}]
func buildTreeURL(provider, org, repo, branch, subPath, rawURL string) string {
	encodedBranch := url.PathEscape(branch)
	switch provider {
	case "github":
		base := "https://github.com/" + org + "/" + repo + "/tree/" + encodedBranch
		if subPath != "" {
			base += "/" + strings.TrimPrefix(subPath, "/")
		}
		return base
	case "bitbucket":
		base := "https://bitbucket.org/" + org + "/" + repo + "/src/HEAD"
		if subPath != "" {
			base += "/" + strings.TrimPrefix(subPath, "/")
		}
		return base + "?at=" + encodedBranch
	case "gitlab":
		base := "https://gitlab.com/" + org + "/" + repo + "/-/tree/" + encodedBranch
		if subPath != "" {
			base += "/" + strings.TrimPrefix(subPath, "/")
		}
		return base
	default:
		// Fall back to the raw clone URL when the provider is unknown.
		return rawURL
	}
}

// splitRepoURL parses a git URL into [gitProvider, org, repo] components.
// Examples:
//
//	https://github.com/myorg/myrepo.git → ["github", "myorg", "myrepo"]
//	https://gitlab.com/myorg/myrepo     → ["gitlab", "myorg", "myrepo"]
func splitRepoURL(rawURL string) []string {
	// Strip scheme
	s := rawURL
	for _, prefix := range []string{"https://", "http://", "git@", "ssh://"} {
		if after, ok := strings.CutPrefix(s, prefix); ok {
			s = after
			break
		}
	}
	// git@github.com:org/repo → github.com/org/repo
	if idx := strings.Index(s, ":"); idx != -1 && !strings.Contains(s[:idx], "/") {
		s = s[:idx] + "/" + s[idx+1:]
	}
	segs := strings.SplitN(s, "/", 3)
	if len(segs) < 3 {
		return nil
	}
	host := strings.ToLower(segs[0])
	org := segs[1]
	repo := strings.TrimSuffix(segs[2], ".git")

	provider := host
	if idx := strings.Index(host, "."); idx != -1 {
		provider = host[:idx]
	}
	return []string{provider, org, repo}
}

func (s *componentService) GetDeploymentTrack(ctx context.Context, componentName string) (*models.DeploymentTrack, error) {
	track, err := s.client.GetDeploymentTrack(ctx, componentName)
	if err != nil {
		return nil, fmt.Errorf("get deployment track: %w", err)
	}
	return track, nil
}

func (s *componentService) GetCommitHistory(ctx context.Context, orgName, projectName, componentName, branch string) (*models.CommitList, error) {
	// Try ICP GraphQL first (uses the component's K8s UID as componentId).
	if s.icpClient.IsAvailable() {
		comp, err := s.client.GetComponent(ctx, componentName)
		if err == nil && comp.UID != "" {
			data, icpErr := s.icpClient.Query(ctx, commitHistoryQuery, map[string]string{
				"componentId": comp.UID,
				"branch":      branch,
			})
			if icpErr == nil {
				raw, ok := data["commitHistory"]
				if ok {
					var items []models.Commit
					if jsonErr := json.Unmarshal(raw, &items); jsonErr == nil {
						if items == nil {
							items = []models.Commit{}
						}
						return &models.CommitList{Items: items}, nil
					}
				}
			}
			slog.WarnContext(ctx, "icp commit history unavailable, falling back to workflow runs",
				"component", componentName, "error", err)
		}
	}

	// Fallback: derive commit history from OpenChoreo workflow runs.
	// Each run records the git-revision (commit SHA) of the source it built.
	runs, err := s.client.ListWorkflowRuns(ctx, orgName, projectName, componentName, 20, "")
	if err != nil {
		return nil, fmt.Errorf("get commit history: %w", err)
	}
	slog.DebugContext(ctx, "commit history fallback: workflow runs", "component", componentName, "total_runs", len(runs.Items))
	items := make([]models.Commit, 0, len(runs.Items))
	for _, run := range runs.Items {
		if run.Commit == "" {
			continue
		}
		items = append(items, models.Commit{
			SHA:      run.Commit,
			IsLatest: len(items) == 0, // first appended item is latest
			Author: models.CommitAuthor{
				Date: run.StartedAt,
			},
		})
	}

	// If no commits found from workflow runs, fall back to the component's
	// deployment track which stores the configured revision commit SHA.
	if len(items) == 0 {
		track, trackErr := s.client.GetDeploymentTrack(ctx, componentName)
		if trackErr == nil && track.CommitSHA != "" {
			items = append(items, models.Commit{
				SHA:      track.CommitSHA,
				IsLatest: true,
			})
		}

		// Final fallback: fetch commit history from GitHub API using the
		// repository URL and branch from the deployment track.
		if len(items) == 0 && s.ghClient != nil && s.ghClient.IsAvailable() &&
			trackErr == nil && track.URL != "" {
			ghBranch := branch
			if ghBranch == "" {
				ghBranch = track.Branch
			}
			if ghBranch != "" {
				ghCommits, ghErr := s.ghClient.GetCommits(ctx, track.URL, ghBranch, 20)
				if ghErr != nil {
					slog.WarnContext(ctx, "github commit history fallback failed",
						"component", componentName, "repo", track.URL, "error", ghErr)
				} else {
					items = ghCommits
				}
			}
		}
	}

	return &models.CommitList{Items: items}, nil
}

func (s *componentService) GetComponentLabels(ctx context.Context, projectID string, orgID int) (*models.LabelList, error) {
	data, err := s.icpClient.Query(ctx, projectLabelsQuery, map[string]interface{}{
		"projectId": projectID,
		"orgId":     orgID,
	})
	if err != nil {
		return nil, fmt.Errorf("get component labels: %w", err)
	}
	raw, ok := data["projectComponentLabels"]
	if !ok {
		return &models.LabelList{Items: []string{}}, nil
	}
	var items []string
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("get component labels: parse: %w", err)
	}
	if items == nil {
		items = []string{}
	}
	return &models.LabelList{Items: items}, nil
}
