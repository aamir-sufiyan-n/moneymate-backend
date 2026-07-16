package tokenissuer

import (
	"time"

	"github.com/google/uuid"
	jwtutil "github.com/moneymate-2026/moneymate-backend/shared/pkg/jwt"
)


type Issuer struct {
    cfg jwtutil.Config
}

func New(cfg jwtutil.Config) *Issuer {
    return &Issuer{cfg: cfg}
}

func (i *Issuer) IssueAccessToken(userID uuid.UUID, handle string, roles []string, tokenVersion int64) (string, time.Time, error) {
    expiresAt := time.Now().Add(time.Duration(i.cfg.AccessExpiryMins) * time.Minute)

    token, err := jwtutil.GenerateAccessToken(jwtutil.AccessTokenParams{
        UserID:       userID.String(),
        Handle:       handle,
        Roles:        roles,
        TokenVersion: tokenVersion,
    }, i.cfg)
    if err != nil {
        return "", time.Time{}, err
    }
    return token, expiresAt, nil
}

func (i *Issuer) IssueRefreshToken(userID uuid.UUID, deviceID string) (token, tokenHash string, expiresAt time.Time, err error) {
    expiresAt = time.Now().Add(time.Duration(i.cfg.RefreshExpiryHrs) * time.Hour)

    token, tokenHash, err = jwtutil.GenerateRefreshToken(jwtutil.RefreshTokenParams{
        UserID:   userID.String(),
        DeviceID: deviceID,
    }, i.cfg)
    if err != nil {
        return "", "", time.Time{}, err
    }
    return token, tokenHash, expiresAt, nil
}