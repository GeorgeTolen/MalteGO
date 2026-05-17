package greynoise

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	baseURL  = "https://api.greynoise.io"
	apiV2    = baseURL + "/v2"
	apiV3    = baseURL + "/v3"
	apiV2Exp = baseURL + "/v2/experimental"
)

// Client defines all GreyNoise API calls used by transforms.
type Client interface {
	CommunityIP(ctx context.Context, ip string) (*CommunityResponse, error)
	ContextIP(ctx context.Context, ip string) (*ContextResponse, error)
	RIOT(ctx context.Context, ip string) (*RIOTResponse, error)
	SimilarIPs(ctx context.Context, ip string) (*SimilarityResponse, error)
	GNQL(ctx context.Context, query string, size int) (*GNQLResponse, error)
}

type httpClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string, timeout time.Duration) Client {
	return &httpClient{
		apiKey: apiKey,
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

	if resp.StatusCode != http.StatusOK {
		if len(body) > 0 {
			return fmt.Errorf("api error %d: %s", resp.StatusCode, string(body))
		}
		return fmt.Errorf("api error %d: %s", resp.StatusCode, rawURL)
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("json decode: %w", err)
	}
	return nil
}

func (c *httpClient) CommunityIP(ctx context.Context, ip string) (*CommunityResponse, error) {
	var r CommunityResponse
	if err := c.get(ctx, fmt.Sprintf("%s/community/%s", apiV3, ip), &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *httpClient) ContextIP(ctx context.Context, ip string) (*ContextResponse, error) {
	var r ContextResponse
	if err := c.get(ctx, fmt.Sprintf("%s/noise/context/%s", apiV2, ip), &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *httpClient) RIOT(ctx context.Context, ip string) (*RIOTResponse, error) {
	var r RIOTResponse
	if err := c.get(ctx, fmt.Sprintf("%s/riot/%s", apiV2, ip), &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *httpClient) SimilarIPs(ctx context.Context, ip string) (*SimilarityResponse, error) {
	var r SimilarityResponse
	if err := c.get(ctx, fmt.Sprintf("%s/similarity/ips/%s", apiV3, ip), &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *httpClient) GNQL(ctx context.Context, query string, size int) (*GNQLResponse, error) {
	if size <= 0 {
		size = 50
	}
	u := fmt.Sprintf("%s/gnql?query=%s&size=%d", apiV2Exp, url.QueryEscape(query), size)
	var r GNQLResponse
	if err := c.get(ctx, u, &r); err != nil {
		return nil, err
	}
	return &r, nil
}
