package api

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	db "github.com/mousav1/ticket/internal/db/sqlc"
	"github.com/mousav1/ticket/internal/token"
	"github.com/mousav1/ticket/internal/util"
)

// Server serves HTTP requests for the appointments service.
type Server struct {
	Config     util.Config
	Store      *db.Store
	TokenMaker token.Maker
	App        *fiber.App
}

// NewServer creates a new HTTP server and configures global middleware.
func NewServer(config util.Config, store *db.Store) (*Server, error) {
	tokenMaker, err := token.NewJWTMaker(config.TOKENSECRETKEY)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders: "*",
	}))

	// Rate-limit public auth endpoints to 10 requests per minute per IP.
	// This mitigates brute-force on /login and account-enumeration on /register.
	authLimiter := limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "too many requests, please try again later"})
		},
	})
	app.Use("/register", authLimiter)
	app.Use("/login", authLimiter)

	server := &Server{
		Config:     config,
		Store:      store,
		TokenMaker: tokenMaker,
		App:        app,
	}

	return server, nil
}

// Start runs the HTTP server on the configured port.
func (server *Server) Start(address string) error {
	return server.App.Listen(fmt.Sprintf(":%s", address))
}
