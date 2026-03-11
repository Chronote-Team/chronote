package router

import (
	"chronote/controllers"
	"chronote/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

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
		postcards := v1.Group("/postcards")
		postcards.Use(middlewares.JWTAuthMiddlewares())
		{
			postcards.POST("", controllers.CreatePostcard)
			postcards.GET("", controllers.GetPostcards)
			postcards.GET("/:id", controllers.GetPostcardDetail)
			postcards.PUT("/:id", controllers.UpdatePostcard)
			postcards.DELETE("/:id", controllers.DeletePostcard)

			postcards.POST("/:id/media", controllers.UploadMedia)
			postcards.GET("/:id/media", controllers.GetMedias)
			postcards.PUT("/:id/media/reorder", controllers.ReorderMedia)
			postcards.DELETE("/:id/media/:media_id", controllers.DeleteMedia)
		}
	}

	return r
}
