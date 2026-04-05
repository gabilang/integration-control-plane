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

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { authenticatedFetch, refreshAccessToken } from '../auth/tokenManager';
import type { GqlArtifact, GqlComponent, GqlEnvironment } from './queries';
import { icpClient } from './client';
import { mapComponent, mapEnvironment, mapProject, type BffComponent, type BffEnvironment, type BffProject, type SchemaConfigItem } from './queries';
import { toBackendArtifactType } from './artifactToggleMutations';
import type { DeployComponentInput, UpdateBuildpackConfigsInput } from '../types/build';

export interface CreateProjectInput {
  name: string;
  handler: string;
  description: string;
  orgHandler: string;
}

export function useCreateProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateProjectInput) =>
      icpClient
        .post<BffProject>('/projects', {
          name: input.handler,        // K8s resource name (slug)
          displayName: input.name,    // human-readable display name
          description: input.description,
          deploymentPipeline: 'default',
        })
        .then(mapProject),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['projects'] }),
  });
}

// ── Environment CRUD ──

export interface EnvironmentInput {
  name: string;
  description: string;
  critical: boolean;
}

export function useCreateEnvironment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: EnvironmentInput) =>
      icpClient.post<BffEnvironment>('/environments', { name: input.name, displayName: input.name, description: input.description, isProduction: input.critical }).then(mapEnvironment),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['environments'] }),
  });
}

export function useUpdateEnvironment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: EnvironmentInput & { environmentId: string }) =>
      icpClient.put<BffEnvironment>(`/environments/${encodeURIComponent(input.environmentId)}`, { displayName: input.name, isProduction: input.critical }).then(mapEnvironment),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['environments'] }),
  });
}

export function useDeleteEnvironment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (environmentId: string) => icpClient.delete<void>(`/environments/${encodeURIComponent(environmentId)}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['environments'] }),
  });
}

export function useDeleteRuntime() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ runtimeId }: { runtimeId: string; envId: string; projectId: string }) =>
      icpClient.delete<void>(`/runtimes/${encodeURIComponent(runtimeId)}`),
    onSuccess: (_, { envId, projectId }) => {
      qc.invalidateQueries({ queryKey: ['runtimes'] });
      qc.invalidateQueries({ queryKey: ['projectRuntimes', envId, projectId] });
    },
  });
}

// ── Artifact status toggle ──

export interface ArtifactStatusInput {
  envId: string;
  componentId: string;
  artifactType: string;
  artifactName: string;
  status: 'active' | 'inactive';
}

export interface ListenerStateInput {
  runtimeIds: string[];
  listenerName: string;
  action: 'START' | 'STOP';
}

// ── Component CRUD ──

export interface CreateComponentInput {
  displayName: string;
  name: string;
  description: string;
  orgHandler: string;
  projectId: string;
  buildpackType: 'MI' | 'BI';
  componentType: string;
}

interface CreateComponentResponse {
  component: BffComponent;
}

// Maps frontend componentType to OpenChoreo ClusterComponentType name.
const COMPONENT_TYPE_SPEC: Record<string, string> = {
  automation: 'deployment/scheduled-task',
  service: 'deployment/service',
  aiAgent: 'deployment/ai-agent',
  eventIntegration: 'deployment/event-integration',
  fileIntegration: 'deployment/file-integration',
  proxy: 'deployment/proxy',
};

// Maps frontend buildpackType to OpenChoreo ClusterWorkflow name.
const WORKFLOW_NAME: Record<string, string> = {
  MI: 'mi-buildpack-builder',
  BI: 'ballerina-buildpack-builder',
};

export function useCreateComponent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateComponentInput) =>
      icpClient
        .post<CreateComponentResponse>('/components', {
          metadata: {
            name: input.name,
            annotations: {
              'core.choreo.dev/display-name': input.displayName,
              'core.choreo.dev/description': input.description,
            },
          },
          spec: {
            owner: { projectName: input.projectId },
            componentType: {
              kind: 'ClusterComponentType',
              name: COMPONENT_TYPE_SPEC[input.componentType] ?? input.componentType,
            },
            autoBuild: true,
            autoDeploy: false,
            workflow: {
              kind: 'ClusterWorkflow',
              name: WORKFLOW_NAME[input.buildpackType] ?? input.buildpackType,
            },
          },
        })
        .then((d) => mapComponent(d.component)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['components'] }),
  });
}

export function useDeleteComponent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: { orgHandler: string; componentId: string; projectId: string }) =>
      icpClient.delete<void>(`/components/${encodeURIComponent(input.componentId)}?projectName=${encodeURIComponent(input.projectId)}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['components'] }),
  });
}

