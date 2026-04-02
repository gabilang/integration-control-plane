package k8s

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

type k8sJobTemplateSpec struct {
	Metadata k8sObjectMeta `json:"metadata"`
	Spec     k8sJobSpec    `json:"spec"`
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
