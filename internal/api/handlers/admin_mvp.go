package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/unknowncode44/appointments/internal/api/dto"
	"github.com/unknowncode44/appointments/internal/api/response"
	"github.com/unknowncode44/appointments/internal/platform/pagination"
	"github.com/unknowncode44/appointments/internal/services"
)

type AdminMVPHandler struct{ service services.AdminService }

func NewAdminMVPHandler(service services.AdminService) *AdminMVPHandler {
	return &AdminMVPHandler{service: service}
}

func (h *AdminMVPHandler) ListTenants(c *fiber.Ctx) error {
	active, err := queryBool(c, "active")
	if err != nil {
		return response.Error(c, err)
	}
	rsp, err := h.service.ListTenants(c.Context(), c.Query("search"), active, pagination.FromCtx(c))
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(rsp)
}

func (h *AdminMVPHandler) GetTenant(c *fiber.Ctx) error {
	return h.getTenant(c, "id")
}

func (h *AdminMVPHandler) CreateTenant(c *fiber.Ctx) error {
	var req dto.TenantRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.BadRequest(c, err)
	}
	rsp, err := h.service.CreateTenant(c.Context(), req)
	if err != nil {
		return response.Error(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(rsp)
}

func (h *AdminMVPHandler) UpdateTenant(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	var req dto.TenantRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.BadRequest(c, err)
	}
	rsp, err := h.service.UpdateTenant(c.Context(), id, req)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(rsp)
}

func (h *AdminMVPHandler) DeactivateTenant(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	rsp, err := h.service.DeactivateTenant(c.Context(), id)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(rsp)
}

func (h *AdminMVPHandler) ListProviders(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return response.Error(c, response.ErrInvalidInput)
	}
	active, err := queryBool(c, "active")
	if err != nil {
		return response.Error(c, err)
	}
	rsp, err := h.service.ListProviders(c.Context(), tenantID, c.Query("search"), active, pagination.FromCtx(c))
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(rsp)
}

func (h *AdminMVPHandler) GetProvider(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	scope, err := callerTenantScope(c)
	if err != nil {
		return response.Error(c, err)
	}
	rsp, err := h.service.GetProvider(c.Context(), id, scope)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(rsp)
}

func (h *AdminMVPHandler) CreateProvider(c *fiber.Ctx) error {
	scope, err := callerTenantScope(c)
	if err != nil {
		return response.Error(c, err)
	}
	var req dto.ProviderRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.BadRequest(c, err)
	}
	if scope != nil {
		req.TenantID = *scope
	}
	rsp, err := h.service.CreateProvider(c.Context(), req)
	if err != nil {
		return response.Error(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(rsp)
}

func (h *AdminMVPHandler) UpdateProvider(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	scope, err := callerTenantScope(c)
	if err != nil {
		return response.Error(c, err)
	}
	var req dto.ProviderUpdateRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.BadRequest(c, err)
	}
	rsp, err := h.service.UpdateProvider(c.Context(), id, req, scope)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(rsp)
}

func (h *AdminMVPHandler) DeactivateProvider(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	scope, err := callerTenantScope(c)
	if err != nil {
		return response.Error(c, err)
	}
	rsp, err := h.service.DeactivateProvider(c.Context(), id, scope)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(rsp)
}

func (h *AdminMVPHandler) ListServices(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return response.Error(c, response.ErrInvalidInput)
	}
	active, err := queryBool(c, "active")
	if err != nil {
		return response.Error(c, err)
	}
	rsp, err := h.service.ListServices(c.Context(), tenantID, c.Query("search"), active, pagination.FromCtx(c))
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(rsp)
}

func (h *AdminMVPHandler) GetService(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	scope, err := callerTenantScope(c)
	if err != nil {
		return response.Error(c, err)
	}
	rsp, err := h.service.GetService(c.Context(), id, scope)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(rsp)
}

