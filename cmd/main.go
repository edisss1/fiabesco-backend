package main

import (
	"github.com/edisss1/fiabesco-backend/db"
	"github.com/edisss1/fiabesco-backend/handlers/auth"
	"github.com/edisss1/fiabesco-backend/handlers/post"
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

	app.Post("/auth/signup", auth.SignUp)
	app.Post("/auth/login", auth.Login)
	app.Post("/user/:_id/post/create", post.CreatePost)
	app.Get("/user/:_id/post/get-all", post.GetPostsByUser)
	app.Delete("/user/:_id/post/delete/:postID", post.DeletePost)
	app.Get("/post/get-feed", post.GetFeedPosts)

	if PORT == "" {
		PORT = "3000"
	}
	if err := app.Listen(":" + PORT); err != nil {
		log.Fatal(err)
	}

}
