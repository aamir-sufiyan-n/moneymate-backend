package postgres

import (
	"context"

	db "github.com/moneymate-2026/moneymate-backend/auth/sqlc/generated"
	"github.com/google/uuid"
)

type OAuthRepository struct {
	queries *db.Queries
}

func NewOAuthRepository(q *db.Queries) *OAuthRepository {
	return &OAuthRepository{
		queries: q,
	}
}

func (r *OAuthRepository) Create(
	ctx context.Context,
	arg db.CreateOAuthAccountParams,
) (db.AuthOauthAccount, error) {
	return r.queries.CreateOAuthAccount(ctx, arg)
}

func (r *OAuthRepository) GetByProvider(
	ctx context.Context,
	arg db.GetOAuthAccountParams,
) (db.AuthOauthAccount, error) {
	return r.queries.GetOAuthAccount(ctx, arg)
}

func (r *OAuthRepository) GetByUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]db.AuthOauthAccount, error) {
	return r.queries.GetOAuthAccountsByUser(ctx, userID)
}