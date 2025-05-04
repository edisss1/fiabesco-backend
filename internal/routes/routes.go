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
	users := app.Group("/users", middleware.RequireJWT)

	users.Post("/me", user.GetUserData)
	users.Patch("/:_id/photo", user.UpdatePhotoURL)
	users.Get("/profile/:_id", user.GetProfileData)
	users.Put("/:_id/block", user.BlockUser)
	users.Put("/unblock", user.UnblockUser)
	users.Put("/:_id/bio", user.EditBio)
	users.Get("/:_id/following", user.GetFollowing)
	users.Post("/:_id/follow", user.FollowUser)

}

func postRoutes(app *fiber.App) {
	users := app.Group("/users", middleware.RequireJWT)
	posts := app.Group("/posts", middleware.RequireJWT)

	users.Post("/:_id/posts", post.CreatePost)
	users.Get("/:_id/post", post.GetPostsByUser)
	users.Delete("/:_id/posts/:postID", post.DeletePost)
	posts.Get("/feed", post.GetFeedPosts)
	posts.Patch("/:_id/caption", post.UpdatePostCaption)
	posts.Post("/like", post.LikePost)
	posts.Get("/:postID", post.GetPost)
	posts.Post("/:postID/comment", post.CommentPost)
	posts.Get("/:postID/comments", post.GetComments)
}

func messageRoutes(app *fiber.App) {
	conversations := app.Group("/conversations", middleware.RequireJWT)
	message := app.Group("/messages", middleware.RequireJWT)

	conversations.Post("/start", messages.StartConversation)
	conversations.Post("/:conversationID/messages/:senderID", messages.SendMessage)
	conversations.Delete("/:conversationID", messages.DeleteConversation)
	conversations.Get("/conversation/:conversationID", messages.GetConversation)
	conversations.Get("/:userID", messages.GetConversations)
	message.Patch("/messages/:_id", messages.EditMessage)
	message.Delete("/messages/delete", messages.DeleteMessage)
}
