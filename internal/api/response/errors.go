package response

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrInvalidInput = errors.New("invalid input")
)

func Error(c *fiber.Ctx, err error) error {
	status := fiber.StatusInternalServerError
	msg := "internal server error"

	switch {
	case errors.Is(err, ErrNotFound), errors.Is(err, pgx.ErrNoRows):
		status = fiber.StatusNotFound
		msg = "not found"
	case errors.Is(err, ErrConflict):
		status = fiber.StatusConflict
		msg = err.Error()
	case errors.Is(err, ErrInvalidInput):
		status = fiber.StatusBadRequest
		msg = err.Error()
	}

	return c.Status(status).JSON(fiber.Map{"error": msg})
}

func BadRequest(c *fiber.Ctx, err error) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
}
