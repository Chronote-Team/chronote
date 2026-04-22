package http

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRoutes, handler *Handler) {
	router.GET("/health", handler.HealthCheck)
	router.GET("/health/details", handler.HealthDetails)
}
