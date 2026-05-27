package greynoise

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// serviceClient implements Client by calling our greynoise-api microservice
// instead of the GreyNoise API directly.
type serviceClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewServiceClient(serviceURL string, timeout time.Duration) Client {
	return &serviceClient{
		baseURL:    serviceURL,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *serviceClient) get(ctx context.Context, path string, out interface{}) error {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	return doGet(ctx, c.httpClient, req, out)
}

func (c *serviceClient) CommunityIP(ctx context.Context, ip string) (*CommunityResponse, error) {
	var r CommunityResponse
	if err := c.get(ctx, "/community/"+ip, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *serviceClient) ContextIP(ctx context.Context, ip string) (*ContextResponse, error) {
	var r ContextResponse
	if err := c.get(ctx, "/context/"+ip, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *serviceClient) RIOT(ctx context.Context, ip string) (*RIOTResponse, error) {
	var r RIOTResponse
	if err := c.get(ctx, "/riot/"+ip, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *serviceClient) SimilarIPs(ctx context.Context, ip string, minScore, limit int) (*SimilarityResponse, error) {
	if minScore <= 0 {
		minScore = 90
	}
	if limit <= 0 {
		limit = 50
	}
	path := fmt.Sprintf("/similar/%s?min_score=%d&limit=%d", ip, minScore, limit)
	var r SimilarityResponse
	if err := c.get(ctx, path, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *serviceClient) GNQL(ctx context.Context, query string, size int) (*GNQLResponse, error) {
	if size <= 0 {
		size = 50
	}
	path := fmt.Sprintf("/gnql?query=%s&size=%d", url.QueryEscape(query), size)
	var r GNQLResponse
	if err := c.get(ctx, path, &r); err != nil {
		return nil, err
	}
	return &r, nil
}
