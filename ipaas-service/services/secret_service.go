package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wso2/integration-control-plane/ipaas-service/clients/icp"
)

const (
	generateJwtSecretMutation = `mutation GenerateComponentEnvironmentJwtSecret($componentId: String!, $environmentId: String!) { generateComponentEnvironmentJwtSecret(componentId: $componentId, environmentId: $environmentId) }`
	rotateJwtSecretMutation   = `mutation RotateComponentEnvironmentJwtSecret($componentId: String!, $environmentId: String!) { rotateComponentEnvironmentJwtSecret(componentId: $componentId, environmentId: $environmentId) }`
)

type SecretService interface {
	GenerateJwtSecret(ctx context.Context, componentID, environmentID string) (string, error)
	RotateJwtSecret(ctx context.Context, componentID, environmentID string) (string, error)
}

type secretService struct {
	icpClient *icp.Client
}

func NewSecretService(icpClient *icp.Client) SecretService {
	return &secretService{icpClient: icpClient}
}

func (s *secretService) GenerateJwtSecret(ctx context.Context, componentID, environmentID string) (string, error) {
	data, err := s.icpClient.Query(ctx, generateJwtSecretMutation, map[string]string{
		"componentId":   componentID,
		"environmentId": environmentID,
	})
	if err != nil {
		return "", fmt.Errorf("generate jwt secret: %w", err)
	}

	raw, ok := data["generateComponentEnvironmentJwtSecret"]
	if !ok {
		return "", nil
	}

	var result string
	if err := json.Unmarshal(raw, &result); err != nil {
		return string(raw), nil
	}
	return result, nil
}

func (s *secretService) RotateJwtSecret(ctx context.Context, componentID, environmentID string) (string, error) {
	data, err := s.icpClient.Query(ctx, rotateJwtSecretMutation, map[string]string{
		"componentId":   componentID,
		"environmentId": environmentID,
	})
	if err != nil {
		return "", fmt.Errorf("rotate jwt secret: %w", err)
	}

	raw, ok := data["rotateComponentEnvironmentJwtSecret"]
	if !ok {
		return "", nil
	}

	var result string
	if err := json.Unmarshal(raw, &result); err != nil {
		return string(raw), nil
	}
	return result, nil
}
