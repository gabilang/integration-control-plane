package config

type Config struct {
	ServerHost string
	ServerPort int
	LogLevel   string

	PlatformAPI   PlatformAPIConfig
	Observability ObservabilityConfig
	ICP           ICPConfig
	GitHub        GitHubConfig
}

// PlatformAPIConfig holds connection settings for the platform-api-service,
// which acts as the API gateway to the OpenChoreo control plane.
// When calling through the data-plane gateway, HostHeader must be set so the
// gateway can route the request to the correct backend.
type PlatformAPIConfig struct {
	BaseURL    string
	HostHeader string
}

// ObservabilityConfig holds connection settings for the observability service.
// BaseURL is optional; if empty, observability endpoints return an unavailable error.
type ObservabilityConfig struct {
	BaseURL string
}

// ICPConfig holds connection settings for the ICP backends (GraphQL + Auth).
// GraphQLURL is the ICP GraphQL API endpoint; AuthBaseURL is the ICP auth REST API.
// Both are optional; if empty the corresponding proxy endpoints return 503.
type ICPConfig struct {
	GraphQLURL  string
	AuthBaseURL string
}

// GitHubConfig holds connection settings for the GitHub REST API.
// BaseURL defaults to "https://api.github.com". Token is optional; when empty,
// unauthenticated requests are used (subject to lower rate limits).
type GitHubConfig struct {
	BaseURL string
	Token   string
}
