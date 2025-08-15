package repost

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
	"net/http"
	"time"
)

var collection *mongo.Collection

func Repost(c *fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return utils.RespondWithError(c, http.StatusBadRequest, "Invalid user ID")
	}

	var body struct {
		PostID        primitive.ObjectID `json:"postID" bson:"postID"`
		RepostCaption string             `json:"repostCaption" bson:"repostCaption"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, http.StatusBadRequest, "Invalid request body")
	}

	repost := types.Repost{
		RepostedBy:    userID,
		PostID:        body.PostID,
		RepostCaption: body.RepostCaption,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	collection = db.Database.Collection("reposts")

	_, err = collection.InsertOne(context.Background(), repost)
	if err != nil {
		return utils.RespondWithError(c, http.StatusInternalServerError, "Error creating repost: "+err.Error())
	}

	collection = db.Database.Collection("posts")

	filter := bson.M{"_id": body.PostID}
	update := bson.M{"$inc": bson.M{"repostCount": 1}}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return utils.RespondWithError(c, http.StatusInternalServerError, "Error updating post: "+err.Error())
	}

	return c.Status(200).JSON(fiber.Map{"msg": "Repost created successfully"})
}

func EditRepostCaption(c *fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return utils.RespondWithError(c, http.StatusBadRequest, "Invalid user ID")
	}
	var body struct {
		NewRepostCaption string             `json:"newRepostCaption" bson:"newRepostCaption"`
		RepostID         primitive.ObjectID `json:"repostID" bson:"repostID"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, http.StatusBadRequest, "Invalid request body")
	}

	collection = db.Database.Collection("reposts")
	var repost types.Repost

	filter := bson.M{"_id": body.RepostID}

	err = collection.FindOne(context.Background(), filter).Decode(&repost)
	if err != nil {
		return utils.RespondWithError(c, http.StatusInternalServerError, "Error finding repost: "+err.Error())
	}

	if errors.Is(err, mongo.ErrNoDocuments) {
		return utils.RespondWithError(c, http.StatusNotFound, "Repost not found")
	}

	if repost.RepostedBy != userID {
		return utils.RespondWithError(c, http.StatusUnauthorized, "You are not authorized to update this repost")
	}

	update := bson.M{"$set": bson.M{"repostCaption": body.NewRepostCaption}}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return utils.RespondWithError(c, http.StatusInternalServerError, "Error updating repost caption: "+err.Error())
	}

	return c.Status(200).JSON(fiber.Map{"msg": "Repost caption updated successfully"})
}

func DeleteRepost(c *fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return utils.RespondWithError(c, http.StatusBadRequest, "Invalid user ID")
	}

	var body struct {
		RepostID primitive.ObjectID `json:"repostID" bson:"repostID"`
	}

	collection = db.Database.Collection("reposts")

	filter := bson.M{"_id": body.RepostID}

	var repost types.Repost

	err = collection.FindOne(context.Background(), filter).Decode(&repost)

	if err != nil {
		return utils.RespondWithError(c, http.StatusInternalServerError, "Error finding repost: "+err.Error())
	}

	if repost.RepostedBy != userID {
		return utils.RespondWithError(c, http.StatusUnauthorized, "You are not authorized to delete this repost")
	}

	_, err = collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return utils.RespondWithError(c, http.StatusInternalServerError, "Error deleting repost: "+err.Error())
	}

	collection = db.Database.Collection("posts")

	filter = bson.M{"_id": repost.PostID}
	update := bson.M{"$inc": bson.M{"repostCount": -1}}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return utils.RespondWithError(c, http.StatusInternalServerError, "Error updating post: "+err.Error())
	}

	return c.Status(200).JSON(fiber.Map{"msg": "Repost deleted successfully"})

}
