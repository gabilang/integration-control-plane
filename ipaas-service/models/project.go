package models

type Project struct {
	UID                string `json:"uid,omitempty"`
	Name               string `json:"name"`
	NamespaceName      string `json:"namespaceName,omitempty"`
	DisplayName        string `json:"displayName,omitempty"`
	Description        string `json:"description,omitempty"`
	DeploymentPipeline string `json:"deploymentPipeline,omitempty"`
	CreatedAt          string `json:"createdAt,omitempty"`
	Status             string `json:"status,omitempty"`
}

type ProjectList struct {
	Items      []Project `json:"items"`
	TotalCount int       `json:"totalCount,omitempty"`
	Page       int       `json:"page,omitempty"`
	PageSize   int       `json:"pageSize,omitempty"`
}

type CreateProjectRequest struct {
	Name               string `json:"name"`
	DisplayName        string `json:"displayName,omitempty"`
	Description        string `json:"description,omitempty"`
	DeploymentPipeline string `json:"deploymentPipeline"`
}

type UpdateProjectRequest struct {
	DisplayName        string `json:"displayName,omitempty"`
	Description        string `json:"description,omitempty"`
	DeploymentPipeline string `json:"deploymentPipeline,omitempty"`
}

// Contributor represents a user who has contributed commits to a project.
type Contributor struct {
	DisplayName        string `json:"displayName"`
	Email              string `json:"email"`
	AvatarURL          string `json:"avatarUrl,omitempty"`
	TotalContributions int    `json:"totalContributions"`
}

// ContributorList is the response for project contributors.
type ContributorList struct {
	Items []Contributor `json:"items"`
}
