package main

import (
	"github.com/edisss1/fiabesco-backend/handlers/ws"
	"github.com/edisss1/fiabesco-backend/internal/config"
	"github.com/edisss1/fiabesco-backend/internal/server"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"log"
)

func main() {
	config.LoadEnv()
	config.ConnectDB()

	app := server.Setup()

	app.Use(func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return c.SendStatus(fiber.StatusUpgradeRequired)

	})

	go ws.RunHub()

	app.Get("/ws", websocket.New(func(conn *websocket.Conn) {
		defer func() {
			ws.Unregister <- conn
			conn.Close()

		}()
		ws.Register <- conn

		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Println("read error:", err)
				}

				return // Calls the deferred function, i.e. closes the connection on error
			}

			if messageType == websocket.TextMessage {
				// Broadcast the received message
				ws.Broadcast <- string(message)
			} else {
				log.Println("websocket message received of type", messageType)
			}
		}
	}))

	port := config.GetPort()

	log.Printf("Server running on port %s", port)
	log.Fatal(app.Listen(":" + port))

}
