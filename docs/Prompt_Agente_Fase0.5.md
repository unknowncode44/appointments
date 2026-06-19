# Prompt para el agente (Opus · Claude extension for VS Code)

> Copiá todo lo que sigue (a partir de "You are working in...") como mensaje inicial al agente, con el repo `appointments` abierto en VS Code.

---

You are working in the Go backend repository `github.com/unknowncode44/appointments` (Fiber v2, sqlc, pgx, golang-migrate). Branch off the latest `fix/fase0-security-hardening`. Your task is to close two residual multi-tenant isolation gaps (Phase 0.5, points 1 and 2). These are the same vulnerability class already fixed elsewhere in this branch — your job is to extend that exact pattern to the endpoints it missed.

## Context: the established pattern (follow it exactly)
The codebase already enforces tenant isolation in handlers + services like this:
- A helper in `internal/api/handlers/mvp_helpers.go`:
  ```go
  // nil for adminUser (no restriction); the caller's TenantID for tenantUser/user; 403 if a non-admin has no tenant.
  func callerTenantScope(c *fiber.Ctx) (*uuid.UUID, error)
  ```
- Reads/updates/deletes by id fetch the resource and, if `scope != nil && resource.TenantID != *scope`, return `response.ErrNotFound` (404 — never 403, to avoid leaking existence).
- Creates force `req.TenantID = *scope` when `scope != nil`.
- For appointments the check runs **inside the `ExecTx`** so it is race-safe.

Study `internal/services/admin.go`, `internal/services/workflows.go`, `internal/api/handlers/admin_mvp.go` and `internal/api/handlers/workflows_mvp.go` to copy the style precisely (signatures, error handling, `response.ErrNotFound`).

## Scope (do ONLY these two points)

### Point 1 — Tenant-scope the scheduling endpoints
These currently operate without verifying the target belongs to the caller's tenant:
- `POST /providers/:id/availability`  → `SchedulingMVPHandler.CreateAvailability`
- `POST /providers/:id/exceptions`    → `SchedulingMVPHandler.CreateException`
- `POST /slot-generator`              → `SchedulingMVPHandler.GenerateSlots`

Required behavior:
- For `CreateAvailability` and `CreateException`: resolve `callerTenantScope(c)`; if `scope != nil`, verify that the provider in the `:id` path belongs to `*scope`. If not, return **404** (`response.ErrNotFound`). adminUser (`scope == nil`) bypasses.
- For `GenerateSlots`: resolve `callerTenantScope(c)`; if `scope != nil`, **force** the request's tenant to `*scope` (ignore any `tenant_id` in the body), and verify the target provider belongs to `*scope` (404 otherwise).
- Inspect `internal/services/scheduling.go`, its repository interface, and the relevant DTOs (`dto.AvailabilityRequest`, `dto.ExceptionRequest`, `dto.SlotGeneratorRequest`) to wire this in.
- Provider-ownership lookup: if the scheduling service/repository has no way to read a provider's tenant, add a minimal query for it (e.g. reuse `GetProvider` or add a focused `GetProviderTenant`) following the existing sqlc query style in `internal/db/query/mvp.sql`. If you add or change any SQL, regenerate with `sqlc generate` and commit the generated files.
- Read availability/exceptions listing endpoints (`GET .../availability`, `.../exceptions`) — if they are also unscoped, apply the same provider-ownership check for consistency. Keep changes minimal and within the scheduling domain.

### Point 2 — Tenant-scope `POST /conversations/message`
- `ConversationMVPHandler.Message` → `conversationService.StoreMessage`.
- Resolve `callerTenantScope(c)`; if `scope != nil`, force `req.TenantID = *scope` (ignore body tenant). Additionally, verify the `customer_id` in the request belongs to `*scope`; return **404** if not. Mirror the create/ownership pattern used elsewhere.

## Tests (required)
Add table-style tests mirroring `tests/tenant_isolation_test.go` (stubbed repositories, no DB) for the new checks:
- Scheduling: cross-tenant provider on `CreateAvailability` / `CreateException` / `GenerateSlots` → `ErrNotFound`; own-tenant → success; `nil` scope (admin) → bypass.
- Conversations: `StoreMessage` cross-tenant customer → `ErrNotFound`; own-tenant → success.
Extend the existing repo stubs if needed; do not require a live Postgres for these unit tests.

## Constraints / out of scope
- Do NOT implement Row-Level Security, do NOT touch the frontend repo, do NOT change migrations except a new SQL query if strictly needed (no schema changes).
- Do NOT merge, push tags, or alter CI. Keep the diff focused on points 1 and 2.
- Preserve existing public behavior for adminUser (full access) and the existing booking/isolation logic.
- Match repo conventions: `gofmt`/`goimports`, zerolog, existing error helpers in `internal/api/response`.

## Definition of done
1. `go build ./...` passes.
2. `go vet ./...` clean.
3. `go test ./...` passes (the new isolation unit tests run without a DB; booking tests need Postgres — note this in the PR if you can't run them locally).
4. All three scheduling endpoints and `POST /conversations/message` reject cross-tenant access with **404** for non-admin callers, verified by the new tests.
5. Update `CHANGELOG.md` with a short entry, and update `API_REFERENCE.md` if any response/validation behavior changed.
6. Produce a concise PR description (English) summarizing the two fixes, the files touched, and how to test. Title suggestion: `fix(fase0.5): tenant-scope scheduling and conversation-message endpoints`.

Begin by reading the four reference files listed under "Context" and the scheduling service/DTOs, then propose a short plan before editing.
