package ws

import (
	"encoding/json"
	"github.com/edisss1/fiabesco-backend/helpers"
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

			recipientConn, ok := clients[payload.RecipientID]
			if ok {
				err = recipientConn.WriteJSON(message)
				log.Printf("Sent message to recipient: %v\n", message)
				if err != nil {
					log.Println("Error sending message to recipient: ", err)
				}
			} else {
				log.Printf("Recipient not connected: %s\n", payload.RecipientID)
			}

			senderConn, ok := clients[payload.SenderID]
			if ok {
				err = senderConn.WriteJSON(message)
				log.Printf("Sent message to sender: %v\n", message)
				if err != nil {
					log.Println("Error sending message to sender: ", err)
				}
			} else {
				log.Printf("Sender not connected: %s\n", payload.SenderID)
			}

		default:
			log.Printf("Unknown message type: %s\n", base.Type)
		}
	}
}
