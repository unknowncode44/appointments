package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/unknowncode44/appointments/internal/api/dto"
	"github.com/unknowncode44/appointments/internal/api/response"
	"github.com/unknowncode44/appointments/internal/platform/pagination"
	"github.com/unknowncode44/appointments/internal/services"
)

type AppointmentMVPHandler struct{ service services.AppointmentService }
type ConversationMVPHandler struct{ service services.ConversationService }

func NewAppointmentMVPHandler(service services.AppointmentService) *AppointmentMVPHandler {
	return &AppointmentMVPHandler{service: service}
}

func NewConversationMVPHandler(service services.ConversationService) *ConversationMVPHandler {
	return &ConversationMVPHandler{service: service}
}

func (h *AppointmentMVPHandler) Create(c *fiber.Ctx) error {
	scope, err := callerTenantScope(c)
	if err != nil {
		return response.Error(c, err)
	}
	var req dto.AppointmentCreateRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.BadRequest(c, err)
	}
	rsp, err := h.service.Create(c.Context(), req, scope)
	if err != nil {
		return response.Error(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(rsp)
}

func (h *AppointmentMVPHandler) List(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return response.Error(c, response.ErrInvalidInput)
	}
	providerID, err := queryUUID(c, "provider_id")
	if err != nil {
		return response.Error(c, err)
	}
	customerID, err := queryUUID(c, "customer_id")
	if err != nil {
		return response.Error(c, err)
	}
	rsp, err := h.service.List(c.Context(), tenantID, providerID, customerID, c.Query("status"), pagination.FromCtx(c))
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(rsp)
}

func (h *AppointmentMVPHandler) Get(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	scope, err := callerTenantScope(c)
	if err != nil {
		return response.Error(c, err)
	}
	rsp, err := h.service.Get(c.Context(), id, scope)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(rsp)
}

func (h *AppointmentMVPHandler) Update(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	scope, err := callerTenantScope(c)
	if err != nil {
		return response.Error(c, err)
	}
	var req dto.AppointmentUpdateRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.BadRequest(c, err)
	}
	rsp, err := h.service.Update(c.Context(), id, req, scope)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(rsp)
}

func (h *AppointmentMVPHandler) Delete(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	scope, err := callerTenantScope(c)
	if err != nil {
		return response.Error(c, err)
	}
	rsp, err := h.service.Cancel(c.Context(), id, scope)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(rsp)
}

func (h *ConversationMVPHandler) List(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return response.Error(c, response.ErrInvalidInput)
	}
	rsp, err := h.service.ListThreads(c.Context(), tenantID, pagination.FromCtx(c))
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(fiber.Map{"data": rsp})
}

func (h *ConversationMVPHandler) Get(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	scope, err := callerTenantScope(c)
	if err != nil {
		return response.Error(c, err)
	}
	rsp, err := h.service.GetThread(c.Context(), id, scope)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(rsp)
}

func (h *ConversationMVPHandler) Message(c *fiber.Ctx) error {
	scope, err := callerTenantScope(c)
	if err != nil {
		return response.Error(c, err)
	}
	var req dto.ConversationMessageRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.BadRequest(c, err)
	}
	rsp, err := h.service.StoreMessage(c.Context(), req, scope)
	if err != nil {
		return response.Error(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(rsp)
}

func (h *ConversationMVPHandler) InboundMessage(c *fiber.Ctx) error {
	var req dto.InboundMessageRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.BadRequest(c, err)
	}
	rsp, err := h.service.ProcessInboundMessage(c.Context(), req)
	if err != nil {
		return response.Error(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(rsp)
}

func (h *ConversationMVPHandler) EvolutionWebhook(c *fiber.Ctx) error {
	var req dto.EvolutionWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, err)
	}
	if err := h.service.ProcessEvolutionWebhook(c.Context(), req, c.Body()); err != nil {
		return response.Error(c, err)
	}
	return c.SendStatus(fiber.StatusAccepted)
}
