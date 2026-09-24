package config

import (
	"fmt"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"time"
)

type ServerConfig struct{ Address string }
type SMTPConfig struct {
	Host           string
	Port           int
	Username       string
	Password       string
	From           string
	TLSMode        string
	TimeoutSeconds int
}
type Config struct {
	Server         ServerConfig
	MongoURI       string
	MongoDB        string
	RedisURL       string
	JWTSecret      string
	OTPMode        string
	SMTP           SMTPConfig
	Env            string
	AllowedOrigins string
}

func Load() *Config {
	port := getEnv("PORT", "8080")
	if !strings.Contains(port, ":") {
		port = ":" + port
	}
	return &Config{
		Server:   ServerConfig{Address: port},
		MongoURI: getEnv("MONGO_URI", "mongodb://127.0.0.1:27018"), MongoDB: getEnv("MONGO_DB", "unalone"),
		RedisURL: getEnv("REDIS_URL", "127.0.0.1:6380"), JWTSecret: getEnv("JWT_SECRET", "local-development-only-change-before-deploy"),
		OTPMode: getEnv("OTP_MODE", "log"), Env: getEnv("ENV", "development"),
		SMTP: SMTPConfig{
			Host: getEnv("SMTP_HOST", ""), Port: getEnvInt("SMTP_PORT", 587),
			Username: getEnv("SMTP_USERNAME", ""), Password: getEnv("SMTP_PASSWORD", ""),
			From: getEnv("SMTP_FROM", ""), TLSMode: strings.ToLower(getEnv("SMTP_TLS_MODE", "starttls")),
			TimeoutSeconds: getEnvInt("SMTP_TIMEOUT_SECONDS", 10),
		},
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173"),
	}
}

func (c *Config) Validate() error {
	switch c.OTPMode {
	case "log":
		return nil
	case "smtp":
		return c.SMTP.Validate()
	default:
		return fmt.Errorf("OTP_MODE must be log or smtp")
	}
}

func (c SMTPConfig) Validate() error {
	if c.Host == "" || c.Username == "" || c.Password == "" || c.From == "" {
		return fmt.Errorf("OTP_MODE=smtp requires SMTP_HOST, SMTP_USERNAME, SMTP_PASSWORD, and SMTP_FROM")
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("SMTP_PORT must be between 1 and 65535")
	}
	if c.TimeoutSeconds < 1 || c.TimeoutSeconds > 60 {
		return fmt.Errorf("SMTP_TIMEOUT_SECONDS must be between 1 and 60")
	}
	if c.TLSMode != "starttls" && c.TLSMode != "implicit" {
		return fmt.Errorf("SMTP_TLS_MODE must be starttls or implicit")
	}
	address, err := mail.ParseAddress(c.From)
	if err != nil || address.Address == "" {
		return fmt.Errorf("SMTP_FROM must be a valid email address")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
func getEnvInt(key string, fallback int) int {
	value, err := strconv.Atoi(getEnv(key, strconv.Itoa(fallback)))
	if err != nil {
		return 0
	}
	return value
}
func TimeoutDuration() time.Duration { return 10 * time.Second }
