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

import { authApiUrl } from '../config/api';

async function authFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${authApiUrl()}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });
  if (!res.ok) {
    const body = await res.text();
    throw new Error(body || `Request failed (${res.status})`);
  }
  const text = await res.text();
  return text ? JSON.parse(text) : (undefined as T);
}

export function authGet<T>(path: string): Promise<T> {
  return authFetch<T>(path);
}

export function authPost<T>(path: string, body: unknown): Promise<T> {
  return authFetch<T>(path, { method: 'POST', body: JSON.stringify(body) });
}

export function authPut<T>(path: string, body: unknown): Promise<T> {
  return authFetch<T>(path, { method: 'PUT', body: JSON.stringify(body) });
}

export function authDelete<T>(path: string): Promise<T> {
  return authFetch<T>(path, { method: 'DELETE' });
}

// ── Permission fetching ──

interface UserPermissionsResponse {
  userId: string;
  scope: { orgUuid: string; projectUuid?: string; integrationUuid?: string; envUuid?: string };
  permissions: { permissionId: number; permissionName: string; permissionDomain: string; description: string }[];
  permissionNames: string[];
}

export function fetchOrgPermissions(orgHandle: string, userId: string): Promise<UserPermissionsResponse> {
  return authGet<UserPermissionsResponse>(`/orgs/${orgHandle}/users/${userId}/permissions`);
}

export function fetchProjectPermissions(orgHandle: string, userId: string, projectId: string): Promise<UserPermissionsResponse> {
  return authGet<UserPermissionsResponse>(`/orgs/${orgHandle}/users/${userId}/permissions?projectId=${projectId}`);
}

export function fetchComponentPermissions(orgHandle: string, userId: string, projectId: string, componentId: string): Promise<UserPermissionsResponse> {
  return authGet<UserPermissionsResponse>(`/orgs/${orgHandle}/users/${userId}/permissions?projectId=${projectId}&integrationId=${componentId}`);
}

// ── Types ──

export interface User {
  userId: string;
  username: string;
  displayName: string;
  isSuperAdmin: boolean;
  isOidcUser: boolean;
  groups: { groupId: string; groupName: string; groupDescription: string }[];
  groupCount: number;
}

export interface Role {
  roleId: string;
  roleName: string;
  description: string;
  orgId: number;
}

export interface Permission {
  permissionId: string;
  permissionName: string;
  permissionDomain: string;
  resourceType: string;
  action: string;
  description: string;
}

export interface PermissionsResponse {
  permissions: Permission[];
  groupedByDomain: Record<string, Permission[]>;
}

export interface RoleDetail extends Role {
  permissions: Permission[];
}

export interface Group {
  groupId: string;
  groupName: string;
  description: string;
}

export interface GroupRoleMapping {
  id: number;
  groupId: string;
  roleId: string;
  roleName: string;
  roleDescription: string;
  orgUuid: string;
  projectUuid: string | null;
  envUuid: string | null;
  integrationUuid: string | null;
}

export interface GroupUser {
  userId: string;
  username: string;
  displayName: string;
}

export interface RoleGroupMapping {
  id: number;
  groupId: string;
  groupName?: string;
  groupDescription?: string;
  roleId: string;
  orgUuid: string;
  projectUuid: string | null;
  envUuid: string | null;
  integrationUuid: string | null;
}
