package meetups

import (
	"context"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"strings"
	"time"
	"unalone/backend/models"
	"unalone/backend/services"
	"unalone/backend/utils"
	"unalone/backend/ws"
	"unicode/utf8"
)

func CreateMeetupHandler(hub *ws.Hub) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			Title       string   `json:"title"`
			Description string   `json:"description"`
			Lat         *float64 `json:"lat"`
			Lon         *float64 `json:"lon"`
			Time        int64    `json:"time"`
		}
		if c.BodyParser(&req) != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid meetup"})
		}
		title, description := strings.TrimSpace(req.Title), strings.TrimSpace(req.Description)
		if utf8.RuneCountInString(title) < 3 || utf8.RuneCountInString(title) > 100 || utf8.RuneCountInString(description) > 1000 {
			return c.Status(400).JSON(fiber.Map{"error": "Use a title of 3–100 characters and a description of at most 1,000 characters"})
		}
		if req.Lat == nil || req.Lon == nil || !utils.ValidCoordinates(*req.Lat, *req.Lon) {
			return c.Status(400).JSON(fiber.Map{"error": "Choose a valid location"})
		}
		if req.Time <= time.Now().Unix() {
			return c.Status(400).JSON(fiber.Map{"error": "Choose a future meetup time"})
		}
		m := models.Meetup{Title: title, Description: description, Lat: *req.Lat, Lon: *req.Lon, Time: req.Time, CreatedBy: c.Locals("userId").(string)}
		ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
		defer cancel()
		if err := services.CreateMeetup(ctx, &m); err != nil {
			return c.Status(503).JSON(fiber.Map{"error": "Could not save the meetup"})
		}
		event, _ := json.Marshal(ws.MeetupCreated{Type: "MEETUP_CREATED", Data: m})
		hub.Broadcast(event)
		return c.Status(201).JSON(fiber.Map{"meetup": m})
	}
}
