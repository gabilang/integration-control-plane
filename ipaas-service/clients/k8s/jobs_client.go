package k8s

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/wso2/integration-control-plane/ipaas-service/clients/requests"
	"github.com/wso2/integration-control-plane/ipaas-service/models"
)

const (
	saTokenPath  = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	saCACertPath = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	k8sAPIBase   = "https://kubernetes.default.svc.cluster.local"
)

var ErrCronJobNotFound = errors.New("cronjob not found")

// JobsClient manages Kubernetes batch/v1 Jobs and CronJobs using in-cluster credentials.
type JobsClient interface {
	ListJobs(ctx context.Context, labelSelector string) (*models.ExecutionList, error)
	GetCronJob(ctx context.Context, labelSelector string) (*k8sCronJob, error)
	TriggerJob(ctx context.Context, cronjob *k8sCronJob) (*models.Execution, error)
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

func (c *jobsClient) saToken() (string, error) {
	token, err := os.ReadFile(saTokenPath)
	if err != nil {
		return "", fmt.Errorf("read k8s SA token: %w", err)
	}
	return strings.TrimSpace(string(token)), nil
}

// ListJobs returns all batch/v1 Jobs matching the given label selector,
// across all namespaces.
func (c *jobsClient) ListJobs(ctx context.Context, labelSelector string) (*models.ExecutionList, error) {
	token, err := c.saToken()
	if err != nil {
		return nil, err
	}

	req := requests.NewRequest("k8s.ListJobs", http.MethodGet, c.baseURL+"/apis/batch/v1/jobs")
	req.SetHeader("Authorization", "Bearer "+token)
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

// GetCronJob returns the first CronJob matching the given label selector.
// Returns ErrCronJobNotFound if no CronJob matches.
func (c *jobsClient) GetCronJob(ctx context.Context, labelSelector string) (*k8sCronJob, error) {
	token, err := c.saToken()
	if err != nil {
		return nil, err
	}

	req := requests.NewRequest("k8s.GetCronJob", http.MethodGet, c.baseURL+"/apis/batch/v1/cronjobs")
	req.SetHeader("Authorization", "Bearer "+token)
	if labelSelector != "" {
		req.SetQuery("labelSelector", labelSelector)
	}

	result := requests.SendRequest(ctx, c.httpClient, req)
	var raw k8sCronJobList
	if err := result.ScanResponse(&raw, http.StatusOK); err != nil {
		return nil, fmt.Errorf("get cronjob: %w", err)
	}
	if len(raw.Items) == 0 {
		return nil, ErrCronJobNotFound
	}
	return &raw.Items[0], nil
}

// TriggerJob creates a one-off Job from the CronJob's job template and returns the new execution.
func (c *jobsClient) TriggerJob(ctx context.Context, cronjob *k8sCronJob) (*models.Execution, error) {
	token, err := c.saToken()
	if err != nil {
		return nil, err
	}

	jobName := fmt.Sprintf("%s-manual-%d", cronjob.Metadata.Name, time.Now().UnixMilli())
	body := k8sTriggerJobBody{
		Metadata: k8sObjectMeta{
			Name:      jobName,
			Namespace: cronjob.Metadata.Namespace,
			Labels:    cronjob.Spec.JobTemplate.Metadata.Labels,
		},
		Spec: cronjob.Spec.JobTemplate.Spec,
	}

	url := fmt.Sprintf("%s/apis/batch/v1/namespaces/%s/jobs", c.baseURL, cronjob.Metadata.Namespace)
	req := requests.NewRequest("k8s.TriggerJob", http.MethodPost, url)
	req.SetHeader("Authorization", "Bearer "+token)
	req.SetJSON(body)

	result := requests.SendRequest(ctx, c.httpClient, req)
	var raw k8sJob
	if err := result.ScanResponse(&raw, http.StatusCreated); err != nil {
		return nil, fmt.Errorf("trigger job: %w", err)
	}
	execution := normalizeJob(raw)
	return &execution, nil
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
