package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

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
	GetCommitHistory(ctx context.Context, componentID, branch string) (*models.CommitList, error)
	GetComponentLabels(ctx context.Context, projectID string, orgID int) (*models.LabelList, error)
}

type componentService struct {
	client       openchoreo.ComponentClient
	observClient observability.Client
	icpClient    *icp.Client
}

func NewComponentService(client openchoreo.ComponentClient, observClient observability.Client, icpClient *icp.Client) ComponentService {
	return &componentService{client: client, observClient: observClient, icpClient: icpClient}
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
	componentRepositoryQuery = `query GetComponentRepository($projectId: String!, $componentHandler: String!) { component(projectId: $projectId, componentHandler: $componentHandler) { repository { gitProvider, organizationApp, nameApp, branch, appSubPath, bitbucketServerUrl, serverUrl, projectApp } } }`
	commitHistoryQuery       = `query GetCommitHistory($componentId: String!, $branch: String!) { commitHistory(componentId: $componentId, branch: $branch) { sha, message, isLatest, author { name, date, email, avatarUrl } } }`
	projectLabelsQuery       = `query GetProjectComponentLabels($projectId: String!, $orgId: Int!) { projectComponentLabels(projectId: $projectId, orgId: $orgId) }`
)

type componentQueryResp struct {
	Repository *models.ComponentRepository `json:"repository"`
}

func (s *componentService) GetComponentRepository(ctx context.Context, projectID, componentHandler string) (*models.ComponentRepository, error) {
	data, err := s.icpClient.Query(ctx, componentRepositoryQuery, map[string]string{
		"projectId":        projectID,
		"componentHandler": componentHandler,
	})
	if err != nil {
		return nil, fmt.Errorf("get component repository: %w", err)
	}
	raw, ok := data["component"]
	if !ok {
		return nil, fmt.Errorf("get component repository: missing field")
	}
	var resp componentQueryResp
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("get component repository: parse: %w", err)
	}
	if resp.Repository == nil {
		return &models.ComponentRepository{}, nil
	}
	return resp.Repository, nil
}

func (s *componentService) GetCommitHistory(ctx context.Context, componentID, branch string) (*models.CommitList, error) {
	data, err := s.icpClient.Query(ctx, commitHistoryQuery, map[string]string{
		"componentId": componentID,
		"branch":      branch,
	})
	if err != nil {
		return nil, fmt.Errorf("get commit history: %w", err)
	}
	raw, ok := data["commitHistory"]
	if !ok {
		return &models.CommitList{Items: []models.Commit{}}, nil
	}
	var items []models.Commit
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("get commit history: parse: %w", err)
	}
	if items == nil {
		items = []models.Commit{}
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
