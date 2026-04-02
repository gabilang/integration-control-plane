package k8s

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/wso2/integration-control-plane/ipaas-service/clients/requests"
	"github.com/wso2/integration-control-plane/ipaas-service/models"
)

const (
	saTokenPath  = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	saCACertPath = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	k8sAPIBase   = "https://kubernetes.default.svc.cluster.local"
)

// JobsClient lists Kubernetes batch/v1 Jobs using in-cluster credentials.
type JobsClient interface {
	ListJobs(ctx context.Context, labelSelector string) (*models.ExecutionList, error)
}

type jobsClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewJobsClient creates a JobsClient configured with the pod's mounted SA CA cert.
func NewJobsClient() (JobsClient, error) {
	caCert, err := os.ReadFile(saCACertPath)
	if err != nil {
		return nil, fmt.Errorf("read k8s CA cert: %w", err)
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caCert)
	return &jobsClient{
		baseURL: k8sAPIBase,
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{RootCAs: pool},
			},
		},
	}, nil
}

// ListJobs returns all batch/v1 Jobs matching the given label selector,
// across all namespaces.
func (c *jobsClient) ListJobs(ctx context.Context, labelSelector string) (*models.ExecutionList, error) {
	token, err := os.ReadFile(saTokenPath)
	if err != nil {
		return nil, fmt.Errorf("read k8s SA token: %w", err)
	}

	req := requests.NewRequest("k8s.ListJobs", http.MethodGet, c.baseURL+"/apis/batch/v1/jobs")
	req.SetHeader("Authorization", "Bearer "+strings.TrimSpace(string(token)))
	if labelSelector != "" {
		req.SetQuery("labelSelector", labelSelector)
	}

	result := requests.SendRequest(ctx, c.httpClient, req)
	var raw k8sJobList
	if err := result.ScanResponse(&raw, http.StatusOK); err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}

	items := make([]models.Execution, len(raw.Items))
	for i, job := range raw.Items {
		items[i] = normalizeJob(job)
	}
	return &models.ExecutionList{Items: items}, nil
}

func normalizeJob(job k8sJob) models.Execution {
	status := "Unknown"
	switch {
	case job.Status.Active > 0:
		status = "Running"
	case job.Status.Succeeded > 0:
		status = "Succeeded"
	case job.Status.Failed > 0:
		status = "Failed"
	}

	var revisionID string
	for _, c := range job.Spec.Template.Spec.Containers {
		if c.Image != "" {
			revisionID = extractRevisionFromImage(c.Image)
			break
		}
	}

	return models.Execution{
		JobID:          job.Metadata.Name,
		Status:         status,
		StartTime:      job.Status.StartTime,
		CompletionTime: job.Status.CompletionTime,
		RevisionID:     revisionID,
	}
}

// extractRevisionFromImage extracts the git revision from a container image tag.
// Image format: "registry/org/component:v1-3c574505" → "3c574505"
func extractRevisionFromImage(image string) string {
	tag := image
	if idx := strings.LastIndex(image, ":"); idx >= 0 {
		tag = image[idx+1:]
	}
	parts := strings.Split(tag, "-")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return ""
}
