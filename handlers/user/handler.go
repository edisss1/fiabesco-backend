package user

import (
	"context"
	"github.com/edisss1/fiabesco-backend/db"
	"github.com/edisss1/fiabesco-backend/types"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var collection *mongo.Collection

func GetUserData(c *fiber.Ctx) error {
	var body struct {
		Email string `json:"email"`
	}

	var user types.User

	if err := c.BodyParser(&body); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	if body.Email == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Body is required"})
	}

	collection = db.Database.Collection("users")
	filter := bson.M{"email": body.Email}

	err := collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid credentials"})
	}
	user.Password = ""

	return c.Status(200).JSON(user)
}

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

func GetProfileData(c *fiber.Ctx) error {
	id := c.Params("_id")
	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var user types.User

	collection = db.Database.Collection("users")

	filter := bson.M{"_id": objectID}

	err = collection.FindOne(context.Background(), filter).Decode(&user)

	user.Password = ""
	user.Email = ""

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "User not found"})
	}

	return c.Status(200).JSON(user)

}
