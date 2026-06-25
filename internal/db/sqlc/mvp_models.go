package db

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Tenant struct {
	ID        uuid.UUID   `json:"id"`
	Name      string      `json:"name"`
	Timezone  string      `json:"timezone"`
	Active    bool        `json:"active"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	Slug      pgtype.Text `json:"slug"`
}

type Provider struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Name      string    `json:"name"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Service struct {
	ID              uuid.UUID `json:"id"`
	TenantID        uuid.UUID `json:"tenant_id"`
	Name            string    `json:"name"`
	DurationMinutes int32     `json:"duration_minutes"`
	Active          bool      `json:"active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Customer struct {
	ID        uuid.UUID   `json:"id"`
	TenantID  uuid.UUID   `json:"tenant_id"`
	FirstName pgtype.Text `json:"first_name"`
	LastName  pgtype.Text `json:"last_name"`
	Notes     pgtype.Text `json:"notes"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	Phone     pgtype.Text `json:"phone"`
	Email     pgtype.Text `json:"email"`
}

type CustomerChannel struct {
	ID                 uuid.UUID `json:"id"`
	CustomerID         uuid.UUID `json:"customer_id"`
	TenantChannelID    uuid.UUID `json:"tenant_channel_id"`
	ExternalIdentifier string    `json:"external_identifier"`
	CreatedAt          time.Time `json:"created_at"`
}

type TenantChannel struct {
	ID          uuid.UUID   `json:"id"`
	TenantID    uuid.UUID   `json:"tenant_id"`
	ChannelType string      `json:"channel_type"`
	ExternalID  string      `json:"external_id"`
	ExternalKey pgtype.Text `json:"external_key"`
	Active      bool        `json:"active"`
	CreatedAt   time.Time   `json:"created_at"`
}

type ProviderAvailability struct {
	ID         uuid.UUID `json:"id"`
	ProviderID uuid.UUID `json:"provider_id"`
	Weekday    int16     `json:"weekday"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
}

type ProviderException struct {
	ID         uuid.UUID   `json:"id"`
	ProviderID uuid.UUID   `json:"provider_id"`
	StartAt    time.Time   `json:"start_at"`
	EndAt      time.Time   `json:"end_at"`
	Reason     pgtype.Text `json:"reason"`
	CreatedAt  time.Time   `json:"created_at"`
}

type AppointmentSlot struct {
	ID            uuid.UUID     `json:"id"`
	TenantID      uuid.UUID     `json:"tenant_id"`
	ProviderID    uuid.UUID     `json:"provider_id"`
	StartAt       time.Time     `json:"start_at"`
	EndAt         time.Time     `json:"end_at"`
	Status        string        `json:"status"`
	AppointmentID uuid.NullUUID `json:"appointment_id"`
	CreatedAt     time.Time     `json:"created_at"`
}

type Appointment struct {
	ID         uuid.UUID     `json:"id"`
	TenantID   uuid.UUID     `json:"tenant_id"`
	CustomerID uuid.UUID     `json:"customer_id"`
	ProviderID uuid.UUID     `json:"provider_id"`
	ServiceID  uuid.UUID     `json:"service_id"`
	SlotID     uuid.NullUUID `json:"slot_id"`
	StartAt    time.Time     `json:"start_at"`
	EndAt      time.Time     `json:"end_at"`
	Status     string        `json:"status"`
	Notes      pgtype.Text   `json:"notes"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

type AppointmentEvent struct {
	ID            uuid.UUID `json:"id"`
	AppointmentID uuid.UUID `json:"appointment_id"`
	EventType     string    `json:"event_type"`
	Payload       []byte    `json:"payload"`
	CreatedAt     time.Time `json:"created_at"`
}

type ConversationThread struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	CustomerID uuid.UUID `json:"customer_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type ConversationMessage struct {
	ID        uuid.UUID `json:"id"`
	ThreadID  uuid.UUID `json:"thread_id"`
	Direction string    `json:"direction"`
	Message   string    `json:"message"`
	Metadata  []byte    `json:"metadata"`
	CreatedAt time.Time `json:"created_at"`
}

type ConversationState struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	CustomerID uuid.UUID `json:"customer_id"`
	State      string    `json:"state"`
	Data       []byte    `json:"data"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type WebhookLog struct {
	ID        uuid.UUID     `json:"id"`
	TenantID  uuid.NullUUID `json:"tenant_id"`
	Source    string        `json:"source"`
	Payload   []byte        `json:"payload"`
	CreatedAt time.Time     `json:"created_at"`
}
