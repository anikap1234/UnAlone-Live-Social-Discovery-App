package meetups

import (
    "context"
    "strconv"

    "github.com/gofiber/fiber/v2"

    "unalone/backend/services"
)

func GetMeetupsNearHandler(c *fiber.Ctx) error {
    latStr := c.Query("lat")
    lonStr := c.Query("lon")
    if latStr == "" || lonStr == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "lat/lon required"})
    }
    lat, err := strconv.ParseFloat(latStr, 64)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid lat"})
    }
    lon, err := strconv.ParseFloat(lonStr, 64)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid lon"})
    }
    // default radius 1000 meters
    meetups, err := services.NearbyMeetups(context.Background(), lat, lon, 1000)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed"})
    }
    return c.JSON(fiber.Map{"meetups": meetups})
}
