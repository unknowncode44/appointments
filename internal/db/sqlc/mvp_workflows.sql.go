package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateAppointmentParams struct {
	TenantID   uuid.UUID
	CustomerID uuid.UUID
	ProviderID uuid.UUID
	ServiceID  uuid.UUID
	SlotID     uuid.NullUUID
	StartAt    time.Time
	EndAt      time.Time
	Status     string
	Notes      pgtype.Text
}

func (q *Queries) CreateAppointment(ctx context.Context, arg CreateAppointmentParams) (Appointment, error) {
	var i Appointment
	err := q.db.QueryRow(ctx, "INSERT INTO appointments (tenant_id, customer_id, provider_id, service_id, slot_id, start_at, end_at, status, notes) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id, tenant_id, customer_id, provider_id, service_id, slot_id, start_at, end_at, status, notes, created_at, updated_at", arg.TenantID, arg.CustomerID, arg.ProviderID, arg.ServiceID, arg.SlotID, arg.StartAt, arg.EndAt, arg.Status, arg.Notes).Scan(&i.ID, &i.TenantID, &i.CustomerID, &i.ProviderID, &i.ServiceID, &i.SlotID, &i.StartAt, &i.EndAt, &i.Status, &i.Notes, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

type ListAppointmentsParams struct {
	TenantID   uuid.UUID
	ProviderID *uuid.UUID
	CustomerID *uuid.UUID
	Status     string
	Limit      int32
	Offset     int32
}

func (q *Queries) ListAppointments(ctx context.Context, arg ListAppointmentsParams) ([]Appointment, error) {
	rows, err := q.db.Query(ctx, "SELECT id, tenant_id, customer_id, provider_id, service_id, slot_id, start_at, end_at, status, notes, created_at, updated_at FROM appointments WHERE tenant_id = $1 AND ($2::uuid IS NULL OR provider_id = $2) AND ($3::uuid IS NULL OR customer_id = $3) AND ($4::text = '' OR status = $4) ORDER BY start_at DESC LIMIT $5 OFFSET $6", arg.TenantID, arg.ProviderID, arg.CustomerID, arg.Status, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Appointment{}
	for rows.Next() {
		var i Appointment
		if err := rows.Scan(&i.ID, &i.TenantID, &i.CustomerID, &i.ProviderID, &i.ServiceID, &i.SlotID, &i.StartAt, &i.EndAt, &i.Status, &i.Notes, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (q *Queries) CountAppointments(ctx context.Context, arg ListAppointmentsParams) (int64, error) {
	var count int64
	err := q.db.QueryRow(ctx, "SELECT count(*) FROM appointments WHERE tenant_id = $1 AND ($2::uuid IS NULL OR provider_id = $2) AND ($3::uuid IS NULL OR customer_id = $3) AND ($4::text = '' OR status = $4)", arg.TenantID, arg.ProviderID, arg.CustomerID, arg.Status).Scan(&count)
	return count, err
}

func (q *Queries) GetAppointment(ctx context.Context, id uuid.UUID) (Appointment, error) {
	var i Appointment
	err := q.db.QueryRow(ctx, "SELECT id, tenant_id, customer_id, provider_id, service_id, slot_id, start_at, end_at, status, notes, created_at, updated_at FROM appointments WHERE id = $1", id).Scan(&i.ID, &i.TenantID, &i.CustomerID, &i.ProviderID, &i.ServiceID, &i.SlotID, &i.StartAt, &i.EndAt, &i.Status, &i.Notes, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

type UpdateAppointmentParams struct {
	ID         uuid.UUID
	ProviderID uuid.UUID
	ServiceID  uuid.UUID
	SlotID     uuid.NullUUID
	StartAt    time.Time
	EndAt      time.Time
	Status     string
	Notes      pgtype.Text
}

func (q *Queries) UpdateAppointment(ctx context.Context, arg UpdateAppointmentParams) (Appointment, error) {
	var i Appointment
	err := q.db.QueryRow(ctx, "UPDATE appointments SET provider_id = $2, service_id = $3, slot_id = $4, start_at = $5, end_at = $6, status = $7, notes = $8, updated_at = now() WHERE id = $1 RETURNING id, tenant_id, customer_id, provider_id, service_id, slot_id, start_at, end_at, status, notes, created_at, updated_at", arg.ID, arg.ProviderID, arg.ServiceID, arg.SlotID, arg.StartAt, arg.EndAt, arg.Status, arg.Notes).Scan(&i.ID, &i.TenantID, &i.CustomerID, &i.ProviderID, &i.ServiceID, &i.SlotID, &i.StartAt, &i.EndAt, &i.Status, &i.Notes, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

func (q *Queries) CancelAppointment(ctx context.Context, id uuid.UUID) (Appointment, error) {
	var i Appointment
	err := q.db.QueryRow(ctx, "UPDATE appointments SET status = 'cancelled', updated_at = now() WHERE id = $1 RETURNING id, tenant_id, customer_id, provider_id, service_id, slot_id, start_at, end_at, status, notes, created_at, updated_at", id).Scan(&i.ID, &i.TenantID, &i.CustomerID, &i.ProviderID, &i.ServiceID, &i.SlotID, &i.StartAt, &i.EndAt, &i.Status, &i.Notes, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

func (q *Queries) GetAppointmentSlotForUpdate(ctx context.Context, id uuid.UUID) (AppointmentSlot, error) {
	var i AppointmentSlot
	err := q.db.QueryRow(ctx, "SELECT id, tenant_id, provider_id, start_at, end_at, status, appointment_id, created_at FROM appointment_slots WHERE id = $1 FOR UPDATE", id).Scan(&i.ID, &i.TenantID, &i.ProviderID, &i.StartAt, &i.EndAt, &i.Status, &i.AppointmentID, &i.CreatedAt)
	return i, err
}

type ReserveAppointmentSlotParams struct {
	ID            uuid.UUID
	AppointmentID uuid.UUID
}

func (q *Queries) ReserveAppointmentSlot(ctx context.Context, arg ReserveAppointmentSlotParams) (AppointmentSlot, error) {
	var i AppointmentSlot
	err := q.db.QueryRow(ctx, "UPDATE appointment_slots SET status = 'reserved', appointment_id = $2 WHERE id = $1 RETURNING id, tenant_id, provider_id, start_at, end_at, status, appointment_id, created_at", arg.ID, arg.AppointmentID).Scan(&i.ID, &i.TenantID, &i.ProviderID, &i.StartAt, &i.EndAt, &i.Status, &i.AppointmentID, &i.CreatedAt)
	return i, err
}

func (q *Queries) ReleaseAppointmentSlot(ctx context.Context, id uuid.UUID) (AppointmentSlot, error) {
	var i AppointmentSlot
	err := q.db.QueryRow(ctx, "UPDATE appointment_slots SET status = 'available', appointment_id = NULL WHERE id = $1 RETURNING id, tenant_id, provider_id, start_at, end_at, status, appointment_id, created_at", id).Scan(&i.ID, &i.TenantID, &i.ProviderID, &i.StartAt, &i.EndAt, &i.Status, &i.AppointmentID, &i.CreatedAt)
	return i, err
}

type CreateAppointmentEventParams struct {
	AppointmentID uuid.UUID
	EventType     string
	Payload       []byte
}

func (q *Queries) CreateAppointmentEvent(ctx context.Context, arg CreateAppointmentEventParams) (AppointmentEvent, error) {
	var i AppointmentEvent
	err := q.db.QueryRow(ctx, "INSERT INTO appointment_events (appointment_id, event_type, payload) VALUES ($1, $2, $3) RETURNING id, appointment_id, event_type, payload, created_at", arg.AppointmentID, arg.EventType, arg.Payload).Scan(&i.ID, &i.AppointmentID, &i.EventType, &i.Payload, &i.CreatedAt)
	return i, err
}

type ListConversationThreadsParams struct {
	TenantID uuid.UUID
	Limit    int32
	Offset   int32
}

func (q *Queries) ListConversationThreads(ctx context.Context, arg ListConversationThreadsParams) ([]ConversationThread, error) {
	rows, err := q.db.Query(ctx, "SELECT id, tenant_id, customer_id, created_at FROM conversation_threads WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3", arg.TenantID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ConversationThread{}
	for rows.Next() {
		var i ConversationThread
		if err := rows.Scan(&i.ID, &i.TenantID, &i.CustomerID, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (q *Queries) GetConversationThread(ctx context.Context, id uuid.UUID) (ConversationThread, error) {
	var i ConversationThread
	err := q.db.QueryRow(ctx, "SELECT id, tenant_id, customer_id, created_at FROM conversation_threads WHERE id = $1", id).Scan(&i.ID, &i.TenantID, &i.CustomerID, &i.CreatedAt)
	return i, err
}

func (q *Queries) ListConversationMessages(ctx context.Context, threadID uuid.UUID) ([]ConversationMessage, error) {
	rows, err := q.db.Query(ctx, "SELECT id, thread_id, direction, message, metadata, created_at FROM conversation_messages WHERE thread_id = $1 ORDER BY created_at ASC", threadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ConversationMessage{}
	for rows.Next() {
		var i ConversationMessage
		if err := rows.Scan(&i.ID, &i.ThreadID, &i.Direction, &i.Message, &i.Metadata, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

type CreateConversationThreadParams struct {
	TenantID   uuid.UUID
	CustomerID uuid.UUID
}

func (q *Queries) CreateConversationThread(ctx context.Context, arg CreateConversationThreadParams) (ConversationThread, error) {
	var i ConversationThread
	err := q.db.QueryRow(ctx, "INSERT INTO conversation_threads (tenant_id, customer_id) VALUES ($1, $2) RETURNING id, tenant_id, customer_id, created_at", arg.TenantID, arg.CustomerID).Scan(&i.ID, &i.TenantID, &i.CustomerID, &i.CreatedAt)
	return i, err
}

func (q *Queries) GetConversationThreadByCustomer(ctx context.Context, arg CreateConversationThreadParams) (ConversationThread, error) {
	var i ConversationThread
	err := q.db.QueryRow(ctx, "SELECT id, tenant_id, customer_id, created_at FROM conversation_threads WHERE tenant_id = $1 AND customer_id = $2 ORDER BY created_at DESC LIMIT 1", arg.TenantID, arg.CustomerID).Scan(&i.ID, &i.TenantID, &i.CustomerID, &i.CreatedAt)
	return i, err
}

type CreateConversationMessageParams struct {
	ThreadID  uuid.UUID
	Direction string
	Message   string
	Metadata  []byte
}

func (q *Queries) CreateConversationMessage(ctx context.Context, arg CreateConversationMessageParams) (ConversationMessage, error) {
	var i ConversationMessage
	err := q.db.QueryRow(ctx, "INSERT INTO conversation_messages (thread_id, direction, message, metadata) VALUES ($1, $2, $3, $4) RETURNING id, thread_id, direction, message, metadata, created_at", arg.ThreadID, arg.Direction, arg.Message, arg.Metadata).Scan(&i.ID, &i.ThreadID, &i.Direction, &i.Message, &i.Metadata, &i.CreatedAt)
	return i, err
}

type UpsertConversationStateParams struct {
	TenantID   uuid.UUID
	CustomerID uuid.UUID
	State      string
	Data       []byte
}

func (q *Queries) UpsertConversationState(ctx context.Context, arg UpsertConversationStateParams) (ConversationState, error) {
	var i ConversationState
	err := q.db.QueryRow(ctx, "INSERT INTO conversation_state (tenant_id, customer_id, state, data) VALUES ($1, $2, $3, $4) RETURNING id, tenant_id, customer_id, state, data, updated_at", arg.TenantID, arg.CustomerID, arg.State, arg.Data).Scan(&i.ID, &i.TenantID, &i.CustomerID, &i.State, &i.Data, &i.UpdatedAt)
	return i, err
}

type CreateWebhookLogParams struct {
	TenantID uuid.NullUUID
	Source   string
	Payload  []byte
}

func (q *Queries) CreateWebhookLog(ctx context.Context, arg CreateWebhookLogParams) (WebhookLog, error) {
	var i WebhookLog
	err := q.db.QueryRow(ctx, "INSERT INTO webhook_logs (tenant_id, source, payload) VALUES ($1, $2, $3) RETURNING id, tenant_id, source, payload, created_at", arg.TenantID, arg.Source, arg.Payload).Scan(&i.ID, &i.TenantID, &i.Source, &i.Payload, &i.CreatedAt)
	return i, err
}

func (q *Queries) GetTenantChannelByExternalID(ctx context.Context, externalID string) (TenantChannel, error) {
	var i TenantChannel
	err := q.db.QueryRow(ctx, "SELECT id, tenant_id, channel_type, external_id, external_key, active, created_at FROM tenant_channels WHERE external_id = $1 AND active = true", externalID).Scan(&i.ID, &i.TenantID, &i.ChannelType, &i.ExternalID, &i.ExternalKey, &i.Active, &i.CreatedAt)
	return i, err
}

type GetCustomerByChannelParams struct {
	TenantChannelID    uuid.UUID
	ExternalIdentifier string
}

func (q *Queries) GetCustomerByChannel(ctx context.Context, arg GetCustomerByChannelParams) (Customer, error) {
	var i Customer
	err := q.db.QueryRow(ctx, "SELECT c.id, c.tenant_id, c.first_name, c.last_name, c.notes, c.created_at, c.updated_at FROM customers c JOIN customer_channels cc ON cc.customer_id = c.id WHERE cc.tenant_channel_id = $1 AND cc.external_identifier = $2", arg.TenantChannelID, arg.ExternalIdentifier).Scan(&i.ID, &i.TenantID, &i.FirstName, &i.LastName, &i.Notes, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

type CreateCustomerChannelParams struct {
	CustomerID         uuid.UUID
	TenantChannelID    uuid.UUID
	ExternalIdentifier string
}

func (q *Queries) CreateCustomerChannel(ctx context.Context, arg CreateCustomerChannelParams) (CustomerChannel, error) {
	var i CustomerChannel
	err := q.db.QueryRow(ctx, "INSERT INTO customer_channels (customer_id, tenant_channel_id, external_identifier) VALUES ($1, $2, $3) RETURNING id, customer_id, tenant_channel_id, external_identifier, created_at", arg.CustomerID, arg.TenantChannelID, arg.ExternalIdentifier).Scan(&i.ID, &i.CustomerID, &i.TenantChannelID, &i.ExternalIdentifier, &i.CreatedAt)
	return i, err
}
