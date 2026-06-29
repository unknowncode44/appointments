# Changelog

All notable changes to TurnoBot / Appointments MVP are documented here.  
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

---

## [Unreleased] — `fix/tenant-slug-conflict-and-backfill`

### Fixed

- **Duplicate tenant slug returns `409`, not `500`.**
  `POST`/`PUT /api/v1/tenants` with a slug that's already taken now maps the
  Postgres unique-violation on `tenants_slug_idx` to a clean
  `409 Conflict` (`{"error":"slug already in use"}`) instead of a generic
  `500`. `tenants` only has the slug unique constraint, so the mapping is
  unambiguous.
- **Hardened slug backfill in migration `000004` (fresh installs only).**
  Same-base collisions now derive their suffix from the tenant id
  (`left(id::text, 8)`) instead of a per-base counter, so a generated slug can
  no longer collide with a tenant literally named e.g. "Base 2" and break the
  `CREATE UNIQUE INDEX`. _Note:_ `000004` already ran on prod, so golang-migrate
  will not re-apply it there — this only affects new/fresh databases.

---

## [Unreleased] — `feat/fase2-public-booking`

### Added

- **F2.1 — Public, no-auth booking flow.**
  A new `/public/:slug` route group lets an end customer self-book an
  appointment from a shareable link without logging in. The tenant is resolved
  from the URL `slug`; every query is scoped to that tenant (the slug *is* the
  scope). Endpoints:
  - `GET /public/:slug` — shop header (name, timezone, slug).
  - `GET /public/:slug/services` — active services.
  - `GET /public/:slug/providers` — active providers.
  - `GET /public/:slug/availability?date=&provider_id=` — available slots
    (reuses `ListAvailableSlots`; provider-time, not filtered by service).
  - `POST /public/:slug/appointments` — create a **`confirmed`** booking
    (no payment, no `pending`, no slot-hold). Find-or-create customer by phone,
    then run the shared booking core.

- **`tenants.slug`** (migration `000004`) — optional unique public slug
  (`[a-z0-9-]`, 3–63 chars). Existing tenants are backfilled from their name.
  Settable via `POST`/`PUT /api/v1/tenants`.

- **`customers.phone` / `customers.email`** (migration `000005`) — nullable;
  partial unique index on `(tenant_id, phone)` backs a race-safe
  `UpsertCustomerByPhone`.

### Changed

- **Shared booking core.** The reserve-slot transaction in
  `appointmentService.Create` was extracted into `bookSlot`, now shared by the
  authenticated and public booking paths (no duplicated `FOR UPDATE` logic). A
  slot belonging to another tenant (or missing) is reported as `404`; a visible
  slot that is no longer available is `409`.

### Security

- **Dedicated rate limiter** on `/public/*` (25 req/min/IP) in addition to the
  global 200/min limiter, to mitigate scraping and booking abuse.
- Public booking validates that the `service_id` and `slot_id` belong to the
  resolved tenant — no public endpoint can read or write another tenant's data.
- Captcha is recommended as a future follow-up (not implemented).

---

## [Unreleased] — `fix/fase0-security-hardening`

### Security

- **F0-1 — Multi-tenant isolation enforced at service layer.**  
  All by-ID and write operations (`GET /:id`, `PUT /:id`, `DELETE /:id`, `POST`) now verify that the requested resource belongs to the caller's tenant before returning data or performing mutations. A caller outside their tenant receives `404 Not Found` — not `403 Forbidden` — to avoid leaking the existence of resources in other tenants.  
  Affected resources: providers, services, customers, tenant channels, appointments, conversation threads.

- **F0-1 — POST writes force `tenant_id` from JWT.**  
  For `tenantUser` and `user` roles, any `tenant_id` supplied in a POST body is silently replaced with the value from the JWT. This prevents a caller from creating resources under a different tenant by manipulating the request body.

- **F0-2 — `RequireTenant` middleware now applies to the `user` role.**  
  Previously the `user` role bypassed tenant enforcement on list endpoints (only `tenantUser` was checked). The middleware now enforces tenant scope for all non-admin roles. `adminUser` continues to bypass the check.

- **F0.5-1 — Tenant-scope the scheduling endpoints.**  
  `POST /providers/:id/availability`, `POST /providers/:id/exceptions`, `POST /slot-generator`, and the related `GET .../availability` / `GET .../exceptions` listings now verify that the target provider belongs to the caller's tenant before reading or writing. Cross-tenant access returns `404 Not Found`. `POST /slot-generator` additionally forces `tenant_id` from the JWT, ignoring any `tenant_id` in the body. `adminUser` bypasses all checks.

