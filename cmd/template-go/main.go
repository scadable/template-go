package main

import (
	"context"
	"log"
	"net/http" // keep standard lib as is

	"template-go/internal/config"
	delivery "template-go/internal/delivery/http"
	"template-go/internal/otel"
	"template-go/pkg/logger"
)

func main() {
	ctx := context.Background()
	cfg := config.MustLoad()

	logger.Init()
	defer logger.Sync()

	shutdown, err := otel.InitTracer(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to init tracing: %v", err)
	}
	defer shutdown(ctx)

	log.Printf("🚀 Starting server on %s\n", cfg.ListenAddr)
	err = http.ListenAndServe(cfg.ListenAddr, delivery.NewRouter())
	if err != nil {
		log.Fatalf("server error: %v", err)
	}
}
