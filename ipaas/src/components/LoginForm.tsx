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

import type { JSX } from 'react';
import { Box, Button, CircularProgress, Typography } from '@wso2/oxygen-ui';
import { useAsgardeo } from '../auth';

export default function LoginForm(): JSX.Element {
  const { signIn, isLoading } = useAsgardeo();

  return (
    <Box>
      <Box sx={{ mb: 6 }}>
        <Typography variant="h1" gutterBottom>
          Sign In
        </Typography>
        <Typography color="text.secondary">
          Sign in with your WSO2 Integration Platform account.
        </Typography>
      </Box>

      <Button
        fullWidth
        variant="contained"
        color="primary"
        size="large"
        onClick={() => signIn()}
        disabled={isLoading}
        startIcon={isLoading ? <CircularProgress size={20} color="inherit" /> : undefined}>
        {isLoading ? 'Redirecting...' : 'Sign In'}
      </Button>
    </Box>
  );
}
