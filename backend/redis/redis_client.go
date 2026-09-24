package redis

import (
	"context"
	"fmt"
	rdb "github.com/redis/go-redis/v9"
	"strings"
	"time"
	"unalone/backend/config"
)

var Client *rdb.Client

func Init(cfg *config.Config) error {
	options := &rdb.Options{Addr: cfg.RedisURL}
	if strings.Contains(cfg.RedisURL, "://") {
		var err error
		options, err = rdb.ParseURL(cfg.RedisURL)
		if err != nil {
			return fmt.Errorf("invalid Redis URL: %w", err)
		}
	}
	Client = rdb.NewClient(options)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := Client.Ping(ctx).Err(); err != nil {
		return err
	}
	settings, err := Client.ConfigGet(ctx, "notify-keyspace-events").Result()
	if err != nil {
		return fmt.Errorf("read Redis expiry configuration: %w", err)
	}
	events := settings["notify-keyspace-events"]
	if !strings.Contains(events, "E") {
		events += "E"
	}
	if !strings.Contains(events, "x") {
		events += "x"
	}
	if err := Client.ConfigSet(ctx, "notify-keyspace-events", events).Err(); err != nil {
		return err
	}
	// Fail closed: local live-location and OTP storage must not be snapshotted.
	for _, key := range []string{"save", "appendonly"} {
		values, err := Client.ConfigGet(ctx, key).Result()
		if err != nil {
			return err
		}
		if (key == "save" && values[key] != "") || (key == "appendonly" && values[key] != "no") {
			return fmt.Errorf("ephemeral Redis requires save disabled and appendonly no (%s is %q)", key, values[key])
		}
	}
	return nil
}
