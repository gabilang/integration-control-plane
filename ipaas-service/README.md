# integration-platform-service

A Go REST API service that acts as an API layer for the [OpenChoreo](https://github.com/openchoreo/openchoreo) integration platform. It proxies and exposes resource management operations to the Devant Console and other consumers.

## Project Structure

```
integration-platform-service/
├── cmd/
│   └── platform-api/
│       └── main.go                   # Entry point — wires dependencies and starts the HTTP server
├── api/
│   ├── app.go                        # Assembles the HTTP handler and middleware chain
│   ├── project_routes.go             # Project route registration
│   └── component_routes.go           # Component route registration
├── clients/
│   ├── openchoreo/
│   │   ├── project_client.go         # OpenChoreo API client for project operations
│   │   └── component_client.go       # OpenChoreo API client for component operations
│   ├── observability/
│   │   └── client.go                 # Observability service client for build logs
│   └── requests/
│       ├── http_request.go           # Fluent HTTP request builder
│       ├── http_error.go             # HTTP error type
│       └── send_request.go           # HTTP execution and response scanning
├── config/
│   ├── config.go                     # Configuration structs
│   └── config_loader.go              # Environment variable loading
├── controllers/
│   ├── project_controller.go         # HTTP handlers for project endpoints
│   └── component_controller.go       # HTTP handlers for component endpoints
├── middleware/
│   ├── auth_token.go                 # Bearer token extraction
│   ├── correlation_id.go             # Correlation ID propagation
│   ├── panic_recover.go              # Panic recovery
│   ├── jwt/
│   │   └── middleware.go             # JWT claims parsing
│   └── logger/
│       └── request_logger.go         # Structured request logging
├── models/
│   ├── project.go                    # Project request/response types
│   └── component.go                  # Component request/response types
├── services/
│   ├── project_service.go            # Business logic for projects
│   └── component_service.go          # Business logic for components and builds
└── utils/
    └── response.go                   # JSON response helpers
```

## Gateway Route Path

When deployed to OpenChoreo, the external gateway route path is auto-generated from the **component name** and the **endpoint name** defined in `workload.yaml`:

```
/{component-name}-{endpoint-name}
```

For this service, `workload.yaml` defines `name: integration-platform-api-http`, so the full gateway path becomes:

```
/integration-platform-api-service-integration-platform-api-http
```

The external gateway rewrites this prefix to `/integration-platform-api/v1.0`, then the internal API gateway (policy engine) strips that context path before forwarding to the pod. As a result, the Go service receives requests at root level (e.g. `/projects`, `/health`).

## API Endpoints

### Health

| Method | Path      | Description  |
|--------|-----------|--------------|
| `GET`  | `/health` | Health check |

### Projects

| Method   | Path                               | Description      |
|----------|------------------------------------|------------------|
| `GET`    | `/projects`                        | List projects    |
| `POST`   | `/projects`                        | Create a project |
| `GET`    | `/projects/{projectName}`          | Get a project    |
| `PUT`    | `/projects/{projectName}`          | Update a project |
| `DELETE` | `/projects/{projectName}`          | Delete a project |

### Components

| Method | Path                                                                         | Description          |
|--------|------------------------------------------------------------------------------|----------------------|
| `GET`  | `/projects/{projectName}/components`                                         | List components      |
| `POST` | `/projects/{projectName}/components`                                         | Create a component   |
| `GET`  | `/projects/{projectName}/components/{componentName}/builds`                  | List builds          |
| `GET`  | `/projects/{projectName}/components/{componentName}/builds/{buildName}`      | Get build status     |
| `GET`  | `/projects/{projectName}/components/{componentName}/builds/{buildName}/logs` | Get build logs       |

> **Note:** All endpoints except `/health` require a valid JWT Bearer token. The org context is extracted from the `ouHandle` claim.

## Configuration

Configuration is loaded from environment variables. Set `ENV_FILE_PATH` to point to a `.env` file for local development.

| Variable                        | Required | Default   | Description                                                                 |
|---------------------------------|----------|-----------|-----------------------------------------------------------------------------|
| `PLATFORM_API_SERVICE_BASE_URL` | Yes      | —         | Platform API base URL (e.g. `http://host/wso2cloud-dp`)                     |
| `OBSERVABILITY_SERVICE_BASE_URL`| No       | —         | Observability service base URL for build logs. If unset, logs return 503.   |
| `SERVER_HOST`                   | No       | `0.0.0.0` | Server bind address                                                         |
| `SERVER_PORT`                   | No       | `8080`    | Server port                                                                 |
| `LOG_LEVEL`                     | No       | `info`    | Log level: `debug`, `info`, `warn`, `error`                                 |
| `ENV_FILE_PATH`                 | No       | —         | Path to a `.env` file                                                       |

## Testing Changes Locally (Without GitHub CI)

Two options to test changes in the local k3d cluster without committing and waiting for CI.

### Option 1 — Build image locally and patch the deployment

```bash
# 1. Build the Docker image
cd integration-platform-service
docker build -t localhost:10082/devant-integration-platform-api-service-image:dev-local .

# 2. Push to the local k3d registry
docker push localhost:10082/devant-integration-platform-api-service-image:dev-local

# 3. Patch the running deployment to use the new image
kubectl set image deployment/integration-platform-api-service-development-f21e1891 \
  main=host.k3d.internal:10082/devant-integration-platform-api-service-image:dev-local \
  -n dp-wso2cloud-devant-development-0bea7a4d

# 4. Wait for rollout
kubectl rollout status deployment/integration-platform-api-service-development-f21e1891 \
  -n dp-wso2cloud-devant-development-0bea7a4d
```

Test via the gateway as normal:
```bash
curl 'http://development-wso2cloud.openchoreoapis.localhost:19080/integration-platform-api-service-integration-platform-api-http/projects' \
  -H 'Authorization: Bearer <your-token>'
```

To restore the original image when done:
```bash
kubectl set image deployment/integration-platform-api-service-development-f21e1891 \
  main=host.k3d.internal:10082/devant-integration-platform-api-service-image:v1-486a7431 \
  -n dp-wso2cloud-devant-development-0bea7a4d
```

---

### Option 2 — Run locally and port-forward the internal gateway

```bash
# 1. Port-forward the internal API gateway to localhost
kubectl port-forward -n openchoreo-data-plane svc/api-platform-default-gateway-router 18080:8080
```

In a separate terminal, run the service pointed at the forwarded gateway:
```bash
cd integration-platform-service
PLATFORM_API_SERVICE_BASE_URL="http://localhost:18080/cloud-core-api/v1.0/wso2cloud-dp" \
SERVER_PORT=9090 \
  go run ./cmd/platform-api/main.go
```

Test directly against the local process (note: JWT middleware still runs, so include a valid token):
```bash
curl 'http://localhost:9090/projects' \
  -H 'Authorization: Bearer <your-token>'
```

> **Note:** In Option 2 the request goes directly to your local process — it bypasses the kgateway, so the JWT is not validated by the gateway. The service still parses claims from the token to extract `ouHandle`.

---

## Development

```bash
# Tidy dependencies
make tidy

# Build binary
make build
# Output: bin/platform-api

# Run directly
make run

# Remove build artifacts
make clean
```
