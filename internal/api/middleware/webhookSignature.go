package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// webhookSignatureHeader carries the hex-encoded HMAC-SHA256 of the raw request
// body. An optional "sha256=" prefix is accepted for tooling compatibility.
const webhookSignatureHeader = "X-Webhook-Signature"

// VerifyWebhookSignature guards a public webhook with an HMAC-SHA256 signature of
// the raw body keyed by a shared secret. It is fail-closed: if no secret is
// configured server-side the request is rejected, so a misconfiguration can never
// silently disable verification.
func VerifyWebhookSignature(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if secret == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "webhook signature verification not configured"})
		}

		provided := strings.TrimSpace(c.Get(webhookSignatureHeader))
		provided = strings.TrimPrefix(provided, "sha256=")
		if provided == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing webhook signature"})
		}

		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(c.Body())
		expected := mac.Sum(nil)

		// Constant-time compare; decode the provided hex so equal-length raw bytes
		// are compared rather than case-sensitive hex strings.
		providedBytes, err := hex.DecodeString(provided)
		if err != nil || !hmac.Equal(providedBytes, expected) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid webhook signature"})
		}
		return c.Next()
	}
}