- **F0.5-2 — Tenant-scope `POST /conversations/message`.**  
  The endpoint now forces `tenant_id` from the JWT and verifies that the supplied `customer_id` belongs to the caller's tenant; a cross-tenant customer returns `404 Not Found`. `adminUser` bypasses the check.

- **F0-5 — CORS no longer allows all origins (`*`).**  
  The `Access-Control-Allow-Origin` header is now controlled by the `ALLOWED_ORIGINS` configuration variable. Only listed origins are permitted. `AllowHeaders` is also narrowed to `Origin, Content-Type, Accept, Authorization`.

### Added

- **F0-5 — Global rate limiter (200 req/min per IP).**  
  Applied to all routes. The existing auth-specific limiter (10 req/min on `/login` and `/register`) is kept on top of the global one.

- **F0-5 — Health endpoints.**  
  - `GET /healthz` — liveness probe; always `200 OK` if the process is running.  
  - `GET /readyz` — readiness probe; returns `200 OK` if the database connection is alive, `503 Service Unavailable` otherwise.  
  Both endpoints are public (no auth, exempt from rate limiting).

- **F0-4 — `user_id` field in JWT payload.**  
  Access and refresh tokens now carry `user_id` (int32), which is the authenticated user's numeric primary key. Previously only the session UUID (`id`) was available, making it impossible to identify the user from the token alone without a DB query. The `id` field remains the session token UUID and is unchanged.

- **F0-3 — Tenant isolation unit tests (`tests/tenant_isolation_test.go`).**  
  20 focused unit tests using stub repositories (no database required). Cover the three-way matrix for every protected resource type:  
  - Cross-tenant access → `ErrNotFound`  
  - Own-tenant access → success  
  - `adminUser` (nil scope) → success  
  Resources covered: Provider, Service, Customer, TenantChannel, Appointment, ConversationThread. Write operations (Update, Deactivate) also covered.

- **F0-3 — Booking integration tests (`tests/booking_test.go`).**  
  Four integration tests that require a real PostgreSQL database. Set `TEST_DATABASE_URL` to run; tests are skipped gracefully otherwise.  
  - `TestBooking_Create` — booking a slot marks it `reserved` and creates the appointment as `confirmed`.  
  - `TestBooking_Cancel` — cancelling releases the slot back to `available`.  
  - `TestBooking_DoubleBook` — attempting to book an already-reserved slot returns `ErrConflict` (exercises the `SELECT … FOR UPDATE` guard).  
  - `TestBooking_Cancel_CrossTenant_Returns404` — a caller from a different tenant cannot cancel another tenant's appointment.

- **F0-4 — DB `CHECK` constraint on `appointments.status`.**  
  Migration `000003_appointment_status_check` adds:  
  ```sql
  CHECK (status IN ('confirmed', 'cancelled', 'completed', 'no_show'))
  ```  
  Invalid status values are rejected at the database level.

- **`Store.Ping(ctx) error`** — new method on `*db.Store` used by `/readyz` to verify the database connection.

### Changed

- **F0-4 — Appointment status on creation changed from `reserved` to `confirmed`.**  
  New appointments are created with `status = 'confirmed'`. The slot itself transitions to `status = 'reserved'`. Previous documentation incorrectly described the appointment status as `reserved`.

- **F0-4 — `AppointmentUpdateRequest.Status` now validated.**  
  The `status` field on `PATCH /api/v1/appointments/:id` is validated against the enum `confirmed | cancelled | completed | no_show` before reaching the service layer. Unknown values return `400 Bad Request`.

- **F0-6 — Migration startup warning documented in `main.go`.**  
  A comment in `runDBMigration` explains that running migrations at process startup is safe only for single-replica deployments. Multi-replica deployments must use a dedicated migration job or a Postgres advisory lock.

- **F0-7 — Go module renamed.**  
  `github.com/mousav1/ticket` → `github.com/unknowncode44/appointments` across all Go source files and `go.mod`. This matches the canonical GitHub remote (`origin`).

---

## Previous releases

### `fixes` branch

| Area | Change |
|---|---|
| `UserResponse.id` | Changed from `int32` to `uuid.UUID` |
| `/api/v1/users/:id` path param | Changed from integer to UUID |
| `POST /api/v1/slot-generator` | New optional field `timezone` (IANA string) |
| `GET /api/v1/users` | Pagination changed from `limit`/`offset` to `page`/`page_size` |
| `POST /api/v1/webhooks/evolution` | Parses Evolution API v2 nested payload |
