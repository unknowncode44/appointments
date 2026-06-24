package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/unknowncode44/appointments/internal/api/dto"
	"github.com/unknowncode44/appointments/internal/api/response"
	"github.com/unknowncode44/appointments/internal/platform/pagination"
	"github.com/unknowncode44/appointments/internal/services"
)

type SchedulingMVPHandler struct{ service services.SchedulingService }

func NewSchedulingMVPHandler(service services.SchedulingService) *SchedulingMVPHandler {
	return &SchedulingMVPHandler{service: service}
}

func (h *SchedulingMVPHandler) CreateAvailability(c *fiber.Ctx) error {
	providerID, err := parseID(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	scope, err := callerTenantScope(c)
	if err != nil {
		return response.Error(c, err)
	}
	var req dto.AvailabilityRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.BadRequest(c, err)
	}
	rsp, err := h.service.CreateAvailability(c.Context(), providerID, req, scope)
	if err != nil {
		return response.Error(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(rsp)
}

func (h *SchedulingMVPHandler) ListAvailability(c *fiber.Ctx) error {
	providerID, err := parseID(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	scope, err := callerTenantScope(c)
	if err != nil {
		return response.Error(c, err)
	}
	rsp, err := h.service.ListAvailability(c.Context(), providerID, scope)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(fiber.Map{"data": rsp})
}

func (h *SchedulingMVPHandler) CreateException(c *fiber.Ctx) error {
	providerID, err := parseID(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	scope, err := callerTenantScope(c)
	if err != nil {
		return response.Error(c, err)
	}
	var req dto.ExceptionRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.BadRequest(c, err)
	}
	rsp, err := h.service.CreateException(c.Context(), providerID, req, scope)
	if err != nil {
		return response.Error(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(rsp)
}

func (h *SchedulingMVPHandler) ListExceptions(c *fiber.Ctx) error {
	providerID, err := parseID(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	scope, err := callerTenantScope(c)
	if err != nil {
		return response.Error(c, err)
	}
	rsp, err := h.service.ListExceptions(c.Context(), providerID, scope)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(fiber.Map{"data": rsp})
}

func (h *SchedulingMVPHandler) GenerateSlots(c *fiber.Ctx) error {
	scope, err := callerTenantScope(c)
	if err != nil {
		return response.Error(c, err)
	}
	var req dto.SlotGeneratorRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.BadRequest(c, err)
	}
	rsp, err := h.service.GenerateSlots(c.Context(), req, scope)
	if err != nil {
		return response.Error(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(rsp)
}

func (h *SchedulingMVPHandler) Availability(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return response.Error(c, response.ErrInvalidInput)
	}
	providerID, err := queryUUID(c, "provider_id")
	if err != nil {
		return response.Error(c, err)
	}
	serviceID, err := queryUUID(c, "service_id")
	if err != nil {
		return response.Error(c, err)
	}
	date := c.Query("date")
	if date == "" {
		return response.Error(c, response.ErrInvalidInput)
	}
	rsp, err := h.service.ListAvailableSlots(c.Context(), tenantID, providerID, serviceID, date, pagination.FromCtx(c))
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(fiber.Map{"data": rsp})
}
