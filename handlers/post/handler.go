package post

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/edisss1/fiabesco-backend/db"
	"github.com/edisss1/fiabesco-backend/types"
	"github.com/edisss1/fiabesco-backend/uploads"
	"github.com/edisss1/fiabesco-backend/utils"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"strconv"
	"strings"
	"time"
)

// TODO: update other functions to behave like GetPostsByUser

type FeedItem struct {
	Post     types.Post `json:"post"`
	UserName string     `json:"userName" bson:"userName"`
	PhotoURL string     `json:"photoURL" bson:"photoURL"`
	Handle   string     `json:"handle"`
}

var collection *mongo.Collection

func CreatePost(c *fiber.Ctx) error {
	id := c.Params("userID")
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	postJSON := c.FormValue("post")
	var post types.Post

	if err := json.Unmarshal([]byte(postJSON), &post); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body"+err.Error())
	}

	bucket, err := gridfs.NewBucket(db.Database)
	if err != nil {
		return utils.RespondWithError(c, 500, "Error creating bucket")
	}

	for i := range post.Images {
		fieldName := fmt.Sprintf("post-img-%d", i)
		ids, err := uploads.UploadFile(c, fieldName, bucket, false)
		if err != nil {
			continue
		}
		if len(ids) > 0 {
			post.Images[i] = ids[0].Hex()
		}
	}

	post.UserID = userID
	collection = db.Database.Collection("posts")

	_, err = collection.InsertOne(context.Background(), post)
	if err != nil {
		return utils.RespondWithError(c, 500, "Failed to create post "+err.Error())
	}

	fmt.Println(post)

	return c.Status(201).JSON(fiber.Map{"post": post})
}

func GetPostsByUser(c *fiber.Ctx) error {
	id := c.Params("_id")
	userID, err := primitive.ObjectIDFromHex(id)

	pipeline := mongo.Pipeline{
		{{"$match", bson.D{{"userID", userID}}}},
		{{"$lookup", bson.D{
			{"from", "users"},
			{"localField", "userID"},
			{"foreignField", "_id"},
			{"as", "user"},
		}}},
		{{"$unwind", "$user"}},
		{{"$project", bson.D{
			{"post", "$$ROOT"},
			{"user", 1},
		}}},
		{{"$project", bson.D{
			{"user.password", 0},
			{"user.email", 0},
		}}},
	}

	collection = db.Database.Collection("posts")
	var posts []struct {
		Post types.Post `json:"post" bson:"post"`
		User types.User `json:"user" bson:"user"`
	}

	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get post" + err.Error()})
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var res struct {
			Post types.Post `json:"post" bson:"post"`
			User types.User `json:"user" bson:"user"`
		}

		if err = cursor.Decode(&res); err != nil {
			return utils.RespondWithError(c, 500, "Failed to decode post data"+err.Error())
		}

		posts = append(posts, res)

	}

	if err := cursor.Err(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cursor iteration error: " + err.Error()})
	}

	return c.Status(200).JSON(fiber.Map{"posts": posts})
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
	pageParam := c.Query("page", "1")

	l := 10
	p, _ := strconv.Atoi(pageParam)
	skip := int64(p*l - l)

	limit := int64(l)

	//opt := options.FindOptions{Limit: &limit, Skip: &skip}

	pipeline := mongo.Pipeline{

		bson.D{{"$sort", bson.D{{"createdAt", -1}}}},

		// Pagination
		bson.D{{"$skip", skip}},
		bson.D{{"$limit", limit}},

		bson.D{{"$lookup", bson.D{
			{"from", "users"},
			{"localField", "userID"},
			{"foreignField", "_id"},
			{"as", "user"},
		}}},

		bson.D{{"$unwind", bson.D{
			{"path", "$user"},
			{"preserveNullAndEmptyArrays", true},
		}}},

		bson.D{{"$project", bson.D{
			{"post", "$$ROOT"},
			{"userName", bson.D{{"$concat", bson.A{"$user.firstName", " ", "$user.lastName"}}}},

			{"photoURL", "$user.photoURL"},
			{"handle", "$user.handle"},
		}}},
	}

	collection := db.Database.Collection("posts")
	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return utils.RespondWithError(c, 500, "Failed to fetch posts "+err.Error())
	}

	var result []FeedItem

	for cursor.Next(context.Background()) {
		var feedItem FeedItem
		if err := cursor.Decode(&feedItem); err != nil {
			return utils.RespondWithError(c, 500, "Failed to decode post: "+err.Error())
		}
		if feedItem.PhotoURL != "" {
			feedItem.PhotoURL = utils.BuildImgURL(feedItem.PhotoURL)
		}
		if feedItem.Post.Images != nil {
			for i := range feedItem.Post.Images {
				feedItem.Post.Images[i] = utils.BuildImgURL(feedItem.Post.Images[i])
			}
		}

		result = append(result, feedItem)
	}

	return c.Status(200).JSON(result)
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
		return utils.RespondWithError(c, 400, "Invalid ID")
	}

	type CommentWithUser struct {
		Comment types.Comment `json:"comment"`
		User    types.User    `json:"user"`
	}

	commentCollection := db.Database.Collection("comments")
	userCollection := db.Database.Collection("users")

	cursor, err := commentCollection.Find(context.Background(), bson.M{"postID": postID})
	if err != nil {
		return utils.RespondWithError(c, 500, "Failed to fetch comments: "+err.Error())
	}
	defer cursor.Close(context.Background())

	var results []CommentWithUser

	for cursor.Next(context.Background()) {
		var comment types.Comment
		if err := cursor.Decode(&comment); err != nil {
			return utils.RespondWithError(c, 500, "Failed to decode comment: "+err.Error())
		}

		var user types.User
		err := userCollection.FindOne(context.Background(), bson.M{"_id": comment.UserID}).Decode(&user)
		if err != nil {
			return utils.RespondWithError(c, 500, "Error decoding user"+err.Error())
		}

		user.Email = ""
		user.Password = ""

		results = append(results, CommentWithUser{
			Comment: comment,
			User:    user,
		})
	}

	return c.Status(200).JSON(results)
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
