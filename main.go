package main

import (
	"chronote/config"
	"chronote/router"
)

func main() {
	config.InitConfig()
	config.InitDB()
	config.InitRedis()

	r := router.SetupRouter()

	port := ":8080"

	_ = r.Run(port)
}
