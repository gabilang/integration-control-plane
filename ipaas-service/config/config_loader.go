package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type configReader struct {
	errors []error
}

// Load reads configuration from environment variables.
// If ENV_FILE_PATH is set, variables are loaded from that file first.
func Load() (Config, error) {
	if envFile := os.Getenv("ENV_FILE_PATH"); envFile != "" {
		if err := loadEnvFile(envFile); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to load env file %s: %v\n", envFile, err)
		}
	}

	r := &configReader{}
	cfg := Config{
		ServerHost: r.readOptionalString("SERVER_HOST", "0.0.0.0"),
		ServerPort: r.readOptionalInt("SERVER_PORT", 8080),
		LogLevel:   r.readOptionalString("LOG_LEVEL", "info"),
		PlatformAPI: PlatformAPIConfig{
			BaseURL:    r.readRequiredString("PLATFORM_API_SERVICE_BASE_URL"),
			HostHeader: r.readOptionalString("PLATFORM_API_SERVICE_HOST", ""),
		},
		Observability: ObservabilityConfig{
			BaseURL: r.readOptionalString("OBSERVABILITY_SERVICE_BASE_URL", ""),
		},
		ICP: ICPConfig{
			GraphQLURL:  r.readOptionalString("ICP_GRAPHQL_URL", ""),
			AuthBaseURL: r.readOptionalString("ICP_AUTH_BASE_URL", ""),
		},
		GitHub: GitHubConfig{
			BaseURL: r.readOptionalString("GITHUB_API_BASE_URL", "https://api.github.com"),
			Token:   r.readOptionalString("GITHUB_TOKEN", ""),
		},
	}

	if len(r.errors) > 0 {
		msgs := make([]string, len(r.errors))
		for i, e := range r.errors {
			msgs[i] = e.Error()
		}
		return Config{}, fmt.Errorf("configuration errors:\n%s", strings.Join(msgs, "\n"))
	}

	return cfg, nil
}

func (r *configReader) readRequiredString(key string) string {
	val := os.Getenv(key)
	if val == "" {
		r.errors = append(r.errors, fmt.Errorf("%s is required", key))
	}
	return val
}

func (r *configReader) readOptionalString(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func (r *configReader) readOptionalInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		r.errors = append(r.errors, fmt.Errorf("%s must be an integer: %w", key, err))
		return defaultVal
	}
	return n
}

// loadEnvFile parses a simple KEY=VALUE env file and sets variables in the environment.
func loadEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		if os.Getenv(key) == "" {
			os.Setenv(key, value) //nolint:errcheck
		}
	}
	return scanner.Err()
}
