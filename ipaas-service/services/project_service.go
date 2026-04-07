package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sort"

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
	GetProjectContributors(ctx context.Context, orgName, projectName string) (*models.ContributorList, error)
}

type projectService struct {
	client    openchoreo.ProjectClient
	compSvc   ComponentService
}

func NewProjectService(client openchoreo.ProjectClient, compSvc ComponentService) ProjectService {
	return &projectService{client: client, compSvc: compSvc}
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

func (s *projectService) GetProjectContributors(ctx context.Context, orgName, projectName string) (*models.ContributorList, error) {
	// List all components in the project.
	compList, err := s.compSvc.ListComponents(ctx, orgName, projectName, 100, "")
	if err != nil {
		return nil, fmt.Errorf("list components for contributors: %w", err)
	}

	// Aggregate commit authors across all components.
	type entry struct {
		name      string
		avatarURL string
		count     int
	}
	byEmail := make(map[string]*entry)

	for _, comp := range compList.Items {
		commits, err := s.compSvc.GetCommitHistory(ctx, orgName, projectName, comp.Name, "")
		if err != nil {
			slog.WarnContext(ctx, "skip component for contributors", "component", comp.Name, "error", err)
			continue
		}
		for _, c := range commits.Items {
			email := c.Author.Email
			if email == "" {
				continue
			}
			if e, ok := byEmail[email]; ok {
				e.count++
				if e.name == "" && c.Author.Name != "" {
					e.name = c.Author.Name
				}
				if e.avatarURL == "" && c.Author.AvatarURL != "" {
					e.avatarURL = c.Author.AvatarURL
				}
			} else {
				byEmail[email] = &entry{name: c.Author.Name, avatarURL: c.Author.AvatarURL, count: 1}
			}
		}
	}

	contributors := make([]models.Contributor, 0, len(byEmail))
	for email, e := range byEmail {
		contributors = append(contributors, models.Contributor{
			DisplayName:        e.name,
			Email:              email,
			AvatarURL:          e.avatarURL,
			TotalContributions: e.count,
		})
	}

	// Sort by contribution count descending.
	sort.Slice(contributors, func(i, j int) bool {
		return contributors[i].TotalContributions > contributors[j].TotalContributions
	})

	return &models.ContributorList{Items: contributors}, nil
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
