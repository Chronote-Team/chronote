package router

import (
	"chronote/controllers"
	"chronote/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	v1 := r.Group("/v1")
	{
		PublicUser := v1.Group("/user")
		{
			PublicUser.POST("/register", controllers.Register)
			PublicUser.POST("/login", controllers.Login)
			PublicUser.POST("/refresh", controllers.RefreshToken)
		}
		ProtectedUser := v1.Group("/user")
		ProtectedUser.Use(middlewares.JWTAuthMiddlewares())
		{
			ProtectedUser.GET("/info", controllers.UserInfo)
			ProtectedUser.POST("/logout", controllers.Logout)
		}
	}

	return r
}
