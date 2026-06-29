package test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"github.com/unknowncode44/appointments/internal/api/middleware"
)

func sign(secret, body string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyWebhookSignature(t *testing.T) {
	const secret = "test-webhook-secret"
	const body = `{"instance":"turnobot_x","data":{"key":{"id":"ABC"}}}`

	testCases := []struct {
		name       string
		secret     string
		signature  string
		wantStatus int
	}{
		{name: "ValidSignature", secret: secret, signature: sign(secret, body), wantStatus: fiber.StatusOK},
		{name: "ValidWithPrefix", secret: secret, signature: "sha256=" + sign(secret, body), wantStatus: fiber.StatusOK},
		{name: "InvalidSignature", secret: secret, signature: sign("wrong-secret", body), wantStatus: fiber.StatusUnauthorized},
		{name: "MissingSignature", secret: secret, signature: "", wantStatus: fiber.StatusUnauthorized},
		{name: "MalformedHex", secret: secret, signature: "not-hex", wantStatus: fiber.StatusUnauthorized},
		{name: "NoSecretConfiguredFailsClosed", secret: "", signature: sign(secret, body), wantStatus: fiber.StatusUnauthorized},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Post("/webhook",
				middleware.VerifyWebhookSignature(tc.secret),
				func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })

			req := httptest.NewRequest(fiber.MethodPost, "/webhook", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			if tc.signature != "" {
				req.Header.Set("X-Webhook-Signature", tc.signature)
			}

			resp, err := app.Test(req)
			require.NoError(t, err)
			require.Equal(t, tc.wantStatus, resp.StatusCode)
		})
	}
}
