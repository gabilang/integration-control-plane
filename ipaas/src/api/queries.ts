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
import { authenticatedFetch } from '../auth/tokenManager';
import { icpClient } from './client';
import { env } from '../config/env';
import { choreoDevopsApiUrl } from '../config/api';

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

export interface BffDeploymentTrack {
  id: string;
  branch?: string;
  commitSha?: string;
  url?: string;
  appPath?: string;
  componentId?: string;
  latest?: boolean;
  autoDeployEnabled?: boolean;
  createdAt?: string;
  updatedAt?: string;
}

export interface BffComponent {
  uid?: string;
  name: string;
  projectName?: string;
  displayName?: string;
  description?: string;
  buildpackType?: string;
  componentType?: string;
  displayType?: string;
  createdAt?: string;
  status?: string;
  deploymentTracks?: BffDeploymentTrack[];
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
  defaultDeploymentPipelineId: string;
}

export interface GqlComponent {
  projectId: string;
  id: string;
  name: string;
  handler: string;
  displayName: string;
  buildpackType?: string;
  componentType?: string;
  displayType: string;
  description: string;
  status: string;
  componentSubType: string | null;
  version: string;
  createdAt: string;
  lastBuildDate: string;
  labels?: string | string[];
  apiId?: string;
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
    defaultDeploymentPipelineId: '',
  };
}

export function mapComponent(c: BffComponent): GqlComponent {
  return {
    id: c.name,
    projectId: c.projectName ?? '',
    name: c.displayName || c.name,
    handler: c.name,
    displayName: c.displayName ?? c.name,
    buildpackType: c.buildpackType,
    componentType: c.componentType,
    displayType: c.displayType ?? '',
    description: c.description ?? '',
    status: c.status ?? '',
    componentSubType: null,
    version: '',
    createdAt: c.createdAt ?? '',
    lastBuildDate: '',
    labels: [],
    apiId: undefined,
  };
}

// // ── Org / project hooks ──
// const PROJECT_FIELDS = 'id, orgId, name, handler, description, version, createdDate, updatedAt, region, type, defaultDeploymentPipelineId';

// const PROJECTS_QUERY = `
//   query GetProjects($orgId: Int!) {
//     projects(orgId: $orgId) { ${PROJECT_FIELDS} }
//   }`;

// const PROJECT_QUERY = `
//   query GetProject($orgId: Int!, $projectId: String!) {
//     project(orgId: $orgId, projectId: $projectId) { ${PROJECT_FIELDS} }
//   }`;

// const PROJECT_BY_HANDLER_QUERY = `
//   query GetProjectByHandler($orgId: Int!, $projectHandler: String!) {
//     projectByHandler(orgId: $orgId, projectHandler: $projectHandler) { ${PROJECT_FIELDS} }
//   }`;

// // componentType excluded as it is not in the schema
// const COMPONENTS_QUERY = `
//   query GetComponents($orgHandler: String!, $projectId: String!) {
//     components(orgHandler: $orgHandler, projectId: $projectId) {
//       projectId, id, name, handler, displayName, displayType, description, status, componentSubType, version, createdAt, lastBuildDate
//     }
//   }`;

function orgId(): number {
  return env.ICP_ORG_NUMERIC_ID;
}

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

// export function useProjects() {
//   const id = orgId();
//   return useQuery({
//     queryKey: ['projects', id],
//     queryFn: () => gql<{ projects: GqlProject[] }>(PROJECTS_QUERY, { orgId: id }).then((d) => d.projects),
//     enabled: id > 0,
//   });
// }

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

export interface ProjectContributor {
  id: number;
  displayName: string;
  email: string;
  pictureUrl: string | null;
  totalContributions: number;
}

const PROJECT_CONTRIBUTORS_QUERY = `
  query GetProjectContributors($orgId: Int!, $projectId: String!) {
    project(orgId: $orgId, projectId: $projectId) {
      projectContributorsData {
        contributorCount
        contributors { id, pictureUrl, email, displayName, totalContributions }
      }
    }
  }`;

