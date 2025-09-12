package main

import (
	"context"
	"go.uber.org/zap"
	"log"
	"net/http"

	"template-go/internal/config"
	delivery "template-go/internal/delivery/http"
	"template-go/internal/otel"
	"template-go/pkg/logger"

	_ "template-go/docs"
)

// @title           Template Go API
// @version         1.0
// @description     Sample template-go service with Swagger
// @host            localhost:8080
// @BasePath        /
func main() {
	ctx := context.Background()
	cfg := config.MustLoad()

	logger.Init()
	defer logger.Sync()

	shutdown, err := otel.InitOtel(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to init tracing: %v", err)
	}

	defer func() {
		if err := shutdown(ctx); err != nil {
			logger.Error(ctx, "failed to shutdown tracer", zap.Error(err))
		}
	}()

	log.Printf("🚀 Starting server on %s\n", cfg.ListenAddr)
	err = http.ListenAndServe(cfg.ListenAddr, delivery.NewRouter(cfg.OTELServiceName))
	if err != nil {
		log.Fatalf("server error: %v", err)
	}
}