func (h *AdminMVPHandler) CreateService(c *fiber.Ctx) error {
	scope, err := callerTenantScope(c)
	if err != nil {
		return response.Error(c, err)
	}
	var req dto.ServiceRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.BadRequest(c, err)
	}
	if scope != nil {
		req.TenantID = *scope
	}
	rsp, err := h.service.CreateService(c.Context(), req)
	if err != nil {
		return response.Error(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(rsp)
}

func (h *AdminMVPHandler) UpdateService(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	scope, err := callerTenantScope(c)
	if err != nil {
		return response.Error(c, err)
	}
	var req dto.ServiceUpdateRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.BadRequest(c, err)
	}
	rsp, err := h.service.UpdateService(c.Context(), id, req, scope)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(rsp)
}

func (h *AdminMVPHandler) DeactivateService(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	scope, err := callerTenantScope(c)
	if err != nil {
		return response.Error(c, err)
	}
	rsp, err := h.service.DeactivateService(c.Context(), id, scope)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(rsp)
}

func (h *AdminMVPHandler) ListCustomers(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return response.Error(c, response.ErrInvalidInput)
	}
	rsp, err := h.service.ListCustomers(c.Context(), tenantID, c.Query("search"), pagination.FromCtx(c))
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(rsp)
}

func (h *AdminMVPHandler) GetCustomer(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	scope, err := callerTenantScope(c)
	if err != nil {
		return response.Error(c, err)
	}
	rsp, err := h.service.GetCustomer(c.Context(), id, scope)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(rsp)
}

func (h *AdminMVPHandler) CreateCustomer(c *fiber.Ctx) error {
	scope, err := callerTenantScope(c)
	if err != nil {
		return response.Error(c, err)
	}
	var req dto.CustomerRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.BadRequest(c, err)
	}
	if scope != nil {
		req.TenantID = *scope
	}
	rsp, err := h.service.CreateCustomer(c.Context(), req)
	if err != nil {
		return response.Error(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(rsp)
}

func (h *AdminMVPHandler) UpdateCustomer(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	scope, err := callerTenantScope(c)
	if err != nil {
		return response.Error(c, err)
	}
	var req dto.CustomerUpdateRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.BadRequest(c, err)
	}
	rsp, err := h.service.UpdateCustomer(c.Context(), id, req, scope)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(rsp)
}

func (h *AdminMVPHandler) ListTenantChannels(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Query("tenant_id"))
	if err != nil {
		return response.Error(c, response.ErrInvalidInput)
	}
	active, err := queryBool(c, "active")
	if err != nil {
		return response.Error(c, err)
	}
	rsp, err := h.service.ListTenantChannels(c.Context(), tenantID, c.Query("channel_type"), active, pagination.FromCtx(c))
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(rsp)
}

func (h *AdminMVPHandler) GetTenantChannel(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	scope, err := callerTenantScope(c)
	if err != nil {
		return response.Error(c, err)
	}
	rsp, err := h.service.GetTenantChannel(c.Context(), id, scope)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(rsp)
}

func (h *AdminMVPHandler) CreateTenantChannel(c *fiber.Ctx) error {
	scope, err := callerTenantScope(c)
	if err != nil {
		return response.Error(c, err)
	}
	var req dto.TenantChannelRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.BadRequest(c, err)
	}
	if scope != nil {
		req.TenantID = *scope
	}
	rsp, err := h.service.CreateTenantChannel(c.Context(), req)
	if err != nil {
		return response.Error(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(rsp)
}

func (h *AdminMVPHandler) UpdateTenantChannel(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	scope, err := callerTenantScope(c)
	if err != nil {
		return response.Error(c, err)
	}
	var req dto.TenantChannelUpdateRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.BadRequest(c, err)
	}
	rsp, err := h.service.UpdateTenantChannel(c.Context(), id, req, scope)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(rsp)
}

func (h *AdminMVPHandler) DeactivateTenantChannel(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return response.Error(c, err)
	}
	scope, err := callerTenantScope(c)
	if err != nil {
		return response.Error(c, err)
	}
	rsp, err := h.service.DeactivateTenantChannel(c.Context(), id, scope)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(rsp)
}

func (h *AdminMVPHandler) getTenant(c *fiber.Ctx, param string) error {
	id, err := parseID(c, param)
	if err != nil {
		return response.Error(c, err)
	}
	rsp, err := h.service.GetTenant(c.Context(), id)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(rsp)
}
