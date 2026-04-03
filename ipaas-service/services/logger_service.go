package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wso2/integration-control-plane/ipaas-service/clients/icp"
	"github.com/wso2/integration-control-plane/ipaas-service/models"
)

const (
	listLoggersQuery       = `query GetLoggers($environmentId: String!, $componentId: String!) { loggersByEnvironmentAndComponent(environmentId: $environmentId, componentId: $componentId) { componentName, logLevel, runtimeIds } }`
	updateLogLevelMutation = `mutation UpdateLogLevel($input: UpdateLogLevelInput!) { updateLogLevel(input: $input) { success, message, commandIds } }`
)

type LoggerService interface {
	ListLoggers(ctx context.Context, environmentID, componentID string) (*models.LoggerList, error)
	UpdateLogLevel(ctx context.Context, input *models.UpdateLogLevelInput) (*models.UpdateLogLevelResult, error)
}

type loggerService struct {
	icpClient *icp.Client
}

func NewLoggerService(icpClient *icp.Client) LoggerService {
	return &loggerService{icpClient: icpClient}
}

func (s *loggerService) ListLoggers(ctx context.Context, environmentID, componentID string) (*models.LoggerList, error) {
	data, err := s.icpClient.Query(ctx, listLoggersQuery, map[string]string{
		"environmentId": environmentID,
		"componentId":   componentID,
	})
	if err != nil {
		return nil, fmt.Errorf("list loggers: %w", err)
	}

	raw, ok := data["loggersByEnvironmentAndComponent"]
	if !ok {
		return &models.LoggerList{Items: []models.Logger{}}, nil
	}

	var items []models.Logger
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("list loggers: parse: %w", err)
	}
	if items == nil {
		items = []models.Logger{}
	}
	return &models.LoggerList{Items: items}, nil
}

func (s *loggerService) UpdateLogLevel(ctx context.Context, input *models.UpdateLogLevelInput) (*models.UpdateLogLevelResult, error) {
	data, err := s.icpClient.Query(ctx, updateLogLevelMutation, map[string]interface{}{
		"input": input,
	})
	if err != nil {
		return nil, fmt.Errorf("update log level: %w", err)
	}

	raw, ok := data["updateLogLevel"]
	if !ok {
		return &models.UpdateLogLevelResult{Success: true}, nil
	}

	var result models.UpdateLogLevelResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("update log level: parse: %w", err)
	}
	return &result, nil
}
