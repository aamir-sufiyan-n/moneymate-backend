package repo

import (
	"context"

	db "github.com/moneymate-2026/moneymate-backend/auth/sqlc/generated"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserRoleRepository struct {
	queries *db.Queries
}

func NewUserRoleRepository(q *db.Queries) *UserRoleRepository {
	return &UserRoleRepository{
		queries: q,
	}
}

func (r *UserRoleRepository) AssignRole(
	ctx context.Context,
	arg db.AssignRoleToUserParams,
) error {
	return r.queries.AssignRoleToUser(ctx, arg)
}

func (r *UserRoleRepository) RemoveRole(
	ctx context.Context,
	arg db.RemoveRoleFromUserParams,
) error {
	return r.queries.RemoveRoleFromUser(ctx, arg)
}

func (r *UserRoleRepository) GetUserRoles(
	ctx context.Context,
	userID pgtype.UUID,
) ([]db.AuthRole, error) {
	return r.queries.GetUserRoles(ctx, userID)
}
