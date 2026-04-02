package k8s

// Internal types for deserializing Kubernetes batch/v1 Job list responses.

type k8sObjectMeta struct {
	Name string `json:"name"`
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
