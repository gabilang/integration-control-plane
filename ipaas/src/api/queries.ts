/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import { useQuery, useQueryClient } from '@tanstack/react-query';
import { gql } from './graphql';
import { icpClient } from './client';

// ── BFF response shapes (mirrors ipaas-service/models/project.go, component.go) ──

export interface BffProject {
  uid?: string;
  name: string;
  displayName?: string;
  description?: string;
  deploymentPipeline?: string;
  createdAt?: string;
  status?: string;
}

export interface BffProjectList {
  items: BffProject[];
  totalCount?: number;
}

export interface BffComponent {
  uid?: string;
  name: string;
  projectName?: string;
  displayName?: string;
  description?: string;
  type?: string;
  createdAt?: string;
  status?: string;
}

export interface BffComponentList {
  items: BffComponent[];
  totalCount?: number;
}

// ── GQL-compatible output shapes (consumed by existing UI components) ──

export interface GqlProject {
  id: string;
  orgId: number;
  name: string;
  handler: string;
  description: string;
  version: string;
  createdDate: string;
  updatedAt: string;
  region: string;
  type: string;
}

export interface GqlComponent {
  projectId: string;
  id: string;
  name: string;
  handler: string;
  displayName: string;
  displayType: string;
  description: string;
  status: string;
  componentType?: string;
  componentSubType: string | null;
  version: string;
  createdAt: string;
  lastBuildDate: string;
}

// ── Mapping helpers (exported for use by mutations.ts) ──

export function mapProject(p: BffProject): GqlProject {
  return {
    id: p.name,                      // K8s name is the unique identifier and URL slug
    orgId: 0,
    name: p.displayName || p.name,
    handler: p.name,
    description: p.description ?? '',
    version: '',
    createdDate: p.createdAt ?? '',
    updatedAt: p.createdAt ?? '',
    region: '',
    type: '',
  };
}

export function mapComponent(c: BffComponent): GqlComponent {
  return {
    id: c.name,
    projectId: c.projectName ?? '',
    name: c.displayName || c.name,
    handler: c.name,
    displayName: c.displayName ?? c.name,
    displayType: c.type ?? '',
    description: c.description ?? '',
    status: c.status ?? '',
    componentType: c.type,
    componentSubType: null,
    version: '',
    createdAt: c.createdAt ?? '',
    lastBuildDate: '',
  };
}

// ── Org / project hooks ──

export interface OrgEntry {
  handle: string;
  numericId: number;
  uuid: string;
}

/**
 * Returns the current organization. With Thunder auth the org context is
 * derived from JWT claims in the BFF; the frontend does not need a numeric ID.
 */
export function useOrgs() {
  return useQuery({
    queryKey: ['orgs'],
    queryFn: async (): Promise<OrgEntry[]> => [{ handle: 'default', numericId: 0, uuid: '' }],
    staleTime: 5 * 60 * 1000,
  });
}

export function useProjects() {
  return useQuery({
    queryKey: ['projects'],
    queryFn: () => icpClient.get<BffProjectList>('/projects').then((d) => d.items.map(mapProject)),
  });
}

export function useProjectsByOrg(orgHandle: string) {
  return useQuery({
    queryKey: ['projects'],
    queryFn: () => icpClient.get<BffProjectList>('/projects').then((d) => d.items.map(mapProject)),
    enabled: !!orgHandle,
  });
}

export function useProject(projectName: string) {
  return useQuery({
    queryKey: ['project', projectName],
    queryFn: () => icpClient.get<BffProject>(`/projects/${encodeURIComponent(projectName)}`).then(mapProject),
    enabled: !!projectName,
  });
}

const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

export function useProjectByHandler(handler: string) {
  return useQuery({
    queryKey: ['project', 'handler', handler],
    queryFn: () => icpClient.get<BffProject>(`/projects/${encodeURIComponent(handler)}`).then(mapProject),
    // Guard: never call if handler is empty or looks like a UUID (use useProject instead)
    enabled: !!handler && !UUID_RE.test(handler),
  });
}

export function useComponents(orgHandler: string, projectName: string) {
  return useQuery({
    queryKey: ['components', projectName],
    queryFn: () =>
      icpClient.get<BffComponentList>('/components', { projectName }).then((d) => d.items.map(mapComponent)),
    enabled: !!orgHandler && !!projectName,
  });
}

