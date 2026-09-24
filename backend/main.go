package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"unalone/backend/api"
	"unalone/backend/config"
	"unalone/backend/db"
	"unalone/backend/redis"
	"unalone/backend/services"
	"unalone/backend/ws"
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration: %v", err)
	}
	if cfg.Env == "production" {
		log.Fatal("This milestone is local-only; configure production delivery and deployment before enabling production mode")
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	startup, stop := context.WithTimeout(ctx, 20*time.Second)
	defer stop()
	if err := db.Init(startup, cfg); err != nil {
		log.Fatalf("MongoDB initialization: %v", err)
	}
	if err := redis.Init(cfg); err != nil {
		log.Fatalf("Redis initialization: %v", err)
	}
	hub := ws.NewHub()
	hotspots := services.NewHotspotService(hub)
	if err := hotspots.StartExpiryListener(ctx); err != nil {
		log.Fatalf("Presence expiry listener: %v", err)
	}
	app := api.NewApp(cfg, hub, hotspots)
	failures := make(chan error, 1)
	go func() { failures <- app.Listen(cfg.Server.Address) }()
	log.Printf("UnAlone local API listening on %s", cfg.Server.Address)
	select {
	case <-ctx.Done():
	case err := <-failures:
		log.Printf("HTTP server stopped: %v", err)
	}
	cancel()
	hub.Close()
	shutdown, done := context.WithTimeout(context.Background(), 10*time.Second)
	defer done()
	_ = app.ShutdownWithContext(shutdown)
	_ = db.Disconnect(shutdown)
	_ = redis.Client.Close()
}
