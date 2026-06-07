package db

import (
	"context"

	"github.com/google/uuid"
)

// UpsertCustomerByChannelParams holds the arguments for UpsertCustomerByChannel.
type UpsertCustomerByChannelParams struct {
	TenantID           uuid.UUID
	TenantChannelID    uuid.UUID
	ExternalIdentifier string
}

// UpsertCustomerByChannel atomically finds or creates a customer for a given
// channel external identifier. It uses INSERT ... ON CONFLICT DO NOTHING on
// both customers and customer_channels to guarantee idempotency under concurrent
// webhook calls for the same phone number.
func (q *Queries) UpsertCustomerByChannel(ctx context.Context, arg UpsertCustomerByChannelParams) (Customer, error) {
	const sql = `
WITH ins_customer AS (
    INSERT INTO customers (tenant_id)
    VALUES ($1)
    ON CONFLICT DO NOTHING
    RETURNING id
),
ins_channel AS (
    INSERT INTO customer_channels (customer_id, tenant_channel_id, external_identifier)
    SELECT id, $2, $3
    FROM ins_customer
    ON CONFLICT (tenant_channel_id, external_identifier) DO NOTHING
),
existing AS (
    SELECT customer_id AS id
    FROM customer_channels
    WHERE tenant_channel_id = $2
      AND external_identifier = $3
    LIMIT 1
),
new_one AS (
    SELECT id FROM ins_customer
)
SELECT c.id, c.tenant_id, c.first_name, c.last_name, c.notes, c.created_at, c.updated_at
FROM customers c
WHERE c.id = COALESCE(
    (SELECT id FROM existing),
    (SELECT id FROM new_one)
)
LIMIT 1`

	var i Customer
	err := q.db.QueryRow(ctx, sql, arg.TenantID, arg.TenantChannelID, arg.ExternalIdentifier).
		Scan(&i.ID, &i.TenantID, &i.FirstName, &i.LastName, &i.Notes, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}
