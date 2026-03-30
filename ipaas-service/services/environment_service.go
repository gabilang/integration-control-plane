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

var ErrEnvironmentNotFound = errors.New("environment not found")

// EnvironmentService handles business logic for environment operations.
type EnvironmentService interface {
	ListEnvironments(ctx context.Context, orgName string) (*models.EnvironmentList, error)
	GetEnvironment(ctx context.Context, orgName, environmentName string) (*models.Environment, error)
}

type environmentService struct {
	client openchoreo.EnvironmentClient
}

func NewEnvironmentService(client openchoreo.EnvironmentClient) EnvironmentService {
	return &environmentService{client: client}
}

func (s *environmentService) ListEnvironments(ctx context.Context, orgName string) (*models.EnvironmentList, error) {
	list, err := s.client.ListEnvironments(ctx, orgName)
	if err != nil {
		return nil, translateEnvironmentError(err)
	}
	return list, nil
}

func (s *environmentService) GetEnvironment(ctx context.Context, orgName, environmentName string) (*models.Environment, error) {
	env, err := s.client.GetEnvironment(ctx, orgName, environmentName)
	if err != nil {
		return nil, translateEnvironmentError(err)
	}
	return env, nil
}

func translateEnvironmentError(err error) error {
	if err == nil {
		return nil
	}
	var httpErr *requests.HttpError
	if errors.As(err, &httpErr) {
		switch httpErr.StatusCode {
		case http.StatusNotFound:
			return fmt.Errorf("%w: %s", ErrEnvironmentNotFound, httpErr.Body)
		case http.StatusUnauthorized:
			return ErrUnauthorized
		}
	}
	return err
}
