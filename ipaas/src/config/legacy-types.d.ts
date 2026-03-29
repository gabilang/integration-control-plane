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

/**
 * Legacy runtime configuration type — used by the Asgardeo/Choreo auth flow
 * (AuthContext.tsx, tokenManager.ts). Preserved for future use; not populated
 * when running with Thunder IdP.
 */
interface LegacyRuntimeConfig {
  authBaseUrl: string;
  graphqlUrl: string;
  asgardeoClientId: string;
  asgardeoAuthorizeEndpoint: string;
  asgardeoTokenEndpoint: string;
  asgardeoSignInRedirectUrl: string;
  asgardeoScope: string;
  stsTokenEndpoint?: string;
  stsClientId?: string;
  stsScope?: string;
  choreoOrgApiUrl: string;
  asgardeoOrgNumericId: number;
  sysApiPrefix?: string;
}

declare global {
  interface Window {
    /**
     * Legacy runtime config injected at startup via `public/config.json` loader.
     * Only present when running with the Asgardeo/Choreo auth flow.
     * Absent in the Thunder IdP flow (which uses `window._env_` instead).
     */
    API_CONFIG: LegacyRuntimeConfig;
  }
}

export {};
