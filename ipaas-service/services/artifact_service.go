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
	GetArtifactTypes(ctx context.Context, componentID, envID string) (*models.ArtifactTypeList, error)
	GetArtifactSource(ctx context.Context, envID, componentID, artifactType, artifactName string) (string, error)
	GetLocalEntryValue(ctx context.Context, componentID, entryName, envID string) (string, error)
	GetArtifactParams(ctx context.Context, componentID, artifactType, artifactName, envID, runtimeID string) ([]map[string]string, error)
	GetArtifactWsdl(ctx context.Context, componentID, artifactType, artifactName, envID, runtimeID string) (string, error)
	UpdateArtifactStatus(ctx context.Context, input *models.ArtifactStatusInput) (*models.ArtifactStatusResult, error)
	UpdateListenerState(ctx context.Context, input *models.ListenerStateInput) (*models.ListenerStateResult, error)
	TriggerArtifact(ctx context.Context, input *models.TriggerArtifactInput) (*models.TriggerArtifactResult, error)
	UpdateArtifactTracingStatus(ctx context.Context, input *models.ArtifactTracingInput) (*models.ArtifactStatusResult, error)
	UpdateArtifactStatisticsStatus(ctx context.Context, input *models.ArtifactTracingInput) (*models.ArtifactStatusResult, error)
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

const (
	artifactTypesQuery           = `query ComponentArtifactTypes($componentId: String!, $environmentId: String!) { componentArtifactTypes(componentId: $componentId, environmentId: $environmentId) { artifactType, artifactCount } }`
	artifactSourceQuery          = `query GetArtifactSource($environmentId: String!, $componentId: String!, $artifactType: String!, $artifactName: String!) { artifactSourceByComponent(environmentId: $environmentId, componentId: $componentId, artifactType: $artifactType, artifactName: $artifactName) }`
	localEntryValueQuery         = `query LocalEntryValue($componentId: String!, $entryName: String!, $environmentId: String) { localEntryValueByComponent(componentId: $componentId, entryName: $entryName, environmentId: $environmentId) }`
	artifactParamsQuery          = `query ArtifactParams($componentId: String!, $artifactType: String!, $artifactName: String!, $environmentId: String, $runtimeId: String) { artifactParametersByComponent(componentId: $componentId, artifactType: $artifactType, artifactName: $artifactName, environmentId: $environmentId, runtimeId: $runtimeId) { name, value } }`
	artifactWsdlQuery            = `query ArtifactWsdl($componentId: String!, $artifactType: String!, $artifactName: String!, $environmentId: String, $runtimeId: String) { artifactWsdlByComponent(componentId: $componentId, artifactType: $artifactType, artifactName: $artifactName, environmentId: $environmentId, runtimeId: $runtimeId) }`
	updateArtifactStatusMutation = `mutation UpdateArtifactStatus($input: ArtifactStatusChangeInput!) { updateArtifactStatus(input: $input) { status, message } }`
	updateListenerStateMutation  = `mutation UpdateListenerState($input: ListenerControlInput!) { updateListenerState(input: $input) { success, message, commandIds } }`
	triggerArtifactMutation      = `mutation TriggerTask($input: ArtifactTriggerInput!) { triggerArtifact(input: $input) { status, message, successCount, failedCount, details } }`
	updateArtifactTracingMutation = `mutation UpdateArtifactTracingStatus($input: ArtifactTracingStatusInput!) { updateArtifactTracingStatus(input: $input) { status, message } }`
	updateArtifactStatsMutation  = `mutation UpdateArtifactStatisticsStatus($input: ArtifactTracingStatusInput!) { updateArtifactStatisticsStatus(input: $input) { status, message } }`
)

func (s *artifactService) GetArtifactTypes(ctx context.Context, componentID, envID string) (*models.ArtifactTypeList, error) {
	data, err := s.icpClient.Query(ctx, artifactTypesQuery, map[string]string{
		"componentId":   componentID,
		"environmentId": envID,
	})
	if err != nil {
		return nil, fmt.Errorf("get artifact types: %w", err)
	}
	raw, ok := data["componentArtifactTypes"]
	if !ok {
		return &models.ArtifactTypeList{Items: []models.ArtifactType{}}, nil
	}
	var items []models.ArtifactType
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("get artifact types: parse: %w", err)
	}
	if items == nil {
		items = []models.ArtifactType{}
	}
	return &models.ArtifactTypeList{Items: items}, nil
}

func (s *artifactService) GetArtifactSource(ctx context.Context, envID, componentID, artifactType, artifactName string) (string, error) {
	data, err := s.icpClient.Query(ctx, artifactSourceQuery, map[string]string{
		"environmentId": envID,
		"componentId":   componentID,
		"artifactType":  artifactType,
		"artifactName":  artifactName,
	})
	if err != nil {
		return "", fmt.Errorf("get artifact source: %w", err)
	}
	raw, ok := data["artifactSourceByComponent"]
	if !ok {
		return "", nil
	}
	var result string
	if err := json.Unmarshal(raw, &result); err != nil {
		return string(raw), nil
	}
	return result, nil
}