export interface GqlComponentDetail extends GqlComponent {
  orgHandler: string;
}

export function useComponentByHandler(projectName: string, handler: string | undefined) {
  return useQuery({
    queryKey: ['component', projectName, handler],
    queryFn: () =>
      icpClient.get<BffComponentList>('/components', { projectName }).then((d) => {
        const c = d.items.find((x) => x.name === handler);
        if (!c) throw new Error(`Component '${handler}' not found in project '${projectName}'`);
        return { ...mapComponent(c), orgHandler: '' } as GqlComponentDetail;
      }),
    enabled: !!projectName && !!handler,
  });
}

export interface GqlEnvironment {
  id: string;
  name: string;
  critical: boolean;
  dpId?: string;
  description?: string;
  createdAt?: string;
}

const ENVIRONMENTS_QUERY = `
  query GetEnvironments($orgUuid: String!, $projectId: String!) {
    environments(orgUuid: $orgUuid, type: "external", projectId: $projectId) {
      id, name, critical, dpId
    }
  }`;

export function useEnvironments(orgUuid: string, projectId: string) {
  return useQuery({
    queryKey: ['environments', orgUuid, projectId],
    queryFn: () => gql<{ environments: GqlEnvironment[] }>(ENVIRONMENTS_QUERY, { orgUuid, projectId }).then((d) => d.environments),
    enabled: !!orgUuid && !!projectId,
  });
}

const ALL_ENVIRONMENTS_QUERY = `{
  environments { id, name, description, critical, dpId, createdAt }
}`;

export function useAllEnvironments() {
  return useQuery({
    queryKey: ['environments'],
    queryFn: () => gql<{ environments: GqlEnvironment[] }>(ALL_ENVIRONMENTS_QUERY).then((d) => d.environments),
    retry: false,
  });
}


export interface GqlLogger {
  componentName: string;
  logLevel: string;
  runtimeIds: string[];
}

const LOGGERS_BY_ENV_AND_COMPONENT_QUERY = `
  query GetLoggers($environmentId: String!, $componentId: String!) {
    loggersByEnvironmentAndComponent(environmentId: $environmentId, componentId: $componentId) {
      componentName, logLevel, runtimeIds
    }
  }`;

export function useLoggers(environmentId: string, componentId: string) {
  return useQuery({
    queryKey: ['loggers', environmentId, componentId],
    queryFn: () => gql<{ loggersByEnvironmentAndComponent: GqlLogger[] }>(LOGGERS_BY_ENV_AND_COMPONENT_QUERY, { environmentId, componentId }).then((d) => d.loggersByEnvironmentAndComponent),
    enabled: !!environmentId && !!componentId,
  });
}

export interface GqlRuntime {
  runtimeId: string;
  runtimeType: string;
  status: string;
  version: string;
  platformName: string;
  platformVersion: string;
  platformHome: string;
  osName: string;
  osVersion: string;
  registrationTime: string;
  lastHeartbeat: string;
  component?: { displayName: string };
}

const RUNTIMES_QUERY = `
  query GetRuntimes($environmentId: String!, $projectId: String!, $componentId: String!) {
    runtimes(environmentId: $environmentId, projectId: $projectId, componentId: $componentId) {
      runtimeId, runtimeType, status, version,
      platformName, platformVersion, platformHome,
      osName, osVersion, registrationTime, lastHeartbeat
    }
  }`;

export function useRuntimes(envId: string, projectId: string, componentId: string) {
  return useQuery({
    queryKey: ['runtimes', envId, projectId, componentId],
    queryFn: () => gql<{ runtimes: GqlRuntime[] }>(RUNTIMES_QUERY, { environmentId: envId, projectId, componentId }).then((d) => d.runtimes),
    enabled: !!envId && !!projectId && !!componentId,
  });
}

const PROJECT_RUNTIMES_QUERY = `
  query GetProjectRuntimes($environmentId: String!, $projectId: String!) {
    runtimes(environmentId: $environmentId, projectId: $projectId) {
      runtimeId, runtimeType, status, version,
      platformName, platformVersion, platformHome,
      osName, osVersion, registrationTime, lastHeartbeat,
      component { displayName }
    }
  }`;

