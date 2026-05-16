package greynoise

import "context"

// MockClient is a test double for Client. Set the Fn fields to control responses.
type MockClient struct {
	CommunityIPFn func(ctx context.Context, ip string) (*CommunityResponse, error)
	ContextIPFn   func(ctx context.Context, ip string) (*ContextResponse, error)
	RIOTFn        func(ctx context.Context, ip string) (*RIOTResponse, error)
	SimilarIPsFn  func(ctx context.Context, ip string) (*SimilarityResponse, error)
	GNQLFn        func(ctx context.Context, query string, size int) (*GNQLResponse, error)
}

func (m *MockClient) CommunityIP(ctx context.Context, ip string) (*CommunityResponse, error) {
	return m.CommunityIPFn(ctx, ip)
}

func (m *MockClient) ContextIP(ctx context.Context, ip string) (*ContextResponse, error) {
	return m.ContextIPFn(ctx, ip)
}

func (m *MockClient) RIOT(ctx context.Context, ip string) (*RIOTResponse, error) {
	return m.RIOTFn(ctx, ip)
}

func (m *MockClient) SimilarIPs(ctx context.Context, ip string) (*SimilarityResponse, error) {
	return m.SimilarIPsFn(ctx, ip)
}

func (m *MockClient) GNQL(ctx context.Context, query string, size int) (*GNQLResponse, error) {
	return m.GNQLFn(ctx, query, size)
}
