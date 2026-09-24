package auth

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"regexp"
	"time"
	authsvc "unalone/backend/auth"
	"unalone/backend/config"
	"unalone/backend/services"
	"unalone/backend/utils"
)

var codePattern = regexp.MustCompile("^[0-9]{6}$")

func VerifyOTPHandler(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			Email string `json:"email"`
			OTP   string `json:"otp"`
		}
		if c.BodyParser(&req) != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}
		email, valid := utils.NormalizeEmail(req.Email)
		if !valid || !codePattern.MatchString(req.OTP) {
			return c.Status(400).JSON(fiber.Map{"error": "Enter your email and six-digit code"})
		}
		ctx, cancel := context.WithTimeout(c.UserContext(), 10*time.Second)
		defer cancel()
		ok, err := authsvc.VerifyOTP(ctx, email, req.OTP)
		if errors.Is(err, authsvc.ErrRateLimited) {
			return c.Status(429).JSON(fiber.Map{"error": err.Error()})
		}
		if err != nil {
			return c.Status(503).JSON(fiber.Map{"error": "Sign-in is temporarily unavailable"})
		}
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "The code is incorrect or has expired"})
		}
		user, err := services.FindOrCreateUser(ctx, email)
		if err != nil {
			return c.Status(503).JSON(fiber.Map{"error": "Could not load your account. Request a new code and retry"})
		}
		token, err := authsvc.GenerateToken(cfg.JWTSecret, user.ID, user.Email)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Could not create your session"})
		}
		c.Cookie(&fiber.Cookie{Name: authsvc.CookieName, Value: token, HTTPOnly: true, Secure: cfg.Env == "production", SameSite: "Lax", Path: "/", MaxAge: 72 * 3600})
		return c.JSON(fiber.Map{"token": token, "user": user})
	}
}
