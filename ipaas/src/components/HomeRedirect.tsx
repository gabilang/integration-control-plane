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

import { useState, useEffect, type JSX } from 'react';
import { Navigate } from 'react-router';
import { useAsgardeo } from '../auth';
import { loginUrl } from '../paths';

/**
 * Root route handler. Redirects authenticated users to their org home
 * (resolved from the JWT `ouHandle` claim), and unauthenticated users
 * to the login page.
 */
export default function HomeRedirect(): JSX.Element {
  const { isSignedIn, isLoading, getDecodedIdToken, getAccessToken } = useAsgardeo();
  const [orgHandle, setOrgHandle] = useState<string | null>(null);

  useEffect(() => {
    if (!isSignedIn) {
      setOrgHandle(null);
      return;
    }

    let cancelled = false;

    async function resolveOrgHandle() {
      let handle: string | undefined;

      // 1. Try ID token claims first
      try {
        const idToken = await getDecodedIdToken();
        const ouHandle = (idToken as Record<string, unknown>)?.ouHandle;
        if (typeof ouHandle === 'string' && ouHandle.trim()) {
          handle = ouHandle.trim();
        }
      } catch { /* ignore */ }

      // 2. Fallback: decode access token
      // (Thunder may not include ouHandle in id_token due to scope_claims filtering)
      if (!handle) {
        try {
          const accessToken = await getAccessToken();
          if (accessToken) {
            const payload = JSON.parse(
              atob(accessToken.split('.')[1].replace(/-/g, '+').replace(/_/g, '/'))
            ) as Record<string, unknown>;
            const ouHandle = payload?.ouHandle;
            if (typeof ouHandle === 'string' && ouHandle.trim()) {
              handle = ouHandle.trim();
            }
          }
        } catch { /* ignore */ }
      }

      if (!cancelled) {
        setOrgHandle(handle ?? 'default');
      }
    }

    resolveOrgHandle();

    return () => {
      cancelled = true;
    };
  }, [isSignedIn, getDecodedIdToken, getAccessToken]);

  if (isLoading) return <></>;

  if (!isSignedIn) {
    return <Navigate to={loginUrl()} replace />;
  }

  // Still resolving org handle from token
  if (orgHandle === null) return <></>;

  return <Navigate to={`/organizations/${orgHandle}`} replace />;
}
