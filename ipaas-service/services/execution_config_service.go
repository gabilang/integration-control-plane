package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wso2/integration-control-plane/ipaas-service/clients/icp"
	"github.com/wso2/integration-control-plane/ipaas-service/models"
)

const (
	getExecutionConfigsQuery   = `query GetExecutionConfigs($componentId: String!, $releaseId: String!) { executionConfigs(componentId: $componentId, releaseId: $releaseId) { cronjobFrequency, cronjobTimezone, cronjobAllowConcurrency, timeoutSeconds, retryCount } }`
	getExecutionArgumentsQuery = `query GetExecutionArguments($id: String!, $componentId: String!, $releaseId: String!) { execution(input: { id: $id, componentId: $componentId, releaseId: $releaseId }) { arguments { argumentName, argumentValue } } }`
	updateJobConfigsMutation   = `mutation UpdateJobConfigs($input: JobConfigInput!) { updateJobConfigs(input: $input) }`
)

type ExecutionConfigService interface {
	GetExecutionConfigs(ctx context.Context, componentID, releaseID string) (*models.ExecutionConfigs, error)
	GetExecutionArguments(ctx context.Context, runID, componentID, releaseID string) ([]models.ExecutionArgument, error)
	UpdateJobConfigs(ctx context.Context, input *models.UpdateJobConfigsInput) (bool, error)
}

type executionConfigService struct {
	icpClient *icp.Client
}

func NewExecutionConfigService(icpClient *icp.Client) ExecutionConfigService {
	return &executionConfigService{icpClient: icpClient}
}

func (s *executionConfigService) GetExecutionConfigs(ctx context.Context, componentID, releaseID string) (*models.ExecutionConfigs, error) {
	data, err := s.icpClient.Query(ctx, getExecutionConfigsQuery, map[string]string{
		"componentId": componentID,
		"releaseId":   releaseID,
	})
	if err != nil {
		return nil, fmt.Errorf("get execution configs: %w", err)
	}

	raw, ok := data["executionConfigs"]
	if !ok {
		return &models.ExecutionConfigs{}, nil
	}

	var result models.ExecutionConfigs
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("get execution configs: parse: %w", err)
	}
	return &result, nil
}

type executionResponse struct {
	Arguments []models.ExecutionArgument `json:"arguments"`
}

func (s *executionConfigService) GetExecutionArguments(ctx context.Context, runID, componentID, releaseID string) ([]models.ExecutionArgument, error) {
	data, err := s.icpClient.Query(ctx, getExecutionArgumentsQuery, map[string]string{
		"id":          runID,
		"componentId": componentID,
		"releaseId":   releaseID,
	})
	if err != nil {
		return nil, fmt.Errorf("get execution arguments: %w", err)
	}

	raw, ok := data["execution"]
	if !ok {
		return []models.ExecutionArgument{}, nil
	}

	var execResp executionResponse
	if err := json.Unmarshal(raw, &execResp); err != nil {
		return nil, fmt.Errorf("get execution arguments: parse: %w", err)
	}
	if execResp.Arguments == nil {
		return []models.ExecutionArgument{}, nil
	}
	return execResp.Arguments, nil
}

func (s *executionConfigService) UpdateJobConfigs(ctx context.Context, input *models.UpdateJobConfigsInput) (bool, error) {
	_, err := s.icpClient.Query(ctx, updateJobConfigsMutation, map[string]interface{}{
		"input": input,
	})
	if err != nil {
		return false, fmt.Errorf("update job configs: %w", err)
	}
	return true, nil
}