export function useProjectRuntimes(envId: string, projectId: string) {
  return useQuery({
    queryKey: ['projectRuntimes', envId, projectId],
    queryFn: () => gql<{ runtimes: GqlRuntime[] }>(PROJECT_RUNTIMES_QUERY, { environmentId: envId, projectId }).then((d) => d.runtimes),
    enabled: !!envId && !!projectId,
  });
}

export { RUNTIMES_QUERY, PROJECT_RUNTIMES_QUERY };

export interface GqlArtifactType {
  artifactType: string;
  artifactCount: number;
}

export function useArtifactTypes(componentId: string, envId: string) {
  return useQuery({
    queryKey: ['artifactTypes', componentId, envId],
    queryFn: () =>
      gql<{ componentArtifactTypes: GqlArtifactType[] }>(
        `query ComponentArtifactTypes($componentId: String!, $environmentId: String!) {
          componentArtifactTypes(componentId: $componentId, environmentId: $environmentId) {
            artifactType, artifactCount
          }
        }`,
        { componentId, environmentId: envId },
      ).then((d) => d.componentArtifactTypes),
    enabled: !!componentId && !!envId,
  });
}

// Backend uses camelCase query name: e.g. RestApi → restApisByEnvironmentAndComponent

export interface GqlArtifact {
  name: string;
  [key: string]: unknown;
}

// Maps artifactType to its GraphQL query field name and useful display fields
// `fields` = flat scalar fields, `gqlFields` = full GraphQL selection (including nested)
// fields = card columns, gqlFields = full GraphQL selection (including nested)
const ARTIFACT_QUERY_MAP: Record<string, { queryName: string; field: string; fields: string; gqlFields: string }> = {
  RestApi: {
    queryName: 'restApisByEnvironmentAndComponent',
    field: 'restApisByEnvironmentAndComponent',
    fields: 'name, context, version, state',
    gqlFields: 'name, context, version, state, tracing, statistics, carbonApp, url, runtimes { runtimeId, status }, resources { path, methods }',
  },
  ProxyService: { queryName: 'proxyServicesByEnvironmentAndComponent', field: 'proxyServicesByEnvironmentAndComponent', fields: 'name, state', gqlFields: 'name, state, tracing, statistics, carbonApp, endpoints, runtimes { runtimeId, status }' },
  Endpoint: { queryName: 'endpointsByEnvironmentAndComponent', field: 'endpointsByEnvironmentAndComponent', fields: 'name, type, state', gqlFields: 'name, type, state, tracing, statistics, attributes { name, value }, runtimes { runtimeId, status }' },
  InboundEndpoint: {
    queryName: 'inboundEndpointsByEnvironmentAndComponent',
    field: 'inboundEndpointsByEnvironmentAndComponent',
    fields: 'name, protocol',
    gqlFields: 'name, protocol, sequence, onError, state, tracing, statistics, carbonApp, runtimes { runtimeId, status }',
  },
  Sequence: { queryName: 'sequencesByEnvironmentAndComponent', field: 'sequencesByEnvironmentAndComponent', fields: 'name, type, container, state', gqlFields: 'name, type, container, state, tracing, statistics, runtimes { runtimeId, status }' },
  Task: { queryName: 'tasksByEnvironmentAndComponent', field: 'tasksByEnvironmentAndComponent', fields: 'name, group, state', gqlFields: 'name, class, group, state, carbonApp, runtimes { runtimeId, status }' },
  LocalEntry: { queryName: 'localEntriesByEnvironmentAndComponent', field: 'localEntriesByEnvironmentAndComponent', fields: 'name, type', gqlFields: 'name, type, value, state, runtimes { runtimeId, status }' },
  CarbonApp: { queryName: 'carbonAppsByEnvironmentAndComponent', field: 'carbonAppsByEnvironmentAndComponent', fields: 'name, version', gqlFields: 'name, version, state, artifacts { name, type }, runtimes { runtimeId, status }' },
  Connector: { queryName: 'connectorsByEnvironmentAndComponent', field: 'connectorsByEnvironmentAndComponent', fields: 'name, package, state', gqlFields: 'name, package, version, state, runtimes { runtimeId, status }' },
  RegistryResource: { queryName: 'registryResourcesByEnvironmentAndComponent', field: 'registryResourcesByEnvironmentAndComponent', fields: 'name, type', gqlFields: 'name, type, runtimes { runtimeId, status }' },
  Listener: { queryName: 'listenersByEnvironmentAndComponent', field: 'listenersByEnvironmentAndComponent', fields: 'name, package, protocol, host, port, state', gqlFields: 'name, package, protocol, host, port, state, runtimes { runtimeId, status }' },
  Service: {
    queryName: 'servicesByEnvironmentAndComponent',
    field: 'servicesByEnvironmentAndComponent',
    fields: 'name, package, basePath, type',
    gqlFields: 'name, package, basePath, type, runtimes { runtimeId, status }, resources { path, method, url, methods }',
  },
  Automation: {
    queryName: 'automationsByEnvironmentAndComponent',
    field: 'automationsByEnvironmentAndComponent',
    fields: 'packageOrg, packageName, packageVersion',
    gqlFields: 'packageOrg, packageName, packageVersion, runtimeIds, runtimes { runtimeId, status, executionTimestamps }, executionTimestamp',
  },
  MessageStore: {
    queryName: 'messageStoresByEnvironmentAndComponent',
    field: 'messageStoresByEnvironmentAndComponent',
    fields: 'name, type, size',
    gqlFields: 'name, type, size, carbonApp, runtimes { runtimeId, status }',
  },
  MessageProcessor: {
    queryName: 'messageProcessorsByEnvironmentAndComponent',
    field: 'messageProcessorsByEnvironmentAndComponent',
    fields: 'name, type, state',
    gqlFields: 'name, type, state, tracing, carbonApp, runtimes { runtimeId, status }',
  },
  Template: {
    queryName: 'templatesByEnvironmentAndComponent',
    field: 'templatesByEnvironmentAndComponent',
    fields: 'name, type',
    gqlFields: 'name, type, tracing, statistics, carbonApp, runtimes { runtimeId, status }',
  },
  DataService: {
    queryName: 'dataServicesByEnvironmentAndComponent',
    field: 'dataServicesByEnvironmentAndComponent',
    fields: 'name, state',
    gqlFields: 'name, description, state, carbonApp, runtimes { runtimeId, status }',
  },
  DataSource: {
    queryName: 'dataSourcesByEnvironmentAndComponent',
    field: 'dataSourcesByEnvironmentAndComponent',
    fields: 'name, type, state',
    gqlFields: 'name, type, driver, url, username, state, runtimes { runtimeId, status }',
  },
};

