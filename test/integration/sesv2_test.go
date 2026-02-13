//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
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
		t.Fatalf("failed to create email identity: %v", err)
	}

	if createOutput.IdentityType != types.IdentityTypeEmailAddress {
		t.Errorf("identity type mismatch: got %s, want EMAIL_ADDRESS", createOutput.IdentityType)
	}

	t.Logf("Created email identity: %s (type: %s)", emailIdentity, createOutput.IdentityType)

	// Get email identity.
	getOutput, err := client.GetEmailIdentity(ctx, &sesv2.GetEmailIdentityInput{
		EmailIdentity: aws.String(emailIdentity),
	})
	if err != nil {
		t.Fatalf("failed to get email identity: %v", err)
	}

	if getOutput.IdentityType != types.IdentityTypeEmailAddress {
		t.Errorf("identity type mismatch: got %s, want EMAIL_ADDRESS", getOutput.IdentityType)
	}

	t.Logf("Got email identity: type=%s, verified=%v", getOutput.IdentityType, getOutput.VerifiedForSendingStatus)
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
		t.Fatalf("failed to create domain identity: %v", err)
	}

	if createOutput.IdentityType != types.IdentityTypeDomain {
		t.Errorf("identity type mismatch: got %s, want DOMAIN", createOutput.IdentityType)
	}

	t.Logf("Created domain identity: %s (type: %s)", domainIdentity, createOutput.IdentityType)
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
		t.Fatalf("failed to create email identity: %v", err)
	}

	// List email identities.
	listOutput, err := client.ListEmailIdentities(ctx, &sesv2.ListEmailIdentitiesInput{})
	if err != nil {
		t.Fatalf("failed to list email identities: %v", err)
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

	t.Logf("Listed %d email identities", len(listOutput.EmailIdentities))
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
		t.Fatalf("failed to create email identity: %v", err)
	}

	// Delete email identity.
	_, err = client.DeleteEmailIdentity(ctx, &sesv2.DeleteEmailIdentityInput{
		EmailIdentity: aws.String(emailIdentity),
	})
	if err != nil {
		t.Fatalf("failed to delete email identity: %v", err)
	}

	t.Log("Deleted email identity successfully")

	// Verify deletion.
	_, err = client.GetEmailIdentity(ctx, &sesv2.GetEmailIdentityInput{
		EmailIdentity: aws.String(emailIdentity),
	})
	if err == nil {
		t.Error("expected error for deleted email identity")
	}

	t.Log("Verified email identity deletion")
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
		t.Fatalf("failed to create configuration set: %v", err)
	}

	t.Logf("Created configuration set: %s", configSetName)

	// Get configuration set.
	getOutput, err := client.GetConfigurationSet(ctx, &sesv2.GetConfigurationSetInput{
		ConfigurationSetName: aws.String(configSetName),
	})
	if err != nil {
		t.Fatalf("failed to get configuration set: %v", err)
	}

	if *getOutput.ConfigurationSetName != configSetName {
		t.Errorf("configuration set name mismatch: got %s, want %s",
			*getOutput.ConfigurationSetName, configSetName)
	}

	t.Logf("Got configuration set: %s", *getOutput.ConfigurationSetName)
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
		t.Fatalf("failed to create configuration set: %v", err)
	}

	// List configuration sets.
	listOutput, err := client.ListConfigurationSets(ctx, &sesv2.ListConfigurationSetsInput{})
	if err != nil {
		t.Fatalf("failed to list configuration sets: %v", err)
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

	t.Logf("Listed %d configuration sets", len(listOutput.ConfigurationSets))
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
		t.Fatalf("failed to create configuration set: %v", err)
	}

	// Delete configuration set.
	_, err = client.DeleteConfigurationSet(ctx, &sesv2.DeleteConfigurationSetInput{
		ConfigurationSetName: aws.String(configSetName),
	})
	if err != nil {
		t.Fatalf("failed to delete configuration set: %v", err)
	}

	t.Log("Deleted configuration set successfully")

	// Verify deletion.
	_, err = client.GetConfigurationSet(ctx, &sesv2.GetConfigurationSetInput{
		ConfigurationSetName: aws.String(configSetName),
	})
	if err == nil {
		t.Error("expected error for deleted configuration set")
	}

	t.Log("Verified configuration set deletion")
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
		t.Fatalf("failed to create email identity: %v", err)
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
		t.Fatalf("failed to send email: %v", err)
	}

	if sendOutput.MessageId == nil || *sendOutput.MessageId == "" {
		t.Error("expected non-empty message ID")
	}

	t.Logf("Sent email with message ID: %s", *sendOutput.MessageId)
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

	t.Log("Got expected error for non-existent email identity")
}
