//go:build !prod

package router

import (
	"chronote/config"

	_ "chronote/docs" // swagger docs
	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/files"

	"github.com/gin-gonic/gin"
)

// setupSwagger registers Swagger UI routes
// This file is only compiled when the 'prod' build tag is NOT set
func setupSwagger(r *gin.Engine) {
	if config.AppConfig.Swagger.Enabled {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
}
