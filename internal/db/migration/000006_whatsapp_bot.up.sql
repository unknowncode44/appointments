-- WhatsApp bot foundation: idempotent conversation state and inbound dedup.

-- conversation_state must be one row per (tenant_id, customer_id) so the FSM can
-- upsert it. The old UpsertConversationState inserted a new row every call, so
-- collapse any duplicates (keep the most recently updated) before adding the
-- unique index that the ON CONFLICT upsert depends on.
DELETE FROM conversation_state cs
USING conversation_state newer
WHERE cs.tenant_id = newer.tenant_id
  AND cs.customer_id = newer.customer_id
  AND (cs.updated_at < newer.updated_at
       OR (cs.updated_at = newer.updated_at AND cs.id < newer.id));

CREATE UNIQUE INDEX conversation_state_tenant_customer_idx
ON conversation_state(tenant_id, customer_id);

-- Idempotency: dedupe inbound webhook messages by the Evolution message id,
-- scoped per tenant. A unique index makes the dedup insert race-safe.
CREATE TABLE inbound_message_dedup (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    external_message_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX inbound_message_dedup_unique_idx
ON inbound_message_dedup(tenant_id, external_message_id);
