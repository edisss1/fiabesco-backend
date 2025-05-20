package limiters

import (
	"github.com/edisss1/fiabesco-backend/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"net/http"
	"time"
)

func SettingsLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        5,
		Expiration: 10 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			id := c.Params("userID")
			userID, err := utils.ParseHexID(id)
			if err != nil {
				return ""
			}
			return userID.String()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return utils.RespondWithError(c, http.StatusTooManyRequests, "Too many change attempt. Try again later")
		},
	})
}
