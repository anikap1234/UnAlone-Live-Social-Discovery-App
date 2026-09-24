package api

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"strings"
	"time"
	authapi "unalone/backend/api/auth"
	loc "unalone/backend/api/location"
	meetups "unalone/backend/api/meetups"
	"unalone/backend/auth"
	"unalone/backend/config"
	"unalone/backend/redis"
	"unalone/backend/services"
	"unalone/backend/ws"
)

func RegisterRoutes(app *fiber.App, cfg *config.Config, hub *ws.Hub, hotspots *services.HotspotService) {
	app.Get("/health", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"status": "ok"}) })
	app.Post("/auth/send-otp", authapi.SendOTPHandler(cfg))
	app.Post("/auth/verify-otp", authapi.VerifyOTPHandler(cfg))
	protected := auth.JWTMiddleware(cfg.JWTSecret)
	app.Get("/auth/me", protected, func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"user": fiber.Map{"userId": c.Locals("userId"), "email": c.Locals("email")}})
	})
	app.Post("/auth/logout", protected, func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
		defer cancel()
		_ = redis.RemoveLocation(ctx, c.Locals("userId").(string))
		_ = hotspots.Reconcile(ctx)
		c.Cookie(&fiber.Cookie{Name: auth.CookieName, Value: "", Path: "/", HTTPOnly: true, Secure: cfg.Env == "production", SameSite: "Lax", Expires: time.Unix(1, 0), MaxAge: -1})
		return c.JSON(fiber.Map{"ok": true})
	})
	app.Get("/ws", protected, ws.ServeWS(hub, strings.Split(cfg.AllowedOrigins, ",")))
	app.Post("/location/update", protected, loc.UpdateLocationHandler(hotspots))
	app.Delete("/location", protected, func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
		defer cancel()
		if err := redis.RemoveLocation(ctx, c.Locals("userId").(string)); err != nil {
			return c.SendStatus(503)
		}
		if err := hotspots.Reconcile(ctx); err != nil {
			return c.SendStatus(503)
		}
		return c.JSON(fiber.Map{"ok": true})
	})
	app.Get("/hotspots", protected, func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
		defer cancel()
		snapshot, err := hotspots.Snapshot(ctx)
		if err != nil {
			return c.Status(503).JSON(fiber.Map{"error": "Could not load live hotspots"})
		}
		return c.JSON(fiber.Map{"hotspots": snapshot})
	})
	app.Post("/meetups", protected, meetups.CreateMeetupHandler(hub))
	app.Get("/meetups/near", protected, meetups.GetMeetupsNearHandler)
}
