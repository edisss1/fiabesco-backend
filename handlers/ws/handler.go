package ws

import (
	"encoding/json"
	"fmt"
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

type UpdateStatusPayload struct {
	UserID string `json:"userID"`
	Status string `json:"status"`
}

type GetConversationsPayload struct {
	UserID string `json:"userID"`
}

func HandleWS(conn *websocket.Conn) {
	userID := conn.Query("userID")

	mu.Lock()
	clients[userID] = conn
	mu.Unlock()

	defer func() {
		mu.Lock()
		delete(clients, userID)
		mu.Unlock()
		conn.Close()
	}()

	for {
		var base BaseWSMessage
		if err := conn.ReadJSON(&base); err != nil {
			log.Println("read error: ", err)
			break
		}

		switch base.Type {
		case "send_message":
			var payload SendMessagePayload
			if err := json.Unmarshal(base.Data, &payload); err != nil {
				log.Println("Unmarshal error: ", err)
				continue
			}

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

			conversation, err := helpers.GetConversation(conversationID)
			if err != nil {
				log.Println("Error getting conversation: ", err)
			}

			fmt.Println(clients)
			fmt.Println("Participants: ", conversation.Participants)

			for _, user := range conversation.Participants {
				fmt.Println("user: ", user)
				conn, ok := clients[user.ID.Hex()]
				if ok {
					err = conn.WriteJSON(struct {
						Type    string        `json:"type"`
						Message types.Message `json:"message"`
					}{
						Type:    "conversations_update",
						Message: message,
					})
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

			if err != nil {
				log.Println("Error saving message: ", err)
				continue
			}

			conversation, err := helpers.GetConversation(message.ConversationID)
			if err != nil {
				log.Println("Error getting conversation: ", err)
			}

			for _, user := range conversation.Participants {

				if conn, ok := clients[user.ID.Hex()]; ok {
					err = conn.WriteJSON(message)
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
				if err != nil {
					log.Println("Error sending conversations to user: ", err)
				}
			}
		case "update_status":
			var payload UpdateStatusPayload
			if err := json.Unmarshal(base.Data, &payload); err != nil {
				log.Println("Unmarshal error: ", err)
				continue
			}
			userID, err := utils.ParseHexID(payload.UserID)
			if err != nil {
				log.Println("Invalid userID: ", err)
				continue
			}

			err = helpers.UpdateUserStatus(userID, payload.Status)
			if err != nil {
				log.Println("Error updating user status: ", err)
				continue
			}
		default:
			log.Printf("Unknown message type: %s\n", base.Type)

		}
	}
}
