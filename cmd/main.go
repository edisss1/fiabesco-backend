package main

import (
	"github.com/edisss1/fiabesco-backend/db"
	"github.com/edisss1/fiabesco-backend/handlers/auth"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file %v", err)
	}

	db.Connect()

	app := fiber.New()

	PORT := os.Getenv("PORT")

	app.Post("/signup", auth.SignUp)
	app.Post("/login", auth.Login)

	if PORT == "" {
		PORT = "3000"
	}
	if err := app.Listen(":" + PORT); err != nil {
		log.Fatal(err)
	}

}
