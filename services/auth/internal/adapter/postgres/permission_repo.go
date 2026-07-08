package postgres

import (
	"context"

	"github.com/google/uuid"
	db "github.com/moneymate-2026/moneymate-backend/auth/sqlc/generated"
)

type PermissionRepository struct {
	queries *db.Queries
}

func NewPermissionRepository(q *db.Queries) *PermissionRepository {
	return &PermissionRepository{
		queries: q,
	}
}

func (r *PermissionRepository) Create(ctx context.Context, arg db.CreatePermissionParams)(db.AuthPermission,error){
	return r.queries.CreatePermission(ctx,arg)
}

func (r *PermissionRepository) GetByID(ctx context.Context, id uuid.UUID)(db.AuthPermission, error){
	return r.queries.GetPermissionByID(ctx,id)
}

func (r *PermissionRepository) GetByName(ctx context.Context, name string) (db.AuthPermission, error) {
	return r.queries.GetPermissionByName(ctx, name)
}

func (r *PermissionRepository) List(ctx context.Context) ([]db.AuthPermission, error) {
	return r.queries.ListPermissions(ctx)
}

func (r *PermissionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeletePermission(ctx, id)
}