package post

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
	"log"
	"strconv"
	"strings"
	"time"
)

// TODO: find a way to optimize post fetching )

var collection *mongo.Collection

func CreatePost(c *fiber.Ctx) error {
	id := c.Params("_id")
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	collection = db.Database.Collection("posts")

	var body struct {
		Caption string `json:"caption"`
	}
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}

	newPost := types.Post{
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Caption:   body.Caption,
	}

	res, err := collection.InsertOne(context.Background(), newPost)
	newPost.ID = res.InsertedID.(primitive.ObjectID)

	return c.Status(201).JSON(newPost)
}

func GetPostsByUser(c *fiber.Ctx) error {
	id := c.Params("_id")
	userID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid ID")
	}

	var posts []types.Post

	collection = db.Database.Collection("posts")

	filter := bson.M{"userID": userID}

	cursor, err := collection.Find(context.Background(), filter)

	if err != nil {
		return err
	}

	defer func() {
		if err := cursor.Close(context.Background()); err != nil {

			log.Println(err)
		}
	}()

	for cursor.Next(context.Background()) {
		var post types.Post

		if err := cursor.Decode(&post); err != nil {
			return err
		}

		posts = append(posts, post)
	}

	collection = db.Database.Collection("users")

	var user types.User

	userFilter := bson.M{"_id": userID}

	err = collection.FindOne(context.Background(), userFilter).Decode(&user)
	if err != nil {
		return utils.RespondWithError(c, 404, "User not found")
	}
	user.Password = ""
	user.Email = ""

	return c.Status(200).JSON(fiber.Map{"posts": posts, "user": user})
}

func DeletePost(c *fiber.Ctx) error {
	postID := c.Params("postID")

	objectID, err := primitive.ObjectIDFromHex(postID)

	postsCollection := db.Database.Collection("posts")

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	filter := bson.M{"_id": objectID}

	_, err = postsCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Error deleting the post"})
	}

	return c.Status(200).JSON(fiber.Map{"msg": "Post was deleted successfully"})
}

func GetPost(c *fiber.Ctx) error {
	id := c.Params("postID")
	postID, err := utils.ParseHexID(id)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid ID")
	}

	var res struct {
		post types.Post
		user types.User
	}

	collection = db.Database.Collection("posts")

	postFilter := bson.M{"_id": postID}

	err = collection.FindOne(context.Background(), postFilter).Decode(&res.post)
	if err != nil {
		return utils.RespondWithError(c, 500, "Failed to decode post")
	}

	userFilter := bson.M{"_id": res.post.UserID}
	collection = db.Database.Collection("users")

	err = collection.FindOne(context.Background(), userFilter).Decode(&res.user)
	res.user.Password = ""
	res.user.Email = ""

	return c.Status(200).JSON(fiber.Map{"post": res.post, "user": res.user})
}

func GetFeedPosts(c *fiber.Ctx) error {
	sampleSizeParam := c.Query("sample", "0")
	limitParam := c.Query("limit", "10")
	skipParam := c.Query("skip", "0")

	sampleSize, _ := strconv.Atoi(sampleSizeParam)
	limit, _ := strconv.Atoi(limitParam)
	skip, _ := strconv.Atoi(skipParam)

	postsCollection := db.Database.Collection("posts")
	usersCollection := db.Database.Collection("users")

	var feedItems []struct {
		Post types.Post `json:"post"`
		User types.User `json:"user"`
	}

	if sampleSize > 0 {
		effectiveSampleSize := sampleSize + skip + limit

		pipeline := []bson.M{
			{"$sample": bson.M{"size": effectiveSampleSize}},
			{"$sort": bson.M{"createdAt": -1}},
			{"$skip": skip},
			{"$limit": limit},
		}

		cursor, err := postsCollection.Aggregate(context.Background(), pipeline)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch posts"})
		}
		defer cursor.Close(context.Background())

		for cursor.Next(context.Background()) {
			var feedItem struct {
				Post types.Post `json:"post"`
				User types.User `json:"user"`
			}
			if err := cursor.Decode(&feedItem.Post); err != nil {
				return utils.RespondWithError(c, 500, "Failed to decode post: "+err.Error())
			}

			filter := bson.M{"_id": feedItem.Post.UserID}
			if err := usersCollection.FindOne(context.Background(), filter).Decode(&feedItem.User); err != nil {
				return utils.RespondWithError(c, 500, "Failed to fetch user: "+err.Error())
			}
			feedItem.User.Password = ""
			feedItem.User.Email = ""

			feedItems = append(feedItems, feedItem)
		}
	} else {
		opts := options.Find().
			SetLimit(int64(limit)).
			SetSkip(int64(skip)).
			SetSort(bson.D{{Key: "createdAt", Value: -1}})

		cursor, err := postsCollection.Find(context.Background(), bson.M{}, opts)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch posts"})
		}
		defer cursor.Close(context.Background())

		for cursor.Next(context.Background()) {
			var feedItem struct {
				Post types.Post `json:"post"`
				User types.User `json:"user"`
			}
			if err := cursor.Decode(&feedItem.Post); err != nil {
				return utils.RespondWithError(c, 500, "Failed to decode post: "+err.Error())
			}

			filter := bson.M{"_id": feedItem.Post.UserID}
			if err := usersCollection.FindOne(context.Background(), filter).Decode(&feedItem.User); err != nil {
				return utils.RespondWithError(c, 500, "Failed to fetch user: "+err.Error())
			}
			feedItem.User.Password = ""
			feedItem.User.Email = ""

			feedItems = append(feedItems, feedItem)
		}
	}

	totalCount, err := postsCollection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to count posts"})
	}

	hasMore := false
	if sampleSize > 0 {
		hasMore = len(feedItems) >= limit
	} else {
		hasMore = (skip + limit) < int(totalCount)
	}

	return c.Status(200).JSON(fiber.Map{
		"feedItems": feedItems,
		"hasMore":   hasMore,
		"nextSkip":  skip + limit,
	})
}

