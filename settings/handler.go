package settings

import "github.com/gofiber/fiber/v2"

func MockHandler(c *fiber.Ctx) error {
	return c.SendString("Hello, World!")
}
