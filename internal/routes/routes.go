package routes

import (
	"github.com/edisss1/fiabesco-backend/handlers/auth"
	"github.com/edisss1/fiabesco-backend/handlers/comments"
	"github.com/edisss1/fiabesco-backend/handlers/mail"
	"github.com/edisss1/fiabesco-backend/handlers/messages"
	"github.com/edisss1/fiabesco-backend/handlers/portfolio"
	"github.com/edisss1/fiabesco-backend/handlers/post"
	"github.com/edisss1/fiabesco-backend/handlers/settings"
	"github.com/edisss1/fiabesco-backend/handlers/social"
	"github.com/edisss1/fiabesco-backend/handlers/uploads"
	"github.com/edisss1/fiabesco-backend/handlers/user"
	"github.com/edisss1/fiabesco-backend/middleware"
	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
	authRoutes(app)
	userRoutes(app)
	postRoutes(app)
	messageRoutes(app)
	settingsRoutes(app)
	portfolioRoutes(app)
	servingRoutes(app)
	emailRoutes(app)
}

func authRoutes(app *fiber.App) {
	app.Post("/auth/signup", auth.SignUp)
	app.Post("/auth/login", auth.Login)
}

func userRoutes(app *fiber.App) {
	users := app.Group("/users", middleware.RequireJWT)

	users.Get("/me", user.GetUserData)
	users.Get("/profile/:_id", user.GetProfileData)
	users.Post("/:userID/block", social.BlockUser)
	users.Delete("/:userID/unblock", social.UnblockUser)
	users.Put("/:_id/bio", user.EditBio)
	users.Get("/:_id/following", social.GetFollowing)
	users.Post("/:_id/follow", social.FollowUser)
	users.Get("/:userID/blocked", social.GetBlockedUsers)
	users.Put("/:userID/pfp", user.ChangePFP)
	users.Put("/:userID/banner", user.UploadBanner)

}

func postRoutes(app *fiber.App) {
	users := app.Group("/users", middleware.RequireJWT)
	posts := app.Group("/posts", middleware.RequireJWT)

	users.Post("/:userID/posts", post.CreatePost)
	users.Get("/:userID/post", post.GetPostsByUser)
	users.Delete("/:_id/posts/:postID", post.DeletePost)
	posts.Get("/feed", post.GetFeedPosts)
	posts.Patch("/:_id/caption", post.UpdatePostCaption)
	posts.Post("/like", post.LikePost)
	posts.Get("/:postID", post.GetPost)
	posts.Post("/:postID/comment", comments.CommentPost)
	posts.Get("/:postID/comments", comments.GetComments)
	posts.Patch("/:commentID/edit", comments.EditComment)
	posts.Delete("/:commentID", comments.DeleteComment)
}

func messageRoutes(app *fiber.App) {
	conversations := app.Group("/conversations", middleware.RequireJWT)
	message := app.Group("/messages", middleware.RequireJWT)

	conversations.Post("/start", messages.StartConversation)
	conversations.Post("/:conversationID/messages/:senderID", messages.SendMessage)
	conversations.Delete("/:conversationID", messages.DeleteConversation)
	conversations.Get("/conversation/:conversationID", messages.GetConversation)
	conversations.Get("/all", messages.GetConversations)
	message.Patch("/:_id", messages.EditMessage)
	message.Delete("/delete", messages.DeleteMessage)

}

func settingsRoutes(app *fiber.App) {
	setting := app.Group("/settings", middleware.RequireJWT)

	setting.Put("/firstname", settings.ChangeFirstName)
	setting.Put("/lastname", settings.ChangeLastName)
	setting.Put("/email", settings.ChangeEmail)
	setting.Put("/handle", settings.ChangeHandle)
	setting.Put("/password", settings.ChangePassword)
	setting.Put("/theme", settings.ChangeTheme)
	setting.Put("/language", settings.ChangeLanguage)
	setting.Put("/visibility", settings.ChangeProfileVisibility)
	setting.Get("/data", settings.DownloadUserData)
}

func portfolioRoutes(app *fiber.App) {
	portfolios := app.Group("/portfolios/:userID", middleware.RequireJWT)

	portfolios.Post("/create/", portfolio.CreatePortfolio)
	portfolios.Get("/", portfolio.GetPortfolio)
}

func servingRoutes(app *fiber.App) {
	images := app.Group("/images")

	images.Get("/:imageID", uploads.ServeImage)
}

func emailRoutes(app *fiber.App) {
	emails := app.Group("/emails", middleware.RequireJWT)
	emails.Post("/send", mail.SendEmail)
}
