package models

// Schedule is the API response for a scheduled-task deployment in one environment.
type Schedule struct {
	Environment           string `json:"environment"`
	ComponentName         string `json:"componentName,omitempty"`
	ProjectName           string `json:"projectName,omitempty"`
	CronExpression        string `json:"cronExpression"`
	State                 string `json:"state"`           // "Active" | "Undeploy"
	ImagePullPolicy       string `json:"imagePullPolicy,omitempty"`
	ReleaseName           string `json:"releaseName,omitempty"`
	BackoffLimit          *int   `json:"backoffLimit,omitempty"`
	ActiveDeadlineSeconds *int   `json:"activeDeadlineSeconds,omitempty"`
}

// ScheduleList is the collection response for list operations.
type ScheduleList struct {
	Items []Schedule `json:"items"`
}

// UpsertScheduleRequest is the POST body for creating or updating a schedule.
type UpsertScheduleRequest struct {
	Environment           string `json:"environment"`     // required
	CronExpression        string `json:"cronExpression"`  // required
	State                 string `json:"state,omitempty"` // defaults to "Active"
	ReleaseName           string `json:"releaseName,omitempty"`
	BackoffLimit          *int   `json:"backoffLimit,omitempty"`          // component-wide (applies to all environments)
	ActiveDeadlineSeconds *int   `json:"activeDeadlineSeconds,omitempty"` // component-wide (applies to all environments)
}
