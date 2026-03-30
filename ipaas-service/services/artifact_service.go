package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/wso2/integration-control-plane/ipaas-service/clients/icp"
	"github.com/wso2/integration-control-plane/ipaas-service/models"
)

var ErrUnknownArtifactType = errors.New("unknown artifact type")

// ArtifactService fetches MI runtime artifacts by translating the request into
// a GraphQL query against the ICP service.
type ArtifactService interface {
	ListArtifacts(ctx context.Context, artifactType, environmentID, componentID string) (*models.ArtifactList, error)
}

type artifactService struct {
	icpClient *icp.Client
}

func NewArtifactService(icpClient *icp.Client) ArtifactService {
	return &artifactService{icpClient: icpClient}
}

type artifactTypeConfig struct {
	field  string
	fields string
}

// artifactTypeMap mirrors ARTIFACT_QUERY_MAP in the console (queries.ts).
var artifactTypeMap = map[string]artifactTypeConfig{
	"RestApi":          {field: "restApisByEnvironmentAndComponent", fields: "name, context, version, state, tracing, statistics, carbonApp, url, runtimes { runtimeId, status }, resources { path, methods }"},
	"ProxyService":     {field: "proxyServicesByEnvironmentAndComponent", fields: "name, state, tracing, statistics, carbonApp, endpoints, runtimes { runtimeId, status }"},
	"Endpoint":         {field: "endpointsByEnvironmentAndComponent", fields: "name, type, state, tracing, statistics, attributes { name, value }, runtimes { runtimeId, status }"},
	"InboundEndpoint":  {field: "inboundEndpointsByEnvironmentAndComponent", fields: "name, protocol, sequence, onError, state, tracing, statistics, carbonApp, runtimes { runtimeId, status }"},
	"Sequence":         {field: "sequencesByEnvironmentAndComponent", fields: "name, type, container, state, tracing, statistics, runtimes { runtimeId, status }"},
	"Task":             {field: "tasksByEnvironmentAndComponent", fields: "name, class, group, state, carbonApp, runtimes { runtimeId, status }"},
	"LocalEntry":       {field: "localEntriesByEnvironmentAndComponent", fields: "name, type, value, state, runtimes { runtimeId, status }"},
	"CarbonApp":        {field: "carbonAppsByEnvironmentAndComponent", fields: "name, version, state, artifacts { name, type }, runtimes { runtimeId, status }"},
	"Connector":        {field: "connectorsByEnvironmentAndComponent", fields: "name, package, version, state, runtimes { runtimeId, status }"},
	"RegistryResource": {field: "registryResourcesByEnvironmentAndComponent", fields: "name, type, runtimes { runtimeId, status }"},
	"Listener":         {field: "listenersByEnvironmentAndComponent", fields: "name, package, protocol, host, port, state, runtimes { runtimeId, status }"},
	"Service":          {field: "servicesByEnvironmentAndComponent", fields: "name, package, basePath, type, runtimes { runtimeId, status }, resources { path, method, url, methods }"},
	"Automation":       {field: "automationsByEnvironmentAndComponent", fields: "packageOrg, packageName, packageVersion, runtimeIds, runtimes { runtimeId, status, executionTimestamps }, executionTimestamp"},
	"MessageStore":     {field: "messageStoresByEnvironmentAndComponent", fields: "name, type, size, carbonApp, runtimes { runtimeId, status }"},
	"MessageProcessor": {field: "messageProcessorsByEnvironmentAndComponent", fields: "name, type, state, tracing, carbonApp, runtimes { runtimeId, status }"},
	"Template":         {field: "templatesByEnvironmentAndComponent", fields: "name, type, tracing, statistics, carbonApp, runtimes { runtimeId, status }"},
	"DataService":      {field: "dataServicesByEnvironmentAndComponent", fields: "name, description, state, carbonApp, runtimes { runtimeId, status }"},
	"DataSource":       {field: "dataSourcesByEnvironmentAndComponent", fields: "name, type, driver, url, username, state, runtimes { runtimeId, status }"},
}

const artifactQueryTemplate = `query ArtifactQuery($environmentId: String!, $componentId: String!) { %s(environmentId: $environmentId, componentId: $componentId) { %s } }`

func (s *artifactService) ListArtifacts(ctx context.Context, artifactType, environmentID, componentID string) (*models.ArtifactList, error) {
	cfg, ok := artifactTypeMap[artifactType]
	if !ok {
		return nil, ErrUnknownArtifactType
	}

	query := fmt.Sprintf(artifactQueryTemplate, cfg.field, cfg.fields)
	data, err := s.icpClient.Query(ctx, query, map[string]string{
		"environmentId": environmentID,
		"componentId":   componentID,
	})
	if err != nil {
		return nil, fmt.Errorf("list artifacts: %w", err)
	}

	raw, ok := data[cfg.field]
	if !ok {
		return &models.ArtifactList{Items: []json.RawMessage{}}, nil
	}

	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("list artifacts: parse items: %w", err)
	}
	if items == nil {
		items = []json.RawMessage{}
	}
	return &models.ArtifactList{Items: items}, nil
}
