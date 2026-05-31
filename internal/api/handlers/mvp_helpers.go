package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mousav1/ticket/internal/api/response"
	"github.com/mousav1/ticket/internal/api/validators"
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
