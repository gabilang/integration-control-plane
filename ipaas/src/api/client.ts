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
 * ipaas-service REST API client.
 *
 * Thin wrapper around the browser Fetch API targeting the ipaas-service BFF.
 * All REST API modules import from here so auth headers, base URL, and error
 * handling stay in one place.
 *
 * The access token is supplied lazily via setTokenAccessor(), which is wired
 * in App.tsx using useAsgardeo().getAccessToken — the same pattern used by the
 * reference integration-platform console.
 */

import { icpApiBaseUrl } from '../config/api';

// eslint-disable-next-line @typescript-eslint/naming-convention, no-underscore-dangle
let _getAccessToken: (() => Promise<string>) | null = null;

/**
 * Set a lazy accessor for the Bearer token.
 *
 * Pass a function that resolves the current access token (e.g. Asgardeo's
 * getAccessToken). The function is called on every request so the auth
 * library can transparently refresh the token between calls.
 *
 * Call with null to clear the accessor (e.g. after sign-out).
 */
export function setTokenAccessor(fn: (() => Promise<string>) | null): void {
  _getAccessToken = fn;
}

async function request<T>(
  method: string,
  path: string,
  options?: {
    params?: Record<string, string | undefined>;
    body?: unknown;
  },
): Promise<T> {
  const base = icpApiBaseUrl();
  // Support both absolute base URLs (http://host/path) and relative paths (/path).
  const isAbsolute = /^https?:\/\//.test(base);
  const url = isAbsolute ? new URL(`${base}${path}`) : new URL(`${base}${path}`, window.location.origin);

  if (options?.params) {
    for (const [k, v] of Object.entries(options.params)) {
      if (v !== undefined && v !== '') url.searchParams.set(k, v);
    }
  }

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    Accept: 'application/json',
  };

  if (_getAccessToken) {
    const token = await _getAccessToken();
    if (token) headers['Authorization'] = `Bearer ${token}`;
  }

  const init: RequestInit = { method, headers };
  if (options?.body !== undefined) {
    init.body = JSON.stringify(options.body);
  }

  // Use relative path for same-origin requests, full URL for cross-origin.
  const fetchUrl = isAbsolute ? url.href : url.pathname + url.search;
  const res = await fetch(fetchUrl, init);

  if (!res.ok) {
    const text = await res.text().catch(() => res.statusText);
    throw new Error(`API error ${res.status}: ${text}`);
  }

  if (res.status === 204) return {} as T;

  return res.json() as Promise<T>;
}

export const icpClient = {
  get: <T>(path: string, params?: Record<string, string | undefined>) => request<T>('GET', path, { params }),
  post: <T>(path: string, body?: unknown) => request<T>('POST', path, { body }),
  put: <T>(path: string, body?: unknown) => request<T>('PUT', path, { body }),
  delete: <T>(path: string) => request<T>('DELETE', path),
};
