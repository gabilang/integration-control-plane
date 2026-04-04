package github

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/wso2/integration-control-plane/ipaas-service/models"
)

// Client provides read-only access to the GitHub REST API for fetching
// commit history from public (or token-accessible) repositories.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a GitHub API client.
// baseURL is typically "https://api.github.com". token is optional;
// when empty, requests are unauthenticated (subject to lower rate limits).
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// IsAvailable returns true when the client has a configured base URL.
func (c *Client) IsAvailable() bool {
	return c.baseURL != ""
}

// ghCommitAuthor is the nested author object inside a GitHub commit.
type ghCommitAuthor struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Date  string `json:"date"`
}

// ghCommitInner is the "commit" sub-object of a GitHub commit response.
type ghCommitInner struct {
	Message string         `json:"message"`
	Author  ghCommitAuthor `json:"author"`
}

// ghAuthor is the top-level author with avatar URL.
type ghAuthor struct {
	AvatarURL string `json:"avatar_url"`
}

// ghCommit represents a single commit from the GitHub List Commits API.
type ghCommit struct {
	SHA    string        `json:"sha"`
	Commit ghCommitInner `json:"commit"`
	Author *ghAuthor     `json:"author"`
}

// ParseRepoOwnerAndName extracts the owner and repo name from a GitHub
// repository URL (e.g. "https://github.com/owner/repo" or
// "https://github.com/owner/repo.git").
func ParseRepoOwnerAndName(repoURL string) (owner, repo string, err error) {
	u, err := url.Parse(repoURL)
	if err != nil {
		return "", "", fmt.Errorf("parse repo URL: %w", err)
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid repo URL path: %s", u.Path)
	}

	owner = parts[0]
	repo = strings.TrimSuffix(parts[1], ".git")
	return owner, repo, nil
}

// GetCommits fetches the most recent commits for a branch from the GitHub API.
// GET /repos/{owner}/{repo}/commits?sha={branch}&per_page={limit}
func (c *Client) GetCommits(ctx context.Context, repoURL, branch string, limit int) ([]models.Commit, error) {
	owner, repo, err := ParseRepoOwnerAndName(repoURL)
	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 20
	}

	apiURL := fmt.Sprintf("%s/repos/%s/%s/commits?sha=%s&per_page=%d",
		c.baseURL,
		url.PathEscape(owner),
		url.PathEscape(repo),
		url.QueryEscape(branch),
		limit,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.WarnContext(ctx, "github API non-200 response",
			"status", resp.StatusCode, "owner", owner, "repo", repo, "branch", branch)
		return nil, fmt.Errorf("github API returned %d", resp.StatusCode)
	}

	var ghCommits []ghCommit
	if err := json.NewDecoder(resp.Body).Decode(&ghCommits); err != nil {
		return nil, fmt.Errorf("decode github response: %w", err)
	}

	commits := make([]models.Commit, 0, len(ghCommits))
	for i, gc := range ghCommits {
		c := models.Commit{
			SHA:      gc.SHA,
			Message:  gc.Commit.Message,
			IsLatest: i == 0,
			Author: models.CommitAuthor{
				Name:  gc.Commit.Author.Name,
				Email: gc.Commit.Author.Email,
				Date:  gc.Commit.Author.Date,
			},
		}
		if gc.Author != nil {
			c.Author.AvatarURL = gc.Author.AvatarURL
		}
		commits = append(commits, c)
	}

	return commits, nil
}
