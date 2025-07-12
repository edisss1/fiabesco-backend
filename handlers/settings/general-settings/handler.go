package settings

import (
	"context"
	"github.com/edisss1/fiabesco-backend/db"
	"github.com/edisss1/fiabesco-backend/handlers/auth"
	"github.com/edisss1/fiabesco-backend/types"
	"github.com/edisss1/fiabesco-backend/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var collection *mongo.Collection

func ChangeFirstName(c *fiber.Ctx) error {
	locals := c.Locals("jwt").(*jwt.Token)
	claims := locals.Claims.(jwt.MapClaims)
	id := claims["id"].(string)

	userID, err := utils.ParseHexID(id)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid ID")
	}

	var body struct {
		FirstName string `json:"firstName"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}

	filter := bson.M{"_id": userID}
	update := bson.M{"firstName": body.FirstName}

	collection = db.Database.Collection("users")

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return utils.RespondWithError(c, 500, "Error updating first name "+err.Error())
	}

	return c.Status(200).JSON(fiber.Map{"msg": "First name updated successfully"})
}

func ChangeLastName(c *fiber.Ctx) error {
	locals := c.Locals("jwt").(*jwt.Token)
	claims := locals.Claims.(jwt.MapClaims)
	id := claims["id"].(string)

	userID, err := utils.ParseHexID(id)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid ID")
	}

	var body struct {
		LastName string `json:"lastName"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}

	filter := bson.M{"_id": userID}
	update := bson.M{"lastName": body.LastName}

	collection = db.Database.Collection("users")

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return utils.RespondWithError(c, 500, "Error updating last name "+err.Error())
	}

	return c.Status(200).JSON(fiber.Map{"msg": "Last name updated successfully"})
}

func ChangeEmail(c *fiber.Ctx) error {
	locals := c.Locals("jwt").(*jwt.Token)
	claims := locals.Claims.(jwt.MapClaims)
	id := claims["id"].(string)

	userID, err := utils.ParseHexID(id)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid ID")
	}

	var body struct {
		Email string `json:"email"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}

	filter := bson.M{"_id": userID}
	update := bson.M{"email": body.Email}

	collection = db.Database.Collection("users")

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return utils.RespondWithError(c, 500, "Error updating email "+err.Error())
	}

	return c.Status(200).JSON(fiber.Map{"msg": "Email updated successfully"})
}

func ChangeHandle(c *fiber.Ctx) error {
	locals := c.Locals("jwt").(*jwt.Token)
	claims := locals.Claims.(jwt.MapClaims)
	id := claims["id"].(string)

	userID, err := utils.ParseHexID(id)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid ID")
	}

	var body struct {
		Handle string `json:"handle"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}

	filter := bson.M{"_id": userID}
	update := bson.M{"handle": body.Handle}

	collection = db.Database.Collection("users")

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return utils.RespondWithError(c, 500, "Error updating handle "+err.Error())
	}

	return c.Status(200).JSON(fiber.Map{"msg": "Handle updated successfully"})

}

func ChangePassword(c *fiber.Ctx) error {
	locals := c.Locals("jwt").(*jwt.Token)
	claims := locals.Claims.(jwt.MapClaims)
	id := claims["id"].(string)

	userID, err := utils.ParseHexID(id)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid ID")
	}

	var body struct {
		Password string `json:"password"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}

	var user types.User
	filter := bson.M{"_id": userID}

	collection = db.Database.Collection("users")

	err = collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		return utils.RespondWithError(c, 404, "User not found")
	}
	matches := auth.CheckPasswordHash(user.Password, body.Password)
	if matches {
		return utils.RespondWithError(c, 400, "New password must be different from old password")
	}

	hashedPassword := auth.HashPassword(body.Password)

	filter = bson.M{"_id": userID}
	update := bson.M{"$set": bson.M{"password": hashedPassword}}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return utils.RespondWithError(c, 500, "Error updating password "+err.Error())
	}

	return c.Status(200).JSON(fiber.Map{"msg": "Password updated successfully"})
}
