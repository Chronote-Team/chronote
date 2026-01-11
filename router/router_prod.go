//go:build prod

package router

import "github.com/gin-gonic/gin"

// setupSwagger is a no-op in production builds
// This file is only compiled when the 'prod' build tag IS set
// Production builds do not include Swagger UI or its dependencies
func setupSwagger(r *gin.Engine) {
	// No Swagger routes in production
}
