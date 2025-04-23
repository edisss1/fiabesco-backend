package messages

import (
	"context"
	"errors"
	"github.com/edisss1/fiabesco-backend/db"
	"github.com/edisss1/fiabesco-backend/helpers"
	"github.com/edisss1/fiabesco-backend/types"
	"github.com/edisss1/fiabesco-backend/utils"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
	"time"
)

var messagesCollection *mongo.Collection
var conversationsCollection *mongo.Collection
var usersCollection *mongo.Collection

func StartConversation(c *fiber.Ctx) error {
	conversationsCollection = db.Database.Collection("conversations")
	usersCollection = db.Database.Collection("users")

	var payload struct {
		SenderID    string `json:"senderID"`
		RecipientID string `json:"recipientID"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}

	senderID, err := primitive.ObjectIDFromHex(payload.SenderID)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid sender ID")
	}
	recipientID, err := primitive.ObjectIDFromHex(payload.RecipientID)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid recipient ID")
	}

	var sender struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	}
	var recipient struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	}

	senderFilter := bson.M{"_id": senderID}
	err = usersCollection.FindOne(context.Background(), senderFilter).Decode(&sender)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid sender ID")
	}

	recipientFilter := bson.M{"_id": recipientID}
	err = usersCollection.FindOne(context.Background(), recipientFilter).Decode(&recipient)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid recipient ID")
	}

	var conversation types.Conversation
	filter := bson.M{
		"isGroup": false,
		"participants": bson.M{
			"$all": []primitive.ObjectID{senderID, recipientID},
		},
	}

	err = conversationsCollection.FindOne(context.Background(), filter).Decode(&conversation)
	if err == nil {
		return c.JSON(fiber.Map{
			"conversationID": conversation.ID.Hex(),
		})
	} else if !errors.Is(err, mongo.ErrNoDocuments) {
		return utils.RespondWithError(c, 500, "DB error")
	}

	senderFullName := strings.TrimSpace(sender.FirstName + " " + sender.LastName)
	recipientFullName := strings.TrimSpace(recipient.FirstName + " " + recipient.LastName)

	newConversation := types.Conversation{
		IsGroup:      false,
		Participants: []primitive.ObjectID{senderID, recipientID},
		Names:        []string{senderFullName, recipientFullName},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	result, err := conversationsCollection.InsertOne(context.Background(), newConversation)
	if err != nil {
		return utils.RespondWithError(c, 500, "DB error")
	}

	return c.Status(201).JSON(fiber.Map{"conversationID": result.InsertedID.(primitive.ObjectID).Hex()})
}

func SendMessage(c *fiber.Ctx) error {
	conversationIDParam := c.Params("conversationID")
	senderIDParam := c.Params("senderID")

	conversationID, err := primitive.ObjectIDFromHex(conversationIDParam)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid conversation ID")
	}

	senderID, err := primitive.ObjectIDFromHex(senderIDParam)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid sender ID")
	}

	var msg struct {
		Content string `json:"content"`
	}

	if err := c.BodyParser(&msg); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}

	message, err := helpers.SaveMessage(senderID, conversationID, msg.Content)
	if err != nil {
		return utils.RespondWithError(c, 400, "Error sending message")
	}

	return c.Status(201).JSON(fiber.Map{"newMessage": message})
}

func DeleteMessage(c *fiber.Ctx) error {
	conversationsCollection = db.Database.Collection("conversations")

	var payload struct {
		ID string `json:"id"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}

	messageID, err := primitive.ObjectIDFromHex(payload.ID)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid message ID")
	}

	filter := bson.M{"_id": messageID}

	_, err = conversationsCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		return utils.RespondWithError(c, 400, "Failed to delete message")
	}

	return c.Status(200).JSON(fiber.Map{"msg": "Message deleted"})
}

func DeleteConversation(c *fiber.Ctx) error {
	conversationsCollection = db.Database.Collection("conversations")
	messagesCollection = db.Database.Collection("messages")

	id := c.Params("conversationID")

	conversationID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid conversation ID")
	}

	messagesFilter := bson.M{"conversationID": conversationID}
	conversationFilter := bson.M{"_id": conversationID}

	_, err = messagesCollection.DeleteMany(context.Background(), messagesFilter)
	if err != nil {
		return utils.RespondWithError(c, 500, "Failed to delete messages")
	}

	_, err = conversationsCollection.DeleteOne(context.Background(), conversationFilter)
	if err != nil {
		return utils.RespondWithError(c, 500, "Failed to delete conversation")
	}

	return c.Status(200).JSON(fiber.Map{"msg": "Conversation deleted successfully"})
}

func EditMessage(c *fiber.Ctx) error {
	messagesCollection = db.Database.Collection("messages")

	id := c.Params("_id")
	var payload struct {
		NewContent string `json:"newContent"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}

	messageID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid message ID")
	}

	filter := bson.M{"_id": messageID}

	err = messagesCollection.FindOne(context.Background(), filter).Decode(&payload)
	if err != nil {
		return utils.RespondWithError(c, 404, "Message not found")
	}

	update := bson.M{"$set": bson.M{"content": payload.NewContent}}

	_, err = messagesCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return utils.RespondWithError(c, 500, "Failed to update message")
	}

	return c.Status(200).JSON(fiber.Map{"msg": "Message updated"})
}

func GetConversation(c *fiber.Ctx) error {
	conversationsCollection = db.Database.Collection("conversations")
	messagesCollection = db.Database.Collection("messages")
	id := c.Params("conversationID")
	conversationID, err := utils.ParseHexID(id)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid ID")
	}
	var conversation types.Conversation

	err = conversationsCollection.FindOne(context.Background(), bson.M{"_id": conversationID}).Decode(&conversation)
	if err != nil {
		return utils.RespondWithError(c, 404, "Conversation not found")
	}

	filter := bson.M{"conversationID": conversationID}
	opts := options.Find().SetSort(bson.D{{"createdAt", -1}})

	cursor, err := messagesCollection.Find(context.Background(), filter, opts)
	if err != nil {
		return utils.RespondWithError(c, 500, "Couldn't get messages")
	}
	var messages []types.Message
	if err := cursor.All(context.Background(), &messages); err != nil {
		return utils.RespondWithError(c, 500, "Error decoding messages")
	}

	return c.Status(200).JSON(fiber.Map{"conversation": conversation, "messages": messages})
}

func GetConversations(c *fiber.Ctx) error {
	id := c.Params("userID")
	userID, err := utils.ParseHexID(id)

	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid ID")
	}

	var conversations []types.Conversation
	conversationsCollection = db.Database.Collection("conversations")

	filter := bson.M{"participants": userID}
	cursor, err := conversationsCollection.Find(context.Background(), filter)

	if err != nil {
		return utils.RespondWithError(c, 500, "Failed to get conversations")
	}

	if err := cursor.All(context.Background(), &conversations); err != nil {
		return utils.RespondWithError(c, 500, "Error decoding conversations")
	}

	return c.Status(200).JSON(conversations)
}
