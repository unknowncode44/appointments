# TurnoBot API Reference

**Base URL:** `https://turnobot.hvdevs.com`

All dates are ISO 8601 / RFC 3339 in UTC (e.g. `"2025-06-08T14:30:00Z"`).  
All IDs are UUIDs unless stated otherwise.

---

## Authentication

Protected endpoints require a Bearer token in the `Authorization` header:

```
Authorization: Bearer <access_token>
```

Tokens are obtained at login. Three roles exist:

| Role | Description |
|---|---|
| `adminUser` | Full platform access; tenant scope is never enforced |
| `tenantUser` | Scoped to their own tenant; drives the WhatsApp setup flow |
| `user` | Scoped to their own tenant; read/write access to customers, appointments, and availability |

### JWT payload fields

Every decoded access token contains:

| Field | Type | Description |
|---|---|---|
| `id` | uuid | Session token ID (not the user ID) |
| `user_id` | int32 | The authenticated user's numeric database ID |
| `username` | string | |
| `role` | string | `adminUser`, `tenantUser`, or `user` |
| `tenant_id` | uuid \| null | Present for `tenantUser` and `user`; absent for `adminUser` |
| `issued_at` | datetime | |
| `expired_at` | datetime | |

> `user_id` was added in Fase 0. Clients that previously read `decoded.id` as the user ID must switch to `decoded.user_id`.

---

## Rate limiting

| Scope | Limit |
|---|---|
| All routes (global) | 200 requests / minute per IP |
| `/register` and `/login` | 10 requests / minute per IP (brute-force protection) |

Exceeding either limit returns `429 Too Many Requests`:

```json
{ "error": "too many requests, please try again later" }
```

---

## CORS

Allowed origins are controlled by the `ALLOWED_ORIGINS` server configuration variable (comma-separated). Requests from unlisted origins are rejected by the browser. Allowed methods: `GET, POST, HEAD, PUT, DELETE, PATCH, OPTIONS`. Allowed headers: `Origin, Content-Type, Accept, Authorization`.

---

## Error format

All errors return a JSON body:

```json
{ "error": "human-readable message" }
```

Common status codes:

| Code | Meaning |
|---|---|
| 400 | Bad request / validation failure |
| 401 | Missing or invalid token |
| 403 | Role not allowed for this endpoint |
| 404 | Resource not found — also returned when a non-admin accesses a resource belonging to a different tenant (prevents existence leakage) |
| 409 | Conflict (duplicate booking, slot already reserved) |
| 429 | Rate limit exceeded |
| 502 | Upstream service (EVO API) failure |

---

## Pagination

List endpoints accept query parameters `page` (default `1`) and `page_size` (default `20`, max `100`).  
Paginated responses have this envelope:

```json
{
  "data":      [...],
  "page":      1,
  "page_size": 20,
  "total":     42
}
```

---

## 0. Health — no auth required

### GET `/healthz`

Liveness probe. Returns `200 OK` immediately without checking dependencies. Use this to verify the process is running.

**Response `200`** — no body.

---

### GET `/readyz`

Readiness probe. Returns `200 OK` only when the database is reachable. Use this in load-balancer health checks.

**Response `200`** — no body.

**Response `503`** (database unavailable):

```json
{ "error": "db unavailable" }
```

---

## 1. Public — Authentication

### POST `/register`

Create a new user account (role defaults to `user`).

**Request body**

```json
{
  "username":  "string (min 3)",
  "password":  "string (min 6)",
  "full_name": "string"
}
```

**Response `201`**

```json
{
  "username":           "string",
  "full_name":          "string",
  "password_changed_at":"datetime",
  "created_at":         "datetime"
}
```

---

### POST `/login`

**Request body**

```json
{
  "username": "string",
  "password": "string"
}
```

**Response `200`**

```json
{
  "session_id":              "uuid",
  "access_token":            "string",
  "access_token_expires_at": "datetime",
  "refresh_token":           "string",
  "refresh_token_expires_at":"datetime",
  "user": {
    "username":           "string",
    "full_name":          "string",
    "password_changed_at":"datetime",
    "created_at":         "datetime"
  }
}
```

---

### POST `/tokens/renew_access`

Exchange a valid refresh token for a new access token.

**Request body**

```json
{ "refresh_token": "string" }
```

**Response `200`**

```json
{
  "access_token":            "string",
  "access_token_expires_at": "datetime"
}
```

---

## 2. Current User — any authenticated role

### GET `/user/info`

Returns the profile of the authenticated user.

**Response `200`**

