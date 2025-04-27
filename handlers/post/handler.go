package post

import (
	"context"
	"fmt"
	"github.com/edisss1/fiabesco-backend/db"
	"github.com/edisss1/fiabesco-backend/types"
	"github.com/edisss1/fiabesco-backend/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/imroc/req/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
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

	if err := c.BodyParser(&post); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if post.Caption == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing caption"})
	}

	err = usersCollection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "User not found"})
	}

	file, err := c.FormFile("file")
	if err == nil && file != nil {
		src, err := file.Open()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to open file"})
		}
		defer src.Close()

		fileBytes, err := io.ReadAll(src)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to read file"})
		}

		uniqueFileName := fmt.Sprintf("%s-%s", uuid.New().String(), file.Filename)
		supabaseUploadURL := fmt.Sprintf("https://fiabesco-storage.supabase.co/storage/v1/object/posts/%s", uniqueFileName)

		client := req.C().SetCommonBearerAuthToken(os.Getenv("SUPABASE_SERVICE_KEY"))

		resp, err := client.R().
			SetHeader("Content-Type", file.Header.Get("Content-Type")).
			SetBody(fileBytes).
			Put(supabaseUploadURL)

		if err != nil || resp.StatusCode >= 400 {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to upload to Supabase"})
		}

		publicURL := fmt.Sprintf("https://<your-project>.supabase.co/storage/v1/object/public/posts/%s", uniqueFileName)
		post.Files = append(post.Files, publicURL)
	}

	post.UserID = objectID
	post.CreatedAt = time.Now()
	post.UpdatedAt = time.Now()
	post.UserFirstName = user.FirstName
	post.UserLastName = user.LastName
	post.UserHandle = user.Handle
	post.UserPhotoURL = user.PhotoURL

	res, err := postsCollection.InsertOne(context.Background(), post)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "An error occurred while saving post"})
	}
	post.ID = res.InsertedID.(primitive.ObjectID)

	return c.Status(201).JSON(post)
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

func GetPost(c *fiber.Ctx) error {
	id := c.Params("postID")
	postID, err := utils.ParseHexID(id)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid ID")
	}
	var post types.Post

	collection = db.Database.Collection("posts")

	filter := bson.M{"_id": postID}

	err = collection.FindOne(context.Background(), filter).Decode(&post)
	if err != nil {
		return utils.RespondWithError(c, 500, "Failed to decode post")
	}

	return c.Status(200).JSON(post)

}

func GetFeedPosts(c *fiber.Ctx) error {
	sampleSizeParam := c.Query("sample", "0") // Default to 0 (no sampling)
	limitParam := c.Query("limit", "10")
	skipParam := c.Query("skip", "0")

	sampleSize, _ := strconv.Atoi(sampleSizeParam)
	limit, _ := strconv.Atoi(limitParam)
	skip, _ := strconv.Atoi(skipParam)

	collection := db.Database.Collection("posts")
	var posts []types.Post

	if sampleSize > 0 {

		effectiveSampleSize := sampleSize + skip + limit

		pipeline := []bson.M{
			{"$sample": bson.M{"size": effectiveSampleSize}},
			{"$sort": bson.M{"createdAt": -1}},
			{"$skip": skip},
			{"$limit": limit},
		}

		cursor, err := collection.Aggregate(context.Background(), pipeline)
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
	} else {
		opts := options.Find().
			SetLimit(int64(limit)).
			SetSkip(int64(skip)).
			SetSort(bson.D{{Key: "createdAt", Value: -1}})

		cursor, err := collection.Find(context.Background(), bson.M{}, opts)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch posts"})
		}
		defer cursor.Close(context.Background())

		for cursor.Next(context.Background()) {
			var post types.Post
			if err := cursor.Decode(&post); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "Failed to decode post" + err.Error()})
			}
			posts = append(posts, post)
		}
	}

	totalCount, err := collection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to count posts"})
	}

	hasMore := false
	if sampleSize > 0 {
		hasMore = len(posts) >= limit
	} else {
		hasMore = (skip + limit) < int(totalCount)
	}

	return c.Status(200).JSON(fiber.Map{
		"posts":    posts,
		"hasMore":  hasMore,
		"nextSkip": skip + limit,
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

	// Check if the user exists
	err = usersCollection.FindOne(context.Background(), userFilter).Decode(&user)
	if err != nil {
		return utils.RespondWithError(c, 404, "User not found")
	}

	// Get user's full name
	userName := strings.TrimSpace(user.FirstName + " " + user.LastName)

	// Fetch the post document
	err = postsCollection.FindOne(context.Background(), postFilter).Decode(&post)
	if err != nil {
		return utils.RespondWithError(c, 500, "Failed to retrieve post: "+err.Error())
	}

	err = likesCollection.FindOne(context.Background(), likeFilter).Decode(&like)

	if err == nil {
		// If already liked, unlike the post
		_, err = likesCollection.DeleteOne(context.Background(), likeFilter)
		if err != nil {
			return utils.RespondWithError(c, 500, "Failed to unlike the post: "+err.Error())
		}

		// Decrement the like count
		update = bson.M{"$inc": bson.M{"likesCount": -1}}

		// Apply the update to the post
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
