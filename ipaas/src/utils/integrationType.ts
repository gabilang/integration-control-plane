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

import type { DisplayType } from '../types/component';
import type { WorkspaceIntegrationType } from '../types/project';

/** Source technology of an integration: Ballerina Integrator or Micro Integrator. */
export type IntegrationTechnology = 'BI' | 'MI';

/**
 * An integration's identity is carried by two fields, not one. Automation,
 * Integration as API and Event Integration are distinguished by `displayType`;
 * File Integration, AI Agent and MCP Server share a generic service
 * `displayType` and are told apart by `componentSubType`. Resolving only the
 * first collapses four of the six types into Integration as API.
 *
 * These take `WorkspaceIntegrationType`, whose Integration-as-API member is
 * spelled `service`. ImportIntegration keeps its own copy because it works in
 * the wider `IntegrationType` vocabulary, where the same member is spelled
 * `integration-as-api` and several values have no workspace equivalent.
 */
export function resolveDisplayType(technology: IntegrationTechnology, integrationType: WorkspaceIntegrationType): DisplayType {
  if (technology === 'BI') {
    if (integrationType === 'automation') return 'scheduledTask';
    if (integrationType === 'event-integration') return 'ballerinaEventHandler';
    return 'ballerinaService';
  }
  if (integrationType === 'automation') return 'miCronjob';
  if (integrationType === 'event-integration') return 'miEventHandler';
  return 'miApiService';
}

export function resolveComponentSubType(technology: IntegrationTechnology, integrationType: WorkspaceIntegrationType): string | undefined {
  if (integrationType === 'file-integration') return technology === 'BI' ? 'ballerinaFileIntegration' : 'miFileIntegration';
  if (integrationType === 'ai-agent') return 'aiAgent';
  if (integrationType === 'mcp-server') return 'MCP';
  return undefined;
}
