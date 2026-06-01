package middleware

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/mousav1/ticket/internal/token"
)

type Permission struct {
	Method string
	Path   string
	Roles  []string
}

var RoutePermissions = []Permission{
	// Admin only routes
	{Method: "GET", Path: "/api/v1/tenants", Roles: []string{"adminUser"}},
	{Method: "POST", Path: "/api/v1/tenants", Roles: []string{"adminUser"}},
	{Method: "GET", Path: "/api/v1/tenants/:id", Roles: []string{"adminUser"}},
	{Method: "PUT", Path: "/api/v1/tenants/:id", Roles: []string{"adminUser"}},
	{Method: "DELETE", Path: "/api/v1/tenants/:id", Roles: []string{"adminUser"}},
	{Method: "GET", Path: "/api/v1/users", Roles: []string{"adminUser"}},
	{Method: "POST", Path: "/api/v1/users", Roles: []string{"adminUser"}},
	{Method: "GET", Path: "/api/v1/users/:id", Roles: []string{"adminUser"}},
	{Method: "PUT", Path: "/api/v1/users/:id", Roles: []string{"adminUser"}},
	{Method: "DELETE", Path: "/api/v1/users/:id", Roles: []string{"adminUser"}},
	{Method: "POST", Path: "/api/v1/users/:id/tenant", Roles: []string{"adminUser"}},
	{Method: "POST", Path: "/api/v1/users/:id/provider", Roles: []string{"adminUser", "tenantUser"}},

	// TenantUser and Admin routes
	{Method: "GET", Path: "/api/v1/providers", Roles: []string{"adminUser", "tenantUser"}},
	{Method: "POST", Path: "/api/v1/providers", Roles: []string{"adminUser", "tenantUser"}},
	{Method: "GET", Path: "/api/v1/providers/:id", Roles: []string{"adminUser", "tenantUser"}},
	{Method: "PUT", Path: "/api/v1/providers/:id", Roles: []string{"adminUser", "tenantUser"}},
	{Method: "DELETE", Path: "/api/v1/providers/:id", Roles: []string{"adminUser", "tenantUser"}},
	{Method: "GET", Path: "/api/v1/services", Roles: []string{"adminUser", "tenantUser"}},
	{Method: "POST", Path: "/api/v1/services", Roles: []string{"adminUser", "tenantUser"}},
	{Method: "GET", Path: "/api/v1/services/:id", Roles: []string{"adminUser", "tenantUser"}},
	{Method: "PUT", Path: "/api/v1/services/:id", Roles: []string{"adminUser", "tenantUser"}},
	{Method: "DELETE", Path: "/api/v1/services/:id", Roles: []string{"adminUser", "tenantUser"}},

	// All authenticated users
	{Method: "GET", Path: "/api/v1/customers", Roles: []string{"adminUser", "tenantUser", "user"}},
	{Method: "POST", Path: "/api/v1/customers", Roles: []string{"adminUser", "tenantUser", "user"}},
	{Method: "GET", Path: "/api/v1/customers/:id", Roles: []string{"adminUser", "tenantUser", "user"}},
	{Method: "PUT", Path: "/api/v1/customers/:id", Roles: []string{"adminUser", "tenantUser", "user"}},
	{Method: "POST", Path: "/api/v1/providers/:id/availability", Roles: []string{"adminUser", "tenantUser", "user"}},
	{Method: "GET", Path: "/api/v1/providers/:id/availability", Roles: []string{"adminUser", "tenantUser", "user"}},
	{Method: "POST", Path: "/api/v1/providers/:id/exceptions", Roles: []string{"adminUser", "tenantUser", "user"}},
	{Method: "GET", Path: "/api/v1/providers/:id/exceptions", Roles: []string{"adminUser", "tenantUser", "user"}},
	{Method: "POST", Path: "/api/v1/slot-generator", Roles: []string{"adminUser", "tenantUser", "user"}},
	{Method: "GET", Path: "/api/v1/availability", Roles: []string{"adminUser", "tenantUser", "user"}},
	{Method: "POST", Path: "/api/v1/appointments", Roles: []string{"adminUser", "tenantUser", "user"}},
	{Method: "GET", Path: "/api/v1/appointments", Roles: []string{"adminUser", "tenantUser", "user"}},
	{Method: "GET", Path: "/api/v1/appointments/:id", Roles: []string{"adminUser", "tenantUser", "user"}},
	{Method: "PATCH", Path: "/api/v1/appointments/:id", Roles: []string{"adminUser", "tenantUser", "user"}},
	{Method: "DELETE", Path: "/api/v1/appointments/:id", Roles: []string{"adminUser", "tenantUser", "user"}},
}

func RoleMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		payload, ok := c.Locals(AuthorizationPayloadKey).(*token.Payload)

		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Role Middleware: invalid authorization payload"})
		}

		method := c.Method()
		path := c.Path()

		// Check if route requires authorization
		for _, perm := range RoutePermissions {
			if matchRoute(perm.Path, path) && perm.Method == method {
				// Check if user has required role
				allowed := false
				for _, role := range perm.Roles {
					if role == payload.Role {
						allowed = true
						break
					}
				}

				if !allowed {
					return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
						"error": fmt.Sprintf("role %s is not allowed to access this resource", payload.Role),
					})
				}
			}
		}

		return c.Next()
	}
}

func matchRoute(pattern, path string) bool {
	if pattern == path {
		return true
	}

	// Simple param matching for patterns like /api/v1/providers/:id
	patternParts := splitPath(pattern)
	pathParts := splitPath(path)

	if len(patternParts) != len(pathParts) {
		return false
	}

	for i, p := range patternParts {
		if p != pathParts[i] && !isParam(p) {
			return false
		}
	}

	return true
}

func splitPath(path string) []string {
	if path == "" {
		return []string{}
	}
	parts := []string{}
	current := ""
	for _, ch := range path {
		if ch == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func isParam(part string) bool {
	return len(part) > 0 && part[0] == ':'
}

func RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		payload, ok := c.Locals(AuthorizationPayloadKey).(*token.Payload)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Require Role: invalid authorization payload"})
		}

		for _, role := range allowedRoles {
			if payload.Role == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": fmt.Sprintf("role %s is not allowed", payload.Role),
		})
	}
}

func ExtractUserFromContext(c *fiber.Ctx) (*token.Payload, error) {
	payload, ok := c.Locals(AuthorizationPayloadKey).(*token.Payload)
	if !ok {
		return nil, errors.New("ExtractUserFromContext: invalid authorization payload")
	}
	return payload, nil
}
