package redis

import (
    "context"
    "fmt"
    "strings"

    rdb "github.com/redis/go-redis/v9"
    "unalone/backend/config"
)

var Client *rdb.Client

func Init(cfg *config.Config) error {
    addr := cfg.RedisURL
    // allow redis://proto if provided
    if strings.HasPrefix(addr, "redis://") {
        addr = strings.TrimPrefix(addr, "redis://")
    }
    Client = rdb.NewClient(&rdb.Options{Addr: addr})
    if err := Client.Ping(context.Background()).Err(); err != nil {
        return fmt.Errorf("redis ping: %w", err)
    }
    return nil
}
