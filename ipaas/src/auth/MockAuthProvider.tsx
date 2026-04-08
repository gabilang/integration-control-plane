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

import { createContext, useContext, type ReactNode } from 'react';
import { env } from '../config/env';

export const MOCK_USER = {
  sub: 'dev-user-uuid-12345',
  email: 'john.doe@dev.local',
  name: 'John Doe',
  given_name: 'John',
  family_name: 'Doe',
  // SDK KnownUser fields
  username: 'john.doe@dev.local',
  displayName: 'John Doe',
  givenName: 'John',
  familyName: 'Doe',
  // Thunder org claims
  ouName: 'default',
  ouHandle: 'default',
};

export function isBypassEnabled(): boolean {
  return env.VITE_DEV_BYPASS_AUTH === 'true';
}

interface HttpRequestConfig {
  url: string;
  method: string;
  headers?: Record<string, string>;
  data?: unknown;
}

interface MockAuthState {
  isLoading: boolean;
  isSignedIn: boolean;
  user: typeof MOCK_USER | null;
  signIn: () => Promise<void>;
  signOut: () => Promise<void>;
  clearSession: () => void;
  getDecodedIdToken: () => Promise<Record<string, unknown>>;
  getAccessToken: () => Promise<string>;
  http: {
    request: <T = unknown>(config: HttpRequestConfig) => Promise<{ data: T }>;
  };
}

const MockAuthContext = createContext<MockAuthState | null>(null);

export function useMockAsgardeo(): MockAuthState {
  const context = useContext(MockAuthContext);
  if (!context) {
    throw new Error('useMockAsgardeo must be used within MockAuthProvider');
  }
  return context;
}

export function MockUser({ children }: { children: (user: typeof MOCK_USER) => ReactNode }) {
  return <>{children(MOCK_USER)}</>;
}

export function mockNavigate(path: string): void {
  window.location.href = path;
}

export function MockAuthProvider({ children }: { children: ReactNode }) {
  if (typeof window !== 'undefined') {
    console.warn(
      '%c AUTH BYPASS ACTIVE ',
      'background: #ff9800; color: black; font-weight: bold; padding: 4px 8px; border-radius: 4px;',
      '\nUsing mock authentication: John Doe / default',
      '\nSet VITE_DEV_BYPASS_AUTH=false to disable',
    );
  }

  const mockState: MockAuthState = {
    isLoading: false,
    isSignedIn: true,
    user: MOCK_USER,

    signIn: async () => {
      console.log('[MockAuth] signIn called - no-op in bypass mode');
    },

    signOut: async () => {
      console.log('[MockAuth] signOut called - redirecting to /login');
      window.location.href = '/login';
    },

    clearSession: () => {
      console.log('[MockAuth] clearSession called - redirecting to /login');
      window.location.href = '/login';
    },

    getDecodedIdToken: async () => ({
      sub: MOCK_USER.sub,
      email: MOCK_USER.email,
      name: MOCK_USER.name,
      given_name: MOCK_USER.given_name,
      family_name: MOCK_USER.family_name,
      ouName: MOCK_USER.ouName,
      ouHandle: MOCK_USER.ouHandle,
      exp: Math.floor(Date.now() / 1000) + 3600,
    }),

    getAccessToken: async () => 'mock-access-token',

    http: {
      request: async function <T = unknown>(config: HttpRequestConfig) {
        const response = await fetch(config.url, {
          method: config.method,
          headers: config.headers,
          body: config.data ? JSON.stringify(config.data) : undefined,
        });

        let data: T;
        const contentType = response.headers.get('content-type') || '';
        if (contentType.includes('application/json')) {
          data = (await response.json()) as T;
        } else {
          data = (await response.text()) as T;
        }

        return { data };
      },
    },
  };

  return <MockAuthContext.Provider value={mockState}>{children}</MockAuthContext.Provider>;
}
