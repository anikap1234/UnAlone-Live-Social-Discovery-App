package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/logger"

    "unalone/backend/api"
    "unalone/backend/config"
    dbpkg "unalone/backend/db"
    "unalone/backend/redis"
    "unalone/backend/ws"
)

func main() {
    cfg := config.Load()

    // Initialize services
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := dbpkg.Init(ctx, cfg); err != nil {
        log.Fatalf("mongodb init: %v", err)
    }

    if err := redis.Init(cfg); err != nil {
        log.Fatalf("redis init: %v", err)
    }

    hub := ws.NewHub()
    go hub.Run()

    app := fiber.New()
    app.Use(logger.New())

    api.RegisterRoutes(app, cfg, hub)

    // Graceful shutdown
    go func() {
        if err := app.Listen(cfg.Server.Address); err != nil {
            log.Printf("server error: %v", err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("shutting down")
    ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancelShutdown()
    _ = app.Shutdown()
    dbpkg.Disconnect(ctxShutdown)
}
