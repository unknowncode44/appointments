package routes

import (
	"github.com/mousav1/ticket/internal/api"
	"github.com/mousav1/ticket/internal/api/handlers"
	"github.com/mousav1/ticket/internal/api/middleware"
	"github.com/mousav1/ticket/internal/repositories"
	"github.com/mousav1/ticket/internal/services"
)

func SetupRoutes(server *api.Server) error {

	server.App.Post("/register", handlers.NewUserHandler(server.Store, server.TokenMaker, server.Config).RegisterUser)
	server.App.Post("/login", handlers.NewUserHandler(server.Store, server.TokenMaker, server.Config).LoginUser)
	server.App.Post("/tokens/renew_access", handlers.NewTokenHandler(server.Store, server.TokenMaker, server.Config).RenewAccessToken)

	adminRepo := repositories.NewAdminRepository(server.Store)
	schedulingRepo := repositories.NewSchedulingRepository(server.Store)
	workflowRepo := repositories.NewWorkflowRepository(server.Store)

	adminHandler := handlers.NewAdminMVPHandler(services.NewAdminService(adminRepo))
	schedulingHandler := handlers.NewSchedulingMVPHandler(services.NewSchedulingService(schedulingRepo))
	appointmentHandler := handlers.NewAppointmentMVPHandler(services.NewAppointmentService(workflowRepo))
	conversationHandler := handlers.NewConversationMVPHandler(services.NewConversationService(workflowRepo))

	server.App.Post("/api/v1/webhooks/evolution", conversationHandler.EvolutionWebhook)

	// Grouped routes that require authentication and role authorization
	authGroup := server.App.Group("/", middleware.AuthMiddleware(server.TokenMaker), middleware.RoleMiddleware())
	authGroup.Get("/user/info", handlers.NewUserHandler(server.Store, server.TokenMaker, server.Config).GetUserProfile)
	authGroup.Put("/user/update", handlers.NewUserHandler(server.Store, server.TokenMaker, server.Config).UpdateUserProfile)
	authGroup.Post("/user/password_change", handlers.NewUserHandler(server.Store, server.TokenMaker, server.Config).ChangePassword)

	apiV1 := authGroup.Group("/api/v1")

	// Admin only routes
	apiV1.Get("/tenants", adminHandler.ListTenants)
	apiV1.Post("/tenants", adminHandler.CreateTenant)
	apiV1.Get("/tenants/:id", adminHandler.GetTenant)
	apiV1.Put("/tenants/:id", adminHandler.UpdateTenant)
	apiV1.Delete("/tenants/:id", adminHandler.DeactivateTenant)

	// User management routes (admin only)
	apiV1.Get("/users", handlers.NewUserHandler(server.Store, server.TokenMaker, server.Config).ListUsers)
	apiV1.Post("/users", handlers.NewUserHandler(server.Store, server.TokenMaker, server.Config).CreateUserAdmin)
	apiV1.Get("/users/:id", handlers.NewUserHandler(server.Store, server.TokenMaker, server.Config).GetUserByID)
	apiV1.Put("/users/:id", handlers.NewUserHandler(server.Store, server.TokenMaker, server.Config).UpdateUser)
	apiV1.Delete("/users/:id", handlers.NewUserHandler(server.Store, server.TokenMaker, server.Config).DeleteUser)
	apiV1.Post("/users/:id/tenant", handlers.NewUserHandler(server.Store, server.TokenMaker, server.Config).LinkUserToTenant)
	apiV1.Get("/users/:id/providers", handlers.NewUserHandler(server.Store, server.TokenMaker, server.Config).GetUserProviders)
	apiV1.Post("/users/:id/provider", handlers.NewUserHandler(server.Store, server.TokenMaker, server.Config).LinkUserToProvider)
	apiV1.Delete("/users/:id/provider", handlers.NewUserHandler(server.Store, server.TokenMaker, server.Config).RemoveUserFromProvider)

	// TenantUser and Admin routes
	apiV1.Get("/providers", adminHandler.ListProviders)
	apiV1.Post("/providers", adminHandler.CreateProvider)
	apiV1.Get("/providers/:id", adminHandler.GetProvider)
	apiV1.Put("/providers/:id", adminHandler.UpdateProvider)
	apiV1.Delete("/providers/:id", adminHandler.DeactivateProvider)
	apiV1.Post("/providers/:id/availability", schedulingHandler.CreateAvailability)
	apiV1.Get("/providers/:id/availability", schedulingHandler.ListAvailability)
	apiV1.Post("/providers/:id/exceptions", schedulingHandler.CreateException)
	apiV1.Get("/providers/:id/exceptions", schedulingHandler.ListExceptions)

	apiV1.Get("/services", adminHandler.ListServices)
	apiV1.Post("/services", adminHandler.CreateService)
	apiV1.Get("/services/:id", adminHandler.GetService)
	apiV1.Put("/services/:id", adminHandler.UpdateService)
	apiV1.Delete("/services/:id", adminHandler.DeactivateService)

	// All authenticated users
	apiV1.Get("/customers", adminHandler.ListCustomers)
	apiV1.Post("/customers", adminHandler.CreateCustomer)
	apiV1.Get("/customers/:id", adminHandler.GetCustomer)
	apiV1.Put("/customers/:id", adminHandler.UpdateCustomer)

	apiV1.Get("/tenant-channels", adminHandler.ListTenantChannels)
	apiV1.Post("/tenant-channels", adminHandler.CreateTenantChannel)
	apiV1.Get("/tenant-channels/:id", adminHandler.GetTenantChannel)
	apiV1.Put("/tenant-channels/:id", adminHandler.UpdateTenantChannel)
	apiV1.Delete("/tenant-channels/:id", adminHandler.DeactivateTenantChannel)

	apiV1.Post("/slot-generator", schedulingHandler.GenerateSlots)
	apiV1.Get("/availability", schedulingHandler.Availability)

	apiV1.Post("/appointments", appointmentHandler.Create)
	apiV1.Get("/appointments", appointmentHandler.List)
	apiV1.Get("/appointments/:id", appointmentHandler.Get)
	apiV1.Patch("/appointments/:id", appointmentHandler.Update)
	apiV1.Delete("/appointments/:id", appointmentHandler.Delete)

	apiV1.Get("/conversations", conversationHandler.List)
	apiV1.Get("/conversations/:id", conversationHandler.Get)
	apiV1.Post("/conversations/message", conversationHandler.Message)
	apiV1.Post("/inbound-messages", conversationHandler.InboundMessage)

	return nil
}
