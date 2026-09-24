package location

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"time"
	"unalone/backend/redis"
	"unalone/backend/services"
	"unalone/backend/utils"
)

func UpdateLocationHandler(hotspots *services.HotspotService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			Lat *float64 `json:"lat"`
			Lon *float64 `json:"lon"`
		}
		if c.BodyParser(&req) != nil || req.Lat == nil || req.Lon == nil || !utils.ValidCoordinates(*req.Lat, *req.Lon) {
			return c.Status(400).JSON(fiber.Map{"error": "Supply valid latitude and longitude"})
		}
		ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
		defer cancel()
		err := hotspots.Update(ctx, c.Locals("userId").(string), *req.Lat, *req.Lon)
		if errors.Is(err, redis.ErrLocationRateLimited) {
			return c.Status(429).JSON(fiber.Map{"error": err.Error()})
		}
		if err != nil {
			return c.Status(503).JSON(fiber.Map{"error": "Could not share your location"})
		}
		return c.JSON(fiber.Map{"ok": true})
	}
}
