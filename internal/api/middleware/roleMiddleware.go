package middleware

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mousav1/ticket/internal/token"
)

const AuthorizationPayloadKey = "authorization_payload"

// RoleMiddleware is kept as a no-op pass-through so existing route wiring
// compiles unchanged. Per-route authorization is enforced by RequireRole()
// applied inline in web.go.
func RoleMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Next()
	}
}

// RequireRole rejects requests whose JWT payload role is not in allowedRoles.
func RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		payload, ok := c.Locals(AuthorizationPayloadKey).(*token.Payload)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid authorization payload"})
		}
		for _, role := range allowedRoles {
			if payload.Role == role {
				return c.Next()
			}
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": fmt.Sprintf("role %s is not allowed to access this resource", payload.Role),
		})
	}
}

// RequireTenant enforces that tenantUser requests can only operate on their
// own tenant. It reads the tenant_id from:
//  1. The query param named by queryParam (e.g. "tenant_id"), OR
//  2. Falls back to the JSON body field if the query param is absent.
//
// adminUser bypasses the check. Regular "user" role bypasses too (they are
// scoped at the appointment/customer level, not at tenant level).
func RequireTenant(queryParam string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		payload, ok := c.Locals(AuthorizationPayloadKey).(*token.Payload)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid authorization payload"})
		}

		// Only enforce for tenantUser; admin sees everything.
		if payload.Role != "tenantUser" {
			return c.Next()
		}

		if payload.TenantID == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "tenant user has no tenant assigned"})
		}

		// Try query param first, then fall back to body field of same name.
		raw := c.Query(queryParam)
		if raw == "" {
			// Try to read from the parsed body map (works for JSON bodies).
			var body map[string]interface{}
			if err := c.BodyParser(&body); err == nil {
				if v, ok := body[queryParam].(string); ok {
					raw = v
				}
			}
		}

		if raw == "" {
			// No tenant_id provided — let the downstream handler reject it.
			return c.Next()
		}

		requestedTenantID, err := uuid.Parse(raw)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid tenant_id"})
		}

		if requestedTenantID != *payload.TenantID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden: tenant mismatch"})
		}

		return c.Next()
	}
}

// ExtractUserFromContext returns the JWT payload stored by AuthMiddleware.
func ExtractUserFromContext(c *fiber.Ctx) (*token.Payload, error) {
	payload, ok := c.Locals(AuthorizationPayloadKey).(*token.Payload)
	if !ok {
		return nil, errors.New("invalid authorization payload")
	}
	return payload, nil
}
