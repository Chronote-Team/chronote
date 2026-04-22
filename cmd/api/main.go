package main

import (
	"log"

	appplatform "chronote-refactor/internal/platform/app"
	platformhttp "chronote-refactor/internal/platform/http"
)

func main() {
	app, err := appplatform.New()
	if err != nil {
		log.Fatalf("build app: %v", err)
	}

	server := platformhttp.NewServer(":8080", app.Router())
	if err := server.Run(); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
