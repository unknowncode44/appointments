package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mousav1/ticket/internal/token"
)

const (
	authorizationHeaderKey  = "authorization"
	AuthorizationPayloadKey = "authorization_payload"
)

// AuthMiddleware creates a gin middleware for authorization
func AuthMiddleware(tokenMaker token.Maker) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract the Authorization header
		authHeader := c.Get(authorizationHeaderKey)
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing authorization header"})
		}

		// Verify if the header is not empty and start with "Bearer "
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing authorization header"})
		}

		tokenStr := authHeader[7:]

		// Verify the token
		payload, err := tokenMaker.VerifyToken(tokenStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
		}

		// Attach user information to the request context
		c.Locals(AuthorizationPayloadKey, payload)
		c.Locals("authorizationPayloadKey", payload)

		// Proceed to the next handler
		return c.Next()
	}
}
