package sesv2

import (
	"context"
	"errors"
	"testing"
)

func TestSendEmail_RawEmailWithoutDestination(t *testing.T) {
	storage := &MemoryStorage{}
	ctx := context.Background()

	rawMessage := "From: sender@example.com\r\n" +
		"To: recipient@example.com\r\n" +
		"Cc: cc@example.com\r\n" +
		"Subject: Test Subject\r\n" +
		"Content-Type: text/plain\r\n" +
		"\r\n" +
		"Test body"

	req := &SendEmailRequest{
		FromEmailAddress: "sender@example.com",
		// Destination is intentionally nil
		Content: &EmailContent{
			Raw: &RawEmail{
				Data: []byte(rawMessage),
			},
		},
	}

	messageID, err := storage.SendEmail(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if messageID == "" {
		t.Fatal("expected non-empty message ID")
	}

	// Verify that the sent email was stored
	sentEmails, err := storage.GetSentEmails(ctx)
	if err != nil {
		t.Fatalf("failed to get sent emails: %v", err)
	}

	if len(sentEmails) == 0 {
		t.Fatal("expected sent email to be stored")
	}

	email := sentEmails[0]
	if email.MessageID != messageID {
		t.Errorf("expected message ID %s, got %s", messageID, email.MessageID)
	}

	// Verify destination was extracted from MIME headers
	if email.Destination == nil {
		t.Fatal("expected destination to be extracted from MIME headers")
	}

	if len(email.Destination.ToAddresses) != 1 || email.Destination.ToAddresses[0] != "recipient@example.com" {
		t.Errorf("expected To: recipient@example.com, got %v", email.Destination.ToAddresses)
	}

	if len(email.Destination.CcAddresses) != 1 || email.Destination.CcAddresses[0] != "cc@example.com" {
		t.Errorf("expected Cc: cc@example.com, got %v", email.Destination.CcAddresses)
	}

	if email.Subject != "Test Subject" {
		t.Errorf("expected subject 'Test Subject', got '%s'", email.Subject)
	}
}

func TestSendEmail_SimpleEmailWithoutDestination_ShouldFail(t *testing.T) {
	storage := &MemoryStorage{}
	ctx := context.Background()

	req := &SendEmailRequest{
		FromEmailAddress: "sender@example.com",
		// Destination is nil
		Content: &EmailContent{
			Simple: &SimpleEmail{
				Subject: &Content{
					Data: "Test",
				},
				Body: &Body{
					Text: &Content{
						Data: "Test body",
					},
				},
			},
		},
	}

	_, err := storage.SendEmail(ctx, req)
	if err == nil {
		t.Fatal("expected error for simple email without destination")
	}

	var identityErr *IdentityError
	if !errors.As(err, &identityErr) {
		t.Fatalf("expected IdentityError, got %T", err)
	}

	if identityErr.Message != "Destination is required" {
		t.Errorf("expected 'Destination is required', got '%s'", identityErr.Message)
	}
}
