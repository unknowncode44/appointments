package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/unknowncode44/appointments/internal/api/middleware"
	"github.com/unknowncode44/appointments/internal/api/response"
	"github.com/unknowncode44/appointments/internal/api/validators"
)

func parseID(c *fiber.Ctx, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params(name))
	if err != nil {
		return uuid.Nil, response.ErrInvalidInput
	}
	return id, nil
}

func queryUUID(c *fiber.Ctx, name string) (*uuid.UUID, error) {
	raw := c.Query(name)
	if raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, response.ErrInvalidInput
	}
	return &id, nil
}

func queryBool(c *fiber.Ctx, name string) (*bool, error) {
	raw := c.Query(name)
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return nil, response.ErrInvalidInput
	}
	return &value, nil
}

func bindAndValidate(c *fiber.Ctx, req any) error {
	if err := c.BodyParser(req); err != nil {
		return err
	}
	return validators.Struct(req)
}

// callerTenantScope returns nil for adminUser (no restriction) or the caller's
// TenantID from the JWT for tenantUser/user roles.
func callerTenantScope(c *fiber.Ctx) (*uuid.UUID, error) {
	payload, err := middleware.ExtractUserFromContext(c)
	if err != nil {
		return nil, fiber.ErrUnauthorized
	}
	if payload.Role == "adminUser" {
		return nil, nil
	}
	if payload.TenantID == nil {
		return nil, response.ErrForbidden
	}
	return payload.TenantID, nil
}
