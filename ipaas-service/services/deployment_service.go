package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wso2/integration-control-plane/ipaas-service/clients/icp"
	"github.com/wso2/integration-control-plane/ipaas-service/models"
)

const (
	getComponentDeploymentQuery   = `query GetComponentDeployment($orgHandler: String!, $orgUuid: String!, $componentId: String!, $versionId: String!, $environmentId: String!) { componentDeployment(orgHandler: $orgHandler, orgUuid: $orgUuid, componentId: $componentId, versionId: $versionId, environmentId: $environmentId) { releaseId, cron, cronTimezone, build { buildId } } }`
	getDeploymentStatusQuery      = `query GetDeploymentStatus($versionId: String!, $componentId: String!) { deploymentStatusByVersion(versionId: $versionId, componentId: $componentId) { id, sha, started_at, completed_at, status, conclusion, conclusionV2, isAutoDeploy, name, failureReason, sourceCommitId, buildRef } }`
	deployDeploymentTrackMutation = `mutation deployDeploymentTrack($input: DeployDeploymentTrackInput!) { deployDeploymentTrack(input: $input) }`
	promoteMutation               = `mutation promote($componentId: String!, $promoteSchema: Promote!) { promote(componentId: $componentId, promoteSchema: $promoteSchema) }`
	stopDeploymentMutation        = `mutation StopDeployment($orgHandler: String!, $componentId: String!, $releaseId: String!, $type: String!, $clearCron: Boolean!) { stopDeployment(orgHandler: $orgHandler, componentId: $componentId, releaseId: $releaseId, type: $type, clearCron: $clearCron) }`
)

type DeploymentService interface {
	GetComponentDeployment(ctx context.Context, orgHandler, orgUUID, componentID, versionID, environmentID string) (*models.ComponentDeployment, error)
	GetDeploymentStatus(ctx context.Context, componentID, versionID string) ([]models.DeploymentStatus, error)
	DeployDeploymentTrack(ctx context.Context, input *models.DeployDeploymentTrackInput) (string, error)
	Promote(ctx context.Context, componentID string, input *models.PromoteInput) (string, error)
	StopDeployment(ctx context.Context, input *models.StopDeploymentInput) (string, error)
}

type deploymentService struct {
	icpClient *icp.Client
}

func NewDeploymentService(icpClient *icp.Client) DeploymentService {
	return &deploymentService{icpClient: icpClient}
}

func (s *deploymentService) GetComponentDeployment(ctx context.Context, orgHandler, orgUUID, componentID, versionID, environmentID string) (*models.ComponentDeployment, error) {
	data, err := s.icpClient.Query(ctx, getComponentDeploymentQuery, map[string]string{
		"orgHandler":    orgHandler,
		"orgUuid":       orgUUID,
		"componentId":   componentID,
		"versionId":     versionID,
		"environmentId": environmentID,
	})
	if err != nil {
		return nil, fmt.Errorf("get component deployment: %w", err)
	}

	raw, ok := data["componentDeployment"]
	if !ok {
		return nil, fmt.Errorf("get component deployment: missing field")
	}

	var result models.ComponentDeployment
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("get component deployment: parse: %w", err)
	}
	return &result, nil
}

func (s *deploymentService) GetDeploymentStatus(ctx context.Context, componentID, versionID string) ([]models.DeploymentStatus, error) {
	data, err := s.icpClient.Query(ctx, getDeploymentStatusQuery, map[string]string{
		"componentId": componentID,
		"versionId":   versionID,
	})
	if err != nil {
		return nil, fmt.Errorf("get deployment status: %w", err)
	}

	raw, ok := data["deploymentStatusByVersion"]
	if !ok {
		return []models.DeploymentStatus{}, nil
	}

	var items []models.DeploymentStatus
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("get deployment status: parse: %w", err)
	}
	if items == nil {
		items = []models.DeploymentStatus{}
	}
	return items, nil
}

func (s *deploymentService) DeployDeploymentTrack(ctx context.Context, input *models.DeployDeploymentTrackInput) (string, error) {
	data, err := s.icpClient.Query(ctx, deployDeploymentTrackMutation, map[string]interface{}{
		"input": input,
	})
	if err != nil {
		return "", fmt.Errorf("deploy deployment track: %w", err)
	}

	raw, ok := data["deployDeploymentTrack"]
	if !ok {
		return "", nil
	}

	var result string
	if err := json.Unmarshal(raw, &result); err != nil {
		return string(raw), nil
	}
	return result, nil
}

func (s *deploymentService) Promote(ctx context.Context, componentID string, input *models.PromoteInput) (string, error) {
	data, err := s.icpClient.Query(ctx, promoteMutation, map[string]interface{}{
		"componentId":   componentID,
		"promoteSchema": input,
	})
	if err != nil {
		return "", fmt.Errorf("promote: %w", err)
	}

	raw, ok := data["promote"]
	if !ok {
		return "", nil
	}

	var result string
	if err := json.Unmarshal(raw, &result); err != nil {
		return string(raw), nil
	}
	return result, nil
}

func (s *deploymentService) StopDeployment(ctx context.Context, input *models.StopDeploymentInput) (string, error) {
	data, err := s.icpClient.Query(ctx, stopDeploymentMutation, map[string]interface{}{
		"orgHandler":  input.OrgHandler,
		"componentId": input.ComponentID,
		"releaseId":   input.ReleaseID,
		"type":        "scheduledTask",
		"clearCron":   true,
	})
	if err != nil {
		return "", fmt.Errorf("stop deployment: %w", err)
	}

	raw, ok := data["stopDeployment"]
	if !ok {
		return "", nil
	}

	var result string
	if err := json.Unmarshal(raw, &result); err != nil {
		return string(raw), nil
	}
	return result, nil
}
