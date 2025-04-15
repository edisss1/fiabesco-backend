package routes

import (
	"github.com/edisss1/fiabesco-backend/handlers/auth"
	"github.com/edisss1/fiabesco-backend/handlers/messages"
	"github.com/edisss1/fiabesco-backend/handlers/post"
	"github.com/edisss1/fiabesco-backend/handlers/user"
	"github.com/edisss1/fiabesco-backend/middleware"
	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
	authRoutes(app)
	userRoutes(app)
	postRoutes(app)
	messageRoutes(app)
}

func authRoutes(app *fiber.App) {
	app.Post("/auth/signup", auth.SignUp)
	app.Post("/auth/login", auth.Login)
}

func userRoutes(app *fiber.App) {
	app.Post("/users/me", middleware.RequireJWT, user.GetUserData)
	app.Patch("/users/:_id/photo", middleware.RequireJWT, user.UpdatePhotoURL)
	app.Get("/users/profile/:_id", middleware.RequireJWT, user.GetProfileData)
	app.Put("/users/:_id", middleware.RequireJWT, user.BlockUser)
}

func postRoutes(app *fiber.App) {
	app.Post("/users/:_id/posts", middleware.RequireJWT, post.CreatePost)
	app.Get("/users/:_id/post", middleware.RequireJWT, post.GetPostsByUser)
	app.Delete("/users/:_id/posts/:postID", middleware.RequireJWT, post.DeletePost)
	app.Get("/posts/feed", middleware.RequireJWT, post.GetFeedPosts)
	app.Patch("/posts/:_id", middleware.RequireJWT, post.UpdatePostCaption)
}

func messageRoutes(app *fiber.App) {
	app.Post("/conversations/start", middleware.RequireJWT, messages.StartConversation)
	app.Post("/conversations/:conversationID/messages/:senderID", middleware.RequireJWT, messages.SendMessage)
	app.Delete("/messages/delete", middleware.RequireJWT, messages.DeleteMessage)
	app.Delete("/conversations/:conversationID", middleware.RequireJWT, messages.DeleteConversation)
	app.Patch("/messages/:_id", middleware.RequireJWT, messages.EditMessage)
}