export function useUpdateArtifactStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: ArtifactStatusInput) =>
      icpClient.put<{ status: string; message: string }>('/artifacts/status', {
        componentId: input.componentId,
        artifactType: toBackendArtifactType(input.artifactType),
        artifactName: input.artifactName,
        status: input.status,
      }),
    onMutate: async (input) => {
      const scope = (q: { queryKey: readonly unknown[] }) => q.queryKey[2] === input.envId && q.queryKey[3] === input.componentId;
      await qc.cancelQueries({ queryKey: ['artifacts', input.artifactType], predicate: scope });
      const previousArtifacts = qc.getQueriesData<GqlArtifact[]>({ queryKey: ['artifacts', input.artifactType], predicate: scope });
      const newState = input.status === 'active' ? 'enabled' : 'disabled';
      qc.setQueriesData<GqlArtifact[]>({ queryKey: ['artifacts', input.artifactType], predicate: scope }, (old) => old?.map((a) => (a.name === input.artifactName ? { ...a, state: newState } : a)));
      return { previousArtifacts, scope };
    },
    onError: (_err, input, context) => {
      if (context?.previousArtifacts) {
        for (const [queryKey, data] of context.previousArtifacts) {
          qc.setQueryData<GqlArtifact[]>(queryKey, data);
        }
      }
    },
    onSettled: (_data, _err, input) => {
      const scope = (q: { queryKey: readonly unknown[] }) => q.queryKey[2] === input.envId && q.queryKey[3] === input.componentId;
      qc.invalidateQueries({ queryKey: ['artifacts', input.artifactType], predicate: scope });
    },
  });
}

export function useUpdateListenerState() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: ListenerStateInput) =>
      icpClient.put<{ success: boolean; message: string; commandIds: string[] }>('/artifacts/listener-state', {
        runtimeIds: input.runtimeIds,
        listenerName: input.listenerName,
        action: input.action,
      }),
    onSuccess: () => {
      // Invalidate all listener queries to refetch the updated state
      qc.invalidateQueries({ queryKey: ['artifacts', 'Listener'] });
    },
  });
}

// ── Logger mutations ──

export interface UpdateLogLevelInput {
  runtimeIds: string[];
  componentName: string;
  logLevel: 'INFO' | 'DEBUG' | 'WARN' | 'ERROR';
}

export function useUpdateLogLevel() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: UpdateLogLevelInput) =>
      icpClient.put<{ success: boolean; message: string; commandIds: string[] }>(
        `/components/${encodeURIComponent(input.componentName)}/loggers`,
        { runtimeIds: input.runtimeIds, componentName: input.componentName, logLevel: input.logLevel },
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['loggers'] });
    },
  });
}

// ── Component-Environment JWT Secrets ──

export function useGenerateComponentEnvironmentJwtSecret() {
  return useMutation({
    mutationFn: ({ componentId, environmentId }: { componentId: string; environmentId: string }) =>
      icpClient
        .post<{ secret: string }>(`/components/${encodeURIComponent(componentId)}/environments/${encodeURIComponent(environmentId)}/jwt-secret`)
        .then((d) => d.secret),
  });
}

