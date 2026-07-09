package repo

import (
	"context"

	db "github.com/moneymate-2026/moneymate-backend/auth/sqlc/generated"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository struct {
	queries *db.Queries
}

func NewUserRepository(q *db.Queries) *UserRepository {
	return &UserRepository{
		queries: q,
	}
}

func (r *UserRepository) Create(ctx context.Context, params db.CreateUserParams) (db.AuthUser, error) {
	return r.queries.CreateUser(ctx, params)
}

func (r *UserRepository) GetByID(ctx context.Context, id pgtype.UUID) (db.AuthUser, error) {
	return r.queries.GetUserByID(ctx, id)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (db.AuthUser, error) {
	return r.queries.GetUserByEmail(ctx, email)
}

func (r *UserRepository) VerifyEmail(ctx context.Context, id pgtype.UUID) error {
	return r.queries.VerifyEmail(ctx, id)
}

func (r *UserRepository) UpdatePassword(ctx context.Context, params db.UpdatePasswordParams) error {
	return r.queries.UpdatePassword(ctx, params)
}

func (r *UserRepository) UpdateStatus(ctx context.Context, params db.UpdateUserStatusParams) error {
	return r.queries.UpdateUserStatus(ctx, params)
}

func (r *UserRepository) Delete(ctx context.Context, id pgtype.UUID) error {
	return r.queries.DeleteUser(ctx, id)
}