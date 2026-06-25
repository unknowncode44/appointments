package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/unknowncode44/appointments/internal/api/dto"
	"github.com/unknowncode44/appointments/internal/api/response"
	db "github.com/unknowncode44/appointments/internal/db/sqlc"
	"github.com/unknowncode44/appointments/internal/repositories"
)

// publicListLimit caps how many services/providers the public endpoints return.
const publicListLimit = 200

type PublicService interface {
	ResolveTenant(context.Context, string) (dto.PublicTenant, error)
	ListServices(context.Context, uuid.UUID) ([]dto.PublicServiceResponse, error)
	ListProviders(context.Context, uuid.UUID) ([]dto.PublicProviderResponse, error)
	ListAvailability(context.Context, uuid.UUID, *uuid.UUID, string) ([]dto.PublicSlotResponse, error)
	CreateBooking(context.Context, uuid.UUID, dto.PublicBookingRequest) (dto.PublicAppointmentResponse, error)
}

type publicService struct {
	repo repositories.PublicRepository
}

func NewPublicService(repo repositories.PublicRepository) PublicService {
	return &publicService{repo: repo}
}

// ResolveTenant looks up an active tenant by slug. A missing or inactive slug
// surfaces as pgx.ErrNoRows, which the handler maps to 404.
func (s *publicService) ResolveTenant(ctx context.Context, slug string) (dto.PublicTenant, error) {
	t, err := s.repo.GetTenantBySlug(ctx, repositories.Text(&slug))
	if err != nil {
		return dto.PublicTenant{}, err
	}
	return dto.PublicTenant{ID: t.ID, Name: t.Name, Timezone: t.Timezone, Slug: t.Slug.String}, nil
}

func (s *publicService) ListServices(ctx context.Context, tenantID uuid.UUID) ([]dto.PublicServiceResponse, error) {
	active := true
	items, err := s.repo.ListServices(ctx, db.ListServicesParams{TenantID: tenantID, Active: &active, Limit: publicListLimit})
	return mapSlice(items, mapPublicService), err
}

func (s *publicService) ListProviders(ctx context.Context, tenantID uuid.UUID) ([]dto.PublicProviderResponse, error) {
	active := true
	items, err := s.repo.ListProviders(ctx, db.ListProvidersParams{TenantID: tenantID, Active: &active, Limit: publicListLimit})
	return mapSlice(items, mapPublicProvider), err
}

func (s *publicService) ListAvailability(ctx context.Context, tenantID uuid.UUID, providerID *uuid.UUID, date string) ([]dto.PublicSlotResponse, error) {
	start, end, err := sameDayBounds(date)
	if err != nil {
		return nil, response.ErrInvalidInput
	}
	// Public availability is provider-time and never filtered by service_id,
	// consistent with the authenticated availability endpoint.
	items, err := s.repo.ListAvailableSlots(ctx, db.ListAvailableSlotsParams{
		TenantID:   tenantID,
		ProviderID: providerID,
		StartAt:    start,
		EndAt:      end,
		Limit:      publicListLimit,
	})
	return mapSlice(items, mapPublicSlot), err
}

// CreateBooking performs the full public booking inside a single transaction:
// upsert the customer by phone, then run the shared booking core. The tenant
// always comes from the resolved slug, never from the request body.
func (s *publicService) CreateBooking(ctx context.Context, tenantID uuid.UUID, req dto.PublicBookingRequest) (dto.PublicAppointmentResponse, error) {
	// Reject a service that does not belong to the resolved tenant (or is
	// inactive) before touching the slot — never write another tenant's data.
	svc, err := s.repo.GetService(ctx, req.ServiceID)
	if err != nil {
		return dto.PublicAppointmentResponse{}, err
	}
	if svc.TenantID != tenantID || !svc.Active {
		return dto.PublicAppointmentResponse{}, response.ErrNotFound
	}

	var created db.Appointment
	err = s.repo.Store().ExecTx(ctx, func(q *db.Queries) error {
		customer, err := q.UpsertCustomerByPhone(ctx, db.UpsertCustomerByPhoneParams{
			TenantID:  tenantID,
			FirstName: repositories.Text(&req.FirstName),
			LastName:  repositories.Text(req.LastName),
			Phone:     repositories.Text(&req.Phone),
			Email:     repositories.Text(req.Email),
		})
		if err != nil {
			return err
		}
		appt, err := bookSlot(ctx, q, bookSlotParams{
			TenantID:   tenantID,
			CustomerID: customer.ID,
			ServiceID:  req.ServiceID,
			SlotID:     req.SlotID,
			Notes:      repositories.Text(req.Notes),
		})
		if err != nil {
			return err
		}
		created = appt
		return nil
	})
	return mapPublicAppointment(created), err
}

func mapPublicService(v db.Service) dto.PublicServiceResponse {
	return dto.PublicServiceResponse{ID: v.ID, Name: v.Name, DurationMinutes: v.DurationMinutes}
}

func mapPublicProvider(v db.Provider) dto.PublicProviderResponse {
	return dto.PublicProviderResponse{ID: v.ID, Name: v.Name}
}

func mapPublicSlot(v db.AppointmentSlot) dto.PublicSlotResponse {
	return dto.PublicSlotResponse{ID: v.ID, ProviderID: v.ProviderID, StartAt: v.StartAt, EndAt: v.EndAt}
}

func mapPublicAppointment(v db.Appointment) dto.PublicAppointmentResponse {
	return dto.PublicAppointmentResponse{
		ID:         v.ID,
		Status:     v.Status,
		ProviderID: v.ProviderID,
		ServiceID:  v.ServiceID,
		StartAt:    v.StartAt,
		EndAt:      v.EndAt,
	}
}
