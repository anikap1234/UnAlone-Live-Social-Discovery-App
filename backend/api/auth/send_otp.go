package auth

import (
    "context"
    "strings"

    "github.com/gofiber/fiber/v2"
    cfgpkg "unalone/backend/config"
    authsvc "unalone/backend/auth"
)

type sendOTPReq struct {
    Email string `json:"email"`
}

func SendOTPHandler(cfg *cfgpkg.Config) fiber.Handler {
    return func(c *fiber.Ctx) error {
        var req sendOTPReq
        if err := c.BodyParser(&req); err != nil {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
        }
        email := strings.TrimSpace(req.Email)
        if email == "" {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email required"})
        }
        ctx := context.Background()
        if _, err := authsvc.SendOTP(ctx, email, cfg.OTPMode); err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to send"})
        }
        return c.JSON(fiber.Map{"ok": true})
    }
}
