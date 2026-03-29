# Troubleshooting 401 Unauthorized Errors

## Overview

401 errors in this setup are almost always caused by stale JWT signing keys propagating through the gateway stack. Understanding the request flow helps diagnose which layer is failing.

## Request Flow

Every API call through the gateway follows this path:

```
Browser / curl
    │
    │  Bearer JWT (issued by platform-idp)
    ▼
kgateway (gateway-default, openchoreo-data-plane)
    │  Path rewrite: /integration-platform-api-service → /integration-platform-api/v1.0
    ▼
api-platform-default-gateway-router (openchoreo-data-plane:8080)
    │  JWT validation against platform-idp JWKS
    │  On success: forwards request to backend pod
    ▼
integration-platform-api-service pod (dp-default-devant-development-*)
    │  Extracts ouHandle from JWT claims
    │  Calls platform-api-service with Bearer token
    ▼
platform-api-service pod (dp-default-core-development-*)
    │  Acquires a system token from platform-idp
    │  Calls openchoreo-api with system token
    ▼
openchoreo-api (openchoreo-control-plane)
    │  Validates system token against platform-idp JWKS
    ▼
Response
```

## The Root Cause: platform-idp Key Rotation on Restart

The `platform-idp` pod generates a new RSA signing key pair **every time it restarts**. It does not persist keys across restarts.

Every component that validates JWTs caches the JWKS (JSON Web Key Set) from platform-idp:

| Component | Cache TTL | Impact when key rotates |
|-----------|-----------|------------------------|
| `api-platform-default-gateway-policy-engine` | 5 min (config) | Stops accepting user JWTs signed with the new key |
| `openchoreo-api` | Unknown (internal) | Stops accepting system tokens signed with the new key |
| `platform-api-service` | In-memory (process lifetime) | Caches a system token that becomes invalid after key rotation |

When platform-idp restarts:
1. A new key pair is generated with a new `kid`
2. All tokens issued before the restart are immediately invalid (different `kid`)
3. New tokens are signed with the new key
4. Cached JWKS in downstream components still reference the old `kid` → validation fails → 401

## Symptoms

| Response body | Source | Meaning |
|---|---|---|
| `{"error":"Unauthorized","message":"Authentication failed."}` | `api-platform-default-gateway-policy-engine` | Internal gateway rejected the JWT — JWKS cache is stale |
| `{"error":"Unauthorized","message":"invalid or expired token"}` | `integration-platform-api-service` | Request passed the gateway but the system token that platform-api-service uses to call openchoreo-api was rejected |

## Fix

Restart the internal gateway stack and openchoreo-api to force JWKS cache refresh:

```bash
# Restart the internal API gateway (forces JWKS re-fetch from platform-idp)
kubectl rollout restart \
  deployment/api-platform-default-gateway-router \
  deployment/api-platform-default-gateway-controller \
  deployment/api-platform-default-gateway-policy-engine \
  -n openchoreo-data-plane

# Restart openchoreo-api (clears its JWKS/system token cache)
kubectl rollout restart deployment/openchoreo-api -n openchoreo-control-plane

# Restart platform-api-service (clears its cached system token)
kubectl rollout restart deployment/platform-api-service-development-791c5eec \
  -n dp-default-core-development-32790e9d
```

Wait for all rollouts to complete, then retry the request.

## Diagnosis Steps

If you get a 401, follow these steps to identify the layer:

### Step 1 — Check where the 401 originates

The response body tells you the source:

```bash
curl -s 'http://development-wso2cloud.openchoreoapis.localhost:19080/integration-platform-api-service/projects' \
  -H 'Authorization: Bearer <token>'
```

- `"Authentication failed."` → internal gateway (policy-engine JWKS cache stale)
- `"invalid or expired token"` → our service got 401 from platform-api-service
- `"missing org context"` → JWT is missing the `ouHandle` claim

### Step 2 — Verify the token's kid exists in platform-idp JWKS

```bash
kubectl exec -n dp-default-devant-development-f57c8f69 \
  deployment/integration-platform-api-service-development-f21e1891 -- \
  wget -qO- --no-check-certificate \
  'https://platform-idp-development-69f4ee10.dp-default-core-development-32790e9d.svc.cluster.local:8090/oauth2/jwks'
```

Decode the JWT header (base64) and confirm the `kid` in the token matches a key in the JWKS output. If it does not match, the platform-idp has rotated keys and the token is genuinely invalid — get a new token from the console.

### Step 3 — Check platform-idp restart history

```bash
kubectl get pod -n dp-default-core-development-32790e9d \
  -l openchoreo.dev/component=platform-idp
```

A high restart count (e.g. 12+) means the key has rotated many times. Apply the fix above.

### Step 4 — Check platform-api-service system token failures

```bash
kubectl logs -n dp-default-core-development-32790e9d \
  deployment/platform-api-service-development-791c5eec --tail=50 \
  | grep -E "401|system_token|forward_request"
```

If you see `get_system_token: success` followed by `statusCode: 401` from openchoreo-api, the system token is stale. Restart platform-api-service.

## Prevention

The platform-idp restarts frequently in the local k3d environment because it is running without persistent storage — its database is in-memory and it crashes under certain conditions. This is a known limitation of the local dev setup.

In a production or stable staging environment, the platform-idp is deployed with persistent storage and does not rotate keys on restart, so this issue does not occur.
