package models

// Component is the API response for a component resource.
type Component struct {
	UID              string            `json:"uid,omitempty"`
	Name             string            `json:"name"`
	ProjectName      string            `json:"projectName,omitempty"`
	DisplayName      string            `json:"displayName,omitempty"`
	Description      string            `json:"description,omitempty"`
	Type             string            `json:"type,omitempty"`
	AutoDeploy       bool              `json:"autoDeploy,omitempty"`
	AutoBuild        bool              `json:"autoBuild,omitempty"`
	CreatedAt        string            `json:"createdAt,omitempty"`
	Status           string            `json:"status,omitempty"`
	DeploymentTracks []DeploymentTrack `json:"deploymentTracks,omitempty"`
}

// ComponentList is the paginated list response for components.
type ComponentList struct {
	Items      []Component `json:"items"`
	TotalCount int         `json:"totalCount,omitempty"`
	Page       int         `json:"page,omitempty"`
	PageSize   int         `json:"pageSize,omitempty"`
}

// WorkflowRun represents a triggered component build/workflow run.
type WorkflowRun struct {
	Name          string         `json:"name,omitempty"`
	Status        string         `json:"status,omitempty"`
	StartedAt     string         `json:"startedAt,omitempty"`
	CompletedAt   string         `json:"completedAt,omitempty"`
	ComponentName string         `json:"componentName,omitempty"`
	ProjectName   string         `json:"projectName,omitempty"`
	Image         string         `json:"image,omitempty"`
	Commit        string         `json:"commit,omitempty"`
	Tasks         []WorkflowTask `json:"tasks,omitempty"`
}

// WorkflowTask represents a single step/task within a workflow run.
type WorkflowTask struct {
	Name        string `json:"name"`
	Phase       string `json:"phase"`
	StartedAt   string `json:"startedAt,omitempty"`
	CompletedAt string `json:"completedAt,omitempty"`
}

// WorkflowRunList is the paginated list of workflow runs.
type WorkflowRunList struct {
	Items      []WorkflowRun `json:"items"`
	TotalCount int           `json:"totalCount,omitempty"`
}

// BuildLogEntry is a single log line from a build.
type BuildLogEntry struct {
	Timestamp string `json:"timestamp,omitempty"`
	Log       string `json:"log"`
	Level     string `json:"level,omitempty"`
}

// BuildLogs is the response for build log queries.
type BuildLogs struct {
	Logs       []BuildLogEntry `json:"logs"`
	TotalCount int             `json:"totalCount,omitempty"`
}

// WorkflowRevision is the git revision used by a component workflow.
type WorkflowRevision struct {
	Branch string `json:"branch,omitempty"`
	Commit string `json:"commit,omitempty"`
}

// WorkflowRepository is the repository config embedded in a component workflow.
type WorkflowRepository struct {
	URL      string            `json:"url,omitempty"`
	Revision *WorkflowRevision `json:"revision,omitempty"`
	AppPath  string            `json:"appPath,omitempty"`
}

// UpdateComponentRequest holds the mutable fields of a component resource.
type UpdateComponentRequest struct {
	DisplayName string `json:"displayName,omitempty"`
	Description string `json:"description,omitempty"`
}
// Only the workflow configuration (repository, build type, parameters) is mutable.
type UpdateBuildParametersRequest struct {
	Workflow *ComponentWorkflowSpec `json:"workflow"`
}

// CreateComponentRequest is the K8s-style body sent by the frontend and
// forwarded to the OpenChoreo API. The structure mirrors the OpenChoreo
// Component resource.
type CreateComponentRequest struct {
	Metadata ComponentMetadata `json:"metadata"`
	Spec     ComponentSpec     `json:"spec"`
}

// ComponentMetadata holds the metadata portion of a component create request.
type ComponentMetadata struct {
	Name        string            `json:"name"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// ComponentSpec holds the spec portion of a component create request.
type ComponentSpec struct {
	Owner         *ComponentOwner         `json:"owner,omitempty"`
	ComponentType *ComponentTypeRef       `json:"componentType,omitempty"`
	AutoDeploy    bool                    `json:"autoDeploy,omitempty"`
	AutoBuild     bool                    `json:"autoBuild,omitempty"`
	Workflow      *ComponentWorkflowSpec  `json:"workflow,omitempty"`
}

// ComponentOwner identifies the project that owns the component.
type ComponentOwner struct {
	ProjectName string `json:"projectName"`
}

// ComponentTypeRef references a ClusterComponentType.
type ComponentTypeRef struct {
	Kind string `json:"kind,omitempty"`
	Name string `json:"name"`
}

// ComponentWorkflowSpec is the K8s-style workflow embedded in a component spec.
type ComponentWorkflowSpec struct {
	Kind       string                      `json:"kind,omitempty"`
	Name       string                      `json:"name,omitempty"`
	Parameters *ComponentWorkflowParameters `json:"parameters,omitempty"`
}

// ComponentWorkflowParameters holds workflow parameters (repository config).
type ComponentWorkflowParameters struct {
	Repository *WorkflowRepository `json:"repository,omitempty"`
}

// CreateComponentResponse is returned after the component is created and the
// initial build run has been triggered.
type CreateComponentResponse struct {
	Component *Component   `json:"component"`
	BuildRun  *WorkflowRun `json:"buildRun,omitempty"`
}

type ComponentRepository struct {
	GitProvider        string `json:"gitProvider,omitempty"`
	OrganizationApp    string `json:"organizationApp,omitempty"`
	NameApp            string `json:"nameApp,omitempty"`
	Branch             string `json:"branch,omitempty"`
	AppSubPath         string `json:"appSubPath,omitempty"`
	BitbucketServerURL string `json:"bitbucketServerUrl,omitempty"`
	ServerURL          string `json:"serverUrl,omitempty"`
	ProjectApp         string `json:"projectApp,omitempty"`
	// TreeURL is the browser-navigable URL to the repository tree at the component branch,
	// e.g. https://github.com/org/repo/tree/main
	TreeURL string `json:"treeUrl,omitempty"`
}

type Commit struct {
	SHA      string       `json:"sha"`
	Message  string       `json:"message,omitempty"`
	IsLatest bool         `json:"isLatest,omitempty"`
	Author   CommitAuthor `json:"author"`
}

type CommitAuthor struct {
	Name      string `json:"name,omitempty"`
	Date      string `json:"date,omitempty"`
	Email     string `json:"email,omitempty"`
	AvatarURL string `json:"avatarUrl,omitempty"`
}

type CommitList struct {
	Items []Commit `json:"items"`
}

type LabelList struct {
	Items []string `json:"items"`
}

// DeploymentTrack represents a deployment track (branch-based versioning)
// synthesized from a component's workflow spec in OpenChoreo.
type DeploymentTrack struct {
	ID                string `json:"id,omitempty"`
	Branch            string `json:"branch,omitempty"`
	CommitSHA         string `json:"commitSha,omitempty"`
	URL               string `json:"url,omitempty"`
	AppPath           string `json:"appPath,omitempty"`
	ComponentID       string `json:"componentId,omitempty"`
	Latest            bool   `json:"latest"`
	AutoDeployEnabled bool   `json:"autoDeployEnabled"`
	CreatedAt         string `json:"createdAt,omitempty"`
	UpdatedAt         string `json:"updatedAt,omitempty"`
}
