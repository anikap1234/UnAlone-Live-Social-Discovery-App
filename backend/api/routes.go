package api

import (
    "github.com/gofiber/fiber/v2"
    "unalone/backend/auth"
    "unalone/backend/config"
    "unalone/backend/ws"
    loc "unalone/backend/api/location"
    m "unalone/backend/api/meetups"
)

func RegisterRoutes(app *fiber.App, cfg *config.Config, hub *ws.Hub) {
    api := app.Group("/")

    api.Post("/auth/send-otp", SendOTPHandler(cfg))
    api.Post("/auth/verify-otp", VerifyOTPHandler(cfg))

    // websocket
    app.Get("/ws", ws.ServeWS(hub))

    // protected
    authMiddleware := auth.JWTMiddleware(cfg.JWTSecret)
    api.Post("/location/update", authMiddleware, loc.UpdateLocationHandler(hub))
    api.Post("/meetups", authMiddleware, m.CreateMeetupHandler(hub))
    api.Get("/meetups/near", authMiddleware, m.GetMeetupsNearHandler)
}
