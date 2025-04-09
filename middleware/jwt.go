package middleware

import (
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"os"
)

func RequireJWT(c *fiber.Ctx) error {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(os.Getenv("JWT_SECRET"))},
		ContextKey: "jwt",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": true,
				"msg":   err.Error()})
		},
	})(c)
}
