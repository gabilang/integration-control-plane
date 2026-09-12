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

import { Button } from '@wso2/oxygen-ui';
import { Play } from '@wso2/oxygen-ui-icons-react';
import { useMemo } from 'react';
import type { ReactNode } from 'react';
import { useAppNavigate } from '../../../hooks/useAppNavigate';
import { useQueryClient } from '@tanstack/react-query';
import { useExecutionConfigs, useRuntimeArguments, useTriggerComponent } from '../../../hooks/useExecutions';
import { useSchemaConfig } from '../../../hooks/useConfiguration';
import type { EnvCardActionsProps } from '../../../types/integration';
import { isDeploymentHealthy } from '../../../utils/deploymentStatus';
import { IS_CLOUD } from '../../../features';
import { hasMissingRequiredConfigs } from './configStatus';
import NextRunLabel from '../../NextRunLabel';
import ScheduleButton from './ScheduleButton';

/**
 * Automation's right-header slot. Test only leaves the card when the task takes runtime arguments.
 */
export default function EnvCardActions({
  component,
  env,
  projectId,
  versionId,
  orgHandler,
  projectHandler,
  componentHandler,
  releaseId,
  buildId,
  deploymentPipelineId,
  envTemplateId,
  deployedCommitSha,
  isBuildInProgress,
  deploymentStatusV2,
  onNotify,
  onTrigger,
}: EnvCardActionsProps): ReactNode {
  const queryClient = useQueryClient();
  const navigate = useAppNavigate();

  const { data: scheduleConfig } = useExecutionConfigs(component.id, releaseId, env.id);
  const { data: schemaConfig } = useSchemaConfig(projectId, component.id, envTemplateId, versionId, deployedCommitSha);
  const missingConfigs = useMemo(() => hasMissingRequiredConfigs(schemaConfig), [schemaConfig]);

  // Automation's Run/Schedule are always disabled while a build is in progress.
  const buildDisabled = !!isBuildInProgress;
  // A stopped schedule leaves the CronJob deployed, so this only excludes a stopped deployment.
  const canTest = isDeploymentHealthy(deploymentStatusV2);


  // Cloud has no runtime-arguments endpoint, so the query stays disabled rather than always failing.
  const { data: runtimeArgs, isLoading: runtimeArgsLoading } = useRuntimeArguments(component.id, versionId, deployedCommitSha ?? '', !IS_CLOUD);
  const hasRuntimeArgs = (runtimeArgs?.length ?? 0) > 0;
  const triggerRun = useTriggerComponent();

  const goToTestPage = () => navigate(`/organizations/${orgHandler}/projects/${projectHandler}/components/${componentHandler}/test`);

  const handleTest = () => {
    if (hasRuntimeArgs) {
      goToTestPage();
      return;
    }
    triggerRun.mutate(
      { orgHandler, projectId, componentId: component.id, releaseId, args: [] },
      {
        onSuccess: () => {
          onNotify({ text: 'Execution triggered successfully', severity: 'success' });
          // Surfaces the run in this card's executions table before the list refetches.
          onTrigger(Date.now());
          queryClient.invalidateQueries({ queryKey: ['taskExecutions'] });
        },
        onError: (err) => onNotify({ text: err instanceof Error ? err.message : 'Failed to trigger execution', severity: 'error' }),
      },
    );
  };

  return (
    <>
      <NextRunLabel cron={scheduleConfig?.cronjobFrequency ?? ''} timeZone={scheduleConfig?.cronjobTimezone ?? ''} sx={{ mr: 0.5 }} />
      <ScheduleButton
        envId={env.id}
        envName={env.name}
        componentId={component.id}
        orgHandler={orgHandler}
        releaseId={releaseId}
        buildId={buildId}
        versionId={versionId}
        deploymentPipelineId={deploymentPipelineId}
        hasSchedule={!!scheduleConfig?.cronjobFrequency}
        disabled={missingConfigs || buildDisabled || !releaseId}
        onSaveSuccess={() => onNotify({ text: 'Schedule updated successfully', severity: 'success' })}
        onSaveError={() => onNotify({ text: 'Failed to save schedule. Please try again.', severity: 'error' })}
        onStopSuccess={() => onNotify({ text: 'Schedule stopped successfully', severity: 'success' })}
      />
      <Button variant="contained" size="small" startIcon={<Play size={14} />} disabled={missingConfigs || buildDisabled || !canTest || triggerRun.isPending || runtimeArgsLoading} onClick={handleTest}>
        Test
      </Button>
    </>
  );
}
