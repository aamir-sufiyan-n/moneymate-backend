package postgres

import (
	"context"

	db "github.com/moneymate-2026/moneymate-backend/auth/sqlc/generated"
)

type RefreshTokenRepository struct {
	queries *db.Queries
}

func NewRefreshTokenRepository(q *db.Queries) *RefreshTokenRepository {
	return &RefreshTokenRepository{
		queries: q,
	}
}

func (r *RefreshTokenRepository) Create(
	ctx context.Context,
	arg db.CreateRefreshTokenParams,
) (db.AuthRefreshToken, error) {
	return r.queries.CreateRefreshToken(ctx, arg)
}

func (r *RefreshTokenRepository) GetByTokenHash(
	ctx context.Context,
	tokenHash string,
) (db.AuthRefreshToken, error) {
	return r.queries.GetRefreshToken(ctx, tokenHash)
}

func (r *RefreshTokenRepository) Revoke(
	ctx context.Context,
	tokenHash string,
) error {
	return r.queries.RevokeRefreshToken(ctx, tokenHash)
}

func (r *RefreshTokenRepository) DeleteExpired(
	ctx context.Context,
) error {
	return r.queries.DeleteExpiredRefreshTokens(ctx)
}