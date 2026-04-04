package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/wso2/integration-control-plane/ipaas-service/clients/openchoreo"
	"github.com/wso2/integration-control-plane/ipaas-service/models"
)

type DeploymentService interface {
	GetComponentDeployment(ctx context.Context, orgHandler, orgUUID, componentName, versionID, environment string) (*models.ComponentDeployment, error)
	DeployDeploymentTrack(ctx context.Context, componentName, projectName string, input *models.DeployDeploymentTrackInput) (string, error)
	DeployToEnvironment(ctx context.Context, orgName, projectName, componentName, environment string) (*models.ComponentDeployment, error)
	Promote(ctx context.Context, componentName, projectName string, input *models.PromoteInput) (string, error)
	StopDeployment(ctx context.Context, componentName string, input *models.StopDeploymentInput) (string, error)
}

type deploymentService struct {
	scheduleClient openchoreo.ScheduleClient
}

func NewDeploymentService(scheduleClient openchoreo.ScheduleClient) DeploymentService {
	return &deploymentService{scheduleClient: scheduleClient}
}

// GetComponentDeployment looks up the ReleaseBinding for the component in the
// given environment. Returns nil (no error) when no release binding exists yet.
func (s *deploymentService) GetComponentDeployment(ctx context.Context, _, _, componentName, _, environment string) (*models.ComponentDeployment, error) {
	schedule, err := s.scheduleClient.GetReleaseBinding(ctx, "", componentName, environment)
	if err != nil {
		// A "not found" error means no deployment exists yet — return nil.
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return nil, nil
		}
		return nil, fmt.Errorf("get component deployment: %w", err)
	}

	return &models.ComponentDeployment{
		ReleaseID:    schedule.ReleaseName,
		Cron:         schedule.CronExpression,
		CronTimezone: "UTC",
	}, nil
}

// DeployDeploymentTrack creates or updates a ReleaseBinding for the component
// in the specified environment. A ComponentRelease is auto-generated if needed.
func (s *deploymentService) DeployDeploymentTrack(ctx context.Context, componentName, projectName string, input *models.DeployDeploymentTrackInput) (string, error) {
	req := &models.UpsertScheduleRequest{
		Environment:    input.EnvironmentID,
		CronExpression: derefString(input.Cron),
		State:          "Active",
	}

	// Try update first; if not found, create.
	schedule, err := s.scheduleClient.UpdateReleaseBinding(ctx, "", projectName, componentName, req)
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			schedule, err = s.scheduleClient.CreateReleaseBinding(ctx, "", projectName, componentName, req)
			if err != nil {
				return "", fmt.Errorf("create release binding: %w", err)
			}
		} else {
			return "", fmt.Errorf("update release binding: %w", err)
		}
	}

	return schedule.ReleaseName, nil
}

// DeployToEnvironment generates a release from the latest workload and creates
// or updates a ReleaseBinding for the given environment. Used for auto-deploy
// after a successful build.
func (s *deploymentService) DeployToEnvironment(ctx context.Context, _, projectName, componentName, environment string) (*models.ComponentDeployment, error) {
	releaseName, err := s.scheduleClient.GenerateRelease(ctx, "", componentName)
	if err != nil {
		return nil, fmt.Errorf("generate release: %w", err)
	}
	slog.InfoContext(ctx, "generated release", "component", componentName, "release", releaseName)

	req := &models.UpsertScheduleRequest{
		Environment: environment,
		State:       "Active",
		ReleaseName: releaseName,
	}

	// Try update first; create if not found.
	schedule, err := s.scheduleClient.UpdateReleaseBinding(ctx, "", projectName, componentName, req)
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			schedule, err = s.scheduleClient.CreateReleaseBinding(ctx, "", projectName, componentName, req)
			if err != nil {
				return nil, fmt.Errorf("create release binding: %w", err)
			}
		} else {
			return nil, fmt.Errorf("update release binding: %w", err)
		}
	}

	return &models.ComponentDeployment{
		ReleaseID:    schedule.ReleaseName,
		Cron:         schedule.CronExpression,
		CronTimezone: "UTC",
	}, nil
}

// Promote creates a ReleaseBinding in the target environment using the release
// from the source environment.
func (s *deploymentService) Promote(ctx context.Context, componentName, projectName string, input *models.PromoteInput) (string, error) {
	req := &models.UpsertScheduleRequest{
		Environment: input.TargetEnvironmentID,
		State:       "Active",
		ReleaseName: input.SourceReleaseID,
	}

	schedule, err := s.scheduleClient.CreateReleaseBinding(ctx, "", projectName, componentName, req)
	if err != nil {
		if strings.Contains(err.Error(), "409") || strings.Contains(err.Error(), "already exists") {
			schedule, err = s.scheduleClient.UpdateReleaseBinding(ctx, "", projectName, componentName, req)
			if err != nil {
				return "", fmt.Errorf("update release binding for promote: %w", err)
			}
		} else {
			return "", fmt.Errorf("create release binding for promote: %w", err)
		}
	}

	return schedule.ReleaseName, nil
}

// StopDeployment deletes the ReleaseBinding for the component in the given environment.
func (s *deploymentService) StopDeployment(ctx context.Context, componentName string, input *models.StopDeploymentInput) (string, error) {
	if err := s.scheduleClient.DeleteReleaseBinding(ctx, "", componentName, input.Environment); err != nil {
		return "", fmt.Errorf("stop deployment: %w", err)
	}
	return "stopped", nil
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
