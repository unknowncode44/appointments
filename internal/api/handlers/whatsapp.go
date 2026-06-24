package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/unknowncode44/appointments/internal/api/middleware"
	db "github.com/unknowncode44/appointments/internal/db/sqlc"
	"github.com/unknowncode44/appointments/internal/util"
)

type WhatsappHandler struct {
	Store      *db.Store
	Config     util.Config
	httpClient *http.Client
}

func NewWhatsappHandler(store *db.Store, config util.Config) *WhatsappHandler {
	return &WhatsappHandler{
		Store:      store,
		Config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func evoInstanceName(tenantID uuid.UUID) string {
	return "turnobot_" + tenantID.String()
}

// evoCall makes a JSON request to the EVO API and returns (responseBody, statusCode, error).
func (h *WhatsappHandler) evoCall(ctx context.Context, method, path, apiKey string, body any) ([]byte, int, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		bodyReader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, h.Config.EvoAPIURL+path, bodyReader)
	if err != nil {
		return nil, 0, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("apikey", apiKey)
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	return data, resp.StatusCode, err
}

// requireTenantID extracts the tenant UUID from the JWT. Writes the error response and
// returns false if the tenant is missing.
func (h *WhatsappHandler) requireTenantID(c *fiber.Ctx) (uuid.UUID, bool) {
	payload, err := middleware.ExtractUserFromContext(c)
	if err != nil || payload.TenantID == nil {
		_ = c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "tenant not assigned to this account"})
		return uuid.Nil, false
	}
	return *payload.TenantID, true
}

// loadChannel retrieves the tenant's WhatsApp channel. Writes the error response and
// returns false if not found or on DB error.
func (h *WhatsappHandler) loadChannel(c *fiber.Ctx) (uuid.UUID, db.TenantChannel, bool) {
	tenantID, ok := h.requireTenantID(c)
	if !ok {
		return uuid.Nil, db.TenantChannel{}, false
	}
	ch, err := h.Store.GetWhatsappChannelByTenant(c.Context(), tenantID)
	if errors.Is(err, pgx.ErrNoRows) {
		_ = c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "no whatsapp instance found — create one first"})
		return uuid.Nil, db.TenantChannel{}, false
	}
	if err != nil {
		_ = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not fetch channel"})
		return uuid.Nil, db.TenantChannel{}, false
	}
	return tenantID, ch, true
}

// POST /api/v1/whatsapp/instance
// Creates an EVO API WhatsApp instance for the authenticated tenantUser.
func (h *WhatsappHandler) CreateInstance(c *fiber.Ctx) error {
	tenantID, ok := h.requireTenantID(c)
	if !ok {
		return nil
	}

	// 409 if any channel record already exists (active or not).
	// User must DELETE /instance to fully reset before creating again.
	_, err := h.Store.GetWhatsappChannelByTenant(c.Context(), tenantID)
	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "whatsapp instance already exists for this tenant"})
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not check existing channel"})
	}

	name := evoInstanceName(tenantID)
	reqBody := map[string]string{
		"instanceName": name,
		"integration":  "WHATSAPP-BAILEYS",
	}
	data, status, err := h.evoCall(c.Context(), http.MethodPost, "/instance/create", h.Config.EvoAPIKey, reqBody)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "evolution api unreachable"})
	}
	if status >= 300 {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": string(data)})
	}

	var evoResp struct {
		Instance struct {
			InstanceName string `json:"instanceName"`
			Status       string `json:"status"`
		} `json:"instance"`
		Hash string `json:"hash"`
	}
	if err := json.Unmarshal(data, &evoResp); err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "invalid evolution api response"})
	}

	_, err = h.Store.CreateTenantChannel(c.Context(), db.CreateTenantChannelParams{
		TenantID:    tenantID,
		ChannelType: "whatsapp",
		ExternalID:  name,
		ExternalKey: pgtype.Text{String: evoResp.Hash, Valid: evoResp.Hash != ""},
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not save channel"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"instance_name": evoResp.Instance.InstanceName,
		"instance_key":  evoResp.Hash,
		"status":        evoResp.Instance.Status,
	})
}

// GET /api/v1/whatsapp/instance/status
// Returns the EVO API connection state for the tenant's WhatsApp instance.
func (h *WhatsappHandler) GetStatus(c *fiber.Ctx) error {
	tenantID, ch, ok := h.loadChannel(c)
	if !ok {
		return nil
	}

	name := evoInstanceName(tenantID)
	data, status, err := h.evoCall(c.Context(), http.MethodGet, "/instance/connectionState/"+name, ch.ExternalKey.String, nil)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "evolution api unreachable"})
	}
	if status >= 300 {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": string(data)})
	}

	var evoResp struct {
		Instance struct {
			InstanceName string `json:"instanceName"`
			State        string `json:"state"`
		} `json:"instance"`
	}
	if err := json.Unmarshal(data, &evoResp); err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "invalid evolution api response"})
	}

	return c.JSON(fiber.Map{
		"state":         evoResp.Instance.State,
		"instance_name": evoResp.Instance.InstanceName,
	})
}

// GET /api/v1/whatsapp/instance/qr
// Returns the QR code for pairing the WhatsApp instance.
func (h *WhatsappHandler) GetQR(c *fiber.Ctx) error {
	tenantID, ch, ok := h.loadChannel(c)
	if !ok {
		return nil
	}

	name := evoInstanceName(tenantID)
	data, status, err := h.evoCall(c.Context(), http.MethodGet, "/instance/connect/"+name, ch.ExternalKey.String, nil)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "evolution api unreachable"})
	}
	if status >= 300 {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": string(data)})
	}

	var evoResp struct {
		Code   string `json:"code"`
		Base64 string `json:"base64"`
	}
	if err := json.Unmarshal(data, &evoResp); err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "invalid evolution api response"})
	}

	return c.JSON(fiber.Map{
		"qr_code": evoResp.Code,
		"base64":  evoResp.Base64,
	})
}

// DELETE /api/v1/whatsapp/instance/logout
// Logs out of WhatsApp and marks the channel inactive. The EVO instance is kept alive.
func (h *WhatsappHandler) Logout(c *fiber.Ctx) error {
	tenantID, ch, ok := h.loadChannel(c)
	if !ok {
		return nil
	}

	name := evoInstanceName(tenantID)
	data, status, err := h.evoCall(c.Context(), http.MethodDelete, "/instance/logout/"+name, ch.ExternalKey.String, nil)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "evolution api unreachable"})
	}
	if status >= 300 {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": string(data)})
	}

	if err := h.Store.DeactivateWhatsappChannelByTenant(c.Context(), tenantID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not deactivate channel"})
	}

	return c.JSON(fiber.Map{"message": "disconnected"})
}

// DELETE /api/v1/whatsapp/instance
// Fully deletes the EVO API instance and removes the channel record (hard reset).
func (h *WhatsappHandler) DeleteInstance(c *fiber.Ctx) error {
	tenantID, ch, ok := h.loadChannel(c)
	if !ok {
		return nil
	}

	name := evoInstanceName(tenantID)
	data, status, err := h.evoCall(c.Context(), http.MethodDelete, "/instance/delete/"+name, ch.ExternalKey.String, nil)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "evolution api unreachable"})
	}
	if status >= 300 {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": string(data)})
	}

	if err := h.Store.DeleteWhatsappChannelByTenant(c.Context(), tenantID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not delete channel"})
	}

	return c.JSON(fiber.Map{"message": "instance deleted"})
}
