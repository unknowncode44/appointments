# Appointments MVP API

Multi-tenant appointment scheduling SaaS backend for barbershops, salons, beauty centers, and similar small businesses. The MVP exposes administrative APIs, provider availability, slot generation, bookings, conversation storage, and an Evolution/WhatsApp webhook foundation.

## Stack

- Go 1.23+
- Fiber
- PostgreSQL
- SQLC-style query layer
- Viper configuration
- JWT authentication
- Docker and Docker Compose
- golang-migrate

## Server

Configuration is loaded from `app.env`.

Required variables:

```env
APP_PORT=8040
APP_NAME=appointments
APP_DEBUG=true
DB_CONNECTION=postgres
DB_HOST=<host>
DB_PORT=<port>
DB_DATABASE=<database>
DB_USERNAME=<user>
DB_PASSWORD=<password>
TOKEN_SECRET_KEY=<32-byte-secret>
ACCESS_TOKEN_DURATION=15m
REFRESH_TOKEN_DURATION=24h
MIGRATION_URL=file://internal/db/migration
```

Run locally:

```bash
go run main.go
```

Run tests:

```bash
go test ./...
```

If the default Go cache is not writable in your environment:

```bash
GOCACHE=/tmp/go-build-cache go test ./...
```

Docker:

```bash
docker-compose up --build
```

## Project Structure

```text
internal/
  api/
    dto/             Request/response DTOs
    handlers/        Fiber HTTP handlers
    middleware/      JWT auth middleware
    response/        Shared HTTP error mapping
    validators/      Request validation wrapper
    server.go        Fiber server setup
  db/
    migration/       Database migrations
    query/           SQLC query definitions
    sqlc/            SQLC generated/manual query wrappers
  platform/
    pagination/      Pagination helpers
  repositories/      Persistence interfaces/adapters
  routes/            Route registration
  services/          Business logic and transactions
  token/             JWT maker and payload
  util/              Config, password, random helpers
```

## Authentication

Public auth endpoints:

```text
POST /register
POST /login
POST /tokens/renew_access
```

Protected endpoints require:

```http
Authorization: Bearer <access_token>
```

Example:

```bash
curl -X POST http://localhost:8040/register \
  -H "Content-Type: application/json" \
  -d '{"username":"dev_local","password":"devpass123","full_name":"Dev Local"}'

curl -X POST http://localhost:8040/login \
  -H "Content-Type: application/json" \
  -d '{"username":"dev_local","password":"devpass123"}'
```

## API Routes

Administrative routes:

```text
GET    /api/v1/tenants
POST   /api/v1/tenants
GET    /api/v1/tenants/:id
PUT    /api/v1/tenants/:id
DELETE /api/v1/tenants/:id

GET    /api/v1/providers
POST   /api/v1/providers
GET    /api/v1/providers/:id
PUT    /api/v1/providers/:id
DELETE /api/v1/providers/:id

GET    /api/v1/services
POST   /api/v1/services
GET    /api/v1/services/:id
PUT    /api/v1/services/:id
DELETE /api/v1/services/:id

GET    /api/v1/customers
POST   /api/v1/customers
GET    /api/v1/customers/:id
PUT    /api/v1/customers/:id
```

Scheduling routes:

```text
POST /api/v1/providers/:id/availability
GET  /api/v1/providers/:id/availability
POST /api/v1/providers/:id/exceptions
GET  /api/v1/providers/:id/exceptions
POST /api/v1/slot-generator
GET  /api/v1/availability
```

Appointment routes:

```text
POST   /api/v1/appointments
GET    /api/v1/appointments
GET    /api/v1/appointments/:id
PATCH  /api/v1/appointments/:id
DELETE /api/v1/appointments/:id
```

Conversation and webhook routes:

```text
GET  /api/v1/conversations
GET  /api/v1/conversations/:id
POST /api/v1/conversations/message

POST /api/v1/webhooks/evolution
```

`POST /api/v1/webhooks/evolution` is public so Evolution can call it without a JWT. It resolves tenants through `tenant_channels.external_id`.

