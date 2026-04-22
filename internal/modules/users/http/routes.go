package http

import "github.com/gin-gonic/gin"

func RegisterPublicRoutes(group *gin.RouterGroup, handler *Handler) {
	group.POST("/register", handler.Register)
}

func RegisterProtectedRoutes(group *gin.RouterGroup, handler *Handler) {
	group.GET("/info", handler.UserInfo)
	group.POST("/avatar", handler.UploadAvatar)
	group.PUT("/update/displayname", handler.UpdateDisplayName)
	group.PUT("/update/password", handler.UpdatePassword)
}
