//go:build integration

package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/sivchari/golden"
)

func newSESv2Client(t *testing.T) *sesv2.Client {
	t.Helper()

	cfg, err := config.LoadDefaultConfig(t.Context(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test", "test", "",
		)),
	)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	return sesv2.NewFromConfig(cfg, func(o *sesv2.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566/ses")
	})
}

func TestSESv2_CreateAndGetEmailIdentity(t *testing.T) {
	client := newSESv2Client(t)
	ctx := t.Context()

	emailIdentity := "test@example.com"

	// Create email identity.
	createOutput, err := client.CreateEmailIdentity(ctx, &sesv2.CreateEmailIdentityInput{
		EmailIdentity: aws.String(emailIdentity),
	})
	if err != nil {
		t.Fatal(err)
	}
	golden.New(t, golden.WithIgnoreFields("Tokens", "ResultMetadata")).Assert(t.Name()+"_create", createOutput)

	// Get email identity.
	getOutput, err := client.GetEmailIdentity(ctx, &sesv2.GetEmailIdentityInput{
		EmailIdentity: aws.String(emailIdentity),
	})
	if err != nil {
		t.Fatal(err)
	}
	golden.New(t, golden.WithIgnoreFields("Tokens", "ResultMetadata")).Assert(t.Name()+"_get", getOutput)
}

func TestSESv2_CreateDomainIdentity(t *testing.T) {
	client := newSESv2Client(t)
	ctx := t.Context()

	domainIdentity := "example.com"

	// Create domain identity.
	createOutput, err := client.CreateEmailIdentity(ctx, &sesv2.CreateEmailIdentityInput{
		EmailIdentity: aws.String(domainIdentity),
	})
	if err != nil {
		t.Fatal(err)
	}
	golden.New(t, golden.WithIgnoreFields("Tokens", "ResultMetadata")).Assert(t.Name(), createOutput)
}

func TestSESv2_ListEmailIdentities(t *testing.T) {
	client := newSESv2Client(t)
	ctx := t.Context()

	// Create an email identity.
	emailIdentity := "list-test@example.com"
	_, err := client.CreateEmailIdentity(ctx, &sesv2.CreateEmailIdentityInput{
		EmailIdentity: aws.String(emailIdentity),
	})
	if err != nil {
		t.Fatal(err)
	}

	// List email identities.
	listOutput, err := client.ListEmailIdentities(ctx, &sesv2.ListEmailIdentitiesInput{})
	if err != nil {
		t.Fatal(err)
	}

	found := false

	for _, identity := range listOutput.EmailIdentities {
		if identity.IdentityName != nil && *identity.IdentityName == emailIdentity {
			found = true

			break
		}
	}

	if !found {
		t.Error("created email identity not found in list")
	}
}

