package api

import (
    "context"
    "time"

    "github.com/gofiber/fiber/v2"
)

func ctxWithTimeout() (context.Context, context.CancelFunc) {
    return context.WithTimeout(context.Background(), 10*time.Second)
}
