package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/wso2/integration-control-plane/ipaas-service/clients/k8s"
	"github.com/wso2/integration-control-plane/ipaas-service/clients/openchoreo"
	"github.com/wso2/integration-control-plane/ipaas-service/clients/requests"
	"github.com/wso2/integration-control-plane/ipaas-service/models"
)

// ExecutionService handles business logic for CronJob execution history and manual triggers.
type ExecutionService interface {
	ListExecutions(ctx context.Context, orgName, projectName, componentName, environment string) (*models.ExecutionList, error)
	TriggerExecution(ctx context.Context, orgName, projectName, componentName, environment string) (*models.Execution, error)
}

type executionService struct {
	scheduleClient openchoreo.ScheduleClient
	jobsClient     k8s.JobsClient
}

func NewExecutionService(scheduleClient openchoreo.ScheduleClient, jobsClient k8s.JobsClient) ExecutionService {
	return &executionService{scheduleClient: scheduleClient, jobsClient: jobsClient}
}

func (s *executionService) ListExecutions(ctx context.Context, orgName, _, componentName, environment string) (*models.ExecutionList, error) {
	orgNs, err := s.scheduleClient.GetReleaseBindingNamespace(ctx, orgName, componentName, environment)
	if err != nil {
		var httpErr *requests.HttpError
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("%w", ErrScheduleNotFound)
		}
		return nil, translateScheduleHTTPError(err)
	}

	labelSelector := fmt.Sprintf(
		"openchoreo.dev/component=%s,openchoreo.dev/environment=%s,openchoreo.dev/namespace=%s",
		componentName, environment, orgNs,
	)
	list, err := s.jobsClient.ListJobs(ctx, labelSelector)
	if err != nil {
		return nil, fmt.Errorf("list executions: %w", err)
	}
	return list, nil
}

func (s *executionService) TriggerExecution(ctx context.Context, orgName, _, componentName, environment string) (*models.Execution, error) {
	orgNs, err := s.scheduleClient.GetReleaseBindingNamespace(ctx, orgName, componentName, environment)
	if err != nil {
		var httpErr *requests.HttpError
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("%w", ErrScheduleNotFound)
		}
		return nil, translateScheduleHTTPError(err)
	}

	labelSelector := fmt.Sprintf(
		"openchoreo.dev/component=%s,openchoreo.dev/environment=%s,openchoreo.dev/namespace=%s",
		componentName, environment, orgNs,
	)
	cronjob, err := s.jobsClient.GetCronJob(ctx, labelSelector)
	if err != nil {
		return nil, fmt.Errorf("trigger execution: %w", err)
	}
	return s.jobsClient.TriggerJob(ctx, cronjob)
}
