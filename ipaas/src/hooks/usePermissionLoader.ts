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
 * All users are authenticated via Thunder OIDC — org-level permissions are
 * granted in full at login (AppLayout). Project/component permission fetching
 * from the local auth API is no longer needed.
 */
export function useLoadProjectPermissions(_orgHandle: string, _projectId: string) {
  // No-op: Thunder users inherit all permissions from org-level grant.
}

export function useLoadComponentPermissions(_orgHandle: string, _projectId: string, _componentId: string) {
  // No-op: Thunder users inherit all permissions from org-level grant.
}
