package utils

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"math/rand"
)

var baseImgURL = "http://localhost:3000/images"

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func GenerateHandle(l int) string {

	b := make([]rune, l)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func ParseHexID(param string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(param)
}
func RespondWithError(c *fiber.Ctx, code int, msg string) error {
	return c.Status(code).JSON(fiber.Map{"error": msg})
}

func BuildImgURL(imageID string) string {
	return fmt.Sprintf("%s/%s", baseImgURL, imageID)
}

func VerifyOwnership(c *fiber.Ctx, ownerID primitive.ObjectID) (bool, error) {
	user := c.Locals("jwt").(*jwt.Token)

	claims := user.Claims.(jwt.MapClaims)
	userID, err := ParseHexID(claims["id"].(string))
	if err != nil {
		return false, err
	}

	fmt.Printf("Claims: %s\n", claims)
	fmt.Printf("Owner ID: %s\n", ownerID)
	fmt.Printf("User ID: %s\n", userID)

	if userID != ownerID {
		return false, nil
	}

	return true, nil

}

func GetUserID(c *fiber.Ctx) (primitive.ObjectID, error) {
	user := c.Locals("jwt").(*jwt.Token)

	claims := user.Claims.(jwt.MapClaims)
	userID, err := ParseHexID(claims["id"].(string))
	if err != nil {
		return primitive.ObjectID{}, err
	}

	return userID, nil
}
