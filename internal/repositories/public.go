package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/unknowncode44/appointments/internal/db/sqlc"
)

// PublicRepository exposes the read paths and store needed by the public,
// no-auth booking flow. Every call is scoped to a tenant resolved from the slug.
type PublicRepository interface {
	Store() *db.Store
	GetTenantBySlug(context.Context, pgtype.Text) (db.Tenant, error)
	GetService(context.Context, uuid.UUID) (db.Service, error)
	ListServices(context.Context, db.ListServicesParams) ([]db.Service, error)
	ListProviders(context.Context, db.ListProvidersParams) ([]db.Provider, error)
	ListAvailableSlots(context.Context, db.ListAvailableSlotsParams) ([]db.AppointmentSlot, error)
}

type publicRepository struct{ store *db.Store }

func NewPublicRepository(store *db.Store) PublicRepository {
	return &publicRepository{store: store}
}

func (r *publicRepository) Store() *db.Store { return r.store }

func (r *publicRepository) GetTenantBySlug(ctx context.Context, slug pgtype.Text) (db.Tenant, error) {
	return r.store.GetTenantBySlug(ctx, slug)
}
func (r *publicRepository) GetService(ctx context.Context, id uuid.UUID) (db.Service, error) {
	return r.store.GetService(ctx, id)
}
func (r *publicRepository) ListServices(ctx context.Context, arg db.ListServicesParams) ([]db.Service, error) {
	return r.store.ListServices(ctx, arg)
}
func (r *publicRepository) ListProviders(ctx context.Context, arg db.ListProvidersParams) ([]db.Provider, error) {
	return r.store.ListProviders(ctx, arg)
}
func (r *publicRepository) ListAvailableSlots(ctx context.Context, arg db.ListAvailableSlotsParams) ([]db.AppointmentSlot, error) {
	return r.store.ListAvailableSlots(ctx, arg)
}
