package ws

import (
	"encoding/json"
	"github.com/edisss1/fiabesco-backend/helpers"
	"github.com/edisss1/fiabesco-backend/types"
	"github.com/edisss1/fiabesco-backend/utils"
	"github.com/gofiber/websocket/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"sync"
)

var clients = make(map[string]*websocket.Conn)
var mu sync.Mutex

type BaseWSMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type SendMessagePayload struct {
	SenderID       string `json:"senderID"`
	RecipientID    string `json:"recipientID"`
	ConversationID string `json:"conversationID"`
	Content        string `json:"content"`
}

type EditMessagePayload struct {
	MessageID      string `json:"messageID"`
	Content        string `json:"content"`
	ConversationID string `json:"conversationID"`
	SenderID       string `json:"senderID"`
}

type GetConversationsPayload struct {
	UserID string `json:"userID"`
}

func HandleWS(conn *websocket.Conn) {
	userID := conn.Query("userID")
	log.Printf("New connection from userID: %s\n", userID)

	mu.Lock()
	clients[userID] = conn
	mu.Unlock()
	log.Printf("Added user %s to clients map\n", userID)

	defer func() {
		mu.Lock()
		delete(clients, userID)
		mu.Unlock()
		conn.Close()
		log.Printf("Closed connection for userID: %s\n", userID)
	}()

	for {
		var base BaseWSMessage
		if err := conn.ReadJSON(&base); err != nil {
			log.Println("read error: ", err)
			break
		}

		log.Printf("Received message: %v\n", base)

		switch base.Type {
		case "send_message":
			var payload SendMessagePayload
			if err := json.Unmarshal(base.Data, &payload); err != nil {
				log.Println("Unmarshal error: ", err)
				continue
			}

			log.Printf("Received send_message payload: %+v\n", payload)

			senderID, err := primitive.ObjectIDFromHex(payload.SenderID)
			if err != nil {
				log.Println("Invalid senderID: ", err)
				continue
			}
			conversationID, err := primitive.ObjectIDFromHex(payload.ConversationID)
			if err != nil {
				log.Println("Invalid conversationID: ", err)
				continue
			}

			message, err := helpers.SaveMessage(senderID, conversationID, payload.Content)
			if err != nil {
				log.Println("Error saving message: ", err)
				continue
			}
			log.Printf("Saved message: %+v\n", message)

			conversation, err := helpers.GetConversation(conversationID)
			if err != nil {
				log.Println("Error getting conversation: ", err)
			}

			for _, userID := range conversation.Participants {
				conn, ok := clients[userID.Hex()]
				if ok {
					err = conn.WriteJSON(struct {
						Type    string        `json:"type"`
						Message types.Message `json:"message"`
					}{
						Type:    "conversations_update",
						Message: message,
					})
					log.Printf("Sent message to participant: %v\n", message)
				}
			}

		case "edit_message":
			var payload EditMessagePayload

			if err := json.Unmarshal(base.Data, &payload); err != nil {
				log.Println("Unmarshal error: ", err)
				continue
			}

			messageID, err := utils.ParseHexID(payload.MessageID)
			if err != nil {
				log.Println("Invalid messageID: ", err)
			}

			conversationID, err := utils.ParseHexID(payload.ConversationID)
			if err != nil {
				log.Println("Invalid conversationID: ", err)
			}
			senderID, err := utils.ParseHexID(payload.SenderID)
			if err != nil {
				log.Println("Invalid senderID: ", err)
			}

			message, err := helpers.SaveEditedMessage(messageID, payload.Content, conversationID, senderID)
			log.Println("Saved message: ", message)

			log.Printf("Edited message: ID: %s, ConversationID: %s, SenderID: %s, Content: %s, CreatedAt: %v, UpdatedAt: %v\n",
				message.ID, message.ConversationID, message.SenderID, message.Content, message.CreatedAt, message.UpdatedAt)
			if err != nil {
				log.Println("Error saving message: ", err)
				continue
			}

			log.Println("ConversationID: ", message.ConversationID)
			conversation, err := helpers.GetConversation(message.ConversationID)
			if err != nil {
				log.Println("Error getting conversation: ", err)
			}

			for _, userID := range conversation.Participants {

				if conn, ok := clients[userID.Hex()]; ok {
					err = conn.WriteJSON(message)
					log.Printf("Sent message to user: %v\n", message)
					if err != nil {
						log.Println("Error sending message to user: ", err)
					}
				}
			}
		case "get_conversations":
			var payload GetConversationsPayload
			if err := json.Unmarshal(base.Data, &payload); err != nil {
				log.Println("Unmarshal error: ", err)
				continue
			}

			userID, err := utils.ParseHexID(payload.UserID)
			if err != nil {
				log.Println("Invalid userID: ", err)
				continue
			}

			conversations, err := helpers.GetConversations(userID)
			if err != nil {
				log.Println("Error getting conversations: ", err)
				continue
			}

			conn, ok := clients[payload.UserID]
			if ok {
				err = conn.WriteJSON(struct {
					Type          string               `json:"type"`
					Conversations []types.Conversation `json:"conversations"`
				}{
					Type:          "conversations",
					Conversations: conversations,
				})
				log.Printf("Sent conversations to user: %v\n", conversations)
				if err != nil {
					log.Println("Error sending conversations to user: ", err)
				}
			}

		default:
			log.Printf("Unknown message type: %s\n", base.Type)

		}
	}
}
