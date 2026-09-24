package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"log"
	"strings"
	"unalone/backend/config"
	"unalone/backend/services"
	"unalone/backend/ws"
)

func NewApp(cfg *config.Config, hub *ws.Hub, hotspots *services.HotspotService) *fiber.App {
	app := fiber.New(fiber.Config{BodyLimit: 16 * 1024, DisableStartupMessage: true})
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{AllowOrigins: cfg.AllowedOrigins, AllowCredentials: true, AllowMethods: "GET,POST,DELETE,OPTIONS", AllowHeaders: "Content-Type,Authorization"}))
	app.Use(func(c *fiber.Ctx) error {
		if c.Method() != "GET" && c.Method() != "OPTIONS" {
			if origin := c.Get("Origin"); origin != "" {
				allowed := false
				for _, value := range strings.Split(cfg.AllowedOrigins, ",") {
					if strings.TrimSpace(value) == origin {
						allowed = true
					}
				}
				if !allowed {
					return c.Status(403).JSON(fiber.Map{"error": "Origin not allowed"})
				}
			}
		}
		err := c.Next()
		// Paths only: query strings can contain GPS and must not enter logs.
		log.Printf("%s %s %d", c.Method(), c.Path(), c.Response().StatusCode())
		return err
	})
	RegisterRoutes(app, cfg, hub, hotspots)
	return app
}
