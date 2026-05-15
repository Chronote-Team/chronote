package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	appplatform "chronote-refactor/internal/platform/app"
	platformconfig "chronote-refactor/internal/platform/config"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("run worker: %v", err)
	}
}

func run() error {
	cfg, err := platformconfig.Load()
	if err != nil {
		return err
	}
	worker, err := appplatform.NewAnalysisWorkerFromConfig(cfg)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return worker.Run(ctx, appplatform.WorkerOptionsFromConfig(cfg))
}
