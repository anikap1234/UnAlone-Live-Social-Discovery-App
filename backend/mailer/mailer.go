package mailer

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
	"time"
	"unalone/backend/config"
)

type Sender interface {
	DeliverOTP(context.Context, string, string) error
	Mode() string
}

func New(mode string, cfg config.SMTPConfig) (Sender, error) {
	switch mode {
	case "log":
		return LogSender{}, nil
	case "smtp":
		if err := cfg.Validate(); err != nil {
			return nil, err
		}
		return &SMTPSender{config: cfg}, nil
	default:
		return nil, errors.New("unsupported OTP delivery mode")
	}
}

type LogSender struct{ Writer io.Writer }

func (s LogSender) DeliverOTP(_ context.Context, recipient, code string) error {
	writer := s.Writer
	if writer == nil {
		writer = os.Stdout
	}
	_, err := fmt.Fprintf(writer, "Development OTP for %s: %s (expires in 5m0s)\n", recipient, code)
	return err
}
func (LogSender) Mode() string { return "development-log" }

type SMTPSender struct{ config config.SMTPConfig }

func (s *SMTPSender) Mode() string { return "email" }

func (s *SMTPSender) DeliverOTP(ctx context.Context, recipient, code string) error {
	from, err := mail.ParseAddress(s.config.From)
	if err != nil {
		return errors.New("invalid SMTP sender address")
	}
	address := net.JoinHostPort(s.config.Host, fmt.Sprintf("%d", s.config.Port))
	tlsConfig := &tls.Config{ServerName: s.config.Host, MinVersion: tls.VersionTLS12}
	dialer := &net.Dialer{Timeout: time.Duration(s.config.TimeoutSeconds) * time.Second}
	var conn net.Conn
	if s.config.TLSMode == "implicit" {
		conn, err = (&tls.Dialer{NetDialer: dialer, Config: tlsConfig}).DialContext(ctx, "tcp", address)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", address)
	}
	if err != nil {
		return fmt.Errorf("connect SMTP server: %w", err)
	}
	defer conn.Close()
	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return fmt.Errorf("start SMTP client: %w", err)
	}
	defer client.Close()
	if s.config.TLSMode == "starttls" {
		ok, _ := client.Extension("STARTTLS")
		if !ok {
			return errors.New("SMTP server does not support STARTTLS")
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("start SMTP TLS: %w", err)
		}
	}
	if err := client.Auth(smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)); err != nil {
		return fmt.Errorf("authenticate SMTP sender: %w", err)
	}
	if err := client.Mail(from.Address); err != nil {
		return fmt.Errorf("set SMTP sender: %w", err)
	}
	if err := client.Rcpt(recipient); err != nil {
		return fmt.Errorf("set SMTP recipient: %w", err)
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("start SMTP message: %w", err)
	}
	if _, err := io.WriteString(writer, message(s.config.From, recipient, code)); err != nil {
		_ = writer.Close()
		return fmt.Errorf("write SMTP message: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("finish SMTP message: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("finish SMTP session: %w", err)
	}
	return nil
}

func message(from, recipient, code string) string {
	subject := mime.QEncoding.Encode("UTF-8", "Your UnAlone sign-in code")
	body := "Your UnAlone sign-in code is: " + code + "\n\n" +
		"It expires in 5 minutes. If you did not request this code, you can ignore this email.\n"
	return strings.Join([]string{
		"From: " + from,
		"To: " + recipient,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: 8bit",
		"",
		body,
	}, "\n")
}
