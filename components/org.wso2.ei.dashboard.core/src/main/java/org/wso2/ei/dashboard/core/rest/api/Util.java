/*
 * Copyright (c) 2020, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
 *
 * WSO2 Inc. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 *
 *
 */

package org.wso2.ei.dashboard.core.rest.api;

import javax.ws.rs.core.Response;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.wso2.ei.dashboard.core.exception.UnAuthorizedException;

/**
 * Util class to be used by APIs.
 */
public class Util {

    private static final Log log = LogFactory.getLog(Util.class);

    static Response.ResponseBuilder logAndBuildUnAuthorizedResponse(String message, UnAuthorizedException e) {
        log.error(message, e);
        return Response.status(Response.Status.UNAUTHORIZED);
    }
}