export function useRotateComponentEnvironmentJwtSecret() {
  return useMutation({
    mutationFn: ({ componentId, environmentId }: { componentId: string; environmentId: string }) =>
      icpClient
        .put<{ secret: string }>(`/components/${encodeURIComponent(componentId)}/environments/${encodeURIComponent(environmentId)}/jwt-secret/rotate`)
        .then((d) => d.secret),
  });
}

// ── Schedule / Job Configs ──

export interface UpdateJobConfigsInput {
  orgHandler: string;
  componentId: string;
  environmentId: string;
  versionId: string;
  cronFrequency?: string;
  cronTimezone?: string;
  jobTimeoutSeconds?: number;
  cronJobAllowConcurrency?: boolean;
  jobRetryCount?: number;
}

export function useUpdateJobConfigs() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: UpdateJobConfigsInput) =>
      icpClient.put<{ success: boolean }>(`/components/${encodeURIComponent(input.componentId)}/job-configs`, input).then((d) => d.success),
    onSuccess: (_data, input) => {
      qc.invalidateQueries({ queryKey: ['executionConfigs', input.componentId] });
    },
  });
}

// ── Task trigger ──

export interface TriggerTaskInput {
  componentId: string;
  taskName: string;
}

export function useTriggerTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: TriggerTaskInput) =>
      icpClient.post<{ status: string; message: string; successCount: number; failedCount: number; details: string[] }>('/artifacts/trigger', {
        componentId: input.componentId,
        taskName: input.taskName,
      }),
    onSuccess: () => {
      // Invalidate task queries to refetch the updated state
      qc.invalidateQueries({ queryKey: ['artifacts', 'Task'] });
    },
  });
}

// ── Deploy deployment track (triggers automation execution) ──

export interface DeployDeploymentTrackInput {
  componentId: string;
  projectName?: string;
  id: string;
  imageId: string;
  environmentId: string;
  deploymentPipelineId: string;
  cronTimezone?: string;
  cron?: string;
  jobTimeoutSeconds?: number;
  cronJobAllowConcurrency?: boolean;
}

export function useDeployDeploymentTrack() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: DeployDeploymentTrackInput) =>
      icpClient.post<string>(
        `/components/${encodeURIComponent(input.componentId)}/deployments`,
        input,
        { projectName: input.projectName },
      ),
    onSuccess: (_data, input) => {
      qc.invalidateQueries({ queryKey: ['deploymentStatus', input.componentId, input.id] });
      qc.invalidateQueries({ queryKey: ['executionConfigs', input.componentId] });
      qc.invalidateQueries({ queryKey: ['componentDeployment'] });
    },
  });
}

// ── Build trigger (deploy component) ──

const DEPLOY_COMPONENT = `
  mutation DeployComponent(
    $componentId: String!
    $versionId: String!
    $envId: String!
    $sha: String!
    $branch: String
    $shaDate: String
    $gitRefType: String
    $cron: String
    $cronTimezone: String
    $commitTag: String
  ) {
    deployComponent(deployment: {
      componentId: $componentId,
      versionId: $versionId,
      envId: $envId,
      sha: $sha,
      branch: $branch,
      shaDate: $shaDate,
      gitRefType: $gitRefType,
      cron: $cron,
      cronTimezone: $cronTimezone,
      commitTag: $commitTag
    }) {
      message
      success
    }
  }`;

export function useTriggerBuild() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: DeployComponentInput) =>
      gql<{ deployComponent: { message: string; success: boolean } }>(DEPLOY_COMPONENT, {
        componentId: input.componentId,
        versionId: input.versionId,
        envId: input.envId,
        sha: input.sha,
        branch: input.branch,
        shaDate: input.shaDate ?? '',
        gitRefType: input.gitRefType,
        cron: input.cron ?? null,
        cronTimezone: input.cronTimezone ?? '',
        commitTag: input.commitTag ?? null,
      }).then((d) => d.deployComponent),
    onSuccess: (_data, input) => {
      // Immediately refresh build history after a build is triggered
      qc.invalidateQueries({ queryKey: ['deploymentStatus', input.componentId, input.versionId] });
    },
  });
}

