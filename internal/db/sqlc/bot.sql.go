package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// OutboundMessage mirrors the outbound_messages table. Customer and channel are
// nullable so non-conversational sends can omit them.
type OutboundMessage struct {
	ID         uuid.UUID     `json:"id"`
	TenantID   uuid.UUID     `json:"tenant_id"`
	CustomerID uuid.NullUUID `json:"customer_id"`
	ChannelID  uuid.NullUUID `json:"channel_id"`
	Message    string        `json:"message"`
	Status     string        `json:"status"`
	CreatedAt  time.Time     `json:"created_at"`
}

// InboundMessageDedup mirrors the inbound_message_dedup idempotency table.
type InboundMessageDedup struct {
	ID                uuid.UUID `json:"id"`
	TenantID          uuid.UUID `json:"tenant_id"`
	ExternalMessageID string    `json:"external_message_id"`
	CreatedAt         time.Time `json:"created_at"`
}

// GetConversationState returns the single FSM state row for a (tenant, customer).
func (q *Queries) GetConversationState(ctx context.Context, tenantID, customerID uuid.UUID) (ConversationState, error) {
	var i ConversationState
	err := q.db.QueryRow(ctx,
		"SELECT id, tenant_id, customer_id, state, data, updated_at FROM conversation_state WHERE tenant_id = $1 AND customer_id = $2",
		tenantID, customerID,
	).Scan(&i.ID, &i.TenantID, &i.CustomerID, &i.State, &i.Data, &i.UpdatedAt)
	return i, err
}

type InsertInboundDedupParams struct {
	TenantID          uuid.UUID
	ExternalMessageID string
}

// InsertInboundDedup is the race-safe idempotency guard. It returns the inserted
// row the first time a (tenant, external_message_id) pair is seen and pgx.ErrNoRows
// on a duplicate (ON CONFLICT DO NOTHING).
func (q *Queries) InsertInboundDedup(ctx context.Context, arg InsertInboundDedupParams) (InboundMessageDedup, error) {
	var i InboundMessageDedup
	err := q.db.QueryRow(ctx,
		"INSERT INTO inbound_message_dedup (tenant_id, external_message_id) VALUES ($1, $2) ON CONFLICT (tenant_id, external_message_id) DO NOTHING RETURNING id, tenant_id, external_message_id, created_at",
		arg.TenantID, arg.ExternalMessageID,
	).Scan(&i.ID, &i.TenantID, &i.ExternalMessageID, &i.CreatedAt)
	return i, err
}

type CreateOutboundMessageParams struct {
	TenantID   uuid.UUID
	CustomerID uuid.NullUUID
	ChannelID  uuid.NullUUID
	Message    string
	Status     string
}

func (q *Queries) CreateOutboundMessage(ctx context.Context, arg CreateOutboundMessageParams) (OutboundMessage, error) {
	var i OutboundMessage
	err := q.db.QueryRow(ctx,
		"INSERT INTO outbound_messages (tenant_id, customer_id, channel_id, message, status) VALUES ($1, $2, $3, $4, $5) RETURNING id, tenant_id, customer_id, channel_id, message, status, created_at",
		arg.TenantID, arg.CustomerID, arg.ChannelID, arg.Message, arg.Status,
	).Scan(&i.ID, &i.TenantID, &i.CustomerID, &i.ChannelID, &i.Message, &i.Status, &i.CreatedAt)
	return i, err
}

type UpdateOutboundMessageStatusParams struct {
	ID     uuid.UUID
	Status string
}

func (q *Queries) UpdateOutboundMessageStatus(ctx context.Context, arg UpdateOutboundMessageStatusParams) (OutboundMessage, error) {
	var i OutboundMessage
	err := q.db.QueryRow(ctx,
		"UPDATE outbound_messages SET status = $2 WHERE id = $1 RETURNING id, tenant_id, customer_id, channel_id, message, status, created_at",
		arg.ID, arg.Status,
	).Scan(&i.ID, &i.TenantID, &i.CustomerID, &i.ChannelID, &i.Message, &i.Status, &i.CreatedAt)
	return i, err
}
