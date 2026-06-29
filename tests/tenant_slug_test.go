package test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/unknowncode44/appointments/internal/api/dto"
	"github.com/unknowncode44/appointments/internal/api/response"
	db "github.com/unknowncode44/appointments/internal/db/sqlc"
	"github.com/unknowncode44/appointments/internal/services"
)

// stubSlugConflictRepo embeds stubAdminRepo and overrides the tenant
// create/update methods to return a Postgres unique-violation, simulating a
// duplicate slug hitting the tenants_slug_idx unique index.
type stubSlugConflictRepo struct {
	*stubAdminRepo
}

func (r *stubSlugConflictRepo) CreateTenant(_ context.Context, _ db.CreateTenantParams) (db.Tenant, error) {
	return db.Tenant{}, db.ErrUniqueViolation
}
func (r *stubSlugConflictRepo) UpdateTenant(_ context.Context, _ db.UpdateTenantParams) (db.Tenant, error) {
	return db.Tenant{}, db.ErrUniqueViolation
}

func TestTenantSlugConflict(t *testing.T) {
	svc := services.NewAdminService(&stubSlugConflictRepo{&stubAdminRepo{}})
	slug := "barberia-juan"
	req := dto.TenantRequest{Name: "Barbería Juan", Timezone: "America/Argentina/Buenos_Aires", Slug: &slug}

	t.Run("CreateTenant maps unique violation to ErrSlugTaken", func(t *testing.T) {
		_, err := svc.CreateTenant(context.Background(), req)
		require.ErrorIs(t, err, response.ErrSlugTaken)
	})

	t.Run("UpdateTenant maps unique violation to ErrSlugTaken", func(t *testing.T) {
		_, err := svc.UpdateTenant(context.Background(), uuid.New(), req)
		require.ErrorIs(t, err, response.ErrSlugTaken)
	})
}
