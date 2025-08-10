package comments

import (
	"context"
	"github.com/edisss1/fiabesco-backend/db"
	"github.com/edisss1/fiabesco-backend/types"
	"github.com/edisss1/fiabesco-backend/utils"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"strconv"
	"time"
)

var collection *mongo.Collection

type CommentRes struct {
	Comment  types.Comment `json:"comment"`
	UserName string        `json:"userName"`
	PhotoURL string        `json:"photoURL"`
}

func CommentPost(c *fiber.Ctx) error {
	id := c.Params("postID")
	postID, err := utils.ParseHexID(id)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid ID")
	}

	var body struct {
		UserID  string `json:"userID"`
		Content string `json:"content"`
	}
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}

	if body.Content == "" {
		return utils.RespondWithError(c, 400, "Content cannot be an empty string")
	}

	userID, err := utils.ParseHexID(body.UserID)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid user ID")
	}

	collection = db.Database.Collection("comments")

	newComment := types.Comment{
		Content:   body.Content,
		PostID:    postID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	res, err := collection.InsertOne(context.Background(), newComment)
	if err != nil {
		return utils.RespondWithError(c, 500, "Error inserting comment")
	}

	newComment.ID = res.InsertedID.(primitive.ObjectID)

	collection = db.Database.Collection("posts")
	filter := bson.M{"_id": postID}
	update := bson.M{"$inc": bson.M{"commentsCount": 1}}

	_, err = collection.UpdateOne(context.Background(), filter, update)

	return c.Status(201).JSON(newComment)

}

func GetComments(c *fiber.Ctx) error {
	id := c.Params("postID")
	postID, err := utils.ParseHexID(id)
	if err != nil {
		return utils.RespondWithError(c, http.StatusBadRequest, "Invalid ID")
	}

	pageParam := c.Query("page", "1")
	l := 10
	p, _ := strconv.Atoi(pageParam)
	skip := int64(p*l - l)
	limit := int64(l)

	pipeline := utils.NewPipeline().
		Match(bson.D{{"postID", postID}}).
		Sort("createdAt", -1).
		Skip(skip).Limit(limit).
		Lookup("users", "userID", "_id", "user").
		Unwind("$user", true).
		Project(bson.D{
			{"comment", "$$ROOT"},
			{"userName", bson.D{{"$concat", bson.A{"$user.firstName", " ", "$user.lastName"}}}},
			{"photoURL", "$user.photoURL"},
		}).Build()

	collection = db.Database.Collection("comments")
	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get comments "+err.Error())
	}

	var comments []CommentRes

	for cursor.Next(context.Background()) {
		var comment CommentRes
		if err := cursor.Decode(&comment); err != nil {
			return utils.RespondWithError(c, 500, "Failed to decode comment")
		}
		if comment.PhotoURL != "" {
			comment.PhotoURL = utils.BuildImgURL(comment.PhotoURL)
		}
		comments = append(comments, comment)
	}

	return c.Status(http.StatusOK).JSON(comments)

}

func EditComment(c *fiber.Ctx) error {
	id := c.Params("commentID")
	commentID, err := utils.ParseHexID(id)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid ID")
	}

	var body struct {
		UserID     string `json:"userID"`
		NewContent string `bson:"newContent"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}
	collection = db.Database.Collection("comments")

	var comment types.Comment
	filter := bson.M{"_id": commentID}
	update := bson.M{"$set": bson.M{"content": body.NewContent}}

	err = collection.FindOne(context.Background(), filter).Decode(&comment)
	if err != nil {
		return utils.RespondWithError(c, 500, "Error decoding comment"+err.Error())
	}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return utils.RespondWithError(c, 500, "Error updating comment"+err.Error())
	}

	return c.Status(200).JSON(fiber.Map{"msg": "Comment updated successfully"})
}

func DeleteComment(c *fiber.Ctx) error {
	id := c.Params("commentID")
	commentID, err := utils.ParseHexID(id)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid ID")
	}

	collection := db.Database.Collection("comments")
	filter := bson.M{"_id": commentID}

	var comment types.Comment

	err = collection.FindOne(context.Background(), filter).Decode(&comment)
	if err != nil {
		return utils.RespondWithError(c, 500, "Error decoding comment "+err.Error())
	}

	isOwner, err := utils.VerifyOwnership(c, comment.UserID)
	if err != nil {
		return utils.RespondWithError(c, 500, "Error verifying ownership "+err.Error())
	}

	if !isOwner {
		return utils.RespondWithError(c, 401, "Unauthorized")
	}

	collection = db.Database.Collection("posts")
	filter = bson.M{"_id": comment.PostID}
	update := bson.M{"$inc": bson.M{"commentsCount": -1}}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return utils.RespondWithError(c, 500, "Error updating post "+err.Error())
	}

	_, err = collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return utils.RespondWithError(c, 500, "Error deleting post "+err.Error())
	}

	return c.Status(200).JSON(fiber.Map{"msg": "Comment deleted successfully"})

}
