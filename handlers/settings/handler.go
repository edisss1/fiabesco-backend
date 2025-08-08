package settings

import (
	"context"
	"github.com/edisss1/fiabesco-backend/db"
	"github.com/edisss1/fiabesco-backend/handlers/auth"
	"github.com/edisss1/fiabesco-backend/helpers"
	"github.com/edisss1/fiabesco-backend/types"
	"github.com/edisss1/fiabesco-backend/utils"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var collection *mongo.Collection

func ChangeFirstName(c *fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid ID")
	}

	var body struct {
		FirstName string `json:"firstName"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}

	filter := bson.M{"_id": userID}
	update := bson.M{"$set": bson.M{"firstName": body.FirstName}}

	collection = db.Database.Collection("users")

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return utils.RespondWithError(c, 500, "Error updating first name "+err.Error())
	}

	return c.Status(200).JSON(fiber.Map{"msg": "First name updated successfully"})
}

func ChangeLastName(c *fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid ID")
	}

	var body struct {
		LastName string `json:"lastName"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}

	filter := bson.M{"_id": userID}
	update := bson.M{"$set": bson.M{"lastName": body.LastName}}

	collection = db.Database.Collection("users")

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return utils.RespondWithError(c, 500, "Error updating last name "+err.Error())
	}

	return c.Status(200).JSON(fiber.Map{"msg": "Last name updated successfully"})
}

func ChangeEmail(c *fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid ID")
	}

	var body struct {
		Email string `json:"email"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}

	filter := bson.M{"_id": userID}
	update := bson.M{"$set": bson.M{"email": body.Email}}

	collection = db.Database.Collection("users")

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return utils.RespondWithError(c, 500, "Error updating email "+err.Error())
	}

	return c.Status(200).JSON(fiber.Map{"msg": "Email updated successfully"})
}

func ChangeHandle(c *fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid ID")
	}

	var body struct {
		Handle string `json:"handle"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}

	filter := bson.M{"handle": body.Handle}

	collection = db.Database.Collection("users")

	if err := collection.FindOne(context.Background(), filter).Err(); err == nil {
		return utils.RespondWithError(c, 400, "Handle already exists")
	}

	filter = bson.M{"_id": userID}
	update := bson.M{"$set": bson.M{"handle": body.Handle}}

	collection = db.Database.Collection("users")

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return utils.RespondWithError(c, 500, "Error updating handle "+err.Error())
	}

	return c.Status(200).JSON(fiber.Map{"msg": "Handle updated successfully"})

}

func ChangePassword(c *fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid ID")
	}

	var body struct {
		Password string `json:"password"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}

	var user types.User
	filter := bson.M{"_id": userID}

	collection = db.Database.Collection("users")

	err = collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		return utils.RespondWithError(c, 404, "User not found")
	}
	matches := auth.CheckPasswordHash(user.Password, body.Password)
	if matches {
		return utils.RespondWithError(c, 400, "New password must be different from old password")
	}

	hashedPassword := auth.HashPassword(body.Password)

	filter = bson.M{"_id": userID}
	update := bson.M{"$set": bson.M{"password": hashedPassword}}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return utils.RespondWithError(c, 500, "Error updating password "+err.Error())
	}

	return c.Status(200).JSON(fiber.Map{"msg": "Password updated successfully"})
}

func ChangeTheme(c *fiber.Ctx) error {

	var body struct {
		Theme string `json:"theme"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}

	err := helpers.SaveSetting(c, map[string]interface{}{"theme": body.Theme})
	if err != nil {
		return utils.RespondWithError(c, 500, "Error updating theme "+err.Error())
	}

	return c.Status(200).JSON(fiber.Map{"msg": "Theme updated successfully"})

}

func ChangeLanguage(c *fiber.Ctx) error {

	var body struct {
		Language string `json:"language"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}

	err := helpers.SaveSetting(c, map[string]interface{}{"language": body.Language})
	if err != nil {
		return utils.RespondWithError(c, 500, "Error updating language "+err.Error())
	}

	return c.Status(200).JSON(fiber.Map{"msg": "Language updated successfully"})

}

func ChangeProfileVisibility(c *fiber.Ctx) error {

	var body struct {
		ProfileVisibility string `json:"profileVisibility"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}

	err := helpers.SaveSetting(c, map[string]interface{}{"profileVisibility": body.ProfileVisibility})
	if err != nil {
		return utils.RespondWithError(c, 500, "Error updating profile visibility "+err.Error())
	}

	return c.Status(200).JSON(fiber.Map{"msg": "Profile visibility updated successfully"})

}

func DownloadUserData(c *fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid ID")
	}

	var user types.User
	var posts []types.Post
	var comments []types.Comment
	var settings types.Settings
	var likes []types.Like
	var conversations []types.Conversation

	filter := bson.M{"_id": userID}

	collection = db.Database.Collection("users")

	err = collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		return utils.RespondWithError(c, 404, "User not found")
	}

	user.Password = ""

	collection = db.Database.Collection("posts")

	cursor, err := collection.Find(context.Background(), bson.M{"userID": userID})
	if err != nil {
		return utils.RespondWithError(c, 500, "Error finding posts "+err.Error())
	}

	err = cursor.All(context.Background(), &posts)
	if err != nil {
		return utils.RespondWithError(c, 500, "Error decoding posts "+err.Error())
	}

	collection = db.Database.Collection("comments")

	cursor, err = collection.Find(context.Background(), bson.M{"userID": userID})
	if err != nil {
		return utils.RespondWithError(c, 500, "Error finding comments "+err.Error())
	}

	err = cursor.All(context.Background(), &comments)
	if err != nil {
		return utils.RespondWithError(c, 500, "Error decoding comments "+err.Error())
	}

	collection = db.Database.Collection("settings")

	err = collection.FindOne(context.Background(), bson.M{"userID": userID}).Decode(&settings)
	if err != nil {
		return utils.RespondWithError(c, 500, "Error finding settings "+err.Error())
	}

	collection = db.Database.Collection("likes")

	cursor, err = collection.Find(context.Background(), bson.M{"userID": userID})
	if err != nil {
		return utils.RespondWithError(c, 500, "Error finding likes "+err.Error())
	}

	err = cursor.All(context.Background(), &likes)
	if err != nil {
		return utils.RespondWithError(c, 500, "Error decoding likes "+err.Error())
	}

	collection = db.Database.Collection("conversations")

	cursor, err = collection.Find(context.Background(), bson.M{"participants": userID})
	if err != nil {
		return utils.RespondWithError(c, 500, "Error finding conversations "+err.Error())
	}

	err = cursor.All(context.Background(), &conversations)
	if err != nil {
		return utils.RespondWithError(c, 500, "Error decoding conversations "+err.Error())
	}

	return c.Status(200).JSON(fiber.Map{"user": user, "posts": posts, "comments": comments, "settings": settings, "likes": likes, "conversations": conversations})

}
