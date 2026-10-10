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

import { Alert, Button, CircularProgress, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, Skeleton, Stack, Typography } from '@wso2/oxygen-ui';
import { useState, type JSX } from 'react';
import { useAutoscaling, useUpdateAutoscaling } from '../../hooks/useScaling';
import { CLOUD_HPA_CARD, CPU_THRESHOLD, MEMORY_THRESHOLD, NO_AUTOSCALING_CARD, SCALE_TO_ZERO_CARD } from '../../constants/scaling';
import ScaleMethodCard from './ScaleMethodCard';
import RangeInput from './RangeInput';
import ThresholdSlider from './ThresholdSlider';
import type { Autoscaling } from '../../types/scaling';
import { alertSx, cardsRowSx, sectionSx, sectionTitleSx } from './CloudAutoscaling.styles';

const NOT_AVAILABLE = 'Not available yet';
const noop = (): void => undefined;

interface HpaValues {
  minReplicas: number;
  maxReplicas: number;
  cpuUtilizationPercentage: number;
}

// The BFF reports no bounds until autoscaling is first enabled, and a bound written outside the
// console may exceed today's limit; either way the form opens on values it can submit.
function initialValues(a: Autoscaling): HpaValues {
  const maxReplicas = Math.min(a.maxReplicas ?? 2, a.maxReplicasLimit);
  return {
    minReplicas: Math.min(a.minReplicas ?? 1, maxReplicas),
    maxReplicas,
    cpuUtilizationPercentage: a.cpuUtilizationPercentage ?? CPU_THRESHOLD.default,
  };
}

function HpaForm({ autoscaling, canManage, saving, onSubmit }: { autoscaling: Autoscaling; canManage: boolean; saving: boolean; onSubmit: (values: HpaValues) => void }): JSX.Element {
  const [values, setValues] = useState(() => initialValues(autoscaling));
  const { minReplicas, maxReplicas, cpuUtilizationPercentage } = values;
  const locked = !canManage || saving;

  return (
    <Stack gap={2.5}>
      <Stack direction={{ xs: 'column', md: 'row' }} gap={4}>
        <RangeInput label="Min replicas" value={minReplicas} onChange={(v) => setValues({ ...values, minReplicas: Math.min(v, maxReplicas) })} min={1} max={maxReplicas} disabled={locked} />
        <RangeInput label="Max replicas" value={maxReplicas} onChange={(v) => setValues({ ...values, maxReplicas: Math.max(v, minReplicas) })} min={minReplicas} max={autoscaling.maxReplicasLimit} disabled={locked} />
      </Stack>
      <Stack direction={{ xs: 'column', md: 'row' }} gap={2}>
        {/* The autoscaler needs a CPU target, so the metric cannot be switched off. */}
        <ThresholdSlider label="CPU Threshold" enabled value={cpuUtilizationPercentage} min={CPU_THRESHOLD.min} max={CPU_THRESHOLD.max} onToggle={noop} onChange={(v) => setValues({ ...values, cpuUtilizationPercentage: v })} disabled={locked} toggleLocked />
        <ThresholdSlider label="Memory Threshold" enabled={false} value={MEMORY_THRESHOLD.default} min={MEMORY_THRESHOLD.min} max={MEMORY_THRESHOLD.max} onToggle={noop} onChange={noop} disabled note={NOT_AVAILABLE} />
      </Stack>
      {canManage && (
        <Stack direction="row">
          <Button variant="contained" disabled={saving} onClick={() => onSubmit(values)} startIcon={saving ? <CircularProgress size={16} color="inherit" /> : undefined}>
            {saving ? 'Saving…' : autoscaling.enabled ? 'Save' : 'Enable autoscaling'}
          </Button>
        </Stack>
      )}
    </Stack>
  );
}

function HpaStatus({ autoscaling }: { autoscaling: Autoscaling }): JSX.Element {
  const { status } = autoscaling;
  if (!status) {
    return (
      <Typography variant="body2" color="text.secondary">
        The autoscaler starts once the change reaches the environment, usually within a few minutes.
      </Typography>
    );
  }
  // A False condition is the HPA saying why it is not scaling, e.g. no CPU metrics yet.
  const blocked = status.conditions.filter((c) => c.status === 'False');
  // ScalingLimited holds at either bound; only the upper one means demand is going unmet.
  const atMax = status.conditions.some((c) => c.type === 'ScalingLimited' && c.status === 'True' && c.reason === 'TooManyReplicas');
  return (
    <Stack gap={1}>
      <Typography variant="body2">
        Replicas: <strong>{status.currentReplicas}</strong> running, <strong>{status.desiredReplicas}</strong> desired
      </Typography>
      <Typography variant="body2">
        CPU: {status.currentCpuUtilizationPercentage === undefined ? 'not reported yet' : <strong>{status.currentCpuUtilizationPercentage}%</strong>}
        {autoscaling.cpuUtilizationPercentage !== undefined && ` (target ${autoscaling.cpuUtilizationPercentage}%)`}
      </Typography>
      {atMax && <Alert severity="warning">Running at the maximum replica count. The CPU load would scale it further; raise Max replicas to allow more.</Alert>}
      {blocked.map((c) => (
        <Alert key={c.type} severity="info">
          {c.type}: {c.message || c.reason}
        </Alert>
      ))}
    </Stack>
  );
}

interface CloudAutoscalingProps {
  projectId: string;
  componentId: string;
  environmentId: string;
  environmentName: string;
  canManage: boolean;
  onSaved: (message: string) => void;
  onError: (message: string) => void;
}

