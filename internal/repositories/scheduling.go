package repositories

import (
	"context"

	"github.com/google/uuid"
	db "github.com/unknowncode44/appointments/internal/db/sqlc"
)

type SchedulingRepository interface {
	GetProvider(context.Context, uuid.UUID) (db.Provider, error)
	CreateAvailability(context.Context, db.CreateProviderAvailabilityParams) (db.ProviderAvailability, error)
	ListAvailability(context.Context, uuid.UUID) ([]db.ProviderAvailability, error)
	CreateException(context.Context, db.CreateProviderExceptionParams) (db.ProviderException, error)
	ListExceptions(context.Context, uuid.UUID) ([]db.ProviderException, error)
	ListExceptionsBetween(context.Context, db.ListProviderExceptionsBetweenParams) ([]db.ProviderException, error)
	CreateSlot(context.Context, db.CreateAppointmentSlotParams) (db.AppointmentSlot, error)
	ListAvailableSlots(context.Context, db.ListAvailableSlotsParams) ([]db.AppointmentSlot, error)
}

type schedulingRepository struct{ store *db.Store }

func NewSchedulingRepository(store *db.Store) SchedulingRepository {
	return &schedulingRepository{store: store}
}

func (r *schedulingRepository) GetProvider(ctx context.Context, id uuid.UUID) (db.Provider, error) {
	return r.store.GetProvider(ctx, id)
}
func (r *schedulingRepository) CreateAvailability(ctx context.Context, arg db.CreateProviderAvailabilityParams) (db.ProviderAvailability, error) {
	return r.store.CreateProviderAvailability(ctx, arg)
}
func (r *schedulingRepository) ListAvailability(ctx context.Context, providerID uuid.UUID) ([]db.ProviderAvailability, error) {
	return r.store.ListProviderAvailability(ctx, providerID)
}
func (r *schedulingRepository) CreateException(ctx context.Context, arg db.CreateProviderExceptionParams) (db.ProviderException, error) {
	return r.store.CreateProviderException(ctx, arg)
}
func (r *schedulingRepository) ListExceptions(ctx context.Context, providerID uuid.UUID) ([]db.ProviderException, error) {
	return r.store.ListProviderExceptions(ctx, providerID)
}
func (r *schedulingRepository) ListExceptionsBetween(ctx context.Context, arg db.ListProviderExceptionsBetweenParams) ([]db.ProviderException, error) {
	return r.store.ListProviderExceptionsBetween(ctx, arg)
}
func (r *schedulingRepository) CreateSlot(ctx context.Context, arg db.CreateAppointmentSlotParams) (db.AppointmentSlot, error) {
	return r.store.CreateAppointmentSlot(ctx, arg)
}
func (r *schedulingRepository) ListAvailableSlots(ctx context.Context, arg db.ListAvailableSlotsParams) ([]db.AppointmentSlot, error) {
	return r.store.ListAvailableSlots(ctx, arg)
}
