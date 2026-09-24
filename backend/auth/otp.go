package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	rdb "github.com/redis/go-redis/v9"
	"unalone/backend/mailer"
	"unalone/backend/redis"
)

var ErrRateLimited = errors.New("Too many attempts. Please try again in an hour")

func emailKey(email string) string {
	sum := sha256.Sum256([]byte(email))
	return hex.EncodeToString(sum[:])
}

func otpHash(email, code string) string { return emailKey(email + ":" + code) }

var sendScript = rdb.NewScript(`
local attempts = redis.call('INCR', KEYS[1])
if attempts == 1 then redis.call('EXPIRE', KEYS[1], 3600) end
if attempts > 5 then return 0 end
redis.call('SET', KEYS[2], ARGV[1], 'EX', 300)
return 1
`)

var verifyScript = rdb.NewScript(`
local attempts = redis.call('INCR', KEYS[1])
if attempts == 1 then redis.call('EXPIRE', KEYS[1], 3600) end
if attempts > 5 then return -1 end
local stored = redis.call('GET', KEYS[2])
if not stored or stored ~= ARGV[1] then return 0 end
redis.call('DEL', KEYS[2])
return 1
`)

var removeCodeScript = rdb.NewScript(`
if redis.call('GET', KEYS[1]) == ARGV[1] then redis.call('DEL', KEYS[1]) end
return 1
`)

func SendOTP(ctx context.Context, email string, delivery mailer.Sender) (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	code := fmt.Sprintf("%06d", n.Int64())
	suffix := emailKey(email)
	ok, err := sendScript.Run(ctx, redis.Client, []string{"otp:send:" + suffix, "otp:code:" + suffix}, otpHash(email, code)).Int()
	if err != nil {
		return "", err
	}
	if ok == 0 {
		return "", ErrRateLimited
	}
	if err := delivery.DeliverOTP(ctx, email, code); err != nil {
		// Delivery and Redis cannot be one transaction. Remove only this code, never a newer retry.
		_, _ = removeCodeScript.Run(ctx, redis.Client, []string{"otp:code:" + suffix}, otpHash(email, code)).Result()
		return "", fmt.Errorf("deliver OTP: %w", err)
	}
	return code, nil
}

func VerifyOTP(ctx context.Context, email, code string) (bool, error) {
	suffix := emailKey(email)
	result, err := verifyScript.Run(ctx, redis.Client, []string{"otp:verify:" + suffix, "otp:code:" + suffix}, otpHash(email, code)).Int()
	if err != nil {
		return false, err
	}
	if result == -1 {
		return false, ErrRateLimited
	}
	return result == 1, nil
}
