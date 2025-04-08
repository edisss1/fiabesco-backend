package user

import (
	"context"
	"github.com/edisss1/fiabesco-backend/db"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var collection *mongo.Collection

func UpdatePhotoURL(c *fiber.Ctx) error {
	id := c.Params("_id")

	var body struct {
		PhotoURL string `json:"photoURL"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if body.PhotoURL == "" {
		return c.Status(400).JSON(fiber.Map{"error": "URL is required"})
	}

	objectID, err := primitive.ObjectIDFromHex(id)

	collection = db.Database.Collection("users")

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"photoURL": body.PhotoURL}}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"msg": "PFP updated successfully"})
}
