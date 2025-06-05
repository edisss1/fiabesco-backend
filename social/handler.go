package social

import (
	"context"
	"github.com/edisss1/fiabesco-backend/db"
	"github.com/edisss1/fiabesco-backend/types"
	"github.com/edisss1/fiabesco-backend/utils"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

var collection *mongo.Collection

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

func BlockUser(c *fiber.Ctx) error {
	id := c.Params("userID")
	userID, err := utils.ParseHexID(id)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid user ID in params")
	}

	var body struct {
		BlockedID string `json:"blockedID"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}

	collection = db.Database.Collection("blocked_users")

	blockedID, err := utils.ParseHexID(body.BlockedID)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid blocked user ID")
	}

	filter := bson.M{"userID": userID, "blockedID": blockedID}
	count, err := collection.CountDocuments(context.Background(), filter)

	if err != nil {
		return utils.RespondWithError(c, 500, "Database error: "+err.Error())
	}

	if count > 0 {
		return utils.RespondWithError(c, 400, "User already blocked")
	}

	blocked := types.Block{
		UserID:    userID,
		BlockedID: blockedID,
		CreatedAt: time.Now(),
	}

	_, err = collection.InsertOne(context.Background(), blocked)
	if err != nil {
		return utils.RespondWithError(c, 500, "Failed to block the user")
	}

	return c.Status(200).JSON(fiber.Map{"msg": "User blocked"})

}

func UnblockUser(c *fiber.Ctx) error {
	id := c.Params("userID")
	userID, err := utils.ParseHexID(id)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid user ID in params")
	}

	var body struct {
		BlockedID string `json:"blockedID"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}

	blockedID, err := utils.ParseHexID(body.BlockedID)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid blocked user ID")
	}

	collection = db.Database.Collection("blocked_users")
	filter := bson.M{"userID": userID, "blockedID": blockedID}

	res, err := collection.DeleteOne(context.Background(), filter)

	if res.DeletedCount == 0 {
		return utils.RespondWithError(c, 404, "User not found")
	}

	if err != nil {
		return utils.RespondWithError(c, 500, "Failed to unblock the user")
	}

	return c.Status(200).JSON(fiber.Map{"msg": "User unblocked"})

}
