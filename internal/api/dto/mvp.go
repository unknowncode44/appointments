package dto

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type TenantRequest struct {
	Name     string  `json:"name" validate:"required"`
	Timezone string  `json:"timezone" validate:"required"`
	Slug     *string `json:"slug" validate:"omitempty,slug"`
}

type TenantResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Timezone  string    `json:"timezone"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Slug      *string   `json:"slug,omitempty"`
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
	Phone     *string   `json:"phone,omitempty"`
	Email     *string   `json:"email,omitempty"`
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

// SlotGeneratorRequest caps number_of_weeks at 12 to prevent runaway slot generation.
// Timezone must be a valid IANA timezone string (e.g. "America/Argentina/Buenos_Aires").
// If omitted, the server falls back to UTC.
type SlotGeneratorRequest struct {
	TenantID      uuid.UUID `json:"tenant_id" validate:"required"`
	ProviderID    uuid.UUID `json:"provider_id" validate:"required"`
	NumberOfWeeks int       `json:"number_of_weeks" validate:"required,min=1,max=12"`
	SlotMinutes   int       `json:"slot_minutes,omitempty"`
	Timezone      string    `json:"timezone,omitempty"`
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
	Status    *string    `json:"status" validate:"omitempty,oneof=confirmed cancelled completed no_show"`
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

// EvolutionWebhookRequest models the Evolution API v2 payload.
// Top-level fields (instance, external_id, from, phone, message) are kept
// for backward-compat. The nested Data field handles real Evolution v2 messages.
type EvolutionWebhookRequest struct {
	// Legacy / simple fields
	Instance   string `json:"instance"`
	ExternalID string `json:"external_id"`
	From       string `json:"from"`
	Phone      string `json:"phone"`
	Message    string `json:"message"`
	Event      string `json:"event"`

	// Evolution API v2 nested payload
	Data EvolutionData `json:"data"`
}

// EvolutionWebhookResult is the small observability/test-friendly response the
// webhook returns instead of a bare 202. processed is true when the message was
// newly handled (FSM advanced); idempotent is true when it was a duplicate.
type EvolutionWebhookResult struct {
	Processed   bool       `json:"processed"`
	Idempotent  bool       `json:"idempotent"`
	TenantID    *uuid.UUID `json:"tenant_id,omitempty"`
	CustomerID  *uuid.UUID `json:"customer_id,omitempty"`
	CurrentStep string     `json:"current_step,omitempty"`
}

type EvolutionData struct {
	Key         EvolutionKey         `json:"key"`
	Message     EvolutionMessageBody `json:"message"`
	PushName    string               `json:"pushName"`
	MessageType string               `json:"messageType"`
}

type EvolutionKey struct {
	RemoteJid string `json:"remoteJid"`
	FromMe    bool   `json:"fromMe"`
	ID        string `json:"id"`
}

type EvolutionMessageBody struct {
	Conversation        string                 `json:"conversation"`
	ExtendedTextMessage *EvolutionExtendedText `json:"extendedTextMessage,omitempty"`
}

type EvolutionExtendedText struct {
	Text string `json:"text"`
}

// UserCreateRequest — admin creates a user with an explicit role.
type UserCreateRequest struct {
	Username string     `json:"username" validate:"required,min=3"`
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
