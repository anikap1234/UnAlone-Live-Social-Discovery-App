package auth

import (
    "context"
    "strings"

    "github.com/gofiber/fiber/v2"
    cfgpkg "unalone/backend/config"
    authsvc "unalone/backend/auth"
)

type verifyReq struct {
    Email string `json:"email"`
    Code  string `json:"code"`
}

func VerifyOTPHandler(cfg *cfgpkg.Config) fiber.Handler {
    return func(c *fiber.Ctx) error {
        var req verifyReq
        if err := c.BodyParser(&req); err != nil {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
        }
        email := strings.TrimSpace(req.Email)
        code := strings.TrimSpace(req.Code)
        if email == "" || code == "" {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing"})
        }
        ok, err := authsvc.VerifyOTP(context.Background(), email, code)
        if err != nil || !ok {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid code"})
        }
        token, err := authsvc.GenerateToken(cfg.JWTSecret, email)
        if err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "token"})
        }
        return c.JSON(fiber.Map{"token": token})
    }
}
