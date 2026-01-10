//go:build !prod

package main

import (
	"chronote/config"
	"chronote/router"

	_ "chronote/docs" // swagger docs
)

// @title           Chronote API
// @version         1.0
// @description     API documentation for Chronote user authentication system
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	config.InitConfig()
	config.InitDB()
	config.InitRedis()
	config.InitS3()

	r := router.SetupRouter()

	port := ":8080"

	_ = r.Run(port)
}
