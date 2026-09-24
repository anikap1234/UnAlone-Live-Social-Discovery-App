package auth

import (
	"github.com/gofiber/fiber/v2"
	"strings"
)

const CookieName = "unalone_session"

func JWTMiddleware(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		value := c.Cookies(CookieName)
		if header := c.Get("Authorization"); header != "" {
			parts := strings.Fields(header)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				return c.Status(401).JSON(fiber.Map{"error": "Invalid authorization header"})
			}
			value = parts[1]
		}
		claims, err := ParseToken(secret, value)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "Please sign in again"})
		}
		c.Locals("userId", claims.Subject)
		c.Locals("email", claims.Email)
		return c.Next()
	}
}
