//go:build !prod

package main

import (
	"chronote/config"
	"chronote/router"
)

func main() {
	config.InitConfig()
	config.InitDB()
	config.InitRedis()
	config.InitS3()

	r := router.SetupRouter()

	port := ":8080"

	_ = r.Run(port)
}
