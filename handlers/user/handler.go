package user

import (
	"bytes"
	"context"
	"encoding/base64"
	"github.com/edisss1/fiabesco-backend/db"
	"github.com/edisss1/fiabesco-backend/handlers/auth"
	"github.com/edisss1/fiabesco-backend/types"
	"github.com/edisss1/fiabesco-backend/uploads"
	"github.com/edisss1/fiabesco-backend/utils"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"io"
	"strings"
	"time"
)

var collection *mongo.Collection

type MeRes struct {
	ID             primitive.ObjectID `json:"_id" bson:"_id"`
	FirstName      string             `json:"firstName"`
	LastName       string             `json:"lastName"`
	Email          string             `json:"email"`
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
	bucket, err := gridfs.NewBucket(db.Database)
	if err != nil {
		return utils.RespondWithError(c, 500, "Failed to create bucket")
	}
	if user.PhotoURL != "" {
		fileID, err := utils.ParseHexID(user.PhotoURL)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
		}
		var buf bytes.Buffer
		downloadStream, err := bucket.OpenDownloadStream(fileID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
		}
		io.Copy(&buf, downloadStream)
		downloadStream.Close()
		base64Img := base64.StdEncoding.EncodeToString(buf.Bytes())
		user.PhotoURL = "data:image/jpeg;base64," + base64Img
	}

	return c.Status(200).JSON(user)
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
	bucket, err := gridfs.NewBucket(db.Database)
	if err != nil {
		return utils.RespondWithError(c, 500, "Failed to create bucket")
	}
	if user.PhotoURL != "" {
		fileID, err := utils.ParseHexID(user.PhotoURL)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
		}
		var buf bytes.Buffer
		downloadStream, err := bucket.OpenDownloadStream(fileID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
		}
		io.Copy(&buf, downloadStream)
		downloadStream.Close()
		base64Img := base64.StdEncoding.EncodeToString(buf.Bytes())
		user.PhotoURL = "data:image/jpeg;base64," + base64Img
	}

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

func ChangePFP(c *fiber.Ctx) error {
	id := c.Params("userID")
	userID, err := utils.ParseHexID(id)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid user ID "+err.Error())
	}

	bucket, err := gridfs.NewBucket(db.Database)
	if err != nil {
		return utils.RespondWithError(c, 500, "Failed to create bucket")
	}
	collection := db.Database.Collection("users")

	ids, err := uploads.UploadFile(c, "pfp", bucket, false)

	filter := bson.M{"_id": userID}
	update := bson.M{"$set": bson.M{"photoURL": ids[0].Hex()}}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return utils.RespondWithError(c, 500, err.Error())
	}

	return c.Status(200).JSON(fiber.Map{"msg": "PFP updated successfully" + ids[0].Hex()})

}
