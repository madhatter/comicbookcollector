package metron

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type Client struct {
	httpClient *http.Client
	limiter    *rate.Limiter
	username   string
	password   string
	baseURL    string
}

// NewClient creates a new Metron API client with the given credentials and a rate limiter to avoid hitting API limits.
func NewClient(username, password string) (*Client, error) {
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if password == "" {
		return nil, fmt.Errorf("password is required")
	}

	return &Client{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		// conservative limit of 15 req/min = 1 req/4s
		limiter:  rate.NewLimiter(rate.Every(4*time.Second), 1),
		username: username,
		password: password,
		baseURL:  "https://metron.cloud/api",
	}, nil
}

// Validate checks if the provided credentials are valid by making a test request to the API. It returns an error if authentication fails or if there are any issues with the connection.
func (c *Client) Validate(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/publisher/?page_size=1", nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("User-Agent", "ComicBookCollector/1.0 (github.com/madhatter/comicbookcollector)")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized:
		return fmt.Errorf("authentication failed: invalid username or password")
	case http.StatusForbidden:
		return fmt.Errorf("access forbidden: insufficient permissions")
	case http.StatusTooManyRequests:
		return fmt.Errorf("rate limited: too many requests, try again later")
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (c *Client) get(ctx context.Context, endpoint string) (*http.Response, error) {
	// Wait for the rate limiter before making the request
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait: %w", err)
	}

	url := c.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("User-Agent", "ComicBookCollector/1.0 (github.com/madhatter/comicbookcollector)")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	return resp, nil
}

// FindIssueByUPC searches for an issue by its UPC code. It first performs a search to find the issue ID, then retrieves the full details of that issue.
func (c *Client) FindIssueByUPC(ctx context.Context, upc string) (*Issue, error) {
	resp, err := c.get(ctx, fmt.Sprintf("/issue/?upc=%s", upc))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var searchResp IssueSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}

	if searchResp.Count == 0 {
		return nil, fmt.Errorf("no issue found with UPC %s", upc)
	}

	// Fetch complete issue details using the ID from the search results
	return c.GetIssue(ctx, searchResp.Results[0].ID)
}

// GetIssue fetches the full details of an issue by its ID. This is used internally after finding the issue ID via a UPC search, but can also be used directly if you already have the issue ID.
func (c *Client) GetIssue(ctx context.Context, issueID int) (*Issue, error) {
	resp, err := c.get(ctx, fmt.Sprintf("/issue/%d/", issueID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var issue Issue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, err
	}

	return &issue, nil
}
