package user

import (
	"context"
	"errors"
	"github.com/edisss1/fiabesco-backend/db"
	"github.com/edisss1/fiabesco-backend/types"
	"github.com/edisss1/fiabesco-backend/utils"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func BlockUser(c *fiber.Ctx) error {
	collection = db.Database.Collection("users")

	id := c.Params("_id")
	userID, err := utils.ParseHexID(id)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid ID")
	}

	var user types.User

	var blockedUser struct {
		ID primitive.ObjectID `json:"id" bson:"id"`
	}

	if err := c.BodyParser(&blockedUser); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}

	filter := bson.M{"_id": userID}
	err = collection.FindOne(context.Background(), filter).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return utils.RespondWithError(c, 404, "User not found")
	}

	for _, blockedID := range user.BlockedUsers {
		if blockedID == blockedUser.ID {
			return utils.RespondWithError(c, 400, "User is already blocked")
		}
	}

	user.BlockedUsers = append(user.BlockedUsers, blockedUser.ID)

	update := bson.M{"$set": bson.M{"blockedUsers": user.BlockedUsers}}
	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return utils.RespondWithError(c, 500, "Failed to update user")
	}

	return c.Status(200).JSON(fiber.Map{"msg": "User blocked"})

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

func GetFollowing(c *fiber.Ctx) error {
	id := c.Params("_id")
	userID, err := utils.ParseHexID(id)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid user ID")
	}

	collection := db.Database.Collection("users")

	var user types.User
	userFilter := bson.M{"_id": userID}

	err = collection.FindOne(context.Background(), userFilter).Decode(&user)
	if err != nil {
		return utils.RespondWithError(c, 404, "User not found")
	}

	var followedUserIDs []primitive.ObjectID
	for _, followedID := range user.FollowedUsers {
		fid, err := utils.ParseHexID(followedID)
		if err != nil {
			return utils.RespondWithError(c, 400, "Invalid followed user ID")
		}
		followedUserIDs = append(followedUserIDs, fid)
	}

	filter := bson.M{"_id": bson.M{"$in": followedUserIDs}}

	projection := bson.M{
		"firstName": 1,
		"lastName":  1,
		"handle":    1,
		"photoURL":  1,
		"bio":       1,
		"_id":       1,
	}

	opts := options.Find().SetProjection(projection)

	cursor, err := collection.Find(context.Background(), filter, opts)
	if err != nil {
		return utils.RespondWithError(c, 500, "Database error: "+err.Error())
	}
	defer cursor.Close(context.Background())

	var followed []types.User
	if err := cursor.All(context.Background(), &followed); err != nil {
		return utils.RespondWithError(c, 500, "Failed to decode followed users")
	}

	return c.Status(200).JSON(followed)
}

func FollowUser(c *fiber.Ctx) error {
	id := c.Params("_id")
	userID, err := utils.ParseHexID(id)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid user ID")
	}

	var body struct {
		ID string `json:"id"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, 400, "Missing or invalid request body")
	}

	collection := db.Database.Collection("users")

	var user types.User
	filter := bson.M{"_id": userID}
	err = collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		return utils.RespondWithError(c, 404, "User not found")
	}

	for _, followedID := range user.FollowedUsers {
		if followedID == body.ID {
			return utils.RespondWithError(c, 400, "Already following this user")
		}
	}

	update := bson.M{"$push": bson.M{"followedUsers": body.ID}}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return utils.RespondWithError(c, 500, "Failed to follow the user")
	}

	incrementFollowers := bson.M{"$inc": bson.M{"followersCount": 1}}
	incrementFollowed := bson.M{"$inc": bson.M{"followingCount": 1}}

	followerID, err := utils.ParseHexID(body.ID)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid follower ID")
	}

	followerFilter := bson.M{"_id": followerID}

	_, err = collection.UpdateOne(context.Background(), filter, incrementFollowers)
	_, err = collection.UpdateOne(context.Background(), followerFilter, incrementFollowed)

	return c.Status(200).JSON(fiber.Map{"msg": "Successfully followed the user"})
}
