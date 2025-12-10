package meetups

import (
    "context"

    "github.com/gofiber/fiber/v2"
    "encoding/json"

    "unalone/backend/models"
    "unalone/backend/services"
    "unalone/backend/ws"
)

type createReq struct {
    Title string `json:"title"`
    Description string `json:"description"`
    Lat float64 `json:"lat"`
    Lon float64 `json:"lon"`
}

func CreateMeetupHandler(hub *ws.Hub) fiber.Handler {
    return func(c *fiber.Ctx) error {
        var req createReq
        if err := c.BodyParser(&req); err != nil {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid"})
        }
        emailI := c.Locals("email")
        if emailI == nil {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthenticated"})
        }
        m := &models.Meetup{
            Title: req.Title,
            Description: req.Description,
            Lat: req.Lat,
            Lon: req.Lon,
            CreatedBy: emailI.(string),
        }
        if err := services.CreateMeetup(context.Background(), m); err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed"})
        }
        b, _ := json.Marshal(ws.MeetupCreated{Type: "MEETUP_CREATED", Data: m})
        hub.Broadcast(b)
        return c.JSON(fiber.Map{"ok": true, "meetup": m})
    }
}
