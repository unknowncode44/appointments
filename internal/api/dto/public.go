package dto

import (
	"time"

	"github.com/google/uuid"
)

// PublicTenant is the resolved tenant context for a /public/:slug request.
// It is internal to the request lifecycle (stored in the Fiber context) and is
// not serialized to clients directly.
type PublicTenant struct {
	ID       uuid.UUID
	Name     string
	Timezone string
	Slug     string
}

// PublicShopResponse is the public-safe shop header for the booking page.
type PublicShopResponse struct {
	Name     string `json:"name"`
	Timezone string `json:"timezone"`
	Slug     string `json:"slug"`
}

// PublicServiceResponse exposes only booking-relevant service fields.
type PublicServiceResponse struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	DurationMinutes int32     `json:"duration_minutes"`
}

// PublicProviderResponse exposes only booking-relevant provider fields.
type PublicProviderResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// PublicSlotResponse exposes an available time slot without leaking internal
// fields (tenant_id, status, appointment_id).
type PublicSlotResponse struct {
	ID         uuid.UUID `json:"id"`
	ProviderID uuid.UUID `json:"provider_id"`
	StartAt    time.Time `json:"start_at"`
	EndAt      time.Time `json:"end_at"`
}

// PublicBookingRequest is the self-service booking payload. The tenant comes
// from the slug, never from the body.
type PublicBookingRequest struct {
	FirstName string    `json:"first_name" validate:"required"`
	LastName  *string   `json:"last_name"`
	Phone     string    `json:"phone" validate:"required"`
	Email     *string   `json:"email" validate:"omitempty,email"`
	ServiceID uuid.UUID `json:"service_id" validate:"required"`
	SlotID    uuid.UUID `json:"slot_id" validate:"required"`
	Notes     *string   `json:"notes"`
}

// PublicAppointmentResponse is the public-safe confirmation of a booking.
type PublicAppointmentResponse struct {
	ID         uuid.UUID `json:"id"`
	Status     string    `json:"status"`
	ProviderID uuid.UUID `json:"provider_id"`
	ServiceID  uuid.UUID `json:"service_id"`
	StartAt    time.Time `json:"start_at"`
	EndAt      time.Time `json:"end_at"`
}
