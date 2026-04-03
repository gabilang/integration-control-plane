package models

type Execution struct {
	JobID          string `json:"jobId"`
	Status         string `json:"status"`
	StartTime      string `json:"startTime,omitempty"`
	CompletionTime string `json:"completionTime,omitempty"`
	RevisionID     string `json:"revisionId,omitempty"`
}

type ExecutionList struct {
	Items []Execution `json:"items"`
}

type ExecutionConfigs struct {
	CronjobFrequency        string `json:"cronjobFrequency,omitempty"`
	CronjobTimezone         string `json:"cronjobTimezone,omitempty"`
	CronjobAllowConcurrency *bool  `json:"cronjobAllowConcurrency,omitempty"`
	TimeoutSeconds          *int   `json:"timeoutSeconds,omitempty"`
	RetryCount              *int   `json:"retryCount,omitempty"`
}

type ExecutionArgument struct {
	ArgumentName  string `json:"argumentName"`
	ArgumentValue string `json:"argumentValue"`
}

type UpdateJobConfigsInput struct {
	OrgHandler              string  `json:"orgHandler"`
	ComponentID             string  `json:"componentId"`
	EnvironmentID           string  `json:"environmentId"`
	VersionID               string  `json:"versionId"`
	CronFrequency           *string `json:"cronFrequency,omitempty"`
	CronTimezone            *string `json:"cronTimezone,omitempty"`
	JobTimeoutSeconds       *int    `json:"jobTimeoutSeconds,omitempty"`
	CronJobAllowConcurrency *bool   `json:"cronJobAllowConcurrency,omitempty"`
	JobRetryCount           *int    `json:"jobRetryCount,omitempty"`
}
