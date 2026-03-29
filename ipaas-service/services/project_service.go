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

var (
	ErrProjectNotFound = errors.New("project not found")
	ErrUnauthorized    = errors.New("unauthorized")
)

// ProjectService handles business logic for project operations.
type ProjectService interface {
	ListProjects(ctx context.Context, orgName string, limit int, cursor string) (*models.ProjectList, error)
	GetProject(ctx context.Context, orgName, projectName string) (*models.Project, error)
	CreateProject(ctx context.Context, orgName string, req *models.CreateProjectRequest) (*models.Project, error)
	UpdateProject(ctx context.Context, orgName, projectName string, req *models.UpdateProjectRequest) (*models.Project, error)
	DeleteProject(ctx context.Context, orgName, projectName string) error
}

type projectService struct {
	client openchoreo.ProjectClient
}

func NewProjectService(client openchoreo.ProjectClient) ProjectService {
	return &projectService{client: client}
}

func (s *projectService) ListProjects(ctx context.Context, orgName string, limit int, cursor string) (*models.ProjectList, error) {
	list, err := s.client.ListProjects(ctx, orgName, limit, cursor)
	if err != nil {
		return nil, translateHTTPError(err)
	}
	return list, nil
}

func (s *projectService) GetProject(ctx context.Context, orgName, projectName string) (*models.Project, error) {
	project, err := s.client.GetProject(ctx, orgName, projectName)
	if err != nil {
		return nil, translateHTTPError(err)
	}
	return project, nil
}

func (s *projectService) CreateProject(ctx context.Context, orgName string, req *models.CreateProjectRequest) (*models.Project, error) {
	project, err := s.client.CreateProject(ctx, orgName, req)
	if err != nil {
		return nil, translateHTTPError(err)
	}
	return project, nil
}

func (s *projectService) UpdateProject(ctx context.Context, orgName, projectName string, req *models.UpdateProjectRequest) (*models.Project, error) {
	project, err := s.client.UpdateProject(ctx, orgName, projectName, req)
	if err != nil {
		return nil, translateHTTPError(err)
	}
	return project, nil
}

func (s *projectService) DeleteProject(ctx context.Context, orgName, projectName string) error {
	return translateHTTPError(s.client.DeleteProject(ctx, orgName, projectName))
}

func translateHTTPError(err error) error {
	if err == nil {
		return nil
	}
	var httpErr *requests.HttpError
	if errors.As(err, &httpErr) {
		switch httpErr.StatusCode {
		case http.StatusNotFound:
			return fmt.Errorf("%w: %s", ErrProjectNotFound, httpErr.Body)
		case http.StatusUnauthorized:
			return ErrUnauthorized
		}
	}
	return err
}
