package models

import "encoding/json"

// ArtifactList is the REST response for artifact queries.
// Items are kept as raw JSON to avoid defining separate structs for every MI
// artifact type (RestApi, Sequence, ProxyService, etc.).
type ArtifactList struct {
	Items []json.RawMessage `json:"items"`
}

type ArtifactType struct {
	ArtifactType  string `json:"artifactType"`
	ArtifactCount int    `json:"artifactCount"`
}

type ArtifactTypeList struct {
	Items []ArtifactType `json:"items"`
}

type ArtifactStatusInput struct {
	ComponentID  string `json:"componentId"`
	ArtifactType string `json:"artifactType"`
	ArtifactName string `json:"artifactName"`
	Status       string `json:"status"`
	EnvID        string `json:"envId"`
}

type ArtifactStatusResult struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type ListenerStateInput struct {
	RuntimeIDs   []string `json:"runtimeIds"`
	ListenerName string   `json:"listenerName"`
	Action       string   `json:"action"`
}

type ListenerStateResult struct {
	Success    bool     `json:"success"`
	Message    string   `json:"message,omitempty"`
	CommandIDs []string `json:"commandIds,omitempty"`
}

type TriggerArtifactInput struct {
	ComponentID string `json:"componentId"`
	TaskName    string `json:"taskName"`
}

type TriggerArtifactResult struct {
	Status       string   `json:"status"`
	Message      string   `json:"message,omitempty"`
	SuccessCount int      `json:"successCount"`
	FailedCount  int      `json:"failedCount"`
	Details      []string `json:"details,omitempty"`
}

type ArtifactTracingInput struct {
	ComponentID  string `json:"componentId"`
	ArtifactType string `json:"artifactType"`
	ArtifactName string `json:"artifactName"`
	Trace        string `json:"trace,omitempty"`
	Statistics   string `json:"statistics,omitempty"`
}
