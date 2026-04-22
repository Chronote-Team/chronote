package http

import "github.com/gin-gonic/gin"

func RegisterPublicRoutes(group *gin.RouterGroup, handler *Handler) {
	group.GET("", handler.GetPostcards)
	group.GET("/:id", handler.GetPostcardDetail)
}

func RegisterProtectedRoutes(group *gin.RouterGroup, handler *Handler) {
	group.POST("", handler.CreatePostcard)
	group.PUT("/:id", handler.UpdatePostcard)
	group.DELETE("/:id", handler.DeletePostcard)
}
