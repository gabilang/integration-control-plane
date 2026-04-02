package k8s

import "encoding/json"

// Internal types for deserializing Kubernetes batch/v1 Job and CronJob responses.

type k8sObjectMeta struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
}

type k8sContainer struct {
	Image string `json:"image"`
}

type k8sPodSpec struct {
	Containers []k8sContainer `json:"containers"`
}

type k8sPodTemplate struct {
	Spec k8sPodSpec `json:"spec"`
}

type k8sJobSpec struct {
	Template k8sPodTemplate `json:"template"`
}

type k8sJobStatus struct {
	Active         int    `json:"active"`
	Succeeded      int    `json:"succeeded"`
	Failed         int    `json:"failed"`
	StartTime      string `json:"startTime,omitempty"`
	CompletionTime string `json:"completionTime,omitempty"`
}

type k8sJob struct {
	Metadata k8sObjectMeta `json:"metadata"`
	Spec     k8sJobSpec    `json:"spec"`
	Status   k8sJobStatus  `json:"status"`
}

type k8sJobList struct {
	Items []k8sJob `json:"items"`
}

// k8sJobTemplateSpec holds the CronJob's jobTemplate.
// Spec is kept as raw JSON so it is passed through to Job creation unchanged —
// avoids losing required fields (container name, restartPolicy, resources, etc.)
// that are not part of our limited read model.
type k8sJobTemplateSpec struct {
	Metadata k8sObjectMeta   `json:"metadata"`
	Spec     json.RawMessage `json:"spec"`
}

type k8sCronJobSpec struct {
	JobTemplate k8sJobTemplateSpec `json:"jobTemplate"`
}

type k8sCronJob struct {
	Metadata k8sObjectMeta  `json:"metadata"`
	Spec     k8sCronJobSpec `json:"spec"`
}

type k8sCronJobList struct {
	Items []k8sCronJob `json:"items"`
}

// k8sTriggerJobBody is the request body for creating a manual Job from a CronJob.
type k8sTriggerJobBody struct {
	Metadata k8sObjectMeta   `json:"metadata"`
	Spec     json.RawMessage `json:"spec"`
}