## Pagination and Filters

List endpoints support:

```text
page
page_size
search
active
tenant_id
provider_id
customer_id
status
date
```

Common examples:

```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8040/api/v1/providers?tenant_id=$TENANT_ID&page=1&page_size=20"

curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8040/api/v1/availability?tenant_id=$TENANT_ID&provider_id=$PROVIDER_ID&date=2026-06-01"
```

## Booking Flow

1. Create tenant.
2. Create providers, services, and customers for the tenant.
3. Add provider weekly availability.
4. Add provider exceptions for blocked periods.
5. Generate appointment slots.
6. Query available slots.
7. Book an appointment using an available slot.
8. Appointment creation reserves the slot transactionally.
9. Cancellation marks the appointment cancelled and releases the slot.
10. Rescheduling releases the old slot, reserves the new slot, updates the appointment, and creates an appointment event.

## Application Flow

```mermaid
flowchart TD
    Client[Admin or App Client] --> Auth[Register/Login]
    Auth --> JWT[JWT Access Token]
    JWT --> AdminAPI[Administrative APIs]
    AdminAPI --> Tenant[Tenants]
    Tenant --> Provider[Providers]
    Tenant --> Service[Services]
    Tenant --> Customer[Customers]

    Provider --> Availability[Provider Availability]
    Provider --> Exceptions[Provider Exceptions]
    Availability --> SlotGenerator[Slot Generator]
    Exceptions --> SlotGenerator
    SlotGenerator --> Slots[Appointment Slots]

    Client --> AvailabilityQuery[GET /api/v1/availability]
    AvailabilityQuery --> Slots
    Slots --> Booking[Create Appointment]
    Booking --> TxBooking[Transaction]
    TxBooking --> Appointment[Appointment]
    TxBooking --> ReserveSlot[Slot status reserved]
    TxBooking --> AppointmentEvent[Appointment Events]

    Appointment --> Cancel[Cancel Appointment]
    Cancel --> TxCancel[Transaction]
    TxCancel --> ReleaseSlot[Slot status available]
    TxCancel --> CancelEvent[Cancelled Event]

    Appointment --> Reschedule[Patch with new slot]
    Reschedule --> TxReschedule[Transaction]
    TxReschedule --> ReleaseOld[Release old slot]
    TxReschedule --> ReserveNew[Reserve new slot]
    TxReschedule --> RescheduleEvent[Rescheduled Event]

    WhatsApp[Evolution / WhatsApp] --> Webhook[POST /api/v1/webhooks/evolution]
    Webhook --> TenantChannel[Resolve tenant channel]
    TenantChannel --> WebhookLog[Store webhook log]
    TenantChannel --> AutoCustomer[Find or create customer]
    AutoCustomer --> Thread[Conversation thread]
    Thread --> Message[Conversation message]
    Message --> State[Conversation state]
```

## Manual Smoke Test

The API was exercised on `http://localhost:8040` with a generated user:

```text
dev_1780267353
```

Verified:

```text
POST /register                         201
POST /login                            200
POST /tokens/renew_access              201
GET  /user/info                        200
PUT  /user/update                      200
POST /user/password_change             200

Tenant CRUD/list                       200/201
Provider CRUD/list                     200/201
Service CRUD/list                      200/201
Customer CRUD/list                     200/201
Provider availability create/list      201/200
Provider exception create/list         201/200
Slot generator                         201
Availability query                     200
Appointment create/list/get/patch/delete 200/201
Conversation message/list/get          200/201
Evolution webhook                      202
```

## Insomnia Collection

Import this file into Insomnia:

```text
docs/insomnia-appointments-mvp.json
```

After login, copy:

- `response.access_token` into `access_token`
- `response.refresh_token` into `refresh_token`

Then create resources in order and copy returned IDs into:

- `tenant_id`
- `provider_id`
- `service_id`
- `customer_id`
- `slot_id`
- `appointment_id`
- `conversation_id`
- `tenant_channel_external_id`

For webhook testing, `tenant_channel_external_id` must match an active row in `tenant_channels.external_id`.
