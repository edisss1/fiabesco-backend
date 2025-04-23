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

	app.Get("/ws", websocket.New(ws.HandleWS))

	port := config.GetPort()

	log.Printf("Server running on port %s", port)
	log.Fatal(app.Listen(":" + port))

}
