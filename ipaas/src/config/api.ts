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
 * API URL helpers derived from the runtime environment (window._env_).
 * Configuration is injected at container startup via docker-entrypoint.sh → env-config.js,
 * so no /config.json fetch is needed and no rebuild is required when endpoints change.
 */

import { env } from './env';

/** Base URL of the ipaas-service BFF (e.g. /ipaas-service). */
export const icpApiBaseUrl = (): string => env.ICP_API_BASE_URL;

/** GraphQL proxy — POSTed to by all GraphQL queries/mutations. */
export const graphqlApiUrl = (): string => `${icpApiBaseUrl()}/graphql`;

/** Auth proxy — base for all user/role/group management calls. */
export const authApiUrl = (): string => `${icpApiBaseUrl()}/auth`;

/** Observability proxy — base for logs and metrics. */
export const observabilityApiUrl = (): string => `${icpApiBaseUrl()}/observability`;