// ── Deploy to environment (auto-deploy after build success) ──

export interface DeployComponentInput {
  componentId: string;
  projectName: string;
  environment?: string;
}

export function useDeployComponent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: DeployComponentInput) =>
      icpClient.post<{ releaseId: string; cron: string; cronTimezone: string }>(
        `/components/${encodeURIComponent(input.componentId)}/deploy`,
        {},
        { projectName: input.projectName, environment: input.environment ?? 'development' },
      ),
    onSuccess: (_data, input) => {
      qc.invalidateQueries({ queryKey: ['componentDeployment'] });
      qc.invalidateQueries({ queryKey: ['executionConfigs', input.componentId] });
    },
  });
}

// ── Promote ──

export interface PromoteInput {
  componentId: string;
  projectName: string;
  apiVersionId: string;
  sourceReleaseId: string;
  targetEnvironmentId: string;
  deploymentPipelineId: string;
}

export function usePromote() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: PromoteInput) =>
      icpClient.post<string>(
        `/components/${encodeURIComponent(input.componentId)}/deployments/promote`,
        {
          apiVersionId: input.apiVersionId,
          sourceReleaseId: input.sourceReleaseId,
          targetEnvironmentId: input.targetEnvironmentId,
          deploymentPipelineId: input.deploymentPipelineId,
        },
        { projectName: input.projectName },
      ),
    onSuccess: (_data, input) => {
      qc.invalidateQueries({ queryKey: ['componentDeployment'] });
      qc.invalidateQueries({ queryKey: ['deploymentStatus', input.componentId] });
    },
  });
}

// ── Stop Deployment (clears cron schedule) ──

export interface StopDeploymentInput {
  orgHandler: string;
  componentId: string;
  releaseId: string;
  environment: string;
}

export function useStopDeployment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: StopDeploymentInput) =>
      icpClient.delete<string>(
        `/components/${encodeURIComponent(input.componentId)}/deployments`,
        { orgHandler: input.orgHandler, componentId: input.componentId, releaseId: input.releaseId, environment: input.environment },
      ),
    onSuccess: (_data, input) => {
      qc.invalidateQueries({ queryKey: ['executionConfigs', input.componentId] });
      qc.invalidateQueries({ queryKey: ['componentDeployment'] });
    },
  });
}

// ── Update component display name ──

export interface UpdateComponentInput {
  id: string;
  displayName: string;
  description: string;
  version: string;
  projectId: string;
  handler: string;
  labels?: string;
}

export function useUpdateComponent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: UpdateComponentInput) =>
      icpClient
        .put<BffComponent>(`/components/${encodeURIComponent(input.id)}?projectName=${encodeURIComponent(input.projectId)}`, {
          displayName: input.displayName,
          description: input.description,
        })
        .then(mapComponent),
    onSuccess: (_data, input) => {
      qc.invalidateQueries({ queryKey: ['component', input.projectId, input.handler] });
      qc.invalidateQueries({ queryKey: ['components'] });
    },
  });
}

// ── Buildpack build configuration ──

const UPDATE_BUILDPACK_BUILD_CONFIGURATIONS = `
  mutation updateBuildpackBuildConfigurations(
    $componentId: String!,
    $buildContext: String,
    $versionId: String!,
    $languageVersion: String,
    $environmentVariables: [BuildpackBuildEnvironmentVariable!],
    $isUnitTestEnabled: Boolean,
    $pullLatestSubmodules: Boolean,
    $enableTrivyScan: Boolean
  ) {
    updateBuildpackBuildConfigurations(input: {
      componentId: $componentId,
      buildContext: $buildContext,
      versionId: $versionId,
      languageVersion: $languageVersion,
      environmentVariables: $environmentVariables,
      isUnitTestEnabled: $isUnitTestEnabled,
      pullLatestSubmodules: $pullLatestSubmodules,
      enableTrivyScan: $enableTrivyScan
    })
  }`;

