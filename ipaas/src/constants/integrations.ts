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
 * The set of integration display types that are supported by the ipaas.
 * Components whose displayType is not in this set should be treated as read-only / disabled.
 */
export const SUPPORTED_DISPLAY_TYPES = new Set([
  'restAPI',
  'manualTrigger',
  'scheduledTask',
  'webhook',
  'miRestApi',
  'miEventHandler',
  'ballerinaService',
  'miApiService',
  'miCronjob',
  'miJob',
  'miWebhook',
  'ballerinaEventHandler',
  'ballerinaWebhook',
  'ballerinaFileIntegration',
  'miFileIntegration',
  'byoiService',
  // OpenChoreo BFF display types (buildpack + componentType)
  'miAutomation',
  'miService',
  'miAiAgent',
  'miEventIntegration',
  'miFileIntegration',
  'miProxy',
  'biAutomation',
  'biService',
  'biAiAgent',
  'biEventIntegration',
  'biFileIntegration',
  'biProxy'
]);

/**
 * Maps a component's displayType and componentSubType to a human-readable label.
 */
export function getDisplayLabel(displayType: string, componentSubType: string | null): string {
  switch (componentSubType) {
    case 'ballerinaFileIntegration':
    case 'miFileIntegration':
      return 'File Integration';
    case 'aiAgent':
      return 'AI Agent';
    case 'mcpServer':
      return 'MCP Server';
  }
  switch (displayType) {
    case 'ballerinaService':
    case 'buildpackService':
    case 'byoiService':
    case 'byocService':
    case 'miApiService':
    case 'graphql':
    case 'thirdPartyApi':
    case 'prismMockService':
    case 'service':
    case 'miService':
    case 'biService':
      return 'Integration as API';
    case 'scheduledTask':
    case 'byocCronjob':
    case 'byoiCronjob':
    case 'miCronjob':
    case 'buildpackCronJob':
    case 'automation':
    case 'miAutomation':
    case 'biAutomation':
      return 'Automation';
    case 'manualTrigger':
      return 'Manual Trigger';
    case 'proxy':
    case 'gitProxy':
    case 'miProxy':
    case 'biProxy':
      return 'REST API Proxy';
    case 'webhook':
    case 'byocWebhook':
    case 'miWebhook':
    case 'buildpackWebhook':
    case 'ballerinaWebhook':
      return 'Webhook';
    case 'byocEventHandler':
    case 'miEventHandler':
    case 'buildpackEventHandler':
    case 'ballerinaEventHandler':
    case 'eventIntegration':
    case 'miEventIntegration':
    case 'biEventIntegration':
      return 'Event Integration';
    case 'restApi':
    case 'byocRestApi':
    case 'miRestApi':
    case 'buildRestApi':
      return 'Integration as an API';
    case 'byocWebApp':
    case 'byocWebAppsDockerfileLess':
    case 'byoiWebApp':
    case 'buildpackWebApp':
      return 'Web Application';
    case 'byocTestRunner':
    case 'buildpackTestRunner':
    case 'byocTestRunnerDockerfileLess':
      return 'Test Runner';
    case 'externalConsumer':
      return 'External Consumer';
    case 'byocJob':
    case 'byoiJob':
    case 'miJob':
    case 'buildpackJob':
      return 'Manual Task';
    case 'fileIntegration':
    case 'miFileIntegration':
    case 'biFileIntegration':
      return 'File Integration';
    case 'aiAgent':
    case 'miAiAgent':
    case 'biAiAgent':
      return 'AI Agent';
    default:
      return displayType ?? 'Unknown';
  }
}
