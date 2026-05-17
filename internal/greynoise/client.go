package greynoise

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	defaultBaseURL = "https://api.greynoise.io"
)

// Client defines all GreyNoise API calls used by transforms.
type Client interface {
	CommunityIP(ctx context.Context, ip string) (*CommunityResponse, error)
	ContextIP(ctx context.Context, ip string) (*ContextResponse, error)
	RIOT(ctx context.Context, ip string) (*RIOTResponse, error)
	SimilarIPs(ctx context.Context, ip string, minScore, limit int) (*SimilarityResponse, error)
	GNQL(ctx context.Context, query string, size int) (*GNQLResponse, error)
}

type httpClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewClient(apiKey string, timeout time.Duration) Client {
	baseURL := os.Getenv("GREYNOISE_BASE_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &httpClient{
		apiKey: apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *httpClient) get(ctx context.Context, rawURL string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("key", c.apiKey)
	req.Header.Set("User-Agent", "MalteGO/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("not found: %s", rawURL)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("api error %d: %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("json decode: %w", err)
	}
	return nil
}

func (c *httpClient) CommunityIP(ctx context.Context, ip string) (*CommunityResponse, error) {
	var r CommunityResponse
	if err := c.get(ctx, fmt.Sprintf("%s/v3/community/%s", c.baseURL, ip), &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *httpClient) ContextIP(ctx context.Context, ip string) (*ContextResponse, error) {
	var r ContextResponse
	if err := c.get(ctx, fmt.Sprintf("%s/v2/noise/context/%s", c.baseURL, ip), &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *httpClient) RIOT(ctx context.Context, ip string) (*RIOTResponse, error) {
	var r RIOTResponse
	if err := c.get(ctx, fmt.Sprintf("%s/v2/riot/%s", c.baseURL, ip), &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *httpClient) SimilarIPs(ctx context.Context, ip string, minScore, limit int) (*SimilarityResponse, error) {
	if minScore <= 0 {
		minScore = 90
	}
	if limit <= 0 {
		limit = 50
	}
	var r SimilarityResponse
	u := fmt.Sprintf("%s/v1/experimental/gnoise/similar/%s?min_score=%d&limit=%d", c.baseURL, ip, minScore, limit)
	if err := c.get(ctx, u, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *httpClient) GNQL(ctx context.Context, query string, size int) (*GNQLResponse, error) {
	if size <= 0 {
		size = 50
	}
	u := fmt.Sprintf("%s/v2/experimental/gnql?query=%s&size=%d", c.baseURL, url.QueryEscape(query), size)
	var r GNQLResponse
	if err := c.get(ctx, u, &r); err != nil {
		return nil, err
	}
	return &r, nil
}
