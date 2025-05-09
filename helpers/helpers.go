package helpers

import (
	"context"
	"errors"
	"github.com/edisss1/fiabesco-backend/db"
	"github.com/edisss1/fiabesco-backend/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