export function useArtifacts(artifactType: string, envId: string, componentId: string, options?: { enabled?: boolean }) {
  const mapping = ARTIFACT_QUERY_MAP[artifactType];
  return useQuery({
    queryKey: ['artifacts', artifactType, envId, componentId],
    queryFn: async () => {
      if (!mapping) return [];
      const data = await gql<Record<string, GqlArtifact[]>>(`query ArtifactQuery($environmentId: String!, $componentId: String!) { ${mapping.field}(environmentId: $environmentId, componentId: $componentId) { ${mapping.gqlFields} } }`, {
        environmentId: envId,
        componentId,
      });
      return data[mapping.field] ?? [];
    },
    enabled: !!artifactType && !!envId && !!componentId && !!mapping && (options?.enabled ?? true),
  });
}

export { ARTIFACT_QUERY_MAP };

// ── Artifact detail panel queries ──

const ARTIFACT_SOURCE_QUERY = `
  query GetArtifactSource($environmentId: String!, $componentId: String!, $artifactType: String!, $artifactName: String!) {
    artifactSourceByComponent(environmentId: $environmentId, componentId: $componentId, artifactType: $artifactType, artifactName: $artifactName)
  }`;

export function useArtifactSource(envId: string, componentId: string, artifactType: string, artifactName: string) {
  return useQuery({
    queryKey: ['artifactSource', envId, componentId, artifactType, artifactName],
    queryFn: () =>
      gql<{ artifactSourceByComponent: string }>(ARTIFACT_SOURCE_QUERY, {
        environmentId: envId,
        componentId,
        artifactType,
        artifactName,
      }).then((d) => d.artifactSourceByComponent),
    enabled: !!envId && !!componentId && !!artifactType && !!artifactName,
  });
}

