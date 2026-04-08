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

import { Alert, Button, Chip, CircularProgress, IconButton, ListingTable, Snackbar, Stack, TablePagination, Tooltip, Typography } from '@wso2/oxygen-ui';
import { GitCommit, Settings } from '@wso2/oxygen-ui-icons-react';
import { useCallback, useEffect, useState } from 'react';
import { useSearchParams } from 'react-router';
import { useTriggerBuild } from '../../api/mutations';
import { type GqlCommit, type GqlDeploymentStatus, type GqlRepository } from '../../api/queries';
import { BuildDrawerType } from '../../constants/build';
import { useBuildAutoEffects } from '../../hooks/useBuildAutoEffects';
import { formatDistanceToNow } from '../../utils/time';
import BuildRightDrawer from './BuildRightDrawer';
import BuildStatusLabel from './BuildStatusLabel';

interface BuildHistoryProps {
  componentId: string;
  versionId: string;
  componentName: string;
  projectName: string;
  builds: GqlDeploymentStatus[];
  buildsLoading: boolean;
  commits: GqlCommit[];
  commitsLoading: boolean;
  repository: GqlRepository | null;
}

export default function BuildHistory({ componentId, versionId, componentName, projectName, builds, buildsLoading, commits, commitsLoading, repository }: BuildHistoryProps) {
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [drawerType, setDrawerType] = useState<BuildDrawerType | null>(null);
  const [selectedBuild, setSelectedBuild] = useState<GqlDeploymentStatus | null>(null);
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(5);

  const [snackbar, setSnackbar] = useState<{ open: boolean; message: string; severity: 'success' | 'error' }>({ open: false, message: '', severity: 'success' });
  const handleNotify = (message: string, severity: 'success' | 'error') => setSnackbar({ open: true, message, severity });

  const [searchParams, setSearchParams] = useSearchParams();
  const triggerBuild = useTriggerBuild();

  const handleAutoOpen = useCallback((build: GqlDeploymentStatus) => {
    setSelectedBuild(build);
    setDrawerType(BuildDrawerType.BuildLogs);
    setDrawerOpen(true);
  }, []);

  const { justTriggered, setJustTriggered, hasInProgress, markAsOpened } = useBuildAutoEffects({
    builds,
    onAutoOpen: handleAutoOpen,
  });

  useEffect(() => {
    const buildId = Number(searchParams.get('buildId'));
    if (buildId && builds.length > 0) {
      const match = builds.find((b) => b.id === buildId);
      if (match) {
        markAsOpened(match.id);
        setSelectedBuild(match);
        setDrawerType(BuildDrawerType.BuildLogs);
        setDrawerOpen(true);
      }
    }
  }, [builds, searchParams, markAsOpened]);

  const latestCommit = commits.find((c) => c.isLatest) ?? commits[0] ?? null;

  const isBuilding = triggerBuild.isPending || hasInProgress || justTriggered;

  const handleBuildLatest = () => {
    setJustTriggered(true);
    triggerBuild.mutate(
      { componentName, projectName },
      { onError: () => setJustTriggered(false) },
    );
  };

  const handleViewDetails = (build: GqlDeploymentStatus) => {
    setSelectedBuild(build);
    setDrawerType(BuildDrawerType.BuildLogs);
    setDrawerOpen(true);
    const params = new URLSearchParams(searchParams);
    params.set('buildId', String(build.id));
    setSearchParams(params, { replace: true });
  };

  const handleSelectCommit = () => {
    setDrawerType(BuildDrawerType.CommitSelector);
    setDrawerOpen(true);
  };

  const handleOpenConfig = () => {
    setDrawerType(BuildDrawerType.BuildConfig);
    setDrawerOpen(true);
  };

  const handleCommitBuild = (_commit: GqlCommit) => {
    setJustTriggered(true);
    triggerBuild.mutate(
      { componentName, projectName },
      { onError: () => setJustTriggered(false) },
    );
  };

  const handleCloseDrawer = () => {
    setDrawerOpen(false);
    setDrawerType(null);
    const params = new URLSearchParams(searchParams);
    params.delete('buildId');
    setSearchParams(params, { replace: true });
  };

  return (
    <>
      <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{ mb: 3 }}>
        <Typography variant="h1">Build History</Typography>

        <Stack direction="row" alignItems="center" gap={1}>
          <Tooltip title={hasInProgress ? 'A build is already in progress' : !latestCommit ? 'No commits available' : ''}>
            <span>
              <Button variant="outlined" size="small" onClick={handleSelectCommit} disabled={isBuilding || commitsLoading || commits.length === 0}>
                Show Commits
              </Button>
            </span>
          </Tooltip>

          <Tooltip title={hasInProgress ? 'A build is already in progress' : !latestCommit ? 'No commits available' : ''}>
            <span>
              <Button
                variant="contained"
                size="small"
                onClick={handleBuildLatest}
                disabled={isBuilding || !latestCommit}
                startIcon={isBuilding ? <CircularProgress color="inherit" size={14} /> : undefined}>
                {isBuilding ? 'Building' : 'Build Latest'}
              </Button>
            </span>
          </Tooltip>

          <Tooltip title={isBuilding ? 'A build is already in progress' : 'Build configuration'}>
            <span>
              <IconButton size="small" onClick={handleOpenConfig} disabled={!repository || isBuilding}>
                <Settings size={18} />
              </IconButton>
            </span>
          </Tooltip>
        </Stack>
      </Stack>

      <ListingTable.Container elevation={0} sx={{ border: '1px solid', borderColor: 'divider', borderRadius: 1, overflow: 'hidden' }}>
        <ListingTable size="small">
          <ListingTable.Head>
            <ListingTable.Row>
              <ListingTable.Cell>Build ID</ListingTable.Cell>
              <ListingTable.Cell>Commit</ListingTable.Cell>
              <ListingTable.Cell>Status</ListingTable.Cell>
              <ListingTable.Cell>Time</ListingTable.Cell>
              <ListingTable.Cell align="right">Actions</ListingTable.Cell>
            </ListingTable.Row>
          </ListingTable.Head>

          <ListingTable.Body>
            {buildsLoading && builds.length === 0 && (
              <ListingTable.Row>
                <ListingTable.Cell colSpan={5} align="center" sx={{ py: 4 }}>
                  <CircularProgress size={24} color="primary" />
                </ListingTable.Cell>
              </ListingTable.Row>
            )}

            {!buildsLoading && builds.length === 0 && (
              <ListingTable.Row>
                <ListingTable.Cell colSpan={5} align="center" sx={{ py: 4 }}>
                  <Typography color="text.secondary" variant="body2">
                    No builds yet. Click &ldquo;Build Latest&rdquo; to start your first build.
                  </Typography>
                </ListingTable.Cell>
              </ListingTable.Row>
            )}

            {builds.slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage).map((build) => (
              <ListingTable.Row key={build.id} hover>
                <ListingTable.Cell>
                  <Stack direction="row" alignItems="center" gap={0.75}>
                    <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
                      #{build.id}
                    </Typography>
                    {build.isTriggeredAtCreation ? (
                      <Chip
                        label="Initial Build"
                        size="small"
                        variant="outlined"
                        color="default"
                        sx={{ fontSize: '0.65rem', height: (theme) => theme.spacing(2.25) }}
                      />
                    ) : (
                      <Chip
                        label={build.isAutoDeploy ? 'Automatic' : 'Manual'}
                        size="small"
                        variant="outlined"
                        color={build.isAutoDeploy ? 'primary' : 'info'}
                        sx={{ fontSize: '0.65rem', height: (theme) => theme.spacing(2.25) }}
                      />
                    )}
                  </Stack>
                </ListingTable.Cell>

                <ListingTable.Cell>
                  {build.sourceCommitId ? (
                    <Stack direction="row" alignItems="center" gap={0.5}>
                      <GitCommit size={12} style={{ opacity: 0.55 }} />
                      <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
                        {build.sourceCommitId.slice(0, 7)}
                      </Typography>
                    </Stack>
                  ) : (
                    <Typography variant="body2" color="text.disabled">
                      —
                    </Typography>
                  )}
                </ListingTable.Cell>

                <ListingTable.Cell>
                  <BuildStatusLabel status={build.status} conclusion={build.conclusionV2 || build.conclusion} />
                </ListingTable.Cell>

                <ListingTable.Cell>
                  <Typography variant="body2" color="text.secondary">
                    {build.started_at ? formatDistanceToNow(build.started_at) : '—'}
                  </Typography>
                </ListingTable.Cell>

                <ListingTable.Cell align="right">
                  <ListingTable.RowActions>
                    <Button variant="text" size="small" onClick={() => handleViewDetails(build)} sx={{ textTransform: 'none', fontSize: '0.75rem' }}>
                      View Details
                    </Button>
                  </ListingTable.RowActions>
                </ListingTable.Cell>
              </ListingTable.Row>
            ))}
          </ListingTable.Body>
        </ListingTable>

        <TablePagination
          component="div"
          count={builds.length}
          page={page}
          onPageChange={(_e, newPage) => setPage(newPage)}
          rowsPerPage={rowsPerPage}
          onRowsPerPageChange={(e) => { setRowsPerPage(Number(e.target.value)); setPage(0); }}
          rowsPerPageOptions={[5, 10, 25]}
        />
      </ListingTable.Container>

      <BuildRightDrawer
        open={drawerOpen}
        onClose={handleCloseDrawer}
        drawerType={drawerType}
        selectedBuild={selectedBuild ?? (drawerType === BuildDrawerType.BuildLogs ? builds[0] ?? null : null)}
        componentId={componentId}
        versionId={versionId}
        commits={commits}
        commitsLoading={commitsLoading}
        onCommitBuild={handleCommitBuild}
        isBuildTriggering={isBuilding}
        repository={repository}
        onNotify={handleNotify}
      />

      <Snackbar open={snackbar.open} autoHideDuration={3000} onClose={() => setSnackbar((s) => ({ ...s, open: false }))} anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}>
        <Alert onClose={() => setSnackbar((s) => ({ ...s, open: false }))} severity={snackbar.severity} variant="filled" sx={{ width: '100%' }}>
          {snackbar.message}
        </Alert>
      </Snackbar>
    </>
  );
}
