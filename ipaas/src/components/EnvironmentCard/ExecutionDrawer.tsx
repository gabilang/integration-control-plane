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

import { Alert, Box, Button, Checkbox, CircularProgress, Divider, Drawer, FormControlLabel, IconButton, InputAdornment, Link, OutlinedInput, Stack, Tab, Tabs, TextField, Tooltip, Typography } from '@wso2/oxygen-ui';
import { ArrowLeft, CheckCircle2, Copy, Plus, RefreshCcw, Search, Trash2, X, XCircle } from '@wso2/oxygen-ui-icons-react';
import { useEffect, useMemo, useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useExecutionArguments, useExecutionLogs, type BffExecution } from '../../api/queries';
import { useTriggerExecution } from '../../api/mutations';

interface ExecutionDrawerProps {
  execution: BffExecution | null;
  open: boolean;
  onClose: () => void;
  onRunSuccess?: () => void;
  orgHandler: string;
  projectHandler: string;
  componentHandler: string;
  projectId: string;
  componentId: string;
  environmentId: string;
}

function formatTriggeredAt(timestamp: string): string {
  if (!timestamp) return '—';
  const date = new Date(timestamp);
  if (isNaN(date.getTime())) return '—';
  const today = new Date();
  const timeStr = date.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  if (date.toDateString() === today.toDateString()) return `Today at ${timeStr}`;
  const yesterday = new Date(today);
  yesterday.setDate(today.getDate() - 1);
  if (date.toDateString() === yesterday.toDateString()) return `Yesterday at ${timeStr}`;
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

function isTerminal(status: string): boolean {
  const v = status?.toLowerCase();
  return v === 'succeeded' || v === 'success' || v === 'failed' || v === 'failure';
}

function StatusIcon({ status }: { status: string }) {
  const v = status?.toLowerCase();
  if (!isTerminal(status)) return <CircularProgress size={16} />;
  if (v === 'succeeded' || v === 'success') return <CheckCircle2 size={16} color="green" />;
  return <XCircle size={16} color="red" />;
}

function highlightText(text: string, keyword: string): React.ReactNode {
  if (!keyword) return text;
  const escaped = keyword.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  const parts = text.split(new RegExp(`(${escaped})`, 'gi'));
  return parts.map((part, i) =>
    part.toLowerCase() === keyword.toLowerCase() ? (
      <mark key={i} style={{ background: '#fde68a', color: 'inherit', borderRadius: 2, padding: '0 1px' }}>
        {part}
      </mark>
    ) : (
      part
    ),
  );
}

const drawerSx = {
  '& .MuiDrawer-paper': {
    width: 480,
    position: 'fixed',
    top: 64,
    height: 'calc(100% - 64px)',
    display: 'flex',
    flexDirection: 'column',
  },
};

export default function ExecutionDrawer({ execution, open, onClose, onRunSuccess, projectId, componentId, environmentId }: ExecutionDrawerProps) {
  const queryClient = useQueryClient();
  const handleClose = () => {
    (document.activeElement as HTMLElement)?.blur();
    onClose();
  };

  const [view, setView] = useState<'execution' | 'logs'>('execution');
  const [tab, setTab] = useState(0);
  const [copied, setCopied] = useState(false);
  const [args, setArgs] = useState<string[]>(['']);
  const [runError, setRunError] = useState<string | null>(null);

  // Logs view state
  const [logSearch, setLogSearch] = useState('');
  const [logFilterMode, setLogFilterMode] = useState(false);

  const trigger = useTriggerExecution();

  const { data: fetchedArgs, isLoading: argsLoading } = useExecutionArguments(execution?.jobId ?? '', componentId, '', open && tab === 1 && !!execution?.jobId);

  const { data: logs = [], isLoading: logsLoading } = useExecutionLogs(componentId, '', execution?.jobId ?? '', environmentId, open && view === 'logs' && !!execution?.jobId);

  const filteredLogs = useMemo(() => {
    if (!logFilterMode || !logSearch.trim()) return logs;
    const lower = logSearch.toLowerCase();
    return logs.filter((e) => `${e.timestamp} ${e.message}`.toLowerCase().includes(lower));
  }, [logs, logSearch, logFilterMode]);

  // Populate editable args from fetched data
  useEffect(() => {
    if (fetchedArgs) {
      setArgs(fetchedArgs.length > 0 ? fetchedArgs.map((a) => a.argumentValue) : ['']);
    }
  }, [fetchedArgs]);

  // Reset state when drawer closes
  useEffect(() => {
    if (!open) {
      setTab(0);
      setArgs(['']);
      setRunError(null);
      setView('execution');
      setLogSearch('');
      setLogFilterMode(false);
    }
  }, [open]);

  const handleCopy = () => {
    if (!execution?.revisionId) return;
    navigator.clipboard.writeText(execution.revisionId).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    });
  };

  const handleRun = () => {
    setRunError(null);
    trigger.mutate(
      { componentId, projectId, environment: environmentId },
      {
        onSuccess: () => {
          onClose();
          onRunSuccess?.();
        },
        onError: (err) => setRunError(err instanceof Error ? err.message : 'Failed to execute'),
      },
    );
  };

  const handleLogsRefresh = () => {
    queryClient.invalidateQueries({ queryKey: ['executionLogs', componentId, '', execution?.jobId, environmentId] });
  };

  return (
    <Drawer anchor="right" open={open} onClose={handleClose} variant="temporary" sx={drawerSx}>
      {/* Header */}
      <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{ px: 2, py: 1.5, borderBottom: '1px solid', borderColor: 'divider', flexShrink: 0 }}>
        <Typography variant="subtitle1" sx={{ fontWeight: 600 }}>
          {view === 'logs' ? 'Attempt: 1' : execution ? `Execution: ${execution.jobId}` : 'Execution'}
        </Typography>
        <IconButton size="small" aria-label="close" onClick={handleClose}>
          <X size={16} />
        </IconButton>
      </Stack>

      {/* ── LOGS VIEW ── */}
      {view === 'logs' && (
        <>
          <Box sx={{ px: 2, pt: 1.5, pb: 1, flexShrink: 0 }}>
            {/* Back button */}
            <Button size="small" variant="text" startIcon={<ArrowLeft size={14} />} onClick={() => setView('execution')} sx={{ mb: 1, px: 0 }}>
              Back
            </Button>

            {/* Search */}
            <OutlinedInput
              fullWidth
              size="small"
              placeholder="Enter keyword to highlight logs"
              value={logSearch}
              onChange={(e) => setLogSearch(e.target.value)}
              startAdornment={
                <InputAdornment position="start">
                  <Search size={16} />
                </InputAdornment>
              }
              endAdornment={
                logSearch ? (
                  <InputAdornment position="end">
                    <IconButton size="small" onClick={() => setLogSearch('')} edge="end">
                      <X size={13} />
                    </IconButton>
                  </InputAdornment>
                ) : null
              }
              sx={{
                bgcolor: 'action.hover',
                '& .MuiOutlinedInput-notchedOutline': { border: 'none' },
                '&:hover .MuiOutlinedInput-notchedOutline': { border: 'none' },
                '&.Mui-focused .MuiOutlinedInput-notchedOutline': { border: '1px solid', borderColor: 'primary.main' },
                fontSize: '0.875rem',
              }}
            />

            {/* Filter checkbox + Refresh */}
            <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{ mt: 1 }}>
              <FormControlLabel control={<Checkbox size="small" checked={logFilterMode} onChange={(e) => setLogFilterMode(e.target.checked)} />} label={<Typography variant="body2">Filter logs with this search term</Typography>} sx={{ m: 0 }} />
              <Button size="small" variant="text" startIcon={<RefreshCcw size={14} />} onClick={handleLogsRefresh} disabled={logsLoading}>
                Refresh
              </Button>
            </Stack>
          </Box>

          {/* Log content */}
          <Box sx={{ flex: 1, overflow: 'auto', px: 2, py: 1 }}>
            {logsLoading ? (
              <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}>
                <CircularProgress size={28} />
              </Box>
            ) : logs.length === 0 ? (
              <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}>
                <Typography variant="body2" color="text.secondary">
                  No logs available for this execution.
                </Typography>
              </Box>
            ) : filteredLogs.length === 0 ? (
              <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
                <Typography variant="body2" color="text.secondary">
                  No log lines match the search term.
                </Typography>
              </Box>
            ) : (
              <Box component="pre" sx={{ m: 0, fontFamily: 'monospace', fontSize: '0.75rem', lineHeight: 1.7, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
                {filteredLogs.map((entry, i) => {
                  const line = entry.timestamp ? `${entry.timestamp} ${entry.message}` : entry.message;
                  return (
                    <Box key={i} component="span" sx={{ display: 'block' }}>
                      {logSearch && !logFilterMode ? highlightText(line, logSearch) : line}
                    </Box>
                  );
                })}
              </Box>
            )}
          </Box>
        </>
      )}

      {/* ── EXECUTION VIEW ── */}
      {view === 'execution' && execution && (
        <>
          <Box sx={{ flex: 1, overflow: 'auto', display: 'flex', flexDirection: 'column' }}>
            {/* Info card */}
            <Box sx={{ px: 2, pt: 2 }}>
              <Box sx={{ border: '1px solid', borderColor: 'divider', borderRadius: 1, px: 2, py: 1.5 }}>
                <Stack direction="row" gap={4}>
                  <Box>
                    <Typography variant="caption" color="text.secondary">
                      Triggered at
                    </Typography>
                    <Typography variant="body2" sx={{ fontWeight: 600, mt: 0.5 }}>
                      {formatTriggeredAt(execution.startTime ?? '')}
                    </Typography>
                  </Box>
                  <Box>
                    <Typography variant="caption" color="text.secondary">
                      Revision
                    </Typography>
                    <Stack direction="row" alignItems="center" gap={0.5} sx={{ mt: 0.5 }}>
                      <StatusIcon status={execution.status} />
                      <Typography variant="body2" sx={{ fontFamily: 'monospace', fontWeight: 600 }}>
                        {execution.revisionId ? execution.revisionId.substring(0, 10) : '—'}
                      </Typography>
                      {execution.revisionId && (
                        <Tooltip title={copied ? 'Copied!' : 'Copy'}>
                          <IconButton size="small" onClick={handleCopy} aria-label="Copy revision ID">
                            <Copy size={12} />
                          </IconButton>
                        </Tooltip>
                      )}
                    </Stack>
                  </Box>
                </Stack>
              </Box>
            </Box>

            {/* Tabs */}
            <Box sx={{ px: 2, mt: 2 }}>
              <Tabs value={tab} onChange={(_, v) => setTab(v)}>
                <Tab label="Attempts" />
                <Tab label="Arguments" />
              </Tabs>
              <Divider />
            </Box>

            {/* Attempts tab */}
            {tab === 0 && (
              <Box sx={{ px: 2, py: 2 }}>
                <Stack direction="row" sx={{ pb: 1, borderBottom: '1px solid', borderColor: 'divider', mb: 1 }}>
                  <Typography variant="caption" sx={{ fontWeight: 700, flex: '0 0 100px' }}>
                    ATTEMPT ID
                  </Typography>
                  <Typography variant="caption" sx={{ fontWeight: 700, flex: '0 0 80px' }}>
                    DURATION
                  </Typography>
                  <Typography variant="caption" sx={{ fontWeight: 700, flex: 1 }}>
                    TRIGGERED AT
                  </Typography>
                  <Typography variant="caption" sx={{ fontWeight: 700 }}>
                    ACTIONS
                  </Typography>
                </Stack>
                <Stack direction="row" alignItems="center" sx={{ py: 1, borderBottom: '1px solid', borderColor: 'divider' }}>
                  <Stack direction="row" alignItems="center" gap={0.75} sx={{ flex: '0 0 100px' }}>
                    <StatusIcon status={execution.status} />
                    <Typography variant="body2">#1</Typography>
                  </Stack>
                  <Typography variant="body2" color="text.secondary" sx={{ flex: '0 0 80px' }}>
                    {isTerminal(execution.status) ? formatDuration(execution.startTime ?? '', execution.completionTime ?? '') : '--'}
                  </Typography>
                  <Typography variant="body2" color="text.secondary" sx={{ flex: 1 }}>
                    {isTerminal(execution.status) ? formatTriggeredAt(execution.startTime ?? '') : '--'}
                  </Typography>
                  <Button variant="text" size="small" onClick={() => setView('logs')}>
                    View Logs
                  </Button>
                </Stack>
              </Box>
            )}

            {/* Arguments tab */}
            {tab === 1 && (
              <Box sx={{ px: 2, py: 2 }}>
                <Alert severity="info" variant="outlined" sx={{ mb: 2 }}>
                  <Typography variant="body2">
                    To learn more about runtime arguments, see the{' '}
                    <Link href="https://wso2.com/ballerina/icp/docs/" target="_blank" rel="noopener noreferrer" variant="body2">
                      WSO2 Integration Platform Documentation
                    </Link>
                    .
                  </Typography>
                </Alert>

                {argsLoading ? (
                  <CircularProgress size={24} sx={{ display: 'block', mx: 'auto', my: 2 }} />
                ) : (
                  <Stack gap={1.5}>
                    <Box>
                      <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 0.5 }}>
                        Arguments
                      </Typography>
                      <Box sx={{ bgcolor: 'action.hover', borderRadius: 1, px: 1.5, py: 1, fontFamily: 'monospace', fontSize: '0.75rem', wordBreak: 'break-all' }}>{JSON.stringify(args)}</Box>
                    </Box>
                    {args.map((arg, index) => (
                      <Stack key={index} gap={0.5}>
                        <Typography variant="caption" color="text.secondary">
                          Arg {index + 1}
                        </Typography>
                        <Stack direction="row" gap={1} alignItems="center">
                          <TextField size="small" fullWidth placeholder="Enter the argument" value={arg} onChange={(e) => setArgs((prev) => prev.map((a, i) => (i === index ? e.target.value : a)))} />
                          <IconButton size="small" aria-label="Remove argument" disabled={args.length === 1} onClick={() => setArgs((prev) => prev.filter((_, i) => i !== index))}>
                            <Trash2 size={14} />
                          </IconButton>
                        </Stack>
                      </Stack>
                    ))}

                    <Button variant="text" size="small" startIcon={<Plus size={14} />} onClick={() => setArgs((prev) => [...prev, ''])} sx={{ alignSelf: 'flex-start', px: 0 }}>
                      Add Argument
                    </Button>

                    {runError && (
                      <Alert severity="error" sx={{ mt: 1 }}>
                        {runError}
                      </Alert>
                    )}
                  </Stack>
                )}
              </Box>
            )}
          </Box>

          {/* Footer — only on Arguments tab */}
          {tab === 1 && !argsLoading && (
            <Stack direction="row" justifyContent="flex-end" gap={1} sx={{ px: 2, py: 1.5, borderTop: '1px solid', borderColor: 'divider', flexShrink: 0 }}>
              <Button onClick={onClose}>Cancel</Button>
              <Button variant="contained" onClick={handleRun} disabled={trigger.isPending} startIcon={trigger.isPending ? <CircularProgress color="inherit" size={16} /> : undefined}>
                {trigger.isPending ? 'Executing…' : 'Execute'}
              </Button>
            </Stack>
          )}
        </>
      )}
    </Drawer>
  );
}
