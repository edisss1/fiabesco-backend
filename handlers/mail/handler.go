package mail

import (
	"fmt"
	"github.com/edisss1/fiabesco-backend/utils"
	"github.com/gofiber/fiber/v2"
	"gopkg.in/gomail.v2"
	"os"
)

func SendEmail(c *fiber.Ctx) error {
	appEmail := os.Getenv("APP_EMAIL")
	appPassword := os.Getenv("APP_PASSWORD")

	var body struct {
		FromEmail    string `json:"fromEmail"`
		FromUserName string `json:"fromUserName"`
		ToEmail      string `json:"toEmail"`
		Subject      string `json:"subject"`
		Body         string `json:"body"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, fiber.StatusBadRequest, err.Error())
	}

	m := gomail.NewMessage()
	m.SetHeader("From", appEmail)
	m.SetHeader("To", body.ToEmail)
	m.SetHeader("Subject", body.Subject)

	composedBody := fmt.Sprintf("Sent from user: %s\n\n%s", body.FromUserName+" "+body.FromEmail, body.Body)

	m.SetBody("text/plain", composedBody)

	d := gomail.NewDialer("smtp.gmail.com", 587, appEmail, appPassword)
	if err := d.DialAndSend(m); err != nil {
		return utils.RespondWithError(c, fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(200).JSON(fiber.Map{"msg": "Email sent successfully"})
}
