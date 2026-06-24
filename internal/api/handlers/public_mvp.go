package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/unknowncode44/appointments/internal/api/dto"
	"github.com/unknowncode44/appointments/internal/api/response"
	"github.com/unknowncode44/appointments/internal/services"
)

// publicTenantKey is the Fiber Locals key holding the tenant resolved from the
// :slug path segment for the current public request.
const publicTenantKey = "public_tenant"

type PublicHandler struct{ service services.PublicService }

func NewPublicHandler(service services.PublicService) *PublicHandler {
	return &PublicHandler{service: service}
}

// ResolveTenant is group middleware for /public/:slug. It resolves the slug to
// an active tenant and stores it in the request context. A missing/inactive
// slug yields 404. The slug is the only tenant scope on these routes.
func (h *PublicHandler) ResolveTenant(c *fiber.Ctx) error {
	tenant, err := h.service.ResolveTenant(c.Context(), c.Params("slug"))
	if err != nil {
		return response.Error(c, err)
	}
	c.Locals(publicTenantKey, tenant)
	return c.Next()
}

// tenantFromCtx returns the tenant resolved by ResolveTenant. It is always
// present because ResolveTenant runs as group middleware before every handler.
func tenantFromCtx(c *fiber.Ctx) dto.PublicTenant {
	return c.Locals(publicTenantKey).(dto.PublicTenant)
}

// Shop returns the public-safe shop header (name, timezone, slug).
func (h *PublicHandler) Shop(c *fiber.Ctx) error {
	t := tenantFromCtx(c)
	return c.JSON(dto.PublicShopResponse{Name: t.Name, Timezone: t.Timezone, Slug: t.Slug})
}

func (h *PublicHandler) Services(c *fiber.Ctx) error {
	t := tenantFromCtx(c)
	rsp, err := h.service.ListServices(c.Context(), t.ID)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(fiber.Map{"data": rsp})
}

func (h *PublicHandler) Providers(c *fiber.Ctx) error {
	t := tenantFromCtx(c)
	rsp, err := h.service.ListProviders(c.Context(), t.ID)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(fiber.Map{"data": rsp})
}

func (h *PublicHandler) Availability(c *fiber.Ctx) error {
	t := tenantFromCtx(c)
	providerID, err := queryUUID(c, "provider_id")
	if err != nil {
		return response.Error(c, err)
	}
	rsp, err := h.service.ListAvailability(c.Context(), t.ID, providerID, c.Query("date"))
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(fiber.Map{"data": rsp})
}

func (h *PublicHandler) CreateBooking(c *fiber.Ctx) error {
	t := tenantFromCtx(c)
	var req dto.PublicBookingRequest
	if err := bindAndValidate(c, &req); err != nil {
		return response.BadRequest(c, err)
	}
	rsp, err := h.service.CreateBooking(c.Context(), t.ID, req)
	if err != nil {
		return response.Error(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(rsp)
}
