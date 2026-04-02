package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/clients/openchoreo"
	"github.com/wso2/integration-control-plane/ipaas-service/clients/requests"
	"github.com/wso2/integration-control-plane/ipaas-service/models"
)

var ErrScheduleNotFound = errors.New("schedule not found")

// ScheduleService handles business logic for scheduled-task deployment operations.
type ScheduleService interface {
	ListSchedules(ctx context.Context, orgName, projectName, componentName string) (*models.ScheduleList, error)
	UpsertSchedule(ctx context.Context, orgName, projectName, componentName string, req *models.UpsertScheduleRequest) (*models.Schedule, error)
	GetSchedule(ctx context.Context, orgName, projectName, componentName, environment string) (*models.Schedule, error)
	DeleteSchedule(ctx context.Context, orgName, projectName, componentName, environment string) error
}

type scheduleService struct {
	client          openchoreo.ScheduleClient
	componentClient openchoreo.ComponentClient
}

func NewScheduleService(client openchoreo.ScheduleClient, componentClient openchoreo.ComponentClient) ScheduleService {
	return &scheduleService{client: client, componentClient: componentClient}
}

func (s *scheduleService) ListSchedules(ctx context.Context, orgName, projectName, componentName string) (*models.ScheduleList, error) {
	list, err := s.client.ListReleaseBindings(ctx, orgName, projectName, componentName)
	if err != nil {
		return nil, translateScheduleHTTPError(err)
	}
	params, err := s.componentClient.GetComponentParameters(ctx, componentName)
	if err != nil {
		slog.WarnContext(ctx, "failed to get component parameters", "error", err, "component", componentName)
	}
	if params != nil {
		for i := range list.Items {
			list.Items[i].BackoffLimit = params.BackoffLimit
			list.Items[i].ActiveDeadlineSeconds = params.ActiveDeadlineSeconds
		}
	}
	return list, nil
}

func (s *scheduleService) GetSchedule(ctx context.Context, orgName, projectName, componentName, environment string) (*models.Schedule, error) {
	schedule, err := s.client.GetReleaseBinding(ctx, orgName, componentName, environment)
	if err != nil {
		return nil, translateScheduleHTTPError(err)
	}
	params, err := s.componentClient.GetComponentParameters(ctx, componentName)
	if err != nil {
		slog.WarnContext(ctx, "failed to get component parameters", "error", err, "component", componentName)
	}
	if params != nil {
		schedule.BackoffLimit = params.BackoffLimit
		schedule.ActiveDeadlineSeconds = params.ActiveDeadlineSeconds
	}
	return schedule, nil
}

// UpsertSchedule creates the ReleaseBinding if it does not exist, otherwise updates it.
// If BackoffLimit or ActiveDeadlineSeconds are set, the Component parameters are patched too
// (these are component-wide settings, not per-environment).
func (s *scheduleService) UpsertSchedule(ctx context.Context, orgName, projectName, componentName string, req *models.UpsertScheduleRequest) (*models.Schedule, error) {
	existing, err := s.client.GetReleaseBinding(ctx, orgName, componentName, req.Environment)
	if err != nil {
		var httpErr *requests.HttpError
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			schedule, createErr := s.client.CreateReleaseBinding(ctx, orgName, projectName, componentName, req)
			if createErr != nil {
				return nil, translateScheduleHTTPError(createErr)
			}
			s.patchComponentParametersIfSet(ctx, componentName, req)
			return schedule, nil
		}
		return nil, translateScheduleHTTPError(err)
	}

	// Preserve the existing releaseName so the PUT body doesn't clear it.
	if req.ReleaseName == "" && existing.ReleaseName != "" {
		req.ReleaseName = existing.ReleaseName
	}
	schedule, err := s.client.UpdateReleaseBinding(ctx, orgName, projectName, componentName, req)
	if err != nil {
		return nil, translateScheduleHTTPError(err)
	}
	s.patchComponentParametersIfSet(ctx, componentName, req)
	return schedule, nil
}

func (s *scheduleService) patchComponentParametersIfSet(ctx context.Context, componentName string, req *models.UpsertScheduleRequest) {
	if req.BackoffLimit == nil && req.ActiveDeadlineSeconds == nil {
		return
	}
	params := &openchoreo.ComponentParameters{
		BackoffLimit:          req.BackoffLimit,
		ActiveDeadlineSeconds: req.ActiveDeadlineSeconds,
	}
	if err := s.componentClient.PatchComponentParameters(ctx, componentName, params); err != nil {
		slog.WarnContext(ctx, "failed to patch component parameters", "error", err, "component", componentName)
	}
}

func (s *scheduleService) DeleteSchedule(ctx context.Context, orgName, projectName, componentName, environment string) error {
	return translateScheduleHTTPError(s.client.DeleteReleaseBinding(ctx, orgName, componentName, environment))
}

func translateScheduleHTTPError(err error) error {
	if err == nil {
		return nil
	}
	var httpErr *requests.HttpError
	if errors.As(err, &httpErr) {
		switch httpErr.StatusCode {
		case http.StatusNotFound:
			return fmt.Errorf("%w: %s", ErrScheduleNotFound, httpErr.Body)
		case http.StatusUnauthorized:
			return ErrUnauthorized
		}
	}
	return err
}
