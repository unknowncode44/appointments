-- Add role and tenant_id to users table
CREATE TYPE user_role AS ENUM ('adminUser', 'tenantUser', 'user');

ALTER TABLE users
ADD COLUMN role user_role NOT NULL DEFAULT 'user',
ADD COLUMN tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;

-- Create user_providers junction table
CREATE TABLE user_providers (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider_id UUID NOT NULL REFERENCES providers(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, provider_id)
);

CREATE INDEX user_providers_user_idx ON user_providers(user_id);
CREATE INDEX user_providers_provider_idx ON user_providers(provider_id);
