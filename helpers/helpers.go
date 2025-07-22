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
	"strings"
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
	usersCollection := db.Database.Collection("users")

	filter := bson.M{"_id": conversationID}
	var conversation types.Conversation
	err := conversationsCollection.FindOne(context.Background(), filter).Decode(&conversation)
	if err != nil {
		return types.Conversation{}, err
	}

	filter = bson.M{"_id": bson.M{"$in": conversation.ParticipantsIds}}
	cursor, err := usersCollection.Find(context.Background(), filter)
	if err != nil {
		return types.Conversation{}, err
	}
	var participants []types.Participant
	if err := cursor.All(context.Background(), &participants); err != nil {
		return types.Conversation{}, err
	}
	conversation.Participants = participants

	return conversation, nil
}

func GetConversations(userID primitive.ObjectID) ([]types.Conversation, error) {
	conversationsCollection := db.Database.Collection("conversations")
	usersCollection := db.Database.Collection("users")

	filter := bson.M{"participantsIds": userID}
	cursor, err := conversationsCollection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	var conversations []types.Conversation
	if err := cursor.All(context.Background(), &conversations); err != nil {
		return nil, err
	}

	participantIDSet := make(map[primitive.ObjectID]struct{})
	for _, conv := range conversations {
		for _, id := range conv.ParticipantsIds {
			participantIDSet[id] = struct{}{}
		}
	}
	var participantIDs []primitive.ObjectID
	for id := range participantIDSet {
		participantIDs = append(participantIDs, id)
	}

	usersFilter := bson.M{"_id": bson.M{"$in": participantIDs}}
	cursor, err = usersCollection.Find(context.Background(), usersFilter)
	if err != nil {
		return nil, err
	}

	userMap := make(map[primitive.ObjectID]types.Participant)
	for cursor.Next(context.Background()) {
		var user struct {
			ID        primitive.ObjectID `bson:"_id"`
			FirstName string             `bson:"firstName"`
			LastName  string             `bson:"lastName"`
			Photo     string             `bson:"photoURL"`
		}
		if err := cursor.Decode(&user); err != nil {
			continue
		}

		if user.Photo != "" {
			user.Photo = utils.BuildImgURL(user.Photo)
		} else {
			user.Photo = ""
		}

		userMap[user.ID] = types.Participant{
			ID:       user.ID,
			UserName: strings.TrimSpace(user.FirstName + " " + user.LastName),
			PhotoURL: user.Photo,
		}
	}

	for i, conv := range conversations {
		var enriched []types.Participant
		for _, id := range conv.ParticipantsIds {
			if participant, ok := userMap[id]; ok {
				enriched = append(enriched, participant)
			}
		}
		conversations[i].Participants = enriched
	}

	return conversations, nil
}

func SaveSetting(c *fiber.Ctx, setting map[string]interface{}) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid user ID")
	}

	collection := db.Database.Collection("settings")

	allowedFields := map[string]bool{
		"theme":             true,
		"language":          true,
		"profileVisibility": true,
	}

	for key := range setting {
		allowed, exists := allowedFields[key]
		if !exists || !allowed {
			return utils.RespondWithError(c, 400, "Invalid field: "+key)
		}
	}

	filter := bson.M{"userID": userID}
	update := bson.M{
		"$set": setting,
	}

	defaultSettings := types.Settings{
		UserID:            userID,
		Theme:             "light",
		Language:          "en",
		ProfileVisibility: "public",
	}

	count, err := collection.CountDocuments(context.Background(), filter)

	if err != nil {
		return utils.RespondWithError(c, 500, "Error checking document count")
	}

	if count == 0 {
		_, err = collection.InsertOne(context.Background(), defaultSettings)
		if err != nil {
			return utils.RespondWithError(c, 500, "Error inserting default settings")
		}
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			return utils.RespondWithError(c, 500, "Error updating settings")
		}
	} else {
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			return utils.RespondWithError(c, 500, "Error updating settings")
		}
	}

	return c.Status(200).JSON(fiber.Map{"message": "Settings updated successfully"})
}

func UpdateUserStatus(userID primitive.ObjectID, status string) error {
	collection := db.Database.Collection("users")

	filter := bson.M{"_id": userID}
	update := bson.M{"$set": bson.M{
		"status":   status,
		"lastSeen": time.Now(),
	}}

	_, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return nil

}
