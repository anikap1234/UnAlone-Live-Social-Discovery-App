package meetups

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"strconv"
	"time"
	"unalone/backend/services"
	"unalone/backend/utils"
)

func GetMeetupsNearHandler(c *fiber.Ctx) error {
	lat, e1 := strconv.ParseFloat(c.Query("lat"), 64)
	lon, e2 := strconv.ParseFloat(c.Query("lon"), 64)
	radius, e3 := strconv.ParseFloat(c.Query("radius", "1000"), 64)
	if e1 != nil || e2 != nil || e3 != nil || !utils.ValidCoordinates(lat, lon) || !(radius > 0 && radius <= 50000) {
		return c.Status(400).JSON(fiber.Map{"error": "Supply valid coordinates and a radius of 1–50,000 meters"})
	}
	ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
	defer cancel()
	meetups, err := services.NearbyMeetups(ctx, lat, lon, radius)
	if err != nil {
		return c.Status(503).JSON(fiber.Map{"error": "Could not load nearby meetups"})
	}
	return c.JSON(fiber.Map{"meetups": meetups})
}
