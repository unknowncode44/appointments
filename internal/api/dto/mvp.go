package dto

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type TenantRequest struct {
	Name     string `json:"name" validate:"required"`
	Timezone string `json:"timezone" validate:"required"`
}

type TenantResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Timezone  string    `json:"timezone"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProviderRequest struct {
	TenantID uuid.UUID `json:"tenant_id" validate:"required"`
	Name     string    `json:"name" validate:"required"`
}

type ProviderUpdateRequest struct {
	Name string `json:"name" validate:"required"`
}

type ProviderResponse struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Name      string    `json:"name"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ServiceRequest struct {
	TenantID        uuid.UUID `json:"tenant_id" validate:"required"`
	Name            string    `json:"name" validate:"required"`
	DurationMinutes int32     `json:"duration_minutes" validate:"required,min=1"`
}

type ServiceUpdateRequest struct {
	Name            string `json:"name" validate:"required"`
	DurationMinutes int32  `json:"duration_minutes" validate:"required,min=1"`
}

type ServiceResponse struct {
	ID              uuid.UUID `json:"id"`
	TenantID        uuid.UUID `json:"tenant_id"`
	Name            string    `json:"name"`
	DurationMinutes int32     `json:"duration_minutes"`
	Active          bool      `json:"active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CustomerRequest struct {
	TenantID  uuid.UUID `json:"tenant_id" validate:"required"`
	FirstName *string   `json:"first_name"`
	LastName  *string   `json:"last_name"`
	Notes     *string   `json:"notes"`
}

type CustomerUpdateRequest struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Notes     *string `json:"notes"`
}

type CustomerResponse struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	FirstName *string   `json:"first_name,omitempty"`
	LastName  *string   `json:"last_name,omitempty"`
	Notes     *string   `json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TenantChannelRequest struct {
	TenantID    uuid.UUID `json:"tenant_id" validate:"required"`
	ChannelType string    `json:"channel_type" validate:"required"`
	ExternalID  string    `json:"external_id" validate:"required"`
	ExternalKey *string   `json:"external_key"`
}

type TenantChannelUpdateRequest struct {
	ChannelType string  `json:"channel_type" validate:"required"`
	ExternalID  string  `json:"external_id" validate:"required"`
	ExternalKey *string `json:"external_key"`
	Active      bool    `json:"active"`
}

type TenantChannelResponse struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	ChannelType string    `json:"channel_type"`
	ExternalID  string    `json:"external_id"`
	ExternalKey *string   `json:"external_key,omitempty"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
}

type AvailabilityRequest struct {
	Weekday   int16  `json:"weekday" validate:"min=0,max=6"`
	StartTime string `json:"start_time" validate:"required"`
	EndTime   string `json:"end_time" validate:"required"`
}

type AvailabilityResponse struct {
	ID         uuid.UUID `json:"id"`
	ProviderID uuid.UUID `json:"provider_id"`
	Weekday    int16     `json:"weekday"`
	StartTime  string    `json:"start_time"`
	EndTime    string    `json:"end_time"`
}

type ExceptionRequest struct {
	StartAt time.Time `json:"start_at" validate:"required"`
	EndAt   time.Time `json:"end_at" validate:"required"`
	Reason  *string   `json:"reason"`
}

type ExceptionResponse struct {
	ID         uuid.UUID `json:"id"`
	ProviderID uuid.UUID `json:"provider_id"`
	StartAt    time.Time `json:"start_at"`
	EndAt      time.Time `json:"end_at"`
	Reason     *string   `json:"reason,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type SlotGeneratorRequest struct {
	TenantID      uuid.UUID `json:"tenant_id" validate:"required"`
	ProviderID    uuid.UUID `json:"provider_id" validate:"required"`
	NumberOfWeeks int       `json:"number_of_weeks" validate:"required,min=1,max=26"`
	SlotMinutes   int       `json:"slot_minutes,omitempty"`
}

type SlotGeneratorResponse struct {
	Generated int `json:"generated"`
}

