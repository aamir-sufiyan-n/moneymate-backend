package proxy

import "context"

type AuthClient interface {
	VerifyToken(ctx context.Context, token string) (string, error)
}

type MockAuthClient struct{}

func NewMockAuthClient() *MockAuthClient {
	return &MockAuthClient{}
}

func (m *MockAuthClient) VerifyToken(ctx context.Context, token string) (string, error) {
	if token == "invalid" {
		return "", context.DeadlineExceeded 
	}
	return "mock-user-uuid-1234", nil
}