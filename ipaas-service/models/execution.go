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
