package openchoreo

// Internal K8s-style types used when communicating with the OpenChoreo API.
// These are not exposed outside this package — callers use the flat models
// from the models package.

// -- Common ------------------------------------------------------------------

type ocObjectMeta struct {
	Name              string            `json:"name,omitempty"`
	Namespace         string            `json:"namespace,omitempty"`
	UID               string            `json:"uid,omitempty"`
	CreationTimestamp string            `json:"creationTimestamp,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
}

type ocCondition struct {
	Type   string `json:"type"`
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

type ocStatus struct {
	Conditions []ocCondition `json:"conditions,omitempty"`
}

type ocRef struct {
	Name string `json:"name"`
}

func latestConditionReason(conditions []ocCondition) string {
	if len(conditions) == 0 {
		return ""
	}
	return conditions[len(conditions)-1].Reason
}

// -- Project -----------------------------------------------------------------

type ocProjectSpec struct {
	DeploymentPipelineRef *ocRef `json:"deploymentPipelineRef,omitempty"`
}

type ocProject struct {
	Metadata ocObjectMeta  `json:"metadata"`
	Spec     ocProjectSpec `json:"spec"`
	Status   ocStatus      `json:"status,omitempty"`
}

type ocProjectList struct {
	Items []ocProject `json:"items"`
}

// -- Component ---------------------------------------------------------------

type ocComponentTypeRef struct {
	Kind string `json:"kind,omitempty"`
	Name string `json:"name"`
}

type ocOwner struct {
	ProjectName string `json:"projectName"`
}

type ocWorkflowRevision struct {
	Branch string `json:"branch,omitempty"`
	Commit string `json:"commit,omitempty"`
}

type ocWorkflowRepository struct {
	URL      string              `json:"url,omitempty"`
	Revision *ocWorkflowRevision `json:"revision,omitempty"`
	AppPath  string              `json:"appPath,omitempty"`
}

type ocWorkflowParameters struct {
	Repository *ocWorkflowRepository `json:"repository,omitempty"`
}

type ocWorkflow struct {
	Kind       string                `json:"kind,omitempty"`
	Name       string                `json:"name"`
	Parameters *ocWorkflowParameters `json:"parameters,omitempty"`
}

type ocComponentSpec struct {
	Owner         *ocOwner            `json:"owner,omitempty"`
	ComponentType *ocComponentTypeRef `json:"componentType,omitempty"`
	AutoDeploy    bool                `json:"autoDeploy,omitempty"`
	AutoBuild     bool                `json:"autoBuild,omitempty"`
	Workflow      *ocWorkflow         `json:"workflow,omitempty"`
}

type ocComponent struct {
	Metadata ocObjectMeta    `json:"metadata"`
	Spec     ocComponentSpec `json:"spec"`
	Status   ocStatus        `json:"status,omitempty"`
}

type ocComponentList struct {
	Items []ocComponent `json:"items"`
}

// -- Workflow Run ------------------------------------------------------------

type ocParameter struct {
	Name  string `json:"name"`
	Value string `json:"value,omitempty"`
}

type ocTaskOutputs struct {
	Parameters []ocParameter `json:"parameters,omitempty"`
}

type ocTask struct {
	Name    string         `json:"name"`
	Outputs *ocTaskOutputs `json:"outputs,omitempty"`
}

type ocWorkflowRunStatus struct {
	Conditions []ocCondition `json:"conditions,omitempty"`
	Tasks      []ocTask      `json:"tasks,omitempty"`
}

type ocWorkflowRunSpec struct {
	Workflow *ocWorkflow `json:"workflow,omitempty"`
}

type ocWorkflowRun struct {
	Metadata ocObjectMeta        `json:"metadata"`
	Spec     ocWorkflowRunSpec   `json:"spec"`
	Status   ocWorkflowRunStatus `json:"status,omitempty"`
}

type ocWorkflowRunList struct {
	Items []ocWorkflowRun `json:"items"`
}
