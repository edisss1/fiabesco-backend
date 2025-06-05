package user

import (
	"context"
	"github.com/edisss1/fiabesco-backend/db"
	"github.com/edisss1/fiabesco-backend/handlers/auth"
	"github.com/edisss1/fiabesco-backend/types"
	"github.com/edisss1/fiabesco-backend/utils"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"strings"
	"time"
)

var collection *mongo.Collection

type MeRes struct {
	ID             primitive.ObjectID `json:"_id" bson:"_id"`
	FirstName      string             `json:"firstName"`
	LastName       string             `json:"lastName"`
	Handle         string             `json:"handle"`
	PhotoURL       string             `json:"photoURL"`
	Bio            string             `json:"bio"`
	Settings       *types.Settings    `json:"settings"`
	CreatedAt      time.Time          `json:"createdAt"`
	FollowersCount uint32             `json:"followersCount"`
	FollowingCount uint32             `json:"followingCount"`
}

func GetUserData(c *fiber.Ctx) error {

	authHeader := c.Get("Authorization")

	if authHeader == "" {
		return utils.RespondWithError(c, 401, "Unauthorized (in auth header)")
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	token, claims, err := auth.VerifyToken(tokenStr)

	if err != nil || token == nil {
		return utils.RespondWithError(c, 401, "Unauthorized (in token)")
	}

	var user MeRes

	collection = db.Database.Collection("users")
	filter := bson.M{"email": claims.Email}

	err = collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid credentials"})
	}

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

func EditBio(c *fiber.Ctx) error {
	id := c.Params("_id")
	userID, err := utils.ParseHexID(id)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid user ID")
	}

	var body struct {
		Bio string `json:"bio"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}

	if body.Bio == "" {
		return utils.RespondWithError(c, 400, "Bio cannot be empty")
	}

	collection = db.Database.Collection("users")

	filter := bson.M{"_id": userID}
	update := bson.M{"$set": bson.M{"bio": body.Bio}}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return utils.RespondWithError(c, 500, err.Error())

	}

	return c.Status(200).JSON(fiber.Map{"newBio": body.Bio})
}
