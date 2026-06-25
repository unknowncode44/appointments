DROP INDEX IF EXISTS customers_tenant_phone_idx;
ALTER TABLE customers DROP COLUMN IF EXISTS email;
ALTER TABLE customers DROP COLUMN IF EXISTS phone;
