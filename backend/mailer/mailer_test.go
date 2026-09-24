package mailer

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestLogSenderWritesCode(t *testing.T) {
	var output bytes.Buffer
	if err := (LogSender{Writer: &output}).DeliverOTP(context.Background(), "person@example.com", "012345"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "person@example.com") || !strings.Contains(output.String(), "012345") {
		t.Fatalf("development output did not include recipient and code: %q", output.String())
	}
}

func TestMessageHasPlainTextOTP(t *testing.T) {
	value := message("UnAlone <auth@example.com>", "person@example.com", "012345")
	for _, want := range []string{
		"From: UnAlone <auth@example.com>",
		"To: person@example.com",
		"Subject: Your UnAlone sign-in code",
		"Your UnAlone sign-in code is: 012345",
		"It expires in 5 minutes.",
	} {
		if !strings.Contains(value, want) {
			t.Fatalf("email message missing %q", want)
		}
	}
}
