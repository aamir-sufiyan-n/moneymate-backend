package postgres

import (
	"context"

	db "github.com/moneymate-2026/moneymate-backend/auth/sqlc/generated"
	"github.com/google/uuid"
)

type RoleRepository struct {
	queries *db.Queries
}

func NewRoleRepository(q *db.Queries) *RoleRepository {
	return &RoleRepository{
		queries: q,
	}
}

func (r *RoleRepository) Create(ctx context.Context, arg db.CreateRoleParams) (db.AuthRole, error) {
	return r.queries.CreateRole(ctx, arg)
}

func (r *RoleRepository) GetByID(ctx context.Context, id uuid.UUID) (db.AuthRole, error) {
	return r.queries.GetRoleByID(ctx, id)
}

func (r *RoleRepository) GetByName(ctx context.Context, name string) (db.AuthRole, error) {
	return r.queries.GetRoleByName(ctx, name)
}

func (r *RoleRepository) List(ctx context.Context) ([]db.AuthRole, error) {
	return r.queries.ListRoles(ctx)
}

func (r *RoleRepository) Update(ctx context.Context, arg db.UpdateRoleParams) error {
	return r.queries.UpdateRole(ctx, arg)
}

func (r *RoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteRole(ctx, id)
}