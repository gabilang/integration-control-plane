package models

type ComponentDeployment struct {
	ReleaseID    string           `json:"releaseId,omitempty"`
	Cron         string           `json:"cron,omitempty"`
	CronTimezone string           `json:"cronTimezone,omitempty"`
	Build        *DeploymentBuild `json:"build,omitempty"`
}

type DeploymentBuild struct {
	BuildID string `json:"buildId,omitempty"`
}

type DeploymentStatus struct {
	ID             int    `json:"id,omitempty"`
	SHA            string `json:"sha,omitempty"`
	StartedAt      string `json:"started_at,omitempty"`
	CompletedAt    string `json:"completed_at,omitempty"`
	Status         string `json:"status,omitempty"`
	Conclusion     string `json:"conclusion,omitempty"`
	ConclusionV2   string `json:"conclusionV2,omitempty"`
	IsAutoDeploy   bool   `json:"isAutoDeploy,omitempty"`
	Name           string `json:"name,omitempty"`
	FailureReason  int    `json:"failureReason,omitempty"`
	SourceCommitID string `json:"sourceCommitId,omitempty"`
	BuildRef       string `json:"buildRef,omitempty"`
}

type DeployDeploymentTrackInput struct {
	ComponentID             string  `json:"componentId"`
	ID                      string  `json:"id"`
	ImageID                 string  `json:"imageId"`
	EnvironmentID           string  `json:"environmentId"`
	DeploymentPipelineID    string  `json:"deploymentPipelineId"`
	CronTimezone            *string `json:"cronTimezone,omitempty"`
	Cron                    *string `json:"cron,omitempty"`
	JobTimeoutSeconds       *int    `json:"jobTimeoutSeconds,omitempty"`
	CronJobAllowConcurrency *bool   `json:"cronJobAllowConcurrency,omitempty"`
}

type PromoteInput struct {
	APIVersionID         string `json:"apiVersionId"`
	SourceReleaseID      string `json:"sourceReleaseId"`
	TargetEnvironmentID  string `json:"targetEnvironmentId"`
	DeploymentPipelineID string `json:"deploymentPipelineId"`
}

type StopDeploymentInput struct {
	OrgHandler  string `json:"orgHandler"`
	ComponentID string `json:"componentId"`
	ReleaseID   string `json:"releaseId"`
}
