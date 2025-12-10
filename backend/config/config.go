package config

import (
    "os"
    "time"
)

type ServerConfig struct {
    Address string
}

type Config struct {
    Server   ServerConfig
    MongoURI string
    MongoDB  string
    RedisURL string
    JWTSecret string
    OTPMode  string
    Env      string
}

// Load reads from env and returns config with defaults
func Load() *Config {
    cfg := &Config{
        Server: ServerConfig{
            Address: getEnv("PORT", "8080"),
        },
        MongoURI: getEnv("MONGO_URI", "mongodb://localhost:27017"),
        MongoDB:  getEnv("MONGO_DB", "unalone"),
        RedisURL: getEnv("REDIS_URL", "localhost:6379"),
        JWTSecret: getEnv("JWT_SECRET", "supersecretlocal"),
        OTPMode:  getEnv("OTP_MODE", "log"),
        Env:      getEnv("ENV", "development"),
    }

    // for Fiber Listen, include colon
    if cfg.Server.Address[0] != ':' {
        cfg.Server.Address = ":" + cfg.Server.Address
    }

    return cfg
}

func getEnv(k, d string) string {
    if v := os.Getenv(k); v != "" {
        return v
    }
    return d
}

func TimeoutDuration() time.Duration { return 10 * time.Second }