func UpdatePostCaption(c *fiber.Ctx) error {
	id := c.Params("_id")

	var body struct {
		Caption string `json:"caption"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if body.Caption == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Caption is required"})
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	collection := db.Database.Collection("posts")
	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"caption": body.Caption}}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"msg": "Post updated successfully"})
}

func LikePost(c *fiber.Ctx) error {
	var body struct {
		UserID string `json:"userID"`
		PostID string `json:"postID"`
	}

	likesCollection := db.Database.Collection("likes")
	postsCollection := db.Database.Collection("posts")
	usersCollection := db.Database.Collection("users")

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}

	postID, err := utils.ParseHexID(body.PostID)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid post ID")
	}
	userID, err := utils.ParseHexID(body.UserID)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid user ID")
	}

	postFilter := bson.M{"_id": postID}
	likeFilter := bson.M{"postID": postID, "userID": userID}
	userFilter := bson.M{"_id": userID}

	var like types.Like
	var post types.Post
	var update bson.M
	var user types.User

	err = usersCollection.FindOne(context.Background(), userFilter).Decode(&user)
	if err != nil {
		return utils.RespondWithError(c, 404, "User not found")
	}

	userName := strings.TrimSpace(user.FirstName + " " + user.LastName)

	err = postsCollection.FindOne(context.Background(), postFilter).Decode(&post)
	if err != nil {
		return utils.RespondWithError(c, 500, "Failed to retrieve post: "+err.Error())
	}

	err = likesCollection.FindOne(context.Background(), likeFilter).Decode(&like)

	if err == nil {
		_, err = likesCollection.DeleteOne(context.Background(), likeFilter)
		if err != nil {
			return utils.RespondWithError(c, 500, "Failed to unlike the post: "+err.Error())
		}

		update = bson.M{"$inc": bson.M{"likesCount": -1}}

		_, err = postsCollection.UpdateOne(context.Background(), postFilter, update)
		if err != nil {
			return utils.RespondWithError(c, 500, "Failed to update post like count: "+err.Error())
		}

		err = postsCollection.FindOne(context.Background(), postFilter).Decode(&post)
		if err != nil {
			return utils.RespondWithError(c, 500, "Failed to retrieve updated post: "+err.Error())
		}

		return c.Status(200).JSON(fiber.Map{
			"likesCount": post.LikesCount,
		})
	}

	update = bson.M{"$inc": bson.M{"likesCount": 1}}

	newLike := types.Like{
		PostID:    postID,
		UserID:    userID,
		UserName:  userName,
		CreatedAt: time.Now(),
	}

	_, err = likesCollection.InsertOne(context.Background(), newLike)
	if err != nil {
		return utils.RespondWithError(c, 500, "Failed to add like: "+err.Error())
	}

	_, err = postsCollection.UpdateOne(context.Background(), postFilter, update)
	if err != nil {
		return utils.RespondWithError(c, 500, "Failed to update post like count: "+err.Error())
	}

	err = postsCollection.FindOne(context.Background(), postFilter).Decode(&post)
	if err != nil {
		return utils.RespondWithError(c, 500, "Failed to retrieve updated post: "+err.Error())
	}

	return c.Status(200).JSON(fiber.Map{
		"likesCount": post.LikesCount,
	})
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

	return c.Status(201).JSON(newComment)

}