func (s *artifactService) GetLocalEntryValue(ctx context.Context, componentID, entryName, envID string) (string, error) {
	data, err := s.icpClient.Query(ctx, localEntryValueQuery, map[string]string{
		"componentId":   componentID,
		"entryName":     entryName,
		"environmentId": envID,
	})
	if err != nil {
		return "", fmt.Errorf("get local entry value: %w", err)
	}
	raw, ok := data["localEntryValueByComponent"]
	if !ok {
		return "", nil
	}
	var result string
	if err := json.Unmarshal(raw, &result); err != nil {
		return string(raw), nil
	}
	return result, nil
}

func (s *artifactService) GetArtifactParams(ctx context.Context, componentID, artifactType, artifactName, envID, runtimeID string) ([]map[string]string, error) {
	data, err := s.icpClient.Query(ctx, artifactParamsQuery, map[string]string{
		"componentId":   componentID,
		"artifactType":  artifactType,
		"artifactName":  artifactName,
		"environmentId": envID,
		"runtimeId":     runtimeID,
	})
	if err != nil {
		return nil, fmt.Errorf("get artifact params: %w", err)
	}
	raw, ok := data["artifactParametersByComponent"]
	if !ok {
		return []map[string]string{}, nil
	}
	var items []map[string]string
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("get artifact params: parse: %w", err)
	}
	if items == nil {
		items = []map[string]string{}
	}
	return items, nil
}

func (s *artifactService) GetArtifactWsdl(ctx context.Context, componentID, artifactType, artifactName, envID, runtimeID string) (string, error) {
	data, err := s.icpClient.Query(ctx, artifactWsdlQuery, map[string]string{
		"componentId":   componentID,
		"artifactType":  artifactType,
		"artifactName":  artifactName,
		"environmentId": envID,
		"runtimeId":     runtimeID,
	})
	if err != nil {
		return "", fmt.Errorf("get artifact wsdl: %w", err)
	}
	raw, ok := data["artifactWsdlByComponent"]
	if !ok {
		return "", nil
	}
	var result string
	if err := json.Unmarshal(raw, &result); err != nil {
		return string(raw), nil
	}
	return result, nil
}

func (s *artifactService) UpdateArtifactStatus(ctx context.Context, input *models.ArtifactStatusInput) (*models.ArtifactStatusResult, error) {
	data, err := s.icpClient.Query(ctx, updateArtifactStatusMutation, map[string]interface{}{
		"input": input,
	})
	if err != nil {
		return nil, fmt.Errorf("update artifact status: %w", err)
	}
	raw, ok := data["updateArtifactStatus"]
	if !ok {
		return &models.ArtifactStatusResult{Status: "ok"}, nil
	}
	var result models.ArtifactStatusResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("update artifact status: parse: %w", err)
	}
	return &result, nil
}

func (s *artifactService) UpdateListenerState(ctx context.Context, input *models.ListenerStateInput) (*models.ListenerStateResult, error) {
	data, err := s.icpClient.Query(ctx, updateListenerStateMutation, map[string]interface{}{
		"input": input,
	})
	if err != nil {
		return nil, fmt.Errorf("update listener state: %w", err)
	}
	raw, ok := data["updateListenerState"]
	if !ok {
		return &models.ListenerStateResult{Success: true}, nil
	}
	var result models.ListenerStateResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("update listener state: parse: %w", err)
	}
	return &result, nil
}

func (s *artifactService) TriggerArtifact(ctx context.Context, input *models.TriggerArtifactInput) (*models.TriggerArtifactResult, error) {
	data, err := s.icpClient.Query(ctx, triggerArtifactMutation, map[string]interface{}{
		"input": input,
	})
	if err != nil {
		return nil, fmt.Errorf("trigger artifact: %w", err)
	}
	raw, ok := data["triggerArtifact"]
	if !ok {
		return &models.TriggerArtifactResult{Status: "ok"}, nil
	}
	var result models.TriggerArtifactResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("trigger artifact: parse: %w", err)
	}
	return &result, nil
}

func (s *artifactService) UpdateArtifactTracingStatus(ctx context.Context, input *models.ArtifactTracingInput) (*models.ArtifactStatusResult, error) {
	data, err := s.icpClient.Query(ctx, updateArtifactTracingMutation, map[string]interface{}{
		"input": input,
	})
	if err != nil {
		return nil, fmt.Errorf("update artifact tracing: %w", err)
	}
	raw, ok := data["updateArtifactTracingStatus"]
	if !ok {
		return &models.ArtifactStatusResult{Status: "ok"}, nil
	}
	var result models.ArtifactStatusResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("update artifact tracing: parse: %w", err)
	}
	return &result, nil
}

func (s *artifactService) UpdateArtifactStatisticsStatus(ctx context.Context, input *models.ArtifactTracingInput) (*models.ArtifactStatusResult, error) {
	data, err := s.icpClient.Query(ctx, updateArtifactStatsMutation, map[string]interface{}{
		"input": input,
	})
	if err != nil {
		return nil, fmt.Errorf("update artifact statistics: %w", err)
	}
	raw, ok := data["updateArtifactStatisticsStatus"]
	if !ok {
		return &models.ArtifactStatusResult{Status: "ok"}, nil
	}
	var result models.ArtifactStatusResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("update artifact statistics: parse: %w", err)
	}
	return &result, nil
}
