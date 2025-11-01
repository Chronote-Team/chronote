package router

import (
	controller "chronote/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	api := r.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			user := v1.Group("/user")
			{
				// auth.POST("/login", controllers.Login)
				user.POST("/register", controller.Register)
			}
		}
	}
	return r
}