export function useProjectContributors(projectId: string) {
  const id = orgId();
  return useQuery({
    queryKey: ['projectContributors', projectId, id],
    queryFn: () =>
      gql<{ project: { projectContributorsData: { contributorCount: number; contributors: ProjectContributor[] } } }>(PROJECT_CONTRIBUTORS_QUERY, { orgId: id, projectId })
        .then((d) => d.project?.projectContributorsData?.contributors ?? [])
        .catch(() => []),
    enabled: !!projectId && id > 0,
    staleTime: 5 * 60 * 1000,
  });
}

export interface CloudDataPlane {
  id: string;
  external_gateway_virtual_host: string;
  internal_gateway_virtual_host: string;
  region: string;
  is_cilium?: boolean;
}

export function useCloudDataPlanes(orgUuid: string) {
  return useQuery({
    queryKey: ['cloud-data-planes', orgUuid],
    queryFn: async () => {
      const res = await authenticatedFetch(`${choreoDevopsApiUrl()}/api/v1/clusters/clouddataplanes?org_uuid=${encodeURIComponent(orgUuid)}`);
      if (!res.ok) throw new Error(`Failed to fetch cloud data planes: ${res.status}`);
      return res.json() as Promise<CloudDataPlane[]>;
    },
    enabled: !!orgUuid,
    staleTime: 5 * 60 * 1000,
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

export interface GqlApiVersion {
  id: string;
  apiVersion: string;
  branch: string;
  latest: boolean;
  accessibility?: string;
}

export interface GqlComponentDetail extends GqlComponent {
  orgHandler: string;
  deploymentTracks?: { id: string }[];
  apiVersions?: GqlApiVersion[];
}

const COMPONENT_BY_HANDLER_QUERY = `
  query GetComponent($projectId: String!, $componentHandler: String!) {
    component(projectId: $projectId, componentHandler: $componentHandler) {
      projectId, id, name, handler, displayName, displayType,
      description, status, componentSubType,
      version, createdAt, lastBuildDate, orgHandler, labels, apiId,
      deploymentTracks { id }
      apiVersions { id, apiVersion, branch, latest, accessibility }
    }
  }`;

export function useComponentByHandler(projectName: string, handler: string | undefined) {
  return useQuery({
    queryKey: ['component', projectName, handler],
    queryFn: () =>
      icpClient.get<BffComponentList>('/components', { projectName }).then((d) => {
        const c = d.items.find((x) => x.name === handler);
        if (!c) throw new Error(`Component '${handler}' not found in project '${projectName}'`);
        return {
          ...mapComponent(c),
          orgHandler: '',
          deploymentTracks: (c.deploymentTracks ?? []).map((t) => ({ id: t.id })),
        } as GqlComponentDetail;
      }),
    enabled: !!projectName && !!handler,
  });
}

export function useProjectComponentLabels(projectId: string) {
  const id = orgId();
  return useQuery({
    queryKey: ['projectComponentLabels', projectId],
    queryFn: () =>
      icpClient
        .get<{ items: string[] }>(`/components/${encodeURIComponent(projectId)}/labels`, { projectName: projectId, orgId: String(id) })
        .then((d) => d.items ?? []),
    enabled: !!projectId && id > 0,
    staleTime: 5 * 60 * 1000,
  });
}

export interface BffEnvironment {
  uid?: string;
  name: string;
  displayName?: string;
  dataPlaneRef?: string;
  isProduction?: boolean;
  createdAt?: string;
}

export interface BffEnvironmentList {
  items: BffEnvironment[];
}

export interface GqlEnvironment {
  id: string;
  name: string;
  critical: boolean;
  templateId?: string;
  dpId?: string;
  description?: string;
  createdAt?: string;
}

export function mapEnvironment(e: BffEnvironment): GqlEnvironment {
  return {
    id: e.name,
    name: e.displayName || e.name,
    critical: e.isProduction ?? false,
    dpId: e.dataPlaneRef,
    createdAt: e.createdAt,
  };
}

export function useEnvironments(orgUuid: string, projectId: string) {
  // const effectiveOrgUuid = getOrgUuidFromToken() ?? orgUuid;
  return useQuery({
    queryKey: ['environments'],
    queryFn: () => icpClient.get<BffEnvironmentList>('/environments').then((d) => d.items.map(mapEnvironment)),
    enabled: !!orgUuid && !!projectId,
  });
}

export function useAllEnvironments() {
  return useQuery({
    queryKey: ['environments'],
    queryFn: () => icpClient.get<BffEnvironmentList>('/environments').then((d) => d.items.map(mapEnvironment)),
    retry: false,
  });
}


export interface GqlLogger {
  componentName: string;
  logLevel: string;
  runtimeIds: string[];
}

export function useLoggers(environmentId: string, componentId: string) {
  return useQuery({
    queryKey: ['loggers', environmentId, componentId],
    queryFn: () =>
      icpClient
        .get<{ items: GqlLogger[] }>(`/components/${encodeURIComponent(componentId)}/loggers`, { environmentId })
        .then((d) => d.items ?? []),
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

export function useRuntimes(envId: string, projectId: string, componentId: string) {
  return useQuery({
    queryKey: ['runtimes', envId, projectId, componentId],
    queryFn: () =>
      icpClient
        .get<{ items: GqlRuntime[] }>(`/components/${encodeURIComponent(componentId)}/runtimes`, { environmentId: envId, projectName: projectId })
        .then((d) => d.items ?? []),
    enabled: !!envId && !!projectId && !!componentId,
  });
}

export function useProjectRuntimes(envId: string, projectId: string) {
  return useQuery({
    queryKey: ['projectRuntimes', envId, projectId],
    queryFn: () =>
      icpClient
        .get<{ items: GqlRuntime[] }>(`/projects/${encodeURIComponent(projectId)}/runtimes`, { environmentId: envId })
        .then((d) => d.items ?? []),
    enabled: !!envId && !!projectId,
  });
}

export interface GqlArtifactType {
  artifactType: string;
  artifactCount: number;
}

export function useArtifactTypes(componentId: string, envId: string) {
  return useQuery({
    queryKey: ['artifactTypes', componentId, envId],
    queryFn: () =>
      icpClient
        .get<{ items: GqlArtifactType[] }>('/artifacts/types', { componentId, environmentId: envId })
        .then((d) => d.items ?? []),
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
    gqlFields: 'packageOrg, packageName, packageVersion, runtimeIds, runtimes { runtimeId, status }, executionTimestamp',
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
  const known = artifactType in ARTIFACT_QUERY_MAP;
  return useQuery({
    queryKey: ['artifacts', artifactType, envId, componentId],
    queryFn: () =>
      icpClient
        .get<{ items: GqlArtifact[] }>('/artifacts', { artifactType, environmentId: envId, componentId })
        .then((d) => d.items ?? []),
    enabled: known && !!artifactType && !!envId && !!componentId && (options?.enabled ?? true),
  });
}

export { ARTIFACT_QUERY_MAP };

// ── Artifact detail panel queries ──

export function useArtifactSource(envId: string, componentId: string, artifactType: string, artifactName: string) {
  return useQuery({
    queryKey: ['artifactSource', envId, componentId, artifactType, artifactName],
    queryFn: () =>
      icpClient.get<string>('/artifacts/source', { environmentId: envId, componentId, artifactType, artifactName }),
    enabled: !!envId && !!componentId && !!artifactType && !!artifactName,
  });
}

export function useLocalEntryValue(componentId: string, entryName: string, envId: string) {
  return useQuery({
    queryKey: ['localEntryValue', componentId, entryName, envId],
    queryFn: () =>
      icpClient.get<string>('/artifacts/local-entry', { componentId, entryName, environmentId: envId }),
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

export function useArtifactParams(componentId: string, artifactType: string, artifactName: string, envId: string, runtimeId?: string) {
  return useQuery({
    queryKey: ['artifactParams', componentId, artifactType, artifactName, envId, runtimeId],
    queryFn: () =>
      icpClient.get<GqlArtifactParam[]>('/artifacts/params', { componentId, artifactType, artifactName, environmentId: envId, runtimeId }),
    enabled: !!componentId && !!artifactType && !!artifactName && !!envId,
  });
}

export function useArtifactWsdl(componentId: string, artifactType: string, artifactName: string, envId: string, runtimeId?: string) {
  return useQuery({
    queryKey: ['artifactWsdl', componentId, artifactType, artifactName, envId, runtimeId],
    queryFn: () =>
      icpClient.get<string>('/artifacts/wsdl', { componentId, artifactType, artifactName, environmentId: envId, runtimeId }),
    enabled: !!componentId && !!artifactType && !!artifactName && !!envId,
  });
}

// ── Component repository & commit history ──

export interface GqlBuildpackConfig {
  versionId: string;
  buildContext: string;
  isUnitTestEnabled: boolean;
  languageVersion: string;
  pullLatestSubmodules: boolean;
  enableTrivyScan: boolean;
  keyValues?: Array<{ id?: string; key: string; value: string }>;
}

export interface GqlRepository {
  gitProvider: string;
  organizationApp: string;
  nameApp: string;
  branch: string;
  appSubPath: string;
  bitbucketServerUrl?: string;
  serverUrl?: string;
  projectApp?: string;
  isBuildConfigurationMigrated?: boolean;
  buildpackConfig?: GqlBuildpackConfig[];
  treeUrl?: string;
}

export interface GqlCommit {
  sha: string;
  message: string;
  isLatest: boolean;
  author: {
    name: string;
    date: string;
    email: string;
    avatarUrl: string;
  };
}

const COMPONENT_REPOSITORY_QUERY = `
  query GetComponentRepository($projectId: String!, $componentHandler: String!) {
    component(projectId: $projectId, componentHandler: $componentHandler) {
      repository {
        gitProvider, organizationApp, nameApp, branch, appSubPath,
        bitbucketServerUrl, serverUrl, projectApp,
        isBuildConfigurationMigrated,
        buildpackConfig { versionId, buildContext, isUnitTestEnabled, languageVersion, pullLatestSubmodules, enableTrivyScan, keyValues { id, key, value } }
      }
    }
  }`;

export function useComponentRepository(projectId: string, componentHandler: string) {
  return useQuery({
    queryKey: ['componentRepository', projectId, componentHandler],
    queryFn: () =>
      icpClient.get<GqlRepository>(`/components/${encodeURIComponent(componentHandler)}/repository`, { projectName: projectId }),
    enabled: !!projectId && !!componentHandler,
  });
}

export function useCommitHistory(componentId: string, branch: string, projectName?: string) {
  return useQuery({
    queryKey: ['commitHistory', componentId, branch],
    queryFn: () =>
      icpClient
        .get<{ items: GqlCommit[] }>(`/components/${encodeURIComponent(componentId)}/commit-history`, { branch, ...(projectName ? { projectName } : {}) })
        .then((d) => d.items ?? []),
    enabled: !!componentId && !!branch,
  });
}

// ── Execution / Schedule configs ──

export interface GqlExecutionConfigs {
  cronjobFrequency: string;
  cronjobTimezone: string;
  cronjobAllowConcurrency?: boolean;
  timeoutSeconds?: number;
  retryCount?: number;
}

export function useExecutionConfigs(componentId: string, releaseId: string) {
  return useQuery({
    queryKey: ['executionConfigs', componentId, releaseId],
    queryFn: () =>
      icpClient
        .get<GqlExecutionConfigs>(`/components/${encodeURIComponent(componentId)}/execution-configs`, { releaseId })
        .catch(() => null),
    enabled: !!componentId && !!releaseId,
    retry: false,
  });
}

// ── Component Deployment (for real releaseId) ──

export interface GqlComponentDeployment {
  releaseId: string;
  cron: string;
  cronTimezone: string;
  build?: { buildId: string };
}

export function useComponentDeployment(orgHandler: string, orgUuid: string, componentId: string, versionId: string, environmentId: string) {
  return useQuery({
    queryKey: ['componentDeployment', orgHandler, componentId, versionId, environmentId],
    queryFn: () =>
      icpClient
        .get<GqlComponentDeployment>(`/components/${encodeURIComponent(componentId)}/deployments`, { orgHandler, orgUuid, versionId, environmentId })
        .catch(() => null),
    enabled: !!orgHandler && !!componentId && !!environmentId,
    retry: false,
  });
}

// ── Deployment execution history ──

export interface GqlDeploymentStatus {
  id: number;
  sha: string;
  started_at: string;
  completed_at: string;
  status: string;
  conclusion: string;
  conclusionV2: string;
  isAutoDeploy: boolean;
  isTriggeredAtCreation: boolean;
  name: string;
  failureReason: number;
  sourceCommitId: string;
  buildRef?: string;
  tasks?: BffWorkflowTask[];
}

const DEPLOYMENT_STATUS_QUERY = `
  query GetDeploymentStatus($versionId: String!, $componentId: String!) {
    deploymentStatusByVersion(versionId: $versionId, componentId: $componentId) {
      id, sha, started_at, completed_at, status, conclusion, conclusionV2, isAutoDeploy, isTriggeredAtCreation, name, failureReason, sourceCommitId, buildRef
    }
  }`;

export function useDeploymentStatus(componentId: string, versionId: string) {
  return useQuery({
    queryKey: ['deploymentStatus', componentId, versionId],
    queryFn: () =>
      icpClient
        .get<GqlDeploymentStatus[]>(`/components/${encodeURIComponent(componentId)}/deployments/status`, { versionId })
        .catch(() => []),
    enabled: !!componentId && !!versionId,
    retry: false,
    refetchInterval: 15000,
  });
}

// ── Builds (OpenChoreo workflow runs) ──

export interface BffWorkflowTask {
  name: string;
  phase: string;
  startedAt?: string;
  completedAt?: string;
}

export interface BffWorkflowRun {
  name: string;
  status: string;
  startedAt: string;
  completedAt: string;
  componentName: string;
  projectName: string;
  image: string;
  commit: string;
  tasks?: BffWorkflowTask[];
}

interface BffWorkflowRunList {
  items: BffWorkflowRun[];
}

/** Maps OpenChoreo workflow run status to the status/conclusion pair used by BuildCard. */
function mapWorkflowRunToBuildInfo(run: BffWorkflowRun): GqlDeploymentStatus {
  let status = 'queued';
  let conclusion = '';

  switch (run.status) {
    case 'Running':
      status = 'in_progress';
      break;
    case 'Succeeded':
      status = 'completed';
      conclusion = 'success';
      break;
    case 'Failed':
      status = 'completed';
      conclusion = 'failure';
      break;
    case 'Pending':
    default:
      status = 'queued';
      break;
  }

  return {
    id: 0,
    sha: run.commit,
    started_at: run.startedAt,
    completed_at: run.completedAt,
    status,
    conclusion,
    conclusionV2: conclusion,
    isAutoDeploy: false,
    name: run.name,
    failureReason: 0,
    sourceCommitId: run.commit,
    buildRef: run.name,
    tasks: run.tasks,
  };
}

export function useBuilds(componentName: string, projectName: string) {
  return useQuery({
    queryKey: ['builds', componentName, projectName],
    queryFn: () =>
      icpClient
        .get<BffWorkflowRunList>(`/components/${encodeURIComponent(componentName)}/builds`, { projectName })
        .then((d) => (d.items ?? []).map(mapWorkflowRunToBuildInfo))
        .catch(() => []),
    enabled: !!componentName && !!projectName,
    retry: false,
    refetchInterval: 15000,
  });
}

export interface TaskExecution {
  id: string;
  startTime: string;
  completionTime: string;
  runId: string;
  revisionId: string;
  failedReason: string;
  status: string;
  arguments?: string | null;
}

export function useTaskExecutions(releaseId: string) {
  const baseUrl = window.API_CONFIG?.systemApisBaseUrl ?? '';
  return useQuery({
    queryKey: ['taskExecutions', releaseId, baseUrl],
    queryFn: async (): Promise<TaskExecution[]> => {
      if (!baseUrl || !releaseId) return [];
      const url = `${baseUrl}/systemapis/choreoobsapi/0.3.0/tasks/executions?releaseId=${releaseId}&limit=10&verbose=true`;
      const res = await authenticatedFetch(url);
      if (!res.ok) return [];
      return res.json();
    },
    enabled: !!baseUrl && !!releaseId,
    retry: false,
    staleTime: 0,
  });
}

export interface ExecutionArgument {
  argumentName: string;
  argumentValue: string;
}

export function useExecutionArguments(runId: string, componentId: string, releaseId: string, enabled: boolean) {
  return useQuery({
    queryKey: ['executionArguments', runId, componentId, releaseId],
    queryFn: () =>
      icpClient
        .get<ExecutionArgument[]>(`/components/${encodeURIComponent(componentId)}/executions/${encodeURIComponent(runId)}/arguments`, { releaseId })
        .then((d) => d ?? [])
        .catch(() => []),
    enabled: enabled && !!runId && !!componentId && !!releaseId,
    retry: false,
    staleTime: 60000,
  });
}

export interface ExecutionLogEntry {
  timestamp: string;
  message: string;
}

export function useExecutionLogs(componentId: string, deploymentTrackId: string, executionId: string, environmentId: string, enabled: boolean) {
  const baseUrl = window.API_CONFIG?.systemApisBaseUrl ?? '';
  return useQuery({
    queryKey: ['executionLogs', componentId, deploymentTrackId, executionId, environmentId, baseUrl],
    queryFn: async (): Promise<ExecutionLogEntry[]> => {
      if (!baseUrl || !componentId || !deploymentTrackId || !executionId || !environmentId) return [];
      const url = `${baseUrl}/systemapis/choreologgingapi/0.2.0/components/${componentId}/deployment-tracks/${deploymentTrackId}/executions/${executionId}/logs?environmentId=${environmentId}&offset=0&limit=10000`;
      const res = await authenticatedFetch(url);
      if (!res.ok) return [];
      const data: { columns: { name: string }[]; rows: string[][] } = await res.json();
      const logIdx = data.columns?.findIndex((c) => c.name === 'LogEntry') ?? -1;
      const timeIdx = data.columns?.findIndex((c) => c.name === 'TimeGenerated') ?? -1;
      return (data.rows ?? []).map((row) => ({
        timestamp: timeIdx >= 0 ? row[timeIdx] : '',
        message: logIdx >= 0 ? row[logIdx] : (row[0] ?? ''),
      }));
    },
    enabled: enabled && !!baseUrl && !!componentId && !!deploymentTrackId && !!executionId && !!environmentId,
    retry: false,
    staleTime: 30000,
  });
}

export function useTaskExecutionCount(releaseId: string) {
  const baseUrl = window.API_CONFIG?.systemApisBaseUrl ?? '';
  return useQuery({
    queryKey: ['taskExecutionCount', releaseId, baseUrl],
    queryFn: async (): Promise<number | null> => {
      if (!baseUrl || !releaseId) return null;
      const to = new Date();
      const from = new Date(to);
      from.setDate(to.getDate() - 30);
      const url = `${baseUrl}/systemapis/choreoobsapi/0.3.0/tasks/executions/count?releaseId=${releaseId}&from=${from.toISOString()}&to=${to.toISOString()}`;
      const res = await authenticatedFetch(url);
      if (!res.ok) return null;
      const data: { count: number } = await res.json();
      return data.count ?? null;
    },
    enabled: !!baseUrl && !!releaseId,
    retry: false,
    staleTime: 30_000,
    refetchInterval: 30_000,
  });
}

// ── Schema-based configurable values ──

export interface SchemaConfigValue {
  value: string;
  environmentUuid?: string;
}

export interface SchemaConfigItem {
  key: string;
  values: SchemaConfigValue[];
  valueType?: string;
  isRequired?: boolean;
  isSensitive?: boolean;
}

export interface SchemaConfigData {
  jsonSchema?: string; // base64-encoded JSON schema
  mappingId?: string;
  configurations: SchemaConfigItem[];
}

export function useSchemaConfig(projectId: string, componentId: string, envId: string, deploymentTrackId: string, commitHash?: string) {
  return useQuery({
    queryKey: ['schemaConfig', projectId, componentId, envId, deploymentTrackId, commitHash],
    queryFn: async (): Promise<SchemaConfigData | null> => {
      const base = new URL(window.API_CONFIG?.graphqlUrl ?? '').origin;
      const qs = commitHash ? `?commitHash=${encodeURIComponent(commitHash)}` : '';
      const url = `${base}/configuration-schema/v1.0/projects/${projectId}/components/${componentId}/env-template/${envId}/deployment-track/${deploymentTrackId}/configurations${qs}`;
      const res = await authenticatedFetch(url);
      if (!res.ok) return null;
      return res.json();
    },
    enabled: !!projectId && !!componentId && !!envId && !!deploymentTrackId && !!commitHash,
    retry: false,
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
