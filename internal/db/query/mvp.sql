-- name: ListTenants :many
SELECT id, name, timezone, active, created_at, updated_at, slug
FROM tenants
WHERE ($1::text = '' OR name ILIKE '%' || $1 || '%')
  AND ($2::boolean IS NULL OR active = $2)
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountTenants :one
SELECT count(*) FROM tenants
WHERE ($1::text = '' OR name ILIKE '%' || $1 || '%')
  AND ($2::boolean IS NULL OR active = $2);

-- name: GetTenant :one
SELECT id, name, timezone, active, created_at, updated_at, slug
FROM tenants
WHERE id = $1;

-- name: GetTenantBySlug :one
SELECT id, name, timezone, active, created_at, updated_at, slug
FROM tenants
WHERE slug = $1 AND active = true;

-- name: CreateTenant :one
INSERT INTO tenants (name, timezone, slug)
VALUES ($1, $2, $3)
RETURNING id, name, timezone, active, created_at, updated_at, slug;

-- name: UpdateTenant :one
UPDATE tenants
SET name = $2, timezone = $3, slug = $4, updated_at = now()
WHERE id = $1
RETURNING id, name, timezone, active, created_at, updated_at, slug;

-- name: DeactivateTenant :one
UPDATE tenants
SET active = false, updated_at = now()
WHERE id = $1
RETURNING id, name, timezone, active, created_at, updated_at, slug;

-- name: ListProviders :many
SELECT id, tenant_id, name, active, created_at, updated_at
FROM providers
WHERE tenant_id = $1
  AND ($2::text = '' OR name ILIKE '%' || $2 || '%')
  AND ($3::boolean IS NULL OR active = $3)
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: CountProviders :one
SELECT count(*) FROM providers
WHERE tenant_id = $1
  AND ($2::text = '' OR name ILIKE '%' || $2 || '%')
  AND ($3::boolean IS NULL OR active = $3);

-- name: GetProvider :one
SELECT id, tenant_id, name, active, created_at, updated_at
FROM providers
WHERE id = $1;

-- name: CreateProvider :one
INSERT INTO providers (tenant_id, name)
VALUES ($1, $2)
RETURNING id, tenant_id, name, active, created_at, updated_at;

-- name: UpdateProvider :one
UPDATE providers
SET name = $2, updated_at = now()
WHERE id = $1
RETURNING id, tenant_id, name, active, created_at, updated_at;

-- name: DeactivateProvider :one
UPDATE providers
SET active = false, updated_at = now()
WHERE id = $1
RETURNING id, tenant_id, name, active, created_at, updated_at;

-- name: ListServices :many
SELECT id, tenant_id, name, duration_minutes, active, created_at, updated_at
FROM services
WHERE tenant_id = $1
  AND ($2::text = '' OR name ILIKE '%' || $2 || '%')
  AND ($3::boolean IS NULL OR active = $3)
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: CountServices :one
SELECT count(*) FROM services
WHERE tenant_id = $1
  AND ($2::text = '' OR name ILIKE '%' || $2 || '%')
  AND ($3::boolean IS NULL OR active = $3);

-- name: GetService :one
SELECT id, tenant_id, name, duration_minutes, active, created_at, updated_at
FROM services
WHERE id = $1;

-- name: CreateService :one
INSERT INTO services (tenant_id, name, duration_minutes)
VALUES ($1, $2, $3)
RETURNING id, tenant_id, name, duration_minutes, active, created_at, updated_at;

-- name: UpdateService :one
UPDATE services
SET name = $2, duration_minutes = $3, updated_at = now()
WHERE id = $1
RETURNING id, tenant_id, name, duration_minutes, active, created_at, updated_at;

-- name: DeactivateService :one
UPDATE services
SET active = false, updated_at = now()
WHERE id = $1
RETURNING id, tenant_id, name, duration_minutes, active, created_at, updated_at;

-- name: ListCustomers :many
SELECT id, tenant_id, first_name, last_name, notes, created_at, updated_at, phone, email
FROM customers
WHERE tenant_id = $1
  AND ($2::text = '' OR coalesce(first_name, '') ILIKE '%' || $2 || '%' OR coalesce(last_name, '') ILIKE '%' || $2 || '%')
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountCustomers :one
SELECT count(*) FROM customers
WHERE tenant_id = $1
  AND ($2::text = '' OR coalesce(first_name, '') ILIKE '%' || $2 || '%' OR coalesce(last_name, '') ILIKE '%' || $2 || '%');

-- name: GetCustomer :one
SELECT id, tenant_id, first_name, last_name, notes, created_at, updated_at, phone, email
FROM customers
WHERE id = $1;

-- name: CreateCustomer :one
INSERT INTO customers (tenant_id, first_name, last_name, notes)
VALUES ($1, $2, $3, $4)
RETURNING id, tenant_id, first_name, last_name, notes, created_at, updated_at, phone, email;