func TestSESv2_DeleteEmailIdentity(t *testing.T) {
	client := newSESv2Client(t)
	ctx := t.Context()

	emailIdentity := "delete-test@example.com"

	// Create email identity.
	_, err := client.CreateEmailIdentity(ctx, &sesv2.CreateEmailIdentityInput{
		EmailIdentity: aws.String(emailIdentity),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Delete email identity.
	_, err = client.DeleteEmailIdentity(ctx, &sesv2.DeleteEmailIdentityInput{
		EmailIdentity: aws.String(emailIdentity),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Verify deletion.
	_, err = client.GetEmailIdentity(ctx, &sesv2.GetEmailIdentityInput{
		EmailIdentity: aws.String(emailIdentity),
	})
	if err == nil {
		t.Error("expected error for deleted email identity")
	}
}

func TestSESv2_CreateAndGetConfigurationSet(t *testing.T) {
	client := newSESv2Client(t)
	ctx := t.Context()

	configSetName := "test-config-set"

	// Create configuration set.
	_, err := client.CreateConfigurationSet(ctx, &sesv2.CreateConfigurationSetInput{
		ConfigurationSetName: aws.String(configSetName),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Get configuration set.
	getOutput, err := client.GetConfigurationSet(ctx, &sesv2.GetConfigurationSetInput{
		ConfigurationSetName: aws.String(configSetName),
	})
	if err != nil {
		t.Fatal(err)
	}
	golden.New(t, golden.WithIgnoreFields("ResultMetadata")).Assert(t.Name()+"_get", getOutput)
}

func TestSESv2_ListConfigurationSets(t *testing.T) {
	client := newSESv2Client(t)
	ctx := t.Context()

	configSetName := "test-list-config-set"

	// Create configuration set.
	_, err := client.CreateConfigurationSet(ctx, &sesv2.CreateConfigurationSetInput{
		ConfigurationSetName: aws.String(configSetName),
	})
	if err != nil {
		t.Fatal(err)
	}

	// List configuration sets.
	listOutput, err := client.ListConfigurationSets(ctx, &sesv2.ListConfigurationSetsInput{})
	if err != nil {
		t.Fatal(err)
	}

	found := false

	for _, name := range listOutput.ConfigurationSets {
		if name == configSetName {
			found = true

			break
		}
	}

	if !found {
		t.Error("created configuration set not found in list")
	}
}

func TestSESv2_DeleteConfigurationSet(t *testing.T) {
	client := newSESv2Client(t)
	ctx := t.Context()

	configSetName := "test-delete-config-set"

	// Create configuration set.
	_, err := client.CreateConfigurationSet(ctx, &sesv2.CreateConfigurationSetInput{
		ConfigurationSetName: aws.String(configSetName),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Delete configuration set.
	_, err = client.DeleteConfigurationSet(ctx, &sesv2.DeleteConfigurationSetInput{
		ConfigurationSetName: aws.String(configSetName),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Verify deletion.
	_, err = client.GetConfigurationSet(ctx, &sesv2.GetConfigurationSetInput{
		ConfigurationSetName: aws.String(configSetName),
	})
	if err == nil {
		t.Error("expected error for deleted configuration set")
	}
}

func TestSESv2_SendEmail(t *testing.T) {
	client := newSESv2Client(t)
	ctx := t.Context()

	// Create email identity first.
	emailIdentity := "sender@example.com"
	_, err := client.CreateEmailIdentity(ctx, &sesv2.CreateEmailIdentityInput{
		EmailIdentity: aws.String(emailIdentity),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Send email.
	sendOutput, err := client.SendEmail(ctx, &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(emailIdentity),
		Destination: &types.Destination{
			ToAddresses: []string{"recipient@example.com"},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Data: aws.String("Test Subject"),
				},
				Body: &types.Body{
					Text: &types.Content{
						Data: aws.String("Test body content"),
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	golden.New(t, golden.WithIgnoreFields("MessageId", "ResultMetadata")).Assert(t.Name(), sendOutput)
}

func TestSESv2_EmailIdentityNotFound(t *testing.T) {
	client := newSESv2Client(t)
	ctx := t.Context()

	// Try to get non-existent email identity.
	_, err := client.GetEmailIdentity(ctx, &sesv2.GetEmailIdentityInput{
		EmailIdentity: aws.String("nonexistent@example.com"),
	})
	if err == nil {
		t.Fatal("expected error for non-existent email identity")
	}
}

func TestSESv2_GetSentEmails(t *testing.T) {
	client := newSESv2Client(t)
	ctx := t.Context()

	// Create email identity.
	fromEmail := "test-sent@example.com"
	_, err := client.CreateEmailIdentity(ctx, &sesv2.CreateEmailIdentityInput{
		EmailIdentity: aws.String(fromEmail),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Send email.
	_, err = client.SendEmail(ctx, &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(fromEmail),
		Destination: &types.Destination{
			ToAddresses: []string{"recipient@example.com"},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Data: aws.String("Test Subject"),
				},
				Body: &types.Body{
					Text: &types.Content{
						Data: aws.String("Test body"),
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Get sent emails via kumo-specific endpoint.
	resp, err := http.Get("http://localhost:4566/kumo/ses/v2/sent-emails")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusOK, resp.StatusCode, body)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	sentEmails, ok := result["SentEmails"]
	if !ok {
		t.Fatal("SentEmails field not found in response")
	}

	emails, ok := sentEmails.([]interface{})
	if !ok {
		t.Fatal("SentEmails is not an array")
	}

	if len(emails) == 0 {
		t.Fatal("no sent emails found")
	}

	// Verify first email
	firstEmail, ok := emails[0].(map[string]interface{})
	if !ok {
		t.Fatal("email is not an object")
	}

	if fromAddr, ok := firstEmail["FromEmailAddress"]; !ok || fromAddr != fromEmail {
		t.Errorf("FromEmailAddress mismatch: got %v, want %s", fromAddr, fromEmail)
	}

	if subject, ok := firstEmail["Subject"]; !ok || subject != "Test Subject" {
		t.Errorf("Subject mismatch: got %v, want 'Test Subject'", subject)
	}
}