export function useUpdateBuildpackConfigs() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: UpdateBuildpackConfigsInput) =>
      gql<{ updateBuildpackBuildConfigurations: string }>(UPDATE_BUILDPACK_BUILD_CONFIGURATIONS, {
        componentId: input.componentId,
        versionId: input.versionId,
        buildContext: input.buildContext,
        languageVersion: input.languageVersion,
        environmentVariables: input.environmentVariables,
        isUnitTestEnabled: input.isUnitTestEnabled,
        pullLatestSubmodules: input.pullLatestSubmodules,
        enableTrivyScan: input.enableTrivyScan,
      }).then((d) => d.updateBuildpackBuildConfigurations),
    onSuccess: (_data, input) => {
      qc.invalidateQueries({ queryKey: ['componentRepository'] });
      qc.invalidateQueries({ queryKey: ['deploymentStatus', input.componentId] });
    },
  });
}

// ── Run-pod trigger (manual execution with optional arguments) ──

export interface SaveSchemaConfigInput {
  projectId: string;
  componentId: string;
  envId: string;
  deploymentTrackId: string;
  configurations: SchemaConfigItem[];
  mappingId?: string;
  commitHash?: string;
}

export function useSaveSchemaConfig() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (input: SaveSchemaConfigInput) => {
      const base = new URL(window.API_CONFIG.graphqlUrl).origin;
      const url = `${base}/configuration-schema/v1.0/projects/${input.projectId}/components/${input.componentId}/env-template/${input.envId}/deployment-track/${input.deploymentTrackId}/configurations`;
      const res = await authenticatedFetch(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ configurations: input.configurations, ...(input.commitHash ? { commitHash: input.commitHash } : {}) }),
      });
      if (!res.ok) {
        const text = await res.text().catch(() => '');
        throw new Error(text || `HTTP ${res.status}`);
      }
      return res.json().catch(() => ({}));
    },
    onSuccess: (_, vars) => {
      qc.invalidateQueries({ queryKey: ['schemaConfig', vars.projectId, vars.componentId, vars.envId, vars.deploymentTrackId] });
    },
  });
}

export interface TriggerComponentInput {
  orgHandler: string;
  projectId: string;
  componentId: string;
  releaseId: string;
  args?: { argument_name: string; argument_value: string }[];
}

async function runPod(url: string, args: TriggerComponentInput['args']): Promise<Response> {
  return authenticatedFetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ args: args ?? [] }),
  });
}

export function useTriggerComponent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (input: TriggerComponentInput) => {
      const origin = new URL(window.API_CONFIG.graphqlUrl).origin;
      const url = `${origin}/component-mgt/1.0.0/orgs/${input.orgHandler}/projects/${input.projectId}/components/${input.componentId}/releases/${input.releaseId}/run-pod`;
      let res = await runPod(url, input.args);

      // A 403 with scope validation failure means the cached token was issued before
      // component_trigger was added to STS_SCOPE. Force a refresh and retry once.
      if (res.status === 403) {
        const text = await res.text().catch(() => '');
        let isScopeError = false;
        try {
          const parsed = JSON.parse(text);
          isScopeError = parsed?.code === '900910' || !!parsed?.error_description?.includes('Scope validation');
        } catch {
          /* not JSON */
        }

        if (isScopeError) {
          await refreshAccessToken();
          res = await runPod(url, input.args);
        } else {
          throw new Error('Permission denied (403)');
        }
      }

      if (!res.ok) {
        const text = await res.text().catch(() => '');
        if (res.status === 403) {
          throw new Error('You do not have permission to trigger this component');
        }
        throw new Error(text || `HTTP ${res.status}`);
      }
      return res.json().catch(() => ({}));
    },
    onSuccess: (_data, input) => {
      qc.invalidateQueries({ queryKey: ['taskExecutions', input.releaseId] });
    },
  });
}
