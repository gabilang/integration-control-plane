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

import { Button, CircularProgress, IconButton, ListingTable, TablePagination, Typography } from '@wso2/oxygen-ui';
import { CheckCircle2, ChevronRight, XCircle } from '@wso2/oxygen-ui-icons-react';
import { Fragment, useEffect, useState } from 'react';
import type { BffExecution } from '../api/queries';
import ExecutionDrawer from './EnvironmentCard/ExecutionDrawer';
import LogsDrawer from './EnvironmentCard/LogsDrawer';

interface AutomationExecutionsProps {
  executions: BffExecution[];
  projectId: string;
  componentId: string;
  environmentId: string;
  orgHandler: string;
  projectHandler: string;
  componentHandler: string;
  envCritical: boolean;
  pendingTriggerTime?: number | null;
  pendingTriggerArgs?: string[] | null;
  onTriggerResolved?: () => void;
  onRunSuccess?: () => void;
}

const QUEUED_SENTINEL = '__queued__';

function formatTriggeredAt(timestamp: string): string {
  if (!timestamp) return '—';
  const date = new Date(timestamp);
  if (isNaN(date.getTime())) return '—';
  const today = new Date();
  const timeStr = date.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  if (date.toDateString() === today.toDateString()) return `Today at ${timeStr}`;
  return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' }) + ` at ${timeStr}`;
}

function formatDuration(startTime: string, completionTime: string): string {
  if (!startTime || !completionTime) return '—';
  const start = new Date(startTime).getTime();
  const end = new Date(completionTime).getTime();
  if (isNaN(start) || isNaN(end)) return '—';
  const diff = Math.floor((end - start) / 1000);
  if (diff < 0) return '—';
  const minutes = Math.floor(diff / 60);
  const seconds = diff % 60;
  if (minutes === 0) return `${seconds}s`;
  return seconds === 0 ? `${minutes}m` : `${minutes}m ${seconds}s`;
}

function isInProgress(status: string): boolean {
  const val = status?.toLowerCase();
  if (val === 'succeeded' || val === 'success' || val === 'complete' || val === 'failed' || val === 'failure') return false;
  return true;
}

function StatusIcon({ status, inProgress }: { status: string; inProgress: boolean }) {
  if (inProgress) return <CircularProgress size={18} />;
  const val = status?.toLowerCase();
  if (val === 'succeeded' || val === 'success') return <CheckCircle2 size={18} color="green" />;
  if (val === 'failed' || val === 'failure') return <XCircle size={18} color="red" />;
  return <CircularProgress size={18} />;
}

// Synthetic "queued" execution shown immediately after triggering, before the API returns
const QUEUED_EXECUTION: BffExecution = {
  jobId: QUEUED_SENTINEL,
  startTime: '',
  completionTime: '',
  revisionId: '',
  status: 'Queued',
};

