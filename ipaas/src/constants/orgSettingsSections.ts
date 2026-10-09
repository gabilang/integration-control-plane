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

import { IS_CLOUD } from '../features';
import { ALL_USER_MGT_PERMISSIONS, Permissions } from './permissions';

/** A single org-Settings section: the tab label, its route, and the permissions that reveal it. */
export interface SettingsSectionDef {
  /** Stable id; also the active-tab key passed by each section page. */
  id: string;
  label: string;
  /** Route suffix after `/settings/` (may include a default sub-tab). */
  path: string;
  /** Any one of these permissions makes the tab visible. */
  permissions: readonly string[];
}

/**
 * Org Settings sections, in display order. Permission mapping follows the Devant
 * gates, translated to the nearest ICP permission: user-mgt → user view,
 * deployment-manage → environment manage. Credentials + APIM arrive in Phase 1b.
 */
export const SETTINGS_SECTIONS: readonly SettingsSectionDef[] = IS_CLOUD
  ? [
      // Cloud supports neither the user-management nor the security sections yet, so
      // Settings carries only the package registries.
      { id: 'package-registries', label: 'Package Registries', path: 'package-registries', permissions: [Permissions.ENVIRONMENT_MANAGE] },
    ]
  : [
      { id: 'access-control', label: 'Access Control', path: 'access-control/users', permissions: ALL_USER_MGT_PERMISSIONS },
      { id: 'egress-control', label: 'Egress Control', path: 'egress-control', permissions: [Permissions.ENVIRONMENT_MANAGE] },
      { id: 'workflows', label: 'Workflows', path: 'workflows', permissions: [Permissions.USER_VIEW] },
      // Credentials is always available (no permission gate, matching Devant).
      { id: 'credentials', label: 'Credentials', path: 'credentials', permissions: [] },
      { id: 'on-prem-keys', label: 'On-Prem Keys', path: 'on-prem-keys', permissions: [Permissions.ENVIRONMENT_MANAGE] },
      { id: 'application-security', label: 'Application Security', path: 'application-security/identity-providers', permissions: [Permissions.USER_VIEW] },
    ];

/** A section is visible if it has no gate or the user holds one of its permissions. */
export function isSettingsSectionVisible(s: SettingsSectionDef, hasAnyPermission: (perms: string[]) => boolean): boolean {
  return s.permissions.length === 0 || hasAnyPermission([...s.permissions]);
}

/** The first section the user is allowed to see, or `null` if none. */
export function firstAvailableSettingsSection(hasAnyPermission: (perms: string[]) => boolean): SettingsSectionDef | null {
  return SETTINGS_SECTIONS.find((s) => isSettingsSectionVisible(s, hasAnyPermission)) ?? null;
}
