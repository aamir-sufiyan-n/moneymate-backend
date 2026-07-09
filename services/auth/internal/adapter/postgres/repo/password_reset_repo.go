package repo

import (
	"context"

	db "github.com/moneymate-2026/moneymate-backend/auth/sqlc/generated"
)

type PasswordResetRepository struct {
	queries *db.Queries
}

func NewPasswordResetRepository(q *db.Queries) *PasswordResetRepository {
	return &PasswordResetRepository{
		queries: q,
	}
}

func (r *PasswordResetRepository) Create(
	ctx context.Context,
	arg db.CreatePasswordResetTokenParams,
) (db.AuthPasswordResetToken, error) {
	return r.queries.CreatePasswordResetToken(ctx, arg)
}

func (r *PasswordResetRepository) GetByTokenHash(
	ctx context.Context,
	tokenHash string,
) (db.AuthPasswordResetToken, error) {
	return r.queries.GetPasswordResetToken(ctx, tokenHash)
}

func (r *PasswordResetRepository) MarkUsed(
	ctx context.Context,
	tokenHash string,
) error {
	return r.queries.MarkPasswordResetTokenUsed(ctx, tokenHash)
}

func (r *PasswordResetRepository) DeleteExpired(
	ctx context.Context,
) error {
	return r.queries.DeleteExpiredPasswordResetTokens(ctx)
}