package usecase

import (
    "context"
    "crypto/rand"
    "fmt"
    "math/big"
    "regexp"
    "strings"

    "github.com/moneymate-2026/moneymate-backend/auth/internal/domain"
    apperrors "github.com/moneymate-2026/moneymate-backend/shared/pkg/errors"
)

const (
    maxHandleGenAttempts = 8
    handleSuffixDigits   = 4 
)

var nonAlphaNumeric = regexp.MustCompile(`[^a-z0-9]`)

func generateHandle(ctx context.Context, repo domain.UserRepository, email, fullName string) (string, error) {
    base := baseFromEmail(email)
    if base == "" {
        base = baseFromName(fullName)
    }
    if base == "" {
        base = "user" 
    }

    for attempt := 0; attempt < maxHandleGenAttempts; attempt++ {
        suffix, err := randomDigits(handleSuffixDigits)
        if err != nil {
            return "", fmt.Errorf("generate handle suffix: %w", err)
        }

        candidate := base + suffix
        exists, err := repo.HandleExists(ctx, candidate)
        if err != nil {
            return "", fmt.Errorf("check handle exists: %w", err)
        }
        if !exists {
            return candidate, nil
        }
    }
    return "", fmt.Errorf("%w: could not generate unique handle after %d attempts", apperrors.ErrInternal, maxHandleGenAttempts)
}

func baseFromEmail(email string) string {
    at := strings.Index(email, "@")
    if at <= 0 {
        return ""
    }
    local := email[:at]
    return sanitizeAndTruncate(local)
}

func baseFromName(name string) string {
    return sanitizeAndTruncate(name)
}

func sanitizeAndTruncate(s string) string {
    s = strings.ToLower(strings.TrimSpace(s))
    s = nonAlphaNumeric.ReplaceAllString(s, "")
    const maxBaseLen = 15 
    if len(s) > maxBaseLen {
        s = s[:maxBaseLen]
    }
    return s
}

func randomDigits(n int) (string, error) {
    var sb strings.Builder
    for range n {
        d, err := rand.Int(rand.Reader, big.NewInt(10))
        if err != nil {
            return "", err
        }
        sb.WriteString(d.String())
    }
    return sb.String(), nil
}