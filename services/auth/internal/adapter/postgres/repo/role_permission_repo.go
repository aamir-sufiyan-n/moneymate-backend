package repo

import (
	"context"

	db "github.com/moneymate-2026/moneymate-backend/auth/sqlc/generated"
	"github.com/jackc/pgx/v5/pgtype"
)

type RolePermissionRepository struct {
	queries *db.Queries
}

func NewRolePermissionRepository(q *db.Queries) *RolePermissionRepository {
	return &RolePermissionRepository{
		queries: q,
	}
}

func (r *RolePermissionRepository) AssignPermission(
	ctx context.Context,
	arg db.AssignPermissionToRoleParams,
) error {
	return r.queries.AssignPermissionToRole(ctx, arg)
}

func (r *RolePermissionRepository) RemovePermission(
	ctx context.Context,
	arg db.RemovePermissionFromRoleParams,
) error {
	return r.queries.RemovePermissionFromRole(ctx, arg)
}

func (r *RolePermissionRepository) GetRolePermissions(
	ctx context.Context,
	roleID pgtype.UUID,
) ([]db.AuthPermission, error) {
	return r.queries.GetRolePermissions(ctx, roleID)
}