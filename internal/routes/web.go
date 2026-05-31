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

	// Grouped routes that require authentication
	authGroup := server.App.Group("/", middleware.AuthMiddleware(server.TokenMaker))
	authGroup.Get("/user/info", handlers.NewUserHandler(server.Store, server.TokenMaker, server.Config).GetUserProfile)
	authGroup.Put("/user/update", handlers.NewUserHandler(server.Store, server.TokenMaker, server.Config).UpdateUserProfile)
	authGroup.Post("/user/password_change", handlers.NewUserHandler(server.Store, server.TokenMaker, server.Config).ChangePassword)

	adminRepo := repositories.NewAdminRepository(server.Store)
	schedulingRepo := repositories.NewSchedulingRepository(server.Store)
	workflowRepo := repositories.NewWorkflowRepository(server.Store)

	adminHandler := handlers.NewAdminMVPHandler(services.NewAdminService(adminRepo))
	schedulingHandler := handlers.NewSchedulingMVPHandler(services.NewSchedulingService(schedulingRepo))
	appointmentHandler := handlers.NewAppointmentMVPHandler(services.NewAppointmentService(workflowRepo))
	conversationHandler := handlers.NewConversationMVPHandler(services.NewConversationService(workflowRepo))

	apiV1 := authGroup.Group("/api/v1")
	apiV1.Get("/tenants", adminHandler.ListTenants)
	apiV1.Post("/tenants", adminHandler.CreateTenant)
	apiV1.Get("/tenants/:id", adminHandler.GetTenant)
	apiV1.Put("/tenants/:id", adminHandler.UpdateTenant)
	apiV1.Delete("/tenants/:id", adminHandler.DeactivateTenant)

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

	apiV1.Get("/customers", adminHandler.ListCustomers)
	apiV1.Post("/customers", adminHandler.CreateCustomer)
	apiV1.Get("/customers/:id", adminHandler.GetCustomer)
	apiV1.Put("/customers/:id", adminHandler.UpdateCustomer)

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

	server.App.Post("/api/v1/webhooks/evolution", conversationHandler.EvolutionWebhook)
	return nil
}