```json
{
  "username":           "string",
  "full_name":          "string",
  "password_changed_at":"datetime",
  "created_at":         "datetime"
}
```

---

### PUT `/user/update`

**Request body**

```json
{ "full_name": "string" }
```

**Response `200`** — same shape as `GET /user/info`.

---

### POST `/user/password_change`

**Request body**

```json
{ "password": "string (min 6)" }
```

**Response `200`** — same shape as `GET /user/info`.

---

## 3. Tenants — `adminUser` only

### GET `/api/v1/tenants`

**Query params:** `search`, `active` (`true`/`false`), `page`, `page_size`

**Response `200`** — paginated list of tenant objects.

**Tenant object**

```json
{
  "id":         "uuid",
  "name":       "string",
  "timezone":   "string",
  "active":     true,
  "created_at": "datetime",
  "updated_at": "datetime"
}
```

---

### POST `/api/v1/tenants`

**Request body**

```json
{
  "name":     "string",
  "timezone": "string (IANA, e.g. America/Argentina/Buenos_Aires)"
}
```

**Response `201`** — tenant object.

---

### GET `/api/v1/tenants/:id`

**Response `200`** — tenant object.

---

### PUT `/api/v1/tenants/:id`

**Request body** — same as `POST /tenants`.

**Response `200`** — updated tenant object.

---

### DELETE `/api/v1/tenants/:id`

Deactivates the tenant (sets `active = false`).

**Response `200`** — deactivated tenant object.

---

## 4. Users — `adminUser` only

### GET `/api/v1/users`

**Query params:** `page`, `page_size`

**Response `200`** — array of user objects.

**User object**

```json
{
  "id":         123,
  "username":   "string",
  "full_name":  "string",
  "role":       "adminUser | tenantUser | user",
  "tenant_id":  "uuid | null",
  "created_at": "datetime"
}
```

> Note: `id` is an integer, not a UUID.

---

### POST `/api/v1/users`

Creates a user with an explicit role.

**Request body**

```json
{
  "username":  "string (min 3)",
  "password":  "string (min 6)",
  "full_name": "string",
  "role":      "adminUser | tenantUser | user",
  "tenant_id": "uuid | null"
}
```

**Response `201`** — user object.

---

### GET `/api/v1/users/:id`

**Response `200`** — user object.

---

### PUT `/api/v1/users/:id`

Updates role and tenant assignment.

**Request body**

```json
{
  "full_name": "string",
  "role":      "adminUser | tenantUser | user",
  "tenant_id": "uuid | null"
}
```

**Response `200`** — user object.

---

### DELETE `/api/v1/users/:id`

Hard-deletes the user.

**Response `204`** — no body.

---

### POST `/api/v1/users/:id/tenant`

Links a user to a tenant (sets `tenant_id`).

**Request body**

```json
{ "tenant_id": "uuid" }
```

**Response `200`** — user object.

---

### GET `/api/v1/users/:id/providers`

Lists provider UUIDs linked to this user.

**Response `200`**

```json
["uuid", "uuid"]
```

---

### POST `/api/v1/users/:id/provider`

Links a user to a provider. Roles: `adminUser`, `tenantUser`.

**Request body**

```json
{ "provider_id": "uuid" }
```

**Response `204`** — no body.

---

### DELETE `/api/v1/users/:id/provider`

Unlinks a user from a provider. Roles: `adminUser`, `tenantUser`.

**Query params:** `provider_id=uuid`

**Response `204`** — no body.

---

## 5. Providers — `adminUser`, `tenantUser`

> **Tenant isolation:** `tenantUser` and `user` can only access providers belonging to their own tenant. Accessing a provider from a different tenant returns `404` (not `403`) to prevent cross-tenant existence leakage. On `POST`, the `tenant_id` in the request body is silently overridden with the caller's JWT `tenant_id`.

### GET `/api/v1/providers`

**Query params:** `tenant_id` (required), `search`, `active`, `page`, `page_size`

**Response `200`** — paginated list of provider objects.

**Provider object**

```json
{
  "id":         "uuid",
  "tenant_id":  "uuid",
  "name":       "string",
  "active":     true,
  "created_at": "datetime",
  "updated_at": "datetime"
}
```

---

### POST `/api/v1/providers`

**Request body**

```json
{
  "tenant_id": "uuid",
  "name":      "string"
}
```

**Response `201`** — provider object.

---

### GET `/api/v1/providers/:id`

**Response `200`** — provider object.

---

### PUT `/api/v1/providers/:id`

**Request body**

```json
{ "name": "string" }
```

**Response `200`** — updated provider object.

