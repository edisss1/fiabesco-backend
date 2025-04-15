package main

import (
	"github.com/edisss1/fiabesco-backend/internal/config"
	"github.com/edisss1/fiabesco-backend/internal/server"
	"log"
)

func main() {
	config.LoadEnv()
	config.ConnectDB()

	app := server.Setup()

	port := config.GetPort()

	log.Printf("Server running on port %s", port)
	log.Fatal(app.Listen(":" + port))

}
