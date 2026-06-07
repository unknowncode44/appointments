-- name: UpsertCustomerByChannel :one
-- Atomically finds or creates a customer for an inbound channel identifier.
-- This prevents the race condition where two concurrent webhook calls for the
-- same phone number both attempt to create the same customer.
WITH ins_customer AS (
    INSERT INTO customers (tenant_id)
    VALUES (@tenant_id)
    ON CONFLICT DO NOTHING
    RETURNING id
),
ins_channel AS (
    INSERT INTO customer_channels (customer_id, tenant_channel_id, external_identifier)
    SELECT id, @tenant_channel_id, @external_identifier
    FROM ins_customer
    ON CONFLICT (tenant_channel_id, external_identifier) DO NOTHING
),
existing AS (
    SELECT cc.customer_id AS id
    FROM customer_channels cc
    WHERE cc.tenant_channel_id = @tenant_channel_id
      AND cc.external_identifier = @external_identifier
    LIMIT 1
),
new_one AS (
    SELECT id FROM ins_customer
)
SELECT COALESCE(
    (SELECT id FROM existing),
    (SELECT id FROM new_one)
) AS id,
c.tenant_id,
c.first_name,
c.last_name,
c.notes,
c.created_at,
c.updated_at
FROM customers c
WHERE c.id = COALESCE(
    (SELECT id FROM existing),
    (SELECT id FROM new_one)
);
