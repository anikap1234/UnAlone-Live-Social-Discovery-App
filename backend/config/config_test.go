package config

import "testing"

func TestValidateSMTPConfiguration(t *testing.T) {
	valid := &Config{OTPMode: "smtp", SMTP: SMTPConfig{
		Host: "smtp.example.com", Port: 587, Username: "user", Password: "secret",
		From: "UnAlone <auth@example.com>", TLSMode: "starttls", TimeoutSeconds: 10,
	}}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid SMTP configuration rejected: %v", err)
	}
	valid.SMTP.TLSMode = "none"
	if err := valid.Validate(); err == nil {
		t.Fatal("insecure SMTP configuration was accepted")
	}
	valid.OTPMode = "unknown"
	if err := valid.Validate(); err == nil {
		t.Fatal("unknown OTP delivery mode was accepted")
	}
}