const LOCAL_ENTRY_VALUE_QUERY = `
  query LocalEntryValue($componentId: String!, $entryName: String!, $environmentId: String) {
    localEntryValueByComponent(componentId: $componentId, entryName: $entryName, environmentId: $environmentId)
  }`;

export function useLocalEntryValue(componentId: string, entryName: string, envId: string) {
  return useQuery({
    queryKey: ['localEntryValue', componentId, entryName, envId],
    queryFn: () =>
      gql<{ localEntryValueByComponent: string }>(LOCAL_ENTRY_VALUE_QUERY, {
        componentId,
        entryName,
        environmentId: envId,
      }).then((d) => d.localEntryValueByComponent),
    enabled: !!componentId && !!entryName && !!envId,
  });
}

// Maps display artifactType to the backend "type" param used in artifactSourceByComponent
export const ARTIFACT_TYPE_TO_SOURCE_TYPE: Record<string, string> = {
  RestApi: 'api',
  ProxyService: 'proxy-service',
  Endpoint: 'endpoint',
  InboundEndpoint: 'inbound-endpoint',
  Sequence: 'sequence',
  Task: 'task',
  LocalEntry: 'local-entry',
  CarbonApp: 'carbon-app',
  Connector: 'connector',
  RegistryResource: 'registry-resource',
  Listener: 'listener',
  Service: 'service',
  Automation: 'automation',
};

export interface GqlArtifactParam {
  name: string;
  value: string;
}

const ARTIFACT_PARAMS_QUERY = `
  query ArtifactParams($componentId: String!, $artifactType: String!, $artifactName: String!, $environmentId: String, $runtimeId: String) {
    artifactParametersByComponent(
      componentId: $componentId,
      artifactType: $artifactType,
      artifactName: $artifactName,
      environmentId: $environmentId,
      runtimeId: $runtimeId
    ) {
      name
      value
    }
  }`;

export function useArtifactParams(componentId: string, artifactType: string, artifactName: string, envId: string, runtimeId?: string) {
  return useQuery({
    queryKey: ['artifactParams', componentId, artifactType, artifactName, envId, runtimeId],
    queryFn: () =>
      gql<{ artifactParametersByComponent: GqlArtifactParam[] }>(ARTIFACT_PARAMS_QUERY, {
        componentId,
        artifactType,
        artifactName,
        environmentId: envId,
        runtimeId,
      }).then((d) => d.artifactParametersByComponent),
    enabled: !!componentId && !!artifactType && !!artifactName && !!envId,
  });
}

const ARTIFACT_WSDL_QUERY = `
  query ArtifactWsdl($componentId: String!, $artifactType: String!, $artifactName: String!, $environmentId: String, $runtimeId: String) {
    artifactWsdlByComponent(
      componentId: $componentId,
      artifactType: $artifactType,
      artifactName: $artifactName,
      environmentId: $environmentId,
      runtimeId: $runtimeId
    )
  }`;

export function useArtifactWsdl(componentId: string, artifactType: string, artifactName: string, envId: string, runtimeId?: string) {
  return useQuery({
    queryKey: ['artifactWsdl', componentId, artifactType, artifactName, envId, runtimeId],
    queryFn: () =>
      gql<{ artifactWsdlByComponent: string }>(ARTIFACT_WSDL_QUERY, {
        componentId,
        artifactType,
        artifactName,
        environmentId: envId,
        runtimeId,
      }).then((d) => d.artifactWsdlByComponent),
    enabled: !!componentId && !!artifactType && !!artifactName && !!envId,
  });
}

// ── Refresh environment artifacts ──

export function useRefreshEnvironmentArtifacts() {
  const qc = useQueryClient();

  return (envId: string, componentId: string) => {
    return Promise.all([
      qc.invalidateQueries({
        queryKey: ['artifacts'],
        predicate: (query) => {
          const [, , envIdKey, compIdKey] = query.queryKey;
          return envIdKey === envId && compIdKey === componentId;
        },
      }),
      qc.invalidateQueries({
        queryKey: ['artifactTypes', componentId, envId],
      }),
    ]);
  };
}
