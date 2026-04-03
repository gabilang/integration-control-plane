package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wso2/integration-control-plane/ipaas-service/clients/icp"
	"github.com/wso2/integration-control-plane/ipaas-service/models"
)

const (
	listRuntimesQuery        = `query GetRuntimes($environmentId: String!, $projectId: String!, $componentId: String!) { runtimes(environmentId: $environmentId, projectId: $projectId, componentId: $componentId) { runtimeId, runtimeType, status, version, platformName, platformVersion, platformHome, osName, osVersion, registrationTime, lastHeartbeat } }`
	listProjectRuntimesQuery = `query GetProjectRuntimes($environmentId: String!, $projectId: String!) { runtimes(environmentId: $environmentId, projectId: $projectId) { runtimeId, runtimeType, status, version, platformName, platformVersion, platformHome, osName, osVersion, registrationTime, lastHeartbeat, component { displayName } } }`
	deleteRuntimeMutation    = `mutation DeleteRuntime($runtimeId: String!) { deleteRuntime(runtimeId: $runtimeId) }`
)

type RuntimeService interface {
	ListRuntimes(ctx context.Context, envID, projectID, componentID string) (*models.RuntimeList, error)
	ListProjectRuntimes(ctx context.Context, envID, projectID string) (*models.RuntimeList, error)
	DeleteRuntime(ctx context.Context, runtimeID string) error
}

type runtimeService struct {
	icpClient *icp.Client
}

func NewRuntimeService(icpClient *icp.Client) RuntimeService {
	return &runtimeService{icpClient: icpClient}
}

func (s *runtimeService) ListRuntimes(ctx context.Context, envID, projectID, componentID string) (*models.RuntimeList, error) {
	data, err := s.icpClient.Query(ctx, listRuntimesQuery, map[string]string{
		"environmentId": envID,
		"projectId":     projectID,
		"componentId":   componentID,
	})
	if err != nil {
		return nil, fmt.Errorf("list runtimes: %w", err)
	}

	raw, ok := data["runtimes"]
	if !ok {
		return &models.RuntimeList{Items: []models.Runtime{}}, nil
	}

	var items []models.Runtime
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("list runtimes: parse items: %w", err)
	}
	if items == nil {
		items = []models.Runtime{}
	}
	return &models.RuntimeList{Items: items}, nil
}

func (s *runtimeService) ListProjectRuntimes(ctx context.Context, envID, projectID string) (*models.RuntimeList, error) {
	data, err := s.icpClient.Query(ctx, listProjectRuntimesQuery, map[string]string{
		"environmentId": envID,
		"projectId":     projectID,
	})
	if err != nil {
		return nil, fmt.Errorf("list project runtimes: %w", err)
	}

	raw, ok := data["runtimes"]
	if !ok {
		return &models.RuntimeList{Items: []models.Runtime{}}, nil
	}

	var items []models.Runtime
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("list project runtimes: parse items: %w", err)
	}
	if items == nil {
		items = []models.Runtime{}
	}
	return &models.RuntimeList{Items: items}, nil
}

func (s *runtimeService) DeleteRuntime(ctx context.Context, runtimeID string) error {
	_, err := s.icpClient.Query(ctx, deleteRuntimeMutation, map[string]string{
		"runtimeId": runtimeID,
	})
	if err != nil {
		return fmt.Errorf("delete runtime: %w", err)
	}
	return nil
}
