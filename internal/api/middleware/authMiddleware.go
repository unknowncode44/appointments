package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mousav1/ticket/internal/token"
)

const authorizationHeaderKey = "authorization"

// AuthMiddleware validates the Bearer JWT and stores the payload under
// AuthorizationPayloadKey. Only one key is written — the exported constant.
func AuthMiddleware(tokenMaker token.Maker) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get(authorizationHeaderKey)
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing authorization header"})
		}
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing authorization header"})
		}
		tokenStr := authHeader[7:]
		payload, err := tokenMaker.VerifyToken(tokenStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
		}
		// Store only under the exported constant — no duplicate key.
		c.Locals(AuthorizationPayloadKey, payload)
		return c.Next()
	}
}
