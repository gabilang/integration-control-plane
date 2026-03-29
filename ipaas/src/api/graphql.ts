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

import { graphqlApiUrl } from '../config/api';
import { authenticatedFetch, getOrgUuidFromToken } from '../auth/tokenManager';

/**
 * Send a GraphQL request through the ipaas-service BFF proxy.
 * The BFF injects the upstream Bearer token automatically; the browser
 * only needs to include the session cookie / Asgardeo token, which the
 * SDK manages and injects via the Authorization header on every fetch.
 */
export async function gql<T>(query: string, variables?: Record<string, unknown>): Promise<T> {
  let res = await fetch(graphqlApiUrl(), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ query, variables }),
  });
  // Some environments return 404 for auth failures instead of 401.
  // Only refresh when the token is unscoped (STS is configured but the token lacks an org UUID),
  // which means it is a raw Asgardeo token rather than a proper org-scoped STS token.
  // If the token already has an org UUID, a 404 is a genuine resource-not-found — not an auth
  // failure — so we skip the refresh to avoid noisy connection-refused logs and unnecessary churn.
  if (res.status === 404) {
    const stsConfigured = !!window.API_CONFIG.stsTokenEndpoint && !!window.API_CONFIG.stsClientId;
    const tokenIsUnscoped = stsConfigured && !getOrgUuidFromToken();
    if (tokenIsUnscoped) {
      const { refreshAccessToken } = await import('../auth/tokenManager');
      await refreshAccessToken();
      res = await authenticatedFetch(graphqlApiUrl(), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query, variables }),
      });
    }
  }
  if (!res.ok) {
    const body = await res.text().catch(() => '');
    throw new Error(`GraphQL request failed (HTTP ${res.status}): ${body || res.statusText}`);
  }
  const json = await res.json();
  if (json.errors) throw new Error(json.errors[0].message);
  return json.data as T;
}
