package auth

import (
	"context"
	"github.com/edisss1/fiabesco-backend/db"
	"github.com/edisss1/fiabesco-backend/types"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

var collection *mongo.Collection

func SignUp(c *fiber.Ctx) error {
	var input types.User

	collection = db.Database.Collection("users")

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	filter := bson.M{"email": input.Email}

	var existingUser types.User

	err := collection.FindOne(context.Background(), filter).Decode(&existingUser)

	if err == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	hash, _ := HashPassword(input.Password)
	input.Password = hash

	_, err = collection.InsertOne(context.Background(), input)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "DB error"})
	}

	token, err := GenerateJWT(input.Email)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "An error occurred while generating token"})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour * 7),
		HTTPOnly: true,
	})

	input.Token = token

	return c.Status(201).JSON(input)
}

func Login(c *fiber.Ctx) error {

	var input types.User

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	var user types.User

	collection = db.Database.Collection("users")

	filter := bson.M{"email": input.Email}

	err := collection.FindOne(context.Background(), filter).Decode(&user)

	if err != nil || !CheckPasswordHash(input.Password, user.Password) {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	token, _ := GenerateJWT(user.Email)

	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour * 7),
		HTTPOnly: true,
	})

	return c.Status(200).JSON(user)
}
