package location

import (
    "context"

    "github.com/gofiber/fiber/v2"
    "unalone/backend/services"
    "unalone/backend/ws"
)

type locReq struct {
    Lat float64 `json:"lat"`
    Lon float64 `json:"lon"`
}

func UpdateLocationHandler(hub *ws.Hub) fiber.Handler {
    return func(c *fiber.Ctx) error {
        var req locReq
        if err := c.BodyParser(&req); err != nil {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
        }
        emailI := c.Locals("email")
        if emailI == nil {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthenticated"})
        }
        email := emailI.(string)
        ctx := context.Background()
        if err := services.UpdateLocationAndMaybeBroadcast(ctx, hub, email, req.Lat, req.Lon); err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed"})
        }
        return c.JSON(fiber.Map{"ok": true})
    }
}