---

### DELETE `/api/v1/providers/:id`

Deactivates the provider.

**Response `200`** — deactivated provider object.

---

> **Tenant isolation:** the availability and exception endpoints below (`POST`/`GET .../availability`, `POST`/`GET .../exceptions`) verify that the provider in the `:id` path belongs to the caller's tenant. For `tenantUser` and `user`, a provider from a different tenant returns `404` (not `403`). `adminUser` bypasses the check.

### POST `/api/v1/providers/:id/availability`

Adds a weekly availability slot for the provider.

**Request body**

```json
{
  "weekday":    0,
  "start_time": "09:00",
  "end_time":   "17:00"
}
```

> `weekday`: `0` = Sunday, `1` = Monday, …, `6` = Saturday.  
> `start_time` / `end_time`: `HH:MM` 24-hour strings.

**Response `201`**

```json
{
  "id":          "uuid",
  "provider_id": "uuid",
  "weekday":     1,
  "start_time":  "09:00",
  "end_time":    "17:00"
}
```

---

### GET `/api/v1/providers/:id/availability`

**Response `200`**

```json
{
  "data": [ /* availability objects */ ]
}
```

---

### POST `/api/v1/providers/:id/exceptions`

Blocks out a time range (vacation, holiday, etc.).

**Request body**

```json
{
  "start_at": "datetime",
  "end_at":   "datetime",
  "reason":   "string | null"
}
```

**Response `201`**

```json
{
  "id":          "uuid",
  "provider_id": "uuid",
  "start_at":    "datetime",
  "end_at":      "datetime",
  "reason":      "string | null",
  "created_at":  "datetime"
}
```

---

### GET `/api/v1/providers/:id/exceptions`

**Response `200`**

```json
{
  "data": [ /* exception objects */ ]
}
```

---

## 6. Services — `adminUser`, `tenantUser`

### GET `/api/v1/services`

**Query params:** `tenant_id` (required), `search`, `active`, `page`, `page_size`

**Response `200`** — paginated list of service objects.

**Service object**

```json
{
  "id":               "uuid",
  "tenant_id":        "uuid",
  "name":             "string",
  "duration_minutes": 30,
  "active":           true,
  "created_at":       "datetime",
  "updated_at":       "datetime"
}
```

---

### POST `/api/v1/services`

**Request body**

```json
{
  "tenant_id":        "uuid",
  "name":             "string",
  "duration_minutes": 30
}
```

**Response `201`** — service object.

---

### GET `/api/v1/services/:id`

**Response `200`** — service object.

---

### PUT `/api/v1/services/:id`

**Request body**

```json
{
  "name":             "string",
  "duration_minutes": 30
}
```

**Response `200`** — updated service object.

---

### DELETE `/api/v1/services/:id`

Deactivates the service.

**Response `200`** — deactivated service object.

---

## 7. Tenant Channels — `adminUser`, `tenantUser`

Low-level channel management. For WhatsApp, prefer the `/whatsapp` proxy endpoints below.

**Channel object**

```json
{
  "id":           "uuid",
  "tenant_id":    "uuid",
  "channel_type": "whatsapp",
  "external_id":  "string",
  "external_key": "string | null",
  "active":       true,
  "created_at":   "datetime"
}
```

### GET `/api/v1/tenant-channels`

**Query params:** `tenant_id` (required), `channel_type`, `active`, `page`, `page_size`

**Response `200`** — paginated list of channel objects.

---

### POST `/api/v1/tenant-channels`

**Request body**

```json
{
  "tenant_id":    "uuid",
  "channel_type": "string",
  "external_id":  "string",
  "external_key": "string | null"
}
```

**Response `201`** — channel object.

---

### GET `/api/v1/tenant-channels/:id`

**Response `200`** — channel object.

---

### PUT `/api/v1/tenant-channels/:id`

**Request body**

```json
{
  "channel_type": "string",
  "external_id":  "string",
  "external_key": "string | null",
  "active":       true
}
```

**Response `200`** — updated channel object.

---

### DELETE `/api/v1/tenant-channels/:id`

Sets `active = false`.

**Response `200`** — deactivated channel object.

---

## 8. Customers — all authenticated roles

### GET `/api/v1/customers`

**Query params:** `tenant_id` (required), `search`, `page`, `page_size`

**Response `200`** — paginated list of customer objects.

**Customer object**

```json
{
  "id":         "uuid",
  "tenant_id":  "uuid",
  "first_name": "string | null",
  "last_name":  "string | null",
  "notes":      "string | null",
  "created_at": "datetime",
  "updated_at": "datetime"
}
```

