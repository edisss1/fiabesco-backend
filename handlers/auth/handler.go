package auth

import (
	"context"
	"github.com/edisss1/fiabesco-backend/db"
	"github.com/edisss1/fiabesco-backend/types"
	"github.com/edisss1/fiabesco-backend/utils"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

var collection *mongo.Collection

func SignUp(c *fiber.Ctx) error {
	var input types.User

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	collection = db.Database.Collection("users")

	var existingUser types.User

	filter := bson.M{"email": input.Email}

	err := collection.FindOne(context.Background(), filter).Decode(&existingUser)

	if err == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	handle := utils.GenerateHandle(24)

	hash := HashPassword(input.Password)
	input.Password = hash
	input.Handle = handle
	input.CreatedAt = time.Now()

	_, err = collection.InsertOne(context.Background(), input)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "DB error"})
	}

	return c.Status(201).JSON(fiber.Map{"msg": "User created"})
}

func Login(c *fiber.Ctx) error {
	var input types.User

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	collection = db.Database.Collection("users")

	filter := bson.M{"email": input.Email}

	var user types.User
	err := collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "User not found"})
	}

	if !CheckPasswordHash(user.Password, input.Password) {
		return c.Status(400).JSON(fiber.Map{"error": "Incorrect password"})
	}

	token, err := GenerateToken(input.Email)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(200).JSON(fiber.Map{"token": token})
}
