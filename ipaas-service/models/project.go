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
