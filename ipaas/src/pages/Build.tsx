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

import { CircularProgress, PageContent, Typography } from '@wso2/oxygen-ui';
import type { JSX } from 'react';
import { useBuilds, useComponentByHandler, useCommitHistory, useProject, useProjectByHandler, useComponentRepository } from '../api/queries';
import BuildHistory from '../components/Build/BuildHistory';
import type { ComponentScope } from '../nav';

const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

export default function Build(scope: ComponentScope): JSX.Element {
  const isProjectUuid = UUID_RE.test(scope.project);

  const { data: projectByHandler, isLoading: loadingProject } = useProjectByHandler(!isProjectUuid ? scope.project : '');
  const { data: projectByUuid } = useProject(isProjectUuid ? scope.project : '');
  const project = isProjectUuid ? projectByUuid : projectByHandler;
  const projectId = isProjectUuid ? scope.project : (project?.id ?? '');
  const projectName = !isProjectUuid ? scope.project : (project?.id ?? '');

  const { data: component, isLoading: loadingComponent } = useComponentByHandler(projectId, scope.component);
  const componentId = component?.id ?? '';

  const apiVersions = component?.apiVersions ?? [];
  const activeVersion = apiVersions.find((v) => v.latest) ?? apiVersions[0];
  const versionId = activeVersion?.id ?? '';

  const { data: repository, isLoading: loadingRepository } = useComponentRepository(projectId, scope.component);
  const branch = repository?.branch ?? '';

  const { data: builds = [], isLoading: loadingBuilds } = useBuilds(scope.component, projectName);
  const { data: commits = [], isLoading: loadingCommits } = useCommitHistory(componentId, branch);

  const isLoading = (!isProjectUuid && loadingProject) || loadingComponent || loadingRepository;

  if (isLoading) {
    return (
      <PageContent sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', py: 8 }}>
        <CircularProgress />
      </PageContent>
    );
  }

  if (!component) {
    return (
      <PageContent>
        <Typography color="error">Failed to load component information.</Typography>
      </PageContent>
    );
  }

  return (
    <PageContent>
      <BuildHistory
        componentId={componentId}
        versionId={versionId}
        componentName={scope.component}
        projectName={projectName}
        builds={builds}
        buildsLoading={loadingBuilds}
        commits={commits}
        commitsLoading={loadingCommits}
        repository={repository ?? null}
      />
    </PageContent>
  );
}