---

### POST `/api/v1/customers`

**Request body**

```json
{
  "tenant_id":  "uuid",
  "first_name": "string | null",
  "last_name":  "string | null",
  "notes":      "string | null"
}
```

**Response `201`** — customer object.

---

### GET `/api/v1/customers/:id`

**Response `200`** — customer object.

---

### PUT `/api/v1/customers/:id`

**Request body**

```json
{
  "first_name": "string | null",
  "last_name":  "string | null",
  "notes":      "string | null"
}
```

**Response `200`** — updated customer object.

---

## 9. Scheduling — all authenticated roles

### POST `/api/v1/slot-generator`

Pre-generates appointment slots for a provider based on their weekly availability.

**Request body**

```json
{
  "tenant_id":      "uuid",
  "provider_id":    "uuid",
  "number_of_weeks": 4,
  "slot_minutes":   30,
  "timezone":       "America/Argentina/Buenos_Aires"
}
```

> `number_of_weeks`: 1–12.  
> `slot_minutes`: optional, defaults to the service duration.  
> `timezone`: optional, defaults to UTC.

> **Tenant isolation:** for `tenantUser` and `user`, `tenant_id` is forced from the JWT (any value in the body is ignored), and the target `provider_id` must belong to that tenant — otherwise `404` is returned. `adminUser` bypasses the check.

**Response `201`**

```json
{ "generated": 56 }
```

---

### GET `/api/v1/availability`

Returns open appointment slots.

**Query params:**

| Param | Type | Required | Description |
|---|---|---|---|
| `tenant_id` | uuid | yes | |
| `date` | `YYYY-MM-DD` | yes | Day to query |
| `provider_id` | uuid | no | Filter by provider |
| `service_id` | uuid | no | Filter by service compatibility |
| `page` | int | no | |
| `page_size` | int | no | |

**Response `200`**

```json
{
  "data": [
    {
      "id":             "uuid",
      "tenant_id":      "uuid",
      "provider_id":    "uuid",
      "start_at":       "datetime",
      "end_at":         "datetime",
      "status":         "available",
      "appointment_id": null,
      "created_at":     "datetime"
    }
  ]
}
```

---

## 10. Appointments — all authenticated roles

**Appointment object**

```json
{
  "id":          "uuid",
  "tenant_id":   "uuid",
  "customer_id": "uuid",
  "provider_id": "uuid",
  "service_id":  "uuid",
  "slot_id":     "uuid | null",
  "start_at":    "datetime",
  "end_at":      "datetime",
  "status":      "confirmed | cancelled | completed | no_show",
  "notes":       "string | null",
  "created_at":  "datetime",
  "updated_at":  "datetime"
}
```

Valid status values:

| Status | Meaning |
|---|---|
| `confirmed` | Slot reserved; appointment is active (set on creation) |
| `cancelled` | Appointment cancelled; slot released back to available |
| `completed` | Appointment was attended |
| `no_show` | Customer did not show up |

