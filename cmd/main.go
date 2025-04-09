package main

import (
	"github.com/edisss1/fiabesco-backend/db"
	"github.com/edisss1/fiabesco-backend/handlers/auth"
	"github.com/edisss1/fiabesco-backend/handlers/post"
	"github.com/edisss1/fiabesco-backend/handlers/user"
	"github.com/edisss1/fiabesco-backend/middleware"
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
	app.Post("/users/:_id/posts", middleware.RequireJWT, post.CreatePost)
	app.Get("/users/:_id/post", middleware.RequireJWT, post.GetPostsByUser)
	app.Delete("/users/:_id/posts/:postID", middleware.RequireJWT, post.DeletePost)
	app.Get("/posts/feed", middleware.RequireJWT, post.GetFeedPosts)
	app.Patch("/posts/:_id", middleware.RequireJWT, post.UpdatePostCaption)
	app.Patch("/users/:_id/photo", middleware.RequireJWT, user.UpdatePhotoURL)

	if PORT == "" {
		PORT = "3000"
	}
	if err := app.Listen(":" + PORT); err != nil {
		log.Fatal(err)
	}

}
