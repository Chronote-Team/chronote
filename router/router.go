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

	return r
}
