-- Drop user_providers table
DROP TABLE IF EXISTS user_providers CASCADE;

-- Remove role and tenant_id from users table
ALTER TABLE users
DROP COLUMN IF EXISTS role,
DROP COLUMN IF EXISTS tenant_id;

-- Drop user_role enum
DROP TYPE IF EXISTS user_role;
