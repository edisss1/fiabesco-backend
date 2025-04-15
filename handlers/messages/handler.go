package messages

import (
	"context"
	"errors"
	"fmt"
	"github.com/edisss1/fiabesco-backend/db"
	"github.com/edisss1/fiabesco-backend/types"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"strings"
	"time"
)

var messagesCollection *mongo.Collection
var conversationsCollection *mongo.Collection

var usersCollection *mongo.Collection

func StartConversation(c *fiber.Ctx) error {

	conversationsCollection = db.Database.Collection("conversations")

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
		return c.Status(400).JSON(fiber.Map{"error": "Invalid sender ID"})
	}
	recipientFilter := bson.M{"_id": recipientID}
	err = usersCollection.FindOne(context.Background(), recipientFilter).Decode(&recipient)

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
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create conversation"})
	}

	return c.Status(201).JSON(fiber.Map{"conversationID": result.InsertedID.(primitive.ObjectID).Hex()})

}

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

func DeleteMessage(c *fiber.Ctx) error {
	conversationsCollection = db.Database.Collection("conversations")

	var payload struct {
		ID string `json:"id"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	messageID, err := primitive.ObjectIDFromHex(payload.ID)
	fmt.Printf("Message ID: %v", messageID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	filter := bson.M{"_id": messageID}

	_, err = conversationsCollection.DeleteOne(context.Background(), filter)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err})
	}

	return c.Status(200).JSON(fiber.Map{"msg": "Message deleted"})
}

func DeleteConversation(c *fiber.Ctx) error {
	conversationsCollection = db.Database.Collection("conversations")
	messagesCollection = db.Database.Collection("messages")

	id := c.Params("conversationID")

	conversationID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	messagesFilter := bson.M{"conversationID": conversationID}
	conversationFilter := bson.M{"_id": conversationID}

	_, err = messagesCollection.DeleteMany(context.Background(), messagesFilter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "DB error"})
	}

	_, err = conversationsCollection.DeleteOne(context.Background(), conversationFilter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "DB error"})
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
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	messageID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	filter := bson.M{"_id": messageID}

	err = messagesCollection.FindOne(context.Background(), filter).Decode(&payload)

	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Message not found"})
	}

	update := bson.M{"$set": bson.M{"content": payload.NewContent}}

	_, err = messagesCollection.UpdateOne(context.Background(), filter, update)

	return c.Status(200).JSON(fiber.Map{"msg": "Message updated"})
}
