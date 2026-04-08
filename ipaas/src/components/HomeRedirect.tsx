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
import { loginUrl, orgHomeUrl, projectHomeUrl } from '../paths';
import { getAndClearLastProjectUrl } from '../auth/tokenManager';
import { icpClient } from '../api/client';
import type { BffProjectList } from '../api/queries';

/**
 * Root route handler. After sign-in, redirects to:
 * 1. The last visited project (saved in localStorage) if it belongs to the current org.
 * 2. The first project in the org (fetched from the BFF).
 * 3. The org home page as a fallback when no projects exist.
 *
 * Unauthenticated users are sent to the login page.
 */
export default function HomeRedirect(): JSX.Element {
  const { isSignedIn, isLoading, getDecodedIdToken, getAccessToken } = useAsgardeo();
  const [redirectUrl, setRedirectUrl] = useState<string | null>(null);

  useEffect(() => {
    if (!isSignedIn) {
      setRedirectUrl(null);
      return;
    }

    let cancelled = false;

    async function resolveRedirect() {
      // Step 1: Resolve orgHandle from token claims
      let handle: string | undefined;

      try {
        const idToken = await getDecodedIdToken();
        const ouHandle = (idToken as Record<string, unknown>)?.ouHandle;
        if (typeof ouHandle === 'string' && ouHandle.trim()) {
          handle = ouHandle.trim();
        }
      } catch { /* ignore */ }

      // Fallback: decode access token
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

      const orgHandle = handle ?? 'default';
      const orgPrefix = `/organizations/${orgHandle}/`;

      // Step 2: Check localStorage for last visited project URL
      const savedUrl = getAndClearLastProjectUrl();
      if (savedUrl && savedUrl.startsWith(orgPrefix)) {
        if (!cancelled) setRedirectUrl(savedUrl);
        return;
      }

      // Step 3: Fetch first project from BFF
      try {
        const { items } = await icpClient.get<BffProjectList>('/projects');
        if (items && items.length > 0) {
          if (!cancelled) setRedirectUrl(projectHomeUrl(orgHandle, items[0].name));
          return;
        }
      } catch { /* ignore — fall through to org home */ }

      // Step 4: Fall back to org home
      if (!cancelled) setRedirectUrl(orgHomeUrl(orgHandle));
    }

    resolveRedirect();

    return () => {
      cancelled = true;
    };
  }, [isSignedIn, getDecodedIdToken, getAccessToken]);

  if (isLoading) return <></>;

  if (!isSignedIn) {
    return <Navigate to={loginUrl()} replace />;
  }

  // Still resolving destination
  if (redirectUrl === null) return <></>;

  return <Navigate to={redirectUrl} replace />;
}
