package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"testing"
	"time"
)

func TestSessionIdentityAndValidation(t *testing.T) {
	secret := "test-secret"
	value, err := GenerateToken(secret, "user-123", "person@example.com")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := ParseToken(secret, value)
	if err != nil || claims.Subject != "user-123" || claims.Email != "person@example.com" {
		t.Fatalf("identity round trip failed: %v", err)
	}
	if _, err := ParseToken("wrong-secret", value); err == nil {
		t.Fatal("accepted a wrong signing key")
	}
	for _, test := range []struct {
		name   string
		method jwt.SigningMethod
		expiry *jwt.NumericDate
		issuer string
	}{
		{"wrong algorithm", jwt.SigningMethodHS384, jwt.NewNumericDate(time.Now().Add(time.Hour)), "unalone"},
		{"expired", jwt.SigningMethodHS256, jwt.NewNumericDate(time.Now().Add(-time.Minute)), "unalone"},
		{"missing expiration", jwt.SigningMethodHS256, nil, "unalone"},
		{"wrong issuer", jwt.SigningMethodHS256, jwt.NewNumericDate(time.Now().Add(time.Hour)), "another-app"},
	} {
		t.Run(test.name, func(t *testing.T) {
			value, err := jwt.NewWithClaims(test.method, Claims{Email: "person@example.com", RegisteredClaims: jwt.RegisteredClaims{Subject: "user-123", Issuer: test.issuer, ExpiresAt: test.expiry}}).SignedString([]byte(secret))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := ParseToken(secret, value); err == nil {
				t.Fatal("invalid session accepted")
			}
		})
	}
}
