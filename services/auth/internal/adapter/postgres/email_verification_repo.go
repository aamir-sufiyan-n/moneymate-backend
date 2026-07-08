package postgres

import (
	"context"

	db "github.com/moneymate-2026/moneymate-backend/auth/sqlc/generated"
)

type EmailVerificationRepository struct {
	queries *db.Queries
}

func NewEmailVerificationRepository(q *db.Queries) *EmailVerificationRepository {
	return &EmailVerificationRepository{
		queries: q,
	}
}

func (r *EmailVerificationRepository) Create(
	ctx context.Context,
	arg db.CreateEmailVerificationParams,
) (db.AuthEmailVerification, error) {
	return r.queries.CreateEmailVerification(ctx, arg)
}

func (r *EmailVerificationRepository) GetByTokenHash(
	ctx context.Context,
	tokenHash string,
) (db.AuthEmailVerification, error) {
	return r.queries.GetEmailVerification(ctx, tokenHash)
}

func (r *EmailVerificationRepository) MarkUsed(
	ctx context.Context,
	tokenHash string,
) error {
	return r.queries.MarkEmailVerificationUsed(ctx, tokenHash)
}

func (r *EmailVerificationRepository) DeleteExpired(
	ctx context.Context,
) error {
	return r.queries.DeleteExpiredEmailVerifications(ctx)
}