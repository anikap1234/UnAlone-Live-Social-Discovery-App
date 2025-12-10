package auth

import (
    "strings"

    "github.com/gofiber/fiber/v2"
)

// JWTMiddleware extracts Bearer token and sets email in context locals
func JWTMiddleware(secret string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        auth := c.Get("Authorization")
        if auth == "" {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing auth"})
        }
        parts := strings.SplitN(auth, " ", 2)
        if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid auth"})
        }
        token := parts[1]
        claims, err := ParseToken(secret, token)
        if err != nil {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
        }
        c.Locals("email", claims.Email)
        return c.Next()
    }
}
