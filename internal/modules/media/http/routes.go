package http

import "github.com/gin-gonic/gin"

func RegisterPublicRoutes(group *gin.RouterGroup, handler *Handler) {
	group.GET("/:id/media", handler.GetMedias)
}

func RegisterProtectedRoutes(group *gin.RouterGroup, handler *Handler) {
	group.POST("/:id/media", handler.UploadMedia)
	group.PUT("/:id/media/reorder", handler.ReorderMedia)
	group.DELETE("/:id/media/:media_id", handler.DeleteMedia)
}
