package post

import (
	"context"
	"fmt"
	"github.com/edisss1/fiabesco-backend/db"
	"github.com/edisss1/fiabesco-backend/types"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"strconv"
	"time"
)

var collection *mongo.Collection

func CreatePost(c *fiber.Ctx) error {
	userID := c.Params("_id")
	objectID, err := primitive.ObjectIDFromHex(userID)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var post types.Post
	var user types.User

	postsCollection := db.Database.Collection("posts")
	usersCollection := db.Database.Collection("users")

	err = usersCollection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&user)

	fmt.Println(user)

	if err := c.BodyParser(&post); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if post.Caption == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing caption"})
	}

	post.UserID = objectID

	post.CreatedAt = time.Now()
	post.UpdatedAt = time.Now()

	_, err = postsCollection.InsertOne(context.Background(), post)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "An error occurred"})
	}

	return c.Status(201).JSON(fiber.Map{"msg": "Post created", "post": post})
}

func GetPostsByUser(c *fiber.Ctx) error {
	userID := c.Params("_id")
	objectID, err := primitive.ObjectIDFromHex(userID)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var posts []types.Post

	collection = db.Database.Collection("posts")

	filter := bson.M{"userID": objectID}

	cursor, err := collection.Find(context.Background(), filter)

	if err != nil {
		return err
	}

	defer func() {
		if err := cursor.Close(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()

	for cursor.Next(context.Background()) {
		var post types.Post

		if err := cursor.Decode(&post); err != nil {
			return err
		}
		posts = append(posts, post)
	}

	return c.Status(200).JSON(posts)
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

func GetFeedPosts(c *fiber.Ctx) error {
	limitParam := c.Query("limit", "10")
	skipParam := c.Query("skip", "0")

	limit, _ := strconv.Atoi(limitParam)
	skip, _ := strconv.Atoi(skipParam)

	collection = db.Database.Collection("posts")
	opts := options.Find().SetLimit(int64(limit)).SetSkip(int64(skip))

	var posts []types.Post

	cursor, err := collection.Find(context.Background(), bson.M{}, opts)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch posts"})
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var post types.Post
		if err := cursor.Decode(&post); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to decode post"})
		}
		posts = append(posts, post)
	}

	return c.Status(200).JSON(posts)
}
