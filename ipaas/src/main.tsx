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

import { OxygenUIThemeProvider, AcrylicOrangeTheme } from '@wso2/oxygen-ui';
import { AsgardeoProvider } from '@asgardeo/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter } from 'react-router';
import { StrictMode, type ReactNode } from 'react';
import { createRoot } from 'react-dom/client';
import App from './App';
import { isBypassEnabled, MockAuthProvider } from './auth';
import { env } from './config/env';
import { AccessControlProvider } from './contexts/AccessControlContext';
import './index.css';

const queryClient = new QueryClient();

const defaultSignInUrl = env.VITE_THUNDER_AFTER_SIGN_IN_URL || (typeof window !== 'undefined' ? window.location.origin : '');
const defaultSignOutUrl =
  env.VITE_THUNDER_AFTER_SIGN_OUT_URL || (typeof window !== 'undefined' ? `${window.location.origin}/login` : '/login');

const asgardeoConfig = {
  baseUrl: env.VITE_THUNDER_URL,
  clientId: env.VITE_THUNDER_CLIENT_ID || 'IPAAS_CONSOLE',
  ...(env.VITE_THUNDER_CLIENT_SECRET && { clientSecret: env.VITE_THUNDER_CLIENT_SECRET }),
  signInUrl: `${env.VITE_THUNDER_URL}/gate`,
  afterSignInUrl: defaultSignInUrl,
  afterSignOutUrl: defaultSignOutUrl,
  scopes: (env.VITE_THUNDER_SCOPES || 'openid profile email').split(' '),
  platform: 'AsgardeoV2' as const,
  tokenValidation: {
    idToken: {
      validate: false,
    },
  },
  storage: 'localStorage' as const,
};

function AuthProviderWrapper({ children }: { children: ReactNode }) {
  if (isBypassEnabled()) {
    return <MockAuthProvider>{children}</MockAuthProvider>;
  }
  return <AsgardeoProvider {...asgardeoConfig}>{children}</AsgardeoProvider>;
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <AuthProviderWrapper>
      <OxygenUIThemeProvider themes={[{ key: 'acrylicOrange', label: 'Acrylic Orange Theme', theme: AcrylicOrangeTheme }]} initialTheme="acrylicOrange">
        <QueryClientProvider client={queryClient}>
          <BrowserRouter>
            <AccessControlProvider>
              <App />
            </AccessControlProvider>
          </BrowserRouter>
        </QueryClientProvider>
      </OxygenUIThemeProvider>
    </AuthProviderWrapper>
  </StrictMode>,
);
