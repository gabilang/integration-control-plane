package models

// ResourceNode represents a single Kubernetes resource in the resource tree.
type ResourceNode struct {
	Group           string         `json:"group,omitempty"`
	Version         string         `json:"version"`
	Kind            string         `json:"kind"`
	Namespace       string         `json:"namespace,omitempty"`
	Name            string         `json:"name"`
	UID             string         `json:"uid"`
	ResourceVersion string         `json:"resourceVersion,omitempty"`
	CreatedAt       string         `json:"createdAt,omitempty"`
	ParentRefs      []ResourceRef  `json:"parentRefs,omitempty"`
	Object          map[string]any `json:"object,omitempty"`
	Health          *HealthInfo    `json:"health,omitempty"`
}

type ResourceRef struct {
	Group     string `json:"group,omitempty"`
	Version   string `json:"version"`
	Kind      string `json:"kind"`
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name"`
	UID       string `json:"uid"`
}

type HealthInfo struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// ReleaseResourceTree holds all resource nodes for a single rendered release.
type ReleaseResourceTree struct {
	Name        string         `json:"name"`
	TargetPlane string         `json:"targetPlane"`
	Nodes       []ResourceNode `json:"nodes"`
}

// ResourceTreeResponse is the BFF response for the full resource tree.
type ResourceTreeResponse struct {
	RenderedReleases []ReleaseResourceTree `json:"renderedReleases"`
}

// ResourceEvent represents a Kubernetes event for a resource.
type ResourceEvent struct {
	Type           string `json:"type"`
	Reason         string `json:"reason"`
	Message        string `json:"message"`
	Count          int    `json:"count,omitempty"`
	FirstTimestamp string `json:"firstTimestamp,omitempty"`
	LastTimestamp   string `json:"lastTimestamp,omitempty"`
	Source         string `json:"source,omitempty"`
}

// ResourceEventsResponse wraps a list of events for a specific resource.
type ResourceEventsResponse struct {
	Events []ResourceEvent `json:"events"`
}

// PodLogEntry represents a single log line from a pod.
type PodLogEntry struct {
	Timestamp string `json:"timestamp"`
	Log       string `json:"log"`
}

// PodLogsResponse wraps a list of log entries for a pod.
type PodLogsResponse struct {
	LogEntries []PodLogEntry `json:"logEntries"`
}
