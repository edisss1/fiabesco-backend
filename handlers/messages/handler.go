package messages

import (
	"context"
	"errors"
	"github.com/edisss1/fiabesco-backend/db"
	"github.com/edisss1/fiabesco-backend/types"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

var messagesCollection *mongo.Collection
var conversationsCollection *mongo.Collection

func SendMessage(c *fiber.Ctx) error {
	conversationIDParam := c.Params("conversationID")
	senderIDParam := c.Params("senderID")

	conversationID, err := primitive.ObjectIDFromHex(conversationIDParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid conversation ID"})
	}

	senderID, err := primitive.ObjectIDFromHex(senderIDParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid sender ID"})
	}

	var msg struct {
		Content string `json:"content"`
	}

	if err := c.BodyParser(&msg); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	messagesCollection = db.Database.Collection("messages")

	message := types.Message{
		SenderID:       senderID,
		ConversationID: conversationID,
		Content:        msg.Content,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	_, err = messagesCollection.InsertOne(context.Background(), message)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to send message"})
	}

	return c.Status(201).JSON(fiber.Map{"msg": "Message sent"})
}

func StartConversation(c *fiber.Ctx) error {

	conversationsCollection = db.Database.Collection("conversation")

	var payload struct {
		SenderID    string `json:"senderID"`
		RecipientID string `json:"recipientID"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	senderID, err := primitive.ObjectIDFromHex(payload.SenderID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid sender ID"})
	}
	recipientID, err := primitive.ObjectIDFromHex(payload.RecipientID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid recipient ID"})

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
		return c.Status(500).JSON(fiber.Map{"error": "DB error"})
	}

	newConversation := types.Conversation{
		IsGroup:      false,
		Participants: []primitive.ObjectID{senderID, recipientID},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	result, err := conversationsCollection.InsertOne(context.Background(), newConversation)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create conversation"})
	}

	return c.Status(201).JSON(fiber.Map{"conversationID": result.InsertedID.(primitive.ObjectID).Hex()})

}
