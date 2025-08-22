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
	"log"
	"slices"
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
		FirstName    string               `json:"firstName"`
		LastName     string               `json:"lastName"`
		PhotoURL     string               `json:"photoURL" bson:"photoURL"`
		BlockedUsers []primitive.ObjectID `json:"blockedUsers" bson:"blockedUsers"`
	}
	var recipient struct {
		FirstName    string               `json:"firstName"`
		LastName     string               `json:"lastName"`
		PhotoURL     string               `json:"photoURL" bson:"photoURL"`
		BlockedUsers []primitive.ObjectID `json:"blockedUsers" bson:"blockedUsers"`
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

	if slices.Contains(sender.BlockedUsers, recipientID) {
		return c.Status(400).JSON(fiber.Map{"started": false, "msg": "You cannot send messages to blocked users"})
	}

	if slices.Contains(recipient.BlockedUsers, senderID) {
		return c.Status(400).JSON(fiber.Map{"started": false, "msg": "Cannot send messages to users who blocked you"})
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

	newConversation := types.Conversation{
		IsGroup:   false,
		CreatedAt: time.Now(),
		ParticipantsIds: []primitive.ObjectID{
			senderID,
			recipientID,
		},
		UpdatedAt: time.Now(),
	}

	result, err := conversationsCollection.InsertOne(context.Background(), newConversation)
	if err != nil {
		return utils.RespondWithError(c, 500, "DB error")
	}

	return c.Status(201).JSON(fiber.Map{"conversationID": result.InsertedID.(primitive.ObjectID).Hex(), "started": true})
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
	messagesCollection = db.Database.Collection("messages")

	var payload struct {
		ID string `json:"id"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}
	log.Printf("Payload ID: %s", payload.ID)

	messageID, err := primitive.ObjectIDFromHex(payload.ID)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid message ID")
	}

	filter := bson.M{"_id": messageID}

	_, err = messagesCollection.DeleteOne(context.Background(), filter)
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

	update := bson.M{"$set": bson.M{"content": payload.NewContent, "isEdited": true, "updatedAt": time.Now()}}

	_, err = messagesCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return utils.RespondWithError(c, 500, "Failed to update message")
	}

	return c.Status(200).JSON(fiber.Map{"msg": "Message updated"})
}

func GetConversation(c *fiber.Ctx) error {
	conversationsCollection = db.Database.Collection("conversations")
	messagesCollection = db.Database.Collection("messages")
	usersCollection = db.Database.Collection("users")
	id := c.Params("conversationID")
	conversationID, err := utils.ParseHexID(id)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid ID")
	}

	filter := bson.M{"_id": conversationID}
	var conversation types.Conversation
	err = conversationsCollection.FindOne(context.Background(), filter).Decode(&conversation)
	if err != nil {
		return utils.RespondWithError(c, 404, "Conversation not found "+err.Error())
	}

	var messages []types.Message
	messagesFilter := bson.M{"conversationID": conversationID}
	cursor, err := messagesCollection.Find(context.Background(), messagesFilter)
	if err != nil {
		return utils.RespondWithError(c, 500, "DB error "+err.Error())
	}

	for cursor.Next(context.Background()) {
		var message types.Message
		err := cursor.Decode(&message)
		if err != nil {
			return utils.RespondWithError(c, 500, "DB error "+err.Error())
		}
		messages = append(messages, message)
	}

	usersFilter := bson.M{"_id": bson.M{"$in": conversation.ParticipantsIds}}
	cursor, err = usersCollection.Find(context.Background(), usersFilter)
	if err != nil {
		return utils.RespondWithError(c, 500, "DB error "+err.Error())
	}
	var enriched []types.Participant

	for cursor.Next(context.Background()) {
		var user struct {
			ID        primitive.ObjectID `bson:"_id"`
			FirstName string             `bson:"firstName"`
			LastName  string             `bson:"lastName"`
			Photo     string             `bson:"photoURL"`
		}
		_ = cursor.Decode(&user)

		if user.Photo != "" {
			user.Photo = utils.BuildImgURL(user.Photo)
		} else {
			user.Photo = ""
		}

		enriched = append(enriched, types.Participant{
			ID:       user.ID,
			UserName: strings.TrimSpace(user.FirstName + " " + user.LastName),
			PhotoURL: user.Photo,
		})
	}

	conversation.Participants = enriched

	return c.Status(200).JSON(fiber.Map{"conversation": conversation, "messages": messages})

}

func GetConversations(c *fiber.Ctx) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid user ID")
	}

	conversations, err := helpers.GetConversations(userID)
	if err != nil {
		return utils.RespondWithError(c, 500, "Couldn't get conversations")
	}

	return c.Status(200).JSON(conversations)
}

// GetMessage will be primarily used to get a single message that is being replied to and not present
// in the loaded conversation on the frontend
func GetMessage(c *fiber.Ctx) error {
	messagesCollection = db.Database.Collection("messages")
	id := c.Params("messageID")
	messageID, err := utils.ParseHexID(id)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid ID")
	}

	filter := bson.M{"_id": messageID}
	var message types.Message

	err = messagesCollection.FindOne(context.Background(), filter).Decode(&message)
	if err != nil {
		return utils.RespondWithError(c, 404, "Message not found")
	}

	return c.Status(200).JSON(message)
}

func SendReply(c *fiber.Ctx) error {
	var body struct {
		Content string `json:"content"`
		ReplyTo string `json:"replyTo"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondWithError(c, 400, "Invalid request body")
	}

	conversationIDParam := c.Params("conversationID")

	senderID, err := utils.GetUserID(c)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid user ID")
	}
	conversationID, err := utils.ParseHexID(conversationIDParam)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid conversation ID")
	}

	replyTo, err := utils.ParseHexID(body.ReplyTo)
	if err != nil {
		return utils.RespondWithError(c, 400, "Invalid reply to ID")
	}

	reply, err := helpers.SaveReply(senderID, conversationID, body.Content, replyTo)
	if err != nil {
		return utils.RespondWithError(c, 400, "Error sending reply")
	}

	return c.Status(200).JSON(fiber.Map{"newMessage": reply})
}
