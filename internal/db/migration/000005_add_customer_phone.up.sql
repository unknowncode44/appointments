-- Public bookers are identified by phone within a tenant.
-- Both columns are nullable: existing customers (WhatsApp-only) have neither.
ALTER TABLE customers ADD COLUMN phone VARCHAR(32);
ALTER TABLE customers ADD COLUMN email VARCHAR(255);

-- Partial unique index supports a race-safe upsert on (tenant_id, phone) while
-- allowing many existing phone-less customers (NULL phone) to coexist.
CREATE UNIQUE INDEX customers_tenant_phone_idx
ON customers(tenant_id, phone)
WHERE phone IS NOT NULL;