type SlotResponse struct {
	ID            uuid.UUID  `json:"id"`
	TenantID      uuid.UUID  `json:"tenant_id"`
	ProviderID    uuid.UUID  `json:"provider_id"`
	StartAt       time.Time  `json:"start_at"`
	EndAt         time.Time  `json:"end_at"`
	Status        string     `json:"status"`
	AppointmentID *uuid.UUID `json:"appointment_id,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type AppointmentCreateRequest struct {
	TenantID   uuid.UUID `json:"tenant_id" validate:"required"`
	CustomerID uuid.UUID `json:"customer_id" validate:"required"`
	ServiceID  uuid.UUID `json:"service_id" validate:"required"`
	SlotID     uuid.UUID `json:"slot_id" validate:"required"`
	Notes      *string   `json:"notes"`
}

type AppointmentUpdateRequest struct {
	Status    *string    `json:"status"`
	SlotID    *uuid.UUID `json:"slot_id"`
	ServiceID *uuid.UUID `json:"service_id"`
	Notes     *string    `json:"notes"`
}

type AppointmentResponse struct {
	ID         uuid.UUID  `json:"id"`
	TenantID   uuid.UUID  `json:"tenant_id"`
	CustomerID uuid.UUID  `json:"customer_id"`
	ProviderID uuid.UUID  `json:"provider_id"`
	ServiceID  uuid.UUID  `json:"service_id"`
	SlotID     *uuid.UUID `json:"slot_id,omitempty"`
	StartAt    time.Time  `json:"start_at"`
	EndAt      time.Time  `json:"end_at"`
	Status     string     `json:"status"`
	Notes      *string    `json:"notes,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type ConversationThreadResponse struct {
	ID         uuid.UUID                     `json:"id"`
	TenantID   uuid.UUID                     `json:"tenant_id"`
	CustomerID uuid.UUID                     `json:"customer_id"`
	CreatedAt  time.Time                     `json:"created_at"`
	Messages   []ConversationMessageResponse `json:"messages,omitempty"`
}

type ConversationMessageRequest struct {
	TenantID   uuid.UUID       `json:"tenant_id" validate:"required"`
	CustomerID uuid.UUID       `json:"customer_id" validate:"required"`
	Direction  string          `json:"direction" validate:"required,oneof=in out"`
	Message    string          `json:"message" validate:"required"`
	Metadata   json.RawMessage `json:"metadata"`
}

type ConversationMessageResponse struct {
	ID        uuid.UUID       `json:"id"`
	ThreadID  uuid.UUID       `json:"thread_id"`
	Direction string          `json:"direction"`
	Message   string          `json:"message"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

type InboundMessageRequest struct {
	TenantChannelExternalID string          `json:"tenant_channel_external_id" validate:"required"`
	ExternalCustomerID      string          `json:"external_customer_id" validate:"required"`
	Message                 string          `json:"message" validate:"required"`
	Source                  string          `json:"source"`
	Metadata                json.RawMessage `json:"metadata"`
}

type EvolutionWebhookRequest struct {
	Instance   string          `json:"instance"`
	ExternalID string          `json:"external_id"`
	From       string          `json:"from"`
	Phone      string          `json:"phone"`
	Message    string          `json:"message"`
	Event      string          `json:"event"`
	Data       json.RawMessage `json:"data"`
}

type UserCreateRequest struct {
	Username string     `json:"username" validate:"required,alphanum"`
	Password string     `json:"password" validate:"required,min=6"`
	FullName string     `json:"full_name" validate:"required"`
	Role     string     `json:"role" validate:"required,oneof=adminUser tenantUser user"`
	TenantID *uuid.UUID `json:"tenant_id"`
}

type UserUpdateRequest struct {
	FullName string     `json:"full_name"`
	Role     string     `json:"role" validate:"oneof=adminUser tenantUser user"`
	TenantID *uuid.UUID `json:"tenant_id"`
}

type UserResponse struct {
	ID        int32      `json:"id"`
	Username  string     `json:"username"`
	FullName  string     `json:"full_name"`
	Role      string     `json:"role"`
	TenantID  *uuid.UUID `json:"tenant_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type UserTenantRequest struct {
	TenantID uuid.UUID `json:"tenant_id" validate:"required"`
}

type UserProviderRequest struct {
	ProviderID uuid.UUID `json:"provider_id" validate:"required"`
}
