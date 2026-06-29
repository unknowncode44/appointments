-- name: GetConversationState :one
SELECT id, tenant_id, customer_id, state, data, updated_at
FROM conversation_state
WHERE tenant_id = $1 AND customer_id = $2;

-- name: InsertInboundDedup :one
-- Race-safe idempotency guard: succeeds (returns the row) the first time an
-- Evolution message id is seen for a tenant, and returns no rows on a duplicate.
INSERT INTO inbound_message_dedup (tenant_id, external_message_id)
VALUES ($1, $2)
ON CONFLICT (tenant_id, external_message_id) DO NOTHING
RETURNING id, tenant_id, external_message_id, created_at;

-- name: CreateOutboundMessage :one
INSERT INTO outbound_messages (tenant_id, customer_id, channel_id, message, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, tenant_id, customer_id, channel_id, message, status, created_at;

-- name: UpdateOutboundMessageStatus :one
UPDATE outbound_messages
SET status = $2
WHERE id = $1
RETURNING id, tenant_id, customer_id, channel_id, message, status, created_at;
