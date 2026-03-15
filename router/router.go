package router

import (
	"chronote/controllers"
	"chronote/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Health check endpoints - no auth required
	r.GET("/health", controllers.HealthCheck)
	r.GET("/health/details", controllers.HealthDetails)

	// Public routes
	PublicUser := r.Group("/user")
	{
		PublicUser.POST("/register", controllers.Register)
		PublicUser.POST("/login", controllers.Login)
		PublicUser.POST("/refresh", controllers.RefreshToken)
	}

	// Protected routes (require JWT)
	ProtectedUser := r.Group("/user")
	ProtectedUser.Use(middlewares.JWTAuthMiddlewares())
	{
		ProtectedUser.GET("/info", controllers.UserInfo)
		ProtectedUser.POST("/logout", controllers.Logout)
		ProtectedUser.POST("/avatar", controllers.UploadAvatar)
		ProtectedUser.PUT("/update/displayname", controllers.UpdateDisplayName)
		ProtectedUser.PUT("/update/password", controllers.UpdatePassword)
	}

	v1 := r.Group("/v1")
	{
		publicPostcards := v1.Group("/postcards")
		publicPostcards.Use(middlewares.OptionalJWTAuthMiddlewares())
		{
			publicPostcards.GET("", controllers.GetPostcards)
			publicPostcards.GET("/:id", controllers.GetPostcardDetail)
			publicPostcards.GET("/:id/media", controllers.GetMedias)
		}

		protectedPostcards := v1.Group("/postcards")
		protectedPostcards.Use(middlewares.JWTAuthMiddlewares())
		{
			protectedPostcards.POST("", controllers.CreatePostcard)
			protectedPostcards.PUT("/:id", controllers.UpdatePostcard)
			protectedPostcards.DELETE("/:id", controllers.DeletePostcard)

			protectedPostcards.POST("/:id/media", controllers.UploadMedia)
			protectedPostcards.PUT("/:id/media/reorder", controllers.ReorderMedia)
			protectedPostcards.DELETE("/:id/media/:media_id", controllers.DeleteMedia)
		}
	}

	return r
}
