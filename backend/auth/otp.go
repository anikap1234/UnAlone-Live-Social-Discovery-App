package auth

import (
    "context"
    "crypto/rand"
    "fmt"
    "math/big"
    "time"

    rdb "github.com/redis/go-redis/v9"
    "unalone/backend/redis"
)

func generateCode() string {
    max := big.NewInt(10)
    s := ""
    for i := 0; i < 6; i++ {
        n, _ := rand.Int(rand.Reader, max)
        s += n.String()
    }
    return s
}

// SendOTP stores OTP in Redis and returns the code when mode is "log"
func SendOTP(ctx context.Context, email string, mode string) (string, error) {
    code := generateCode()
    key := fmt.Sprintf("otp:%s", email)
    if err := redis.Client.Set(ctx, key, code, 5*time.Minute).Err(); err != nil {
        return "", err
    }
    if mode == "log" {
        // Print to stdout for development
        fmt.Printf("OTP for %s: %s\n", email, code)
    }
    return code, nil
}

// VerifyOTP checks code stored in Redis
func VerifyOTP(ctx context.Context, email, code string) (bool, error) {
    key := fmt.Sprintf("otp:%s", email)
    v, err := redis.Client.Get(ctx, key).Result()
    if err != nil {
        if err == rdb.Nil {
            return false, nil
        }
        return false, err
    }
    if v != code {
        return false, nil
    }
    // delete after use
    _ = redis.Client.Del(ctx, key).Err()
    return true, nil
}
