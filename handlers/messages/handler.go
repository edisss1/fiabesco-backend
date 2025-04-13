package messages

import (
	"context"
	"fmt"
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
	id := c.Params("_id")
	senderID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	messagesCollection = db.Database.Collection("messages")
	conversationsCollection = db.Database.Collection("conversations")

	var msg struct {
		RecipientID string `json:"recipientID,omitempty" bson:"recipientID,omitempty"`
		Content     string `json:"content"`
	}

	if err := c.BodyParser(&msg); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	recipientID, err := primitive.ObjectIDFromHex(msg.RecipientID)
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

	if err == mongo.ErrNoDocuments {
		conversation = types.Conversation{
			IsGroup:      false,
			Participants: []primitive.ObjectID{senderID, recipientID},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		result, err := conversationsCollection.InsertOne(context.Background(), conversation)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create conversation"})
		}
		conversation.ID = result.InsertedID.(primitive.ObjectID)
		fmt.Println("Conversation ID: ", conversation.ID)
		fmt.Println("Result: ", result)
	} else if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to check conversation"})
	}

	message := types.Message{
		SenderID:       senderID,
		ConversationID: conversation.ID,
		Content:        msg.Content,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	_, err = messagesCollection.InsertOne(context.Background(), message)

	return c.Status(201).JSON(fiber.Map{"msg": "Message sent"})

}
