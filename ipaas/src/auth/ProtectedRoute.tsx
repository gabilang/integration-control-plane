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

import { useEffect } from 'react';
import type { JSX } from 'react';
import { Navigate, Outlet, useLocation } from 'react-router';
import { useAsgardeo } from './index';
import { useAccessControl } from '../contexts/AccessControlContext';
import { loginUrl } from '../paths';
import { Permissions } from '../constants/permissions';

export default function ProtectedRoute(): JSX.Element {
  const { isSignedIn, isLoading } = useAsgardeo();
  const { setOrgPermissions } = useAccessControl();
  const { pathname } = useLocation();

  useEffect(() => {
    if (!isSignedIn) {
      setOrgPermissions([]);
      return;
    }
    // All Thunder-authenticated users receive full org-level permissions.
    // Fine-grained access control is enforced at the BFF layer.
    setOrgPermissions(Object.values(Permissions));
  }, [isSignedIn, pathname, setOrgPermissions]);

  // Show nothing while the SDK determines auth state
  if (isLoading) return <></>;

  if (!isSignedIn) {
    return <Navigate to={loginUrl()} replace />;
  }

  return <Outlet />;
}