/** Cloud's scaling body: one per-environment HPA setting, on or off. Scale to zero is shown but cannot be chosen yet. */
export default function CloudAutoscaling({ projectId, componentId, environmentId, environmentName, canManage, onSaved, onError }: CloudAutoscalingProps): JSX.Element {
  const { data: autoscaling, isLoading, isError, error, refetch } = useAutoscaling(projectId, componentId, environmentId);
  const update = useUpdateAutoscaling(projectId);
  // HPA picked but not yet saved: the form shows, while the environment still runs without autoscaling.
  const [hpaPicked, setHpaPicked] = useState(false);
  const [confirmingDisable, setConfirmingDisable] = useState(false);

  if (isLoading) {
    return (
      <Stack gap={2}>
        <Skeleton variant="rounded" height={120} />
        <Skeleton variant="rounded" height={180} />
      </Stack>
    );
  }
  if (isError) {
    return (
      <Alert
        severity="error"
        action={
          <Button color="inherit" size="small" onClick={() => void refetch()}>
            Retry
          </Button>
        }>
        {error instanceof Error ? error.message : 'Failed to load the scaling configuration.'}
      </Alert>
    );
  }
  if (!autoscaling) {
    return <Alert severity="info">Select an environment to configure scaling.</Alert>;
  }
  if (!autoscaling.supported) {
    return <Alert severity="info">Autoscaling isn&apos;t available for this integration type.</Alert>;
  }

  const showHpa = autoscaling.enabled || hpaPicked;

  const selectHpa = () => {
    if (!autoscaling.enabled) setHpaPicked(true);
  };

  const selectNoAutoscaling = () => {
    if (autoscaling.enabled) setConfirmingDisable(true);
    else setHpaPicked(false);
  };

  const enable = (values: HpaValues) => {
    const wasEnabled = autoscaling.enabled;
    update.mutate(
      { componentId, environmentId, data: { enabled: true, ...values } },
      {
        // An HPA change is part of the deployment, so applying it replaces the running replicas once.
        onSuccess: () => onSaved(wasEnabled ? 'Autoscaling settings saved. They take effect within a few minutes as the replicas are replaced.' : 'Autoscaling enabled. It takes effect within a few minutes as the replicas are replaced.'),
        onError: (e) => onError(e instanceof Error ? e.message : 'Failed to save the autoscaling settings.'),
      },
    );
  };

  const disable = () => {
    update.mutate(
      { componentId, environmentId, data: { enabled: false } },
      {
        onSuccess: () => {
          setConfirmingDisable(false);
          setHpaPicked(false);
          onSaved('Autoscaling turned off. The integration returns to a fixed number of replicas within a few minutes.');
        },
        onError: (e) => {
          setConfirmingDisable(false);
          onError(e instanceof Error ? e.message : 'Failed to turn off autoscaling.');
        },
      },
    );
  };

  return (
    <>
      {/* The platform validates the setting when it renders the release, so a saved change can still be rejected. */}
      {autoscaling.syncMessage && (
        <Alert severity="warning" sx={alertSx}>
          {autoscaling.syncMessage}
        </Alert>
      )}
      {!autoscaling.effective && (
        <Alert severity="info" sx={alertSx}>
          The release deployed to {environmentName} was built before this integration could autoscale. Redeploy it to {environmentName} to enable autoscaling.
        </Alert>
      )}

      <Stack direction={{ xs: 'column', md: 'row' }} gap={2} sx={cardsRowSx}>
        <ScaleMethodCard title={SCALE_TO_ZERO_CARD.title} description={SCALE_TO_ZERO_CARD.description} selected={false} disabled note={NOT_AVAILABLE} onSelect={noop} />
        <ScaleMethodCard title={CLOUD_HPA_CARD.title} description={CLOUD_HPA_CARD.description} selected={showHpa} disabled={!canManage || (!autoscaling.enabled && !autoscaling.effective)} onSelect={selectHpa} />
        <ScaleMethodCard title={NO_AUTOSCALING_CARD.title} description={NO_AUTOSCALING_CARD.description} selected={!showHpa} disabled={!canManage} onSelect={selectNoAutoscaling} />
      </Stack>

      <Typography variant="subtitle1" sx={sectionTitleSx}>
        Scaling Configuration
      </Typography>
      <Stack sx={sectionSx}>
        {showHpa ? (
          // Keyed on the saved setting so the form reopens on it after every save, without an effect.
          <HpaForm
            key={`${autoscaling.enabled}:${autoscaling.minReplicas}:${autoscaling.maxReplicas}:${autoscaling.cpuUtilizationPercentage}`}
            autoscaling={autoscaling}
            canManage={canManage && (autoscaling.enabled || autoscaling.effective)}
            saving={update.isPending}
            onSubmit={enable}
          />
        ) : (
          <Typography variant="body2" color="text.secondary">
            This integration runs a fixed number of replicas in {environmentName}. Choose HPA to scale it with CPU usage.
          </Typography>
        )}
      </Stack>

      {autoscaling.enabled && (
        <>
          <Typography variant="subtitle1" sx={sectionTitleSx}>
            Current Status
          </Typography>
          <HpaStatus autoscaling={autoscaling} />
        </>
      )}

      <Dialog open={confirmingDisable} onClose={() => setConfirmingDisable(false)} maxWidth="xs" fullWidth>
        <DialogTitle>Turn off autoscaling?</DialogTitle>
        <DialogContent>
          <DialogContentText>
            The autoscaler is removed from <strong>{environmentName}</strong> and the integration returns to a fixed number of replicas. Continue?
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setConfirmingDisable(false)} disabled={update.isPending}>
            Cancel
          </Button>
          <Button variant="contained" onClick={disable} disabled={update.isPending} startIcon={update.isPending ? <CircularProgress size={16} color="inherit" /> : undefined}>
            Turn off
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
}
