package services

import (
	"context"
	"errors"
	"fmt"
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
	client openchoreo.ScheduleClient
}

func NewScheduleService(client openchoreo.ScheduleClient) ScheduleService {
	return &scheduleService{client: client}
}

func (s *scheduleService) ListSchedules(ctx context.Context, orgName, projectName, componentName string) (*models.ScheduleList, error) {
	list, err := s.client.ListReleaseBindings(ctx, orgName, projectName, componentName)
	if err != nil {
		return nil, translateScheduleHTTPError(err)
	}
	return list, nil
}

func (s *scheduleService) GetSchedule(ctx context.Context, orgName, projectName, componentName, environment string) (*models.Schedule, error) {
	schedule, err := s.client.GetReleaseBinding(ctx, orgName, componentName, environment)
	if err != nil {
		return nil, translateScheduleHTTPError(err)
	}
	return schedule, nil
}

// UpsertSchedule creates the ReleaseBinding if it does not exist, otherwise updates it.
func (s *scheduleService) UpsertSchedule(ctx context.Context, orgName, projectName, componentName string, req *models.UpsertScheduleRequest) (*models.Schedule, error) {
	_, err := s.client.GetReleaseBinding(ctx, orgName, componentName, req.Environment)
	if err != nil {
		var httpErr *requests.HttpError
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			schedule, createErr := s.client.CreateReleaseBinding(ctx, orgName, projectName, componentName, req)
			if createErr != nil {
				return nil, translateScheduleHTTPError(createErr)
			}
			return schedule, nil
		}
		return nil, translateScheduleHTTPError(err)
	}

	schedule, err := s.client.UpdateReleaseBinding(ctx, orgName, projectName, componentName, req)
	if err != nil {
		return nil, translateScheduleHTTPError(err)
	}
	return schedule, nil
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