-- name: UpdateCustomer :one
UPDATE customers
SET first_name = $2, last_name = $3, notes = $4, updated_at = now()
WHERE id = $1
RETURNING id, tenant_id, first_name, last_name, notes, created_at, updated_at, phone, email;

-- name: UpsertCustomerByPhone :one
-- Race-safe find-or-create for public bookers identified by (tenant_id, phone).
-- On conflict we update names/email when new values are provided, keeping the
-- existing customer row (and its id) stable.
INSERT INTO customers (tenant_id, first_name, last_name, phone, email)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (tenant_id, phone) WHERE phone IS NOT NULL
DO UPDATE SET
    first_name = COALESCE(EXCLUDED.first_name, customers.first_name),
    last_name  = COALESCE(EXCLUDED.last_name, customers.last_name),
    email      = COALESCE(EXCLUDED.email, customers.email),
    updated_at = now()
RETURNING id, tenant_id, first_name, last_name, notes, created_at, updated_at, phone, email;

-- name: ListTenantChannels :many
SELECT id, tenant_id, channel_type, external_id, external_key, active, created_at
FROM tenant_channels
WHERE tenant_id = $1
  AND ($2::text = '' OR channel_type = $2)
  AND ($3::boolean IS NULL OR active = $3)
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: CountTenantChannels :one
SELECT count(*)
FROM tenant_channels
WHERE tenant_id = $1
  AND ($2::text = '' OR channel_type = $2)
  AND ($3::boolean IS NULL OR active = $3);

-- name: GetTenantChannel :one
SELECT id, tenant_id, channel_type, external_id, external_key, active, created_at
FROM tenant_channels
WHERE id = $1;

-- name: CreateTenantChannel :one
INSERT INTO tenant_channels (tenant_id, channel_type, external_id, external_key)
VALUES ($1, $2, $3, $4)
RETURNING id, tenant_id, channel_type, external_id, external_key, active, created_at;

-- name: UpdateTenantChannel :one
UPDATE tenant_channels
SET channel_type = $2, external_id = $3, external_key = $4, active = $5
WHERE id = $1
RETURNING id, tenant_id, channel_type, external_id, external_key, active, created_at;

-- name: DeactivateTenantChannel :one
UPDATE tenant_channels
SET active = false
WHERE id = $1
RETURNING id, tenant_id, channel_type, external_id, external_key, active, created_at;

-- name: CreateProviderAvailability :one
INSERT INTO provider_availability (provider_id, weekday, start_time, end_time)
VALUES ($1, $2, $3, $4)
RETURNING id, provider_id, weekday, start_time, end_time;

-- name: ListProviderAvailability :many
SELECT id, provider_id, weekday, start_time, end_time
FROM provider_availability
WHERE provider_id = $1
ORDER BY weekday, start_time;

-- name: CreateProviderException :one
INSERT INTO provider_exceptions (provider_id, start_at, end_at, reason)
VALUES ($1, $2, $3, $4)
RETURNING id, provider_id, start_at, end_at, reason, created_at;

-- name: ListProviderExceptions :many
SELECT id, provider_id, start_at, end_at, reason, created_at
FROM provider_exceptions
WHERE provider_id = $1
ORDER BY start_at DESC;

-- name: ListProviderExceptionsBetween :many
SELECT id, provider_id, start_at, end_at, reason, created_at
FROM provider_exceptions
WHERE provider_id = $1 AND start_at < $3 AND end_at > $2
ORDER BY start_at;

-- name: CreateAppointmentSlot :one
INSERT INTO appointment_slots (tenant_id, provider_id, start_at, end_at, status)
SELECT $1, $2, $3, $4, 'available'
WHERE NOT EXISTS (
  SELECT 1 FROM appointment_slots
  WHERE provider_id = $2 AND start_at = $3 AND end_at = $4
)
RETURNING id, tenant_id, provider_id, start_at, end_at, status, appointment_id, created_at;

-- name: ListAvailableSlots :many
SELECT s.id, s.tenant_id, s.provider_id, s.start_at, s.end_at, s.status, s.appointment_id, s.created_at
FROM appointment_slots s
WHERE s.tenant_id = $1
  AND ($2::uuid IS NULL OR s.provider_id = $2)
  AND s.status = 'available'
  AND s.start_at >= $3
  AND s.start_at < $4
  AND ($5::uuid IS NULL OR EXISTS (
    SELECT 1 FROM provider_services ps
    WHERE ps.provider_id = s.provider_id AND ps.service_id = $5
  ))
ORDER BY s.start_at
LIMIT $6 OFFSET $7;

-- name: GetAppointmentSlotForUpdate :one
SELECT id, tenant_id, provider_id, start_at, end_at, status, appointment_id, created_at
FROM appointment_slots
WHERE id = $1
FOR UPDATE;

-- name: ReserveAppointmentSlot :one
UPDATE appointment_slots
SET status = 'reserved', appointment_id = $2
WHERE id = $1
RETURNING id, tenant_id, provider_id, start_at, end_at, status, appointment_id, created_at;