export default function AutomationExecutions({
  executions,
  projectId,
  componentId,
  environmentId,
  orgHandler,
  projectHandler,
  componentHandler,
  envCritical,
  pendingTriggerTime,
  pendingTriggerArgs: _pendingTriggerArgs,
  onTriggerResolved,
  onRunSuccess,
}: AutomationExecutionsProps) {
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(5);
  const [selectedExecution, setSelectedExecution] = useState<BffExecution | null>(null);
  const [logsExecution, setLogsExecution] = useState<BffExecution | null>(null);

  const hasInProgress = executions.some((e) => isInProgress(e.status));
  const [extendPoll, setExtendPoll] = useState(false);
  const _shouldPoll = !!pendingTriggerTime || hasInProgress || extendPoll;

  // Detect when the real new execution arrives and clear the queued sentinel
  useEffect(() => {
    if (!pendingTriggerTime || executions.length === 0) return;
    const latestStartMs = new Date(executions[0].startTime ?? '').getTime();
    if (!isNaN(latestStartMs) && latestStartMs >= pendingTriggerTime - 5000) {
      onTriggerResolved?.();
    }
  }, [executions, pendingTriggerTime, onTriggerResolved]);

  // Auto-clear sentinel after 60s in case the API never returns the new execution.
  // Keep polling alive via extendPoll so a delayed execution is still detected.
  useEffect(() => {
    if (!pendingTriggerTime) return;
    const timer = setTimeout(() => {
      onTriggerResolved?.();
      setExtendPoll(true);
    }, 60000);
    return () => clearTimeout(timer);
  }, [pendingTriggerTime, onTriggerResolved]);

  // Stop extended polling once an execution arrives or after a further 60s give-up.
  useEffect(() => {
    if (!extendPoll) return;
    if (executions.length > 0) {
      setExtendPoll(false);
      return;
    }
    const timer = setTimeout(() => setExtendPoll(false), 60000);
    return () => clearTimeout(timer);
  }, [extendPoll, executions]);

  // Show the queued sentinel row at position 0 while pendingTriggerTime is set and no new exec arrived
  const showQueued = !!pendingTriggerTime && (executions.length === 0 || new Date(executions[0].startTime ?? '').getTime() < pendingTriggerTime - 5000);
  const allExecutions = showQueued ? [QUEUED_EXECUTION, ...executions] : executions;

  const maxPage = Math.max(0, Math.ceil(allExecutions.length / rowsPerPage) - 1);
  const safePage = Math.min(page, maxPage);
  const paged = allExecutions.slice(safePage * rowsPerPage, safePage * rowsPerPage + rowsPerPage);

  if (allExecutions.length === 0) {
    return (
      <Typography variant="body2" color="text.secondary" sx={{ textAlign: 'center', py: 2 }}>
        No execution data available. Click &apos;{envCritical ? 'Run' : 'Test'}&apos; or use &apos;Schedule&apos; to trigger an execution.
      </Typography>
    );
  }

  return (
    <Fragment>
      <ListingTable.Container>
        <ListingTable density="compact">
          <ListingTable.Head>
            <ListingTable.Row>
              <ListingTable.Cell>Status</ListingTable.Cell>
              <ListingTable.Cell>Triggered At</ListingTable.Cell>
              <ListingTable.Cell>Duration</ListingTable.Cell>
              <ListingTable.Cell>Commit ID</ListingTable.Cell>
              <ListingTable.Cell>Latest Logs</ListingTable.Cell>
              <ListingTable.Cell></ListingTable.Cell>
            </ListingTable.Row>
          </ListingTable.Head>
          <ListingTable.Body>
            {paged.map((e) => {
              const inProgress = isInProgress(e.status);
              return (
                <ListingTable.Row key={e.jobId}>
                  <ListingTable.Cell>
                    <StatusIcon status={e.status} inProgress={inProgress} />
                  </ListingTable.Cell>
                  <ListingTable.Cell>
                    <Typography variant="body2">{formatTriggeredAt(e.startTime ?? '')}</Typography>
                  </ListingTable.Cell>
                  <ListingTable.Cell>
                    <Typography variant="body2">{inProgress ? '--' : formatDuration(e.startTime ?? '', e.completionTime ?? '')}</Typography>
                  </ListingTable.Cell>
                  <ListingTable.Cell>
                    <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
                      {e.revisionId ? e.revisionId.substring(0, 7) : '—'}
                    </Typography>
                  </ListingTable.Cell>
                  <ListingTable.Cell>
                    {e.jobId === QUEUED_SENTINEL ? (
                      <Typography variant="body2" color="text.secondary">
                        --
                      </Typography>
                    ) : (
                      <Button variant="text" size="small" onClick={() => setLogsExecution(e)}>
                        View Logs
                      </Button>
                    )}
                  </ListingTable.Cell>
                  <ListingTable.Cell>
                    {e.jobId !== QUEUED_SENTINEL && (
                      <IconButton size="small" aria-label="View execution details" onClick={() => setSelectedExecution(e)}>
                        <ChevronRight size={16} />
                      </IconButton>
                    )}
                  </ListingTable.Cell>
                </ListingTable.Row>
              );
            })}
          </ListingTable.Body>
        </ListingTable>
        <TablePagination
          sx={{ borderTop: '1px solid', borderColor: 'divider' }}
          component="div"
          count={allExecutions.length}
          page={safePage}
          onPageChange={(_, p) => setPage(p)}
          rowsPerPage={rowsPerPage}
          onRowsPerPageChange={(e) => {
            setRowsPerPage(parseInt(e.target.value, 10));
            setPage(0);
          }}
          rowsPerPageOptions={[5, 10, 25]}
        />
      </ListingTable.Container>

      <ExecutionDrawer
        open={!!selectedExecution}
        execution={selectedExecution}
        onClose={() => setSelectedExecution(null)}
        onRunSuccess={onRunSuccess}
        orgHandler={orgHandler}
        projectHandler={projectHandler}
        componentHandler={componentHandler}
        projectId={projectId}
        componentId={componentId}
        environmentId={environmentId}
      />

      <LogsDrawer open={!!logsExecution} onClose={() => setLogsExecution(null)} executionId={logsExecution?.jobId ?? ''} componentId={componentId} environmentId={environmentId} />
    </Fragment>
  );
}
