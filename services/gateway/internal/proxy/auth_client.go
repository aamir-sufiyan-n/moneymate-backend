package proxy

import (
	"context"
	"fmt"
	"time"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	
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

// httpAuthClient talks to auth-svc over HTTP (JSON).
type httpAuthClient struct {
    baseURL    string
    httpClient *http.Client
}

// NewAuthClient creates an HTTP client for auth-svc.
// addr example: "auth:8081"
func NewAuthClient(addr string) (AuthClient, error) {
    return &httpAuthClient{
        baseURL:    "http://" + addr,
        httpClient: &http.Client{Timeout: 5 * time.Second},
    }, nil
}

func (c *httpAuthClient) VerifyAccessToken(ctx context.Context, token string) (*TokenClaims, error) {
    body, _ := json.Marshal(map[string]string{"token": token})
    resp, err := c.do(ctx, "POST", "/internal/auth/verify-access-token", body)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusUnauthorized {
        return nil, fmt.Errorf("token is invalid or expired")
    }
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("auth-svc returned %d", resp.StatusCode)
    }

    var result struct {
        Data struct {
            Valid      bool   `json:"valid"`
            UserID     string `json:"user_id"`
            Email      string `json:"email"`
            Role       string `json:"role"`
            MerchantID string `json:"merchant_id"`
            ExpiresAt  int64  `json:"expires_at"`
        } `json:"data"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("decode auth-svc response: %w", err)
    }
    if !result.Data.Valid {
        return nil, fmt.Errorf("token is invalid")
    }

    return &TokenClaims{
        UserID:     result.Data.UserID,
        Email:      result.Data.Email,
        Role:       result.Data.Role,
        MerchantID: result.Data.MerchantID,
        ExpiresAt:  time.Unix(result.Data.ExpiresAt, 0),
    }, nil
}

func (c *httpAuthClient) VerifyTransactionToken(ctx context.Context, token, transactionID string) (*TransactionClaims, error) {
    body, _ := json.Marshal(map[string]string{
        "token":          token,
        "transaction_id": transactionID,
    })
    resp, err := c.do(ctx, "POST", "/internal/auth/verify-transaction-token", body)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusUnauthorized {
        return nil, fmt.Errorf("transaction token is invalid or expired")
    }
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("auth-svc returned %d", resp.StatusCode)
    }

    var result struct {
        Data struct {
            Valid         bool   `json:"valid"`
            UserID        string `json:"user_id"`
            TransactionID string `json:"transaction_id"`
        } `json:"data"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("decode auth-svc response: %w", err)
    }
    if !result.Data.Valid {
        return nil, fmt.Errorf("transaction token is invalid")
    }

    return &TransactionClaims{
        UserID:        result.Data.UserID,
        TransactionID: result.Data.TransactionID,
    }, nil
}

func (c *httpAuthClient) GetUserByID(ctx context.Context, userID string) (*UserInfo, error) {
    resp, err := c.do(ctx, "GET", "/internal/auth/users/"+userID, nil)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusNotFound {
        return nil, fmt.Errorf("user not found")
    }
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("auth-svc returned %d", resp.StatusCode)
    }

    var result struct {
        Data struct {
            UserID   string `json:"user_id"`
            Email    string `json:"email"`
            FullName string `json:"full_name"`
            Handle   string `json:"handle"`
            Role     string `json:"role"`
        } `json:"data"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("decode auth-svc response: %w", err)
    }

    return &UserInfo{
        UserID:   result.Data.UserID,
        Email:    result.Data.Email,
        FullName: result.Data.FullName,
        Handle:   result.Data.Handle,
        Role:     result.Data.Role,
    }, nil
}

func (c *httpAuthClient) Close() error {
    c.httpClient.CloseIdleConnections()
    return nil
}

// do is a helper that makes an HTTP request with context.
func (c *httpAuthClient) do(ctx context.Context, method, path string, body []byte) (*http.Response, error) {
    var bodyReader io.Reader
    if body != nil {
        bodyReader = bytes.NewReader(body)
    }

    req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("auth-svc request failed: %w", err)
    }
    return resp, nil
}

// // grpcAuthClient is the production implementation that talks to auth-svc over gRPC.
// type grpcAuthClient struct {
// 	conn   *grpc.ClientConn
// 	client pb.AuthServiceClient
// }

// // NewAuthClient creates a real gRPC connection to auth-svc.
// // addr example: "auth-svc:9091"
// func NewAuthClient(addr string) (AuthClient, error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	conn, err := grpc.DialContext(ctx, addr,
// 		grpc.WithTransportCredentials(insecure.NewCredentials()),
// 		grpc.WithBlock(),
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to connect to auth-svc at %s: %w", addr, err)
// 	}

// 	return &grpcAuthClient{
// 		conn:   conn,
// 		client: pb.NewAuthServiceClient(conn),
// 	}, nil
// }

// func (c *grpcAuthClient) VerifyAccessToken(ctx context.Context, token string) (*TokenClaims, error) {
// 	resp, err := c.client.VerifyAccessToken(ctx, &pb.VerifyAccessTokenRequest{
// 		Token: token,
// 	})
// 	if err != nil {
// 		return nil, fmt.Errorf("auth-svc VerifyAccessToken RPC failed: %w", err)
// 	}
// 	if !resp.Valid {
// 		return nil, fmt.Errorf("token is invalid or expired")
// 	}

// 	return &TokenClaims{
// 		UserID:     resp.UserId,
// 		Email:      resp.Email,
// 		Role:       resp.Role,
// 		MerchantID: resp.MerchantId,
// 		ExpiresAt:  time.Unix(resp.ExpiresAt, 0),
// 	}, nil
// }

// func (c *grpcAuthClient) VerifyTransactionToken(ctx context.Context, token, transactionID string) (*TransactionClaims, error) {
// 	resp, err := c.client.VerifyTransactionToken(ctx, &pb.VerifyTransactionTokenRequest{
// 		Token:         token,
// 		TransactionId: transactionID,
// 	})
// 	if err != nil {
// 		return nil, fmt.Errorf("auth-svc VerifyTransactionToken RPC failed: %w", err)
// 	}
// 	if !resp.Valid {
// 		return nil, fmt.Errorf("transaction token is invalid or expired")
// 	}

// 	return &TransactionClaims{
// 		UserID:        resp.UserId,
// 		TransactionID: resp.TransactionId,
// 	}, nil
// }

// func (c *grpcAuthClient) GetUserByID(ctx context.Context, userID string) (*UserInfo, error) {
// 	resp, err := c.client.GetUserByID(ctx, &pb.GetUserByIDRequest{
// 		UserId: userID,
// 	})
// 	if err != nil {
// 		return nil, fmt.Errorf("auth-svc GetUserByID RPC failed: %w", err)
// 	}

// 	return &UserInfo{
// 		UserID:   resp.UserId,
// 		Email:    resp.Email,
// 		FullName: resp.FullName,
// 		Handle:   resp.Handle,
// 		Role:     resp.Role,
// 	}, nil
// }

// func (c *grpcAuthClient) Close() error {
// 	return c.conn.Close()
// }
