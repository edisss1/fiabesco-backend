package helpers

import (
	"context"
	"errors"
	"github.com/edisss1/fiabesco-backend/db"
	"github.com/edisss1/fiabesco-backend/types"
	"github.com/edisss1/fiabesco-backend/utils"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"time"
)

func SaveMessage(senderID, conversationID primitive.ObjectID, content string) (types.Message, error) {
	message := types.Message{
		SenderID:       senderID,
		ConversationID: conversationID,
		Content:        content,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Read:           false,
	}

	messagesCollection := db.Database.Collection("messages")
	conversationsCollection := db.Database.Collection("conversations")

	count, err := conversationsCollection.CountDocuments(context.Background(), bson.M{"_id": conversationID})
	if err != nil || count == 0 {
		return types.Message{}, errors.New("conversation not found")
	}
	res, err := messagesCollection.InsertOne(context.Background(), message)
	if err != nil {
		return types.Message{}, err
	}
	message.ID = res.InsertedID.(primitive.ObjectID)

	filter := bson.M{"_id": conversationID}
	update := bson.M{"$set": bson.M{"lastMessage": message}}

	_, err = conversationsCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return types.Message{}, err
	}

	return message, nil
}

func SaveEditedMessage(messageID primitive.ObjectID, content string, conversationID primitive.ObjectID, senderID primitive.ObjectID) (types.Message, error) {
	messagesCollection := db.Database.Collection("messages")
	conversationsCollection := db.Database.Collection("conversations")
	filter := bson.M{"_id": messageID}
	update := bson.M{"$set": bson.M{"content": content}}

	_, err := messagesCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return types.Message{}, err
	}

	var updatedMessage types.Message
	err = messagesCollection.FindOne(context.Background(), filter).Decode(&updatedMessage)
	if err != nil {
		return types.Message{}, err
	}

	var conversation types.Conversation
	err = conversationsCollection.FindOne(context.Background(), bson.M{"_id": conversationID}).Decode(&conversation)
	if err != nil {
		log.Println("Error decoding conversation")
	}

	lastMessage := conversation.LastMessage
	if lastMessage.ID == messageID {
		filter := bson.M{"_id": conversationID}
		update := bson.M{"$set": bson.M{"lastMessage": updatedMessage}}
		_, err := conversationsCollection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			return types.Message{}, err
		}
	}

	return updatedMessage, nil
}

func GetConversation(conversationID primitive.ObjectID) (types.Conversation, error) {
	conversationsCollection := db.Database.Collection("conversations")

	filter := bson.M{"_id": conversationID}
	var conversation types.Conversation
	err := conversationsCollection.FindOne(context.Background(), filter).Decode(&conversation)
	if err != nil {
		return types.Conversation{}, err
	}

	return conversation, nil
}

func GetConversations(userID primitive.ObjectID) ([]types.Conversation, error) {
	conversationsCollection := db.Database.Collection("conversations")

	filter := bson.M{"participants": userID}
	var conversatinos []types.Conversation

	cursor, err := conversationsCollection.Find(context.Background(), filter)
	if err != nil {
		return []types.Conversation{}, err
	}

	err = cursor.All(context.Background(), &conversatinos)
	if err != nil {
		return []types.Conversation{}, err
	}

	return conversatinos, nil

}

func SaveSettings(c *fiber.Ctx) error {
	id := c.Params("userID")
	userID, err := utils.ParseHexID(id)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid user ID")
	}

	var body types.Settings

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}

	collection := db.Database.Collection("users")

	var user types.User
	filter := bson.M{"_id": userID}

	if err := collection.FindOne(context.Background(), filter).Decode(&user); err != nil {
		return utils.RespondWithError(c, 404, "User not found")
	}

	if body.Theme != "" {
		_, err = collection.UpdateOne(context.Background(), filter, bson.M{"$set": bson.M{"settings.theme": body.Theme}})

		if err != nil {
			return utils.RespondWithError(c, 500, "Failed to update user settings")
		}

	}

	if body.Language != "" {

		_, err = collection.UpdateOne(context.Background(), filter, bson.M{"$set": bson.M{"settings.language": body.Language}})

		if err != nil {
			return utils.RespondWithError(c, 500, "Failed to update user settings")
		}
	}

	return c.Status(200).JSON(fiber.Map{"msg": "Settings saved successfully"})

}
