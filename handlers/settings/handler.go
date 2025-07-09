package settings

import (
	"context"
	"github.com/edisss1/fiabesco-backend/db"
	"github.com/edisss1/fiabesco-backend/types"
	"github.com/edisss1/fiabesco-backend/utils"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

func SaveSettings(c *fiber.Ctx) error {
	id := c.Params("userID")
	userID, err := utils.ParseHexID(id)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid user ID")
	}

	var body types.Settings

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}

	collection := db.Database.Collection("users")

	var user types.User
	filter := bson.M{"_id": userID}

	if err := collection.FindOne(context.Background(), filter).Decode(&user); err != nil {
		return utils.RespondWithError(c, 404, "User not found")
	}

	if body.Theme != "" {
		_, err = collection.UpdateOne(context.Background(), filter, bson.M{"$set": bson.M{"settings.theme": body.Theme}})

		if err != nil {
			return utils.RespondWithError(c, 500, "Failed to update user settings")
		}

	}

	if body.Language != "" {

		_, err = collection.UpdateOne(context.Background(), filter, bson.M{"$set": bson.M{"settings.language": body.Language}})

		if err != nil {
			return utils.RespondWithError(c, 500, "Failed to update user settings")
		}
	}

	return c.Status(200).JSON(fiber.Map{"msg": "Settings saved successfully"})

}
