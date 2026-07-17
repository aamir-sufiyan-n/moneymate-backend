package proxy

import (
	"context"
	"fmt"
	"time"

	pb "github.com/moneymate-2026/moneymate-backend/shared/proto/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AuthClient defines the contract the gateway uses to talk to auth-svc.
// All middleware and handlers depend on this interface, never the concrete gRPC client.
type AuthClient interface {
	VerifyAccessToken(ctx context.Context, token string) (*TokenClaims, error)
	VerifyTransactionToken(ctx context.Context, token, transactionID string) (*TransactionClaims, error)
	GetUserByID(ctx context.Context, userID string) (*UserInfo, error)
	Close() error
}

// TokenClaims holds the decoded identity from a verified access token.
type TokenClaims struct {
	UserID     string
	Email      string
	Role       string // "user" or "merchant"
	MerchantID string // populated only when Role == "merchant"
	ExpiresAt  time.Time
}

// TransactionClaims holds the decoded identity from a verified transaction token.
type TransactionClaims struct {
	UserID        string
	TransactionID string
}

// UserInfo holds basic user data returned by GetUserByID.
type UserInfo struct {
	UserID   string
	Email    string
	FullName string
	Handle   string
	Role     string
}

// grpcAuthClient is the production implementation that talks to auth-svc over gRPC.
type grpcAuthClient struct {
	conn   *grpc.ClientConn
	client pb.AuthServiceClient
}

// NewAuthClient creates a real gRPC connection to auth-svc.
// addr example: "auth-svc:9091"
func NewAuthClient(addr string) (AuthClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth-svc at %s: %w", addr, err)
	}

	return &grpcAuthClient{
		conn:   conn,
		client: pb.NewAuthServiceClient(conn),
	}, nil
}

func (c *grpcAuthClient) VerifyAccessToken(ctx context.Context, token string) (*TokenClaims, error) {
	resp, err := c.client.VerifyAccessToken(ctx, &pb.VerifyAccessTokenRequest{
		Token: token,
	})
	if err != nil {
		return nil, fmt.Errorf("auth-svc VerifyAccessToken RPC failed: %w", err)
	}
	if !resp.Valid {
		return nil, fmt.Errorf("token is invalid or expired")
	}

	return &TokenClaims{
		UserID:     resp.UserId,
		Email:      resp.Email,
		Role:       resp.Role,
		MerchantID: resp.MerchantId,
		ExpiresAt:  time.Unix(resp.ExpiresAt, 0),
	}, nil
}

func (c *grpcAuthClient) VerifyTransactionToken(ctx context.Context, token, transactionID string) (*TransactionClaims, error) {
	resp, err := c.client.VerifyTransactionToken(ctx, &pb.VerifyTransactionTokenRequest{
		Token:         token,
		TransactionId: transactionID,
	})
	if err != nil {
		return nil, fmt.Errorf("auth-svc VerifyTransactionToken RPC failed: %w", err)
	}
	if !resp.Valid {
		return nil, fmt.Errorf("transaction token is invalid or expired")
	}

	return &TransactionClaims{
		UserID:        resp.UserId,
		TransactionID: resp.TransactionId,
	}, nil
}

func (c *grpcAuthClient) GetUserByID(ctx context.Context, userID string) (*UserInfo, error) {
	resp, err := c.client.GetUserByID(ctx, &pb.GetUserByIDRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("auth-svc GetUserByID RPC failed: %w", err)
	}

	return &UserInfo{
		UserID:   resp.UserId,
		Email:    resp.Email,
		FullName: resp.FullName,
		Handle:   resp.Handle,
		Role:     resp.Role,
	}, nil
}

func (c *grpcAuthClient) Close() error {
	return c.conn.Close()
}