> A DB `CHECK` constraint enforces these four values. Any other status sent via `PATCH` will be rejected with `400 Bad Request`.
```

### POST `/api/v1/appointments`

Reserves a slot and creates an appointment.

**Request body**

```json
{
  "tenant_id":   "uuid",
  "customer_id": "uuid",
  "service_id":  "uuid",
  "slot_id":     "uuid",
  "notes":       "string | null"
}
```

**Response `201`** — appointment object.

---

### GET `/api/v1/appointments`

**Query params:** `tenant_id` (required), `provider_id`, `customer_id`, `status`, `page`, `page_size`

**Response `200`** — paginated list of appointment objects.

---

### GET `/api/v1/appointments/:id`

**Response `200`** — appointment object.

---

### PATCH `/api/v1/appointments/:id`

Updates one or more fields. All fields are optional. Sending `"status": "cancelled"` is equivalent to calling `DELETE` — it triggers the full cancellation flow (slot released, event recorded).

**Request body** (all fields optional)

```json
{
  "status":     "confirmed | cancelled | completed | no_show",
  "slot_id":    "uuid | null",
  "service_id": "uuid | null",
  "notes":      "string | null"
}
```

> Non-admin callers can only patch appointments that belong to their own tenant. Cross-tenant attempts return `404`.

**Response `200`** — updated appointment object.

---

### DELETE `/api/v1/appointments/:id`

Cancels the appointment (sets `status = "cancelled"`). Does not hard-delete.

**Response `200`** — cancelled appointment object.

---

## 11. Conversations — `adminUser`, `tenantUser`

### GET `/api/v1/conversations`

**Query params:** `tenant_id` (required), `page`, `page_size`

**Response `200`**

```json
{
  "data": [
    {
      "id":          "uuid",
      "tenant_id":   "uuid",
      "customer_id": "uuid",
      "created_at":  "datetime",
      "messages":    []
    }
  ]
}
```

---

### GET `/api/v1/conversations/:id`

Returns a thread with its messages.

**Response `200`**

```json
{
  "id":          "uuid",
  "tenant_id":   "uuid",
  "customer_id": "uuid",
  "created_at":  "datetime",
  "messages": [
    {
      "id":        "uuid",
      "thread_id": "uuid",
      "direction": "in | out",
      "message":   "string",
      "metadata":  {},
      "created_at":"datetime"
    }
  ]
}
```

---

### POST `/api/v1/conversations/message`

Stores a message in a conversation thread (creates the thread if needed).

**Request body**

```json
{
  "tenant_id":   "uuid",
  "customer_id": "uuid",
  "direction":   "in | out",
  "message":     "string",
  "metadata":    {}
}
```

> **Tenant isolation:** for `tenantUser`, `tenant_id` is forced from the JWT (any value in the body is ignored), and the supplied `customer_id` must belong to that tenant — otherwise `404` is returned. `adminUser` bypasses the check.

**Response `201`** — message object.

---

### POST `/api/v1/inbound-messages`

Processes an inbound message identified by channel external ID and customer external ID. Creates the customer if they don't exist.

**Request body**

```json
{
  "tenant_channel_external_id": "string",
  "external_customer_id":       "string",
  "message":                    "string",
  "source":                     "string",
  "metadata":                   {}
}
```

**Response `201`** — message object.

---

## 12. WhatsApp Instance — `tenantUser` only

All endpoints are scoped to the authenticated `tenantUser` and require no body parameters identifying the tenant — the JWT carries that context.

Each tenant has exactly one WhatsApp instance. The instance is named `turnobot_<tenant_uuid>` internally; the frontend never needs to know this.

**Typical frontend flow:**

```
1. GET  /status       → if 404: show "Connect" button
2. POST /instance     → receive instance_key; store it client-side for QR polling
3. GET  /qr           → display QR (poll every 25 s)
4. GET  /status       → poll every 3 s until state = "open"
5. Connected — show status card
6. DELETE /logout     → disconnect (instance kept alive for reconnection via QR)
7. DELETE /instance   → hard reset — full teardown; required before calling POST /instance again
```

---

### POST `/api/v1/whatsapp/instance`

Creates the EVO API WhatsApp instance for this tenant.

> Returns `409` if an instance already exists. The user must call `DELETE /instance` to reset before retrying.

**Request body:** none

**Response `201`**

```json
{
  "instance_name": "turnobot_<uuid>",
  "instance_key":  "string",
  "status":        "string"
}
```

> `instance_key` is the per-instance API key. It is safe to store on the client for direct EVO API reads (QR image, etc.) if needed, but all mutations must go through this proxy.

---

### GET `/api/v1/whatsapp/instance/status`

Returns the current EVO API connection state.

**Response `200`**

```json
{
  "instance_name": "turnobot_<uuid>",
  "state":         "open | connecting | close"
}
```

| State | Meaning |
|---|---|
| `open` | Paired and connected |
| `connecting` | Waiting for QR scan |
| `close` | Disconnected |

---

### GET `/api/v1/whatsapp/instance/qr`

Returns the QR code to pair the device. Only meaningful when `state` is `connecting` or `close`.

**Response `200`**

```json
{
  "qr_code": "string",
  "base64":  "data:image/png;base64,..."
}
```

> Render `base64` directly in an `<img src="...">` tag.

---

### DELETE `/api/v1/whatsapp/instance/logout`

Logs out of WhatsApp. The EVO instance is kept alive so the user can reconnect by scanning a new QR without going through `POST /instance` again.

**Response `200`**

```json
{ "message": "disconnected" }
```

---

### DELETE `/api/v1/whatsapp/instance`

**Hard reset.** Deletes the EVO API instance entirely and removes the local channel record. After this call the user must go through `POST /instance` again to reconnect.

**Response `200`**

```json
{ "message": "instance deleted" }
```

---

## 13. Webhooks — public (no auth)

### POST `/api/v1/webhooks/evolution`

Receives incoming events from the Evolution API (WhatsApp). This endpoint is called by EVO API, not by the frontend.

**Request body** — Evolution API v2 payload (handled internally).

**Response `202`** — no body.
