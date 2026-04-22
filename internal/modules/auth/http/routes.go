package http

import "github.com/gin-gonic/gin"

func RegisterPublicRoutes(group *gin.RouterGroup, handler *Handler) {
	group.POST("/login", handler.Login)
	group.POST("/refresh", handler.RefreshToken)
}

func RegisterProtectedRoutes(group *gin.RouterGroup, handler *Handler) {
	group.POST("/logout", handler.Logout)
}