-- name: ReleaseAppointmentSlot :one
UPDATE appointment_slots
SET status = 'available', appointment_id = NULL
WHERE id = $1
RETURNING id, tenant_id, provider_id, start_at, end_at, status, appointment_id, created_at;

-- name: CreateAppointment :one
INSERT INTO appointments (tenant_id, customer_id, provider_id, service_id, slot_id, start_at, end_at, status, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, tenant_id, customer_id, provider_id, service_id, slot_id, start_at, end_at, status, notes, created_at, updated_at;

-- name: ListAppointments :many
SELECT id, tenant_id, customer_id, provider_id, service_id, slot_id, start_at, end_at, status, notes, created_at, updated_at
FROM appointments
WHERE tenant_id = $1
  AND ($2::uuid IS NULL OR provider_id = $2)
  AND ($3::uuid IS NULL OR customer_id = $3)
  AND ($4::text = '' OR status = $4)
ORDER BY start_at DESC
LIMIT $5 OFFSET $6;

-- name: CountAppointments :one
SELECT count(*) FROM appointments
WHERE tenant_id = $1
  AND ($2::uuid IS NULL OR provider_id = $2)
  AND ($3::uuid IS NULL OR customer_id = $3)
  AND ($4::text = '' OR status = $4);

-- name: GetAppointment :one
SELECT id, tenant_id, customer_id, provider_id, service_id, slot_id, start_at, end_at, status, notes, created_at, updated_at
FROM appointments
WHERE id = $1;

-- name: UpdateAppointment :one
UPDATE appointments
SET provider_id = $2, service_id = $3, slot_id = $4, start_at = $5, end_at = $6, status = $7, notes = $8, updated_at = now()
WHERE id = $1
RETURNING id, tenant_id, customer_id, provider_id, service_id, slot_id, start_at, end_at, status, notes, created_at, updated_at;

-- name: CancelAppointment :one
UPDATE appointments
SET status = 'cancelled', updated_at = now()
WHERE id = $1
RETURNING id, tenant_id, customer_id, provider_id, service_id, slot_id, start_at, end_at, status, notes, created_at, updated_at;

-- name: CreateAppointmentEvent :one
INSERT INTO appointment_events (appointment_id, event_type, payload)
VALUES ($1, $2, $3)
RETURNING id, appointment_id, event_type, payload, created_at;

-- name: ListConversationThreads :many
SELECT id, tenant_id, customer_id, created_at
FROM conversation_threads
WHERE tenant_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetConversationThread :one
SELECT id, tenant_id, customer_id, created_at
FROM conversation_threads
WHERE id = $1;

-- name: ListConversationMessages :many
SELECT id, thread_id, direction, message, metadata, created_at
FROM conversation_messages
WHERE thread_id = $1
ORDER BY created_at ASC;

-- name: CreateConversationThread :one
INSERT INTO conversation_threads (tenant_id, customer_id)
VALUES ($1, $2)
RETURNING id, tenant_id, customer_id, created_at;

-- name: GetConversationThreadByCustomer :one
SELECT id, tenant_id, customer_id, created_at
FROM conversation_threads
WHERE tenant_id = $1 AND customer_id = $2
ORDER BY created_at DESC
LIMIT 1;

-- name: CreateConversationMessage :one
INSERT INTO conversation_messages (thread_id, direction, message, metadata)
VALUES ($1, $2, $3, $4)
RETURNING id, thread_id, direction, message, metadata, created_at;

-- name: UpsertConversationState :one
INSERT INTO conversation_state (tenant_id, customer_id, state, data)
VALUES ($1, $2, $3, $4)
ON CONFLICT (tenant_id, customer_id) DO UPDATE
SET state = EXCLUDED.state, data = EXCLUDED.data, updated_at = now()
RETURNING id, tenant_id, customer_id, state, data, updated_at;

-- name: CreateWebhookLog :one
INSERT INTO webhook_logs (tenant_id, source, payload)
VALUES ($1, $2, $3)
RETURNING id, tenant_id, source, payload, created_at;

-- name: GetTenantChannelByExternalID :one
SELECT id, tenant_id, channel_type, external_id, external_key, active, created_at
FROM tenant_channels
WHERE external_id = $1 AND active = true;

-- name: GetCustomerByChannel :one
SELECT c.id, c.tenant_id, c.first_name, c.last_name, c.notes, c.created_at, c.updated_at
FROM customers c
JOIN customer_channels cc ON cc.customer_id = c.id
WHERE cc.tenant_channel_id = $1 AND cc.external_identifier = $2;

-- name: CreateCustomerChannel :one
INSERT INTO customer_channels (customer_id, tenant_channel_id, external_identifier)
VALUES ($1, $2, $3)
RETURNING id, customer_id, tenant_channel_id, external_identifier, created_at;
