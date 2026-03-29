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

import { env } from '../config/env';
import { useAsgardeo as realUseAsgardeo, User as RealUser, navigate as realNavigate } from '@asgardeo/react';
import { useMockAsgardeo, MockUser, mockNavigate, isBypassEnabled, MockAuthProvider, MOCK_USER } from './MockAuthProvider';

const bypassEnabled = env.VITE_DEV_BYPASS_AUTH === 'true';

export const useAsgardeo = (bypassEnabled ? useMockAsgardeo : realUseAsgardeo) as typeof realUseAsgardeo;
export const User = bypassEnabled ? MockUser : RealUser;
export const navigate = bypassEnabled ? mockNavigate : realNavigate;

export { isBypassEnabled, MockAuthProvider, MOCK_USER };
