package auth

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"time"
	authsvc "unalone/backend/auth"
	"unalone/backend/config"
	"unalone/backend/mailer"
	"unalone/backend/utils"
)

func SendOTPHandler(cfg *config.Config) fiber.Handler {
	delivery, configErr := mailer.New(cfg.OTPMode, cfg.SMTP)
	return func(c *fiber.Ctx) error {
		if configErr != nil {
			return c.Status(503).JSON(fiber.Map{"error": "Email delivery is not configured"})
		}
		var req struct {
			Email string `json:"email"`
		}
		if c.BodyParser(&req) != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}
		email, valid := utils.NormalizeEmail(req.Email)
		if !valid {
			return c.Status(400).JSON(fiber.Map{"error": "Enter a valid email address"})
		}
		ctx, cancel := context.WithTimeout(c.UserContext(), 5*time.Second)
		defer cancel()
		_, err := authsvc.SendOTP(ctx, email, delivery)
		if errors.Is(err, authsvc.ErrRateLimited) {
			return c.Status(429).JSON(fiber.Map{"error": err.Error()})
		}
		if err != nil {
			return c.Status(503).JSON(fiber.Map{"error": "Could not send a code. Please try again"})
		}
		return c.JSON(fiber.Map{"ok": true, "delivery": delivery.Mode()})
	}
}
