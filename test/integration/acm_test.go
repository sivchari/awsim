//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/aws/aws-sdk-go-v2/service/acm/types"
)

func newACMClient(t *testing.T) *acm.Client {
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

	return acm.NewFromConfig(cfg, func(o *acm.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestACM_RequestAndDescribeCertificate(t *testing.T) {
	client := newACMClient(t)
	ctx := t.Context()

	// Request certificate.
	requestOutput, err := client.RequestCertificate(ctx, &acm.RequestCertificateInput{
		DomainName: aws.String("example.com"),
		SubjectAlternativeNames: []string{
			"www.example.com",
			"api.example.com",
		},
		ValidationMethod: types.ValidationMethodDns,
	})
	if err != nil {
		t.Fatalf("failed to request certificate: %v", err)
	}

	if requestOutput.CertificateArn == nil {
		t.Fatal("certificate ARN is nil")
	}

	arn := *requestOutput.CertificateArn
	t.Logf("Requested certificate: %s", arn)

	// Describe certificate.
	describeOutput, err := client.DescribeCertificate(ctx, &acm.DescribeCertificateInput{
		CertificateArn: aws.String(arn),
	})
	if err != nil {
		t.Fatalf("failed to describe certificate: %v", err)
	}

	if describeOutput.Certificate == nil {
		t.Fatal("certificate is nil")
	}

	cert := describeOutput.Certificate

	if *cert.CertificateArn != arn {
		t.Errorf("ARN mismatch: got %s, want %s", *cert.CertificateArn, arn)
	}

	if *cert.DomainName != "example.com" {
		t.Errorf("domain mismatch: got %s, want example.com", *cert.DomainName)
	}

	if cert.Status != types.CertificateStatusPendingValidation {
		t.Errorf("status mismatch: got %s, want PENDING_VALIDATION", cert.Status)
	}

	if cert.Type != types.CertificateTypeAmazonIssued {
		t.Errorf("type mismatch: got %s, want AMAZON_ISSUED", cert.Type)
	}

	// Check domain validation options.
	if len(cert.DomainValidationOptions) == 0 {
		t.Fatal("no domain validation options")
	}

	// Should have validation options for main domain and SANs.
	expectedDomains := map[string]bool{
		"example.com":     false,
		"www.example.com": false,
		"api.example.com": false,
	}

	for _, dv := range cert.DomainValidationOptions {
		if dv.DomainName != nil {
			expectedDomains[*dv.DomainName] = true
		}
	}

	for domain, found := range expectedDomains {
		if !found {
			t.Errorf("missing domain validation for: %s", domain)
		}
	}

	t.Logf("Described certificate: %s", arn)
}

func TestACM_ListCertificates(t *testing.T) {
	client := newACMClient(t)
	ctx := t.Context()

	// Request a certificate first.
	requestOutput, err := client.RequestCertificate(ctx, &acm.RequestCertificateInput{
		DomainName: aws.String("list-test.example.com"),
	})
	if err != nil {
		t.Fatalf("failed to request certificate: %v", err)
	}

	arn := *requestOutput.CertificateArn

	// List certificates.
	listOutput, err := client.ListCertificates(ctx, &acm.ListCertificatesInput{
		MaxItems: aws.Int32(10),
	})
	if err != nil {
		t.Fatalf("failed to list certificates: %v", err)
	}

	if len(listOutput.CertificateSummaryList) == 0 {
		t.Fatal("no certificates returned")
	}

	// Find our certificate.
	found := false

	for _, cert := range listOutput.CertificateSummaryList {
		if cert.CertificateArn != nil && *cert.CertificateArn == arn {
			found = true

			if *cert.DomainName != "list-test.example.com" {
				t.Errorf("domain mismatch: got %s, want list-test.example.com", *cert.DomainName)
			}

			break
		}
	}

	if !found {
		t.Errorf("certificate %s not found in list", arn)
	}

	t.Logf("Listed %d certificates", len(listOutput.CertificateSummaryList))
}

func TestACM_ListCertificatesWithStatusFilter(t *testing.T) {
	client := newACMClient(t)
	ctx := t.Context()

	// Request a certificate.
	_, err := client.RequestCertificate(ctx, &acm.RequestCertificateInput{
		DomainName: aws.String("filter-test.example.com"),
	})
	if err != nil {
		t.Fatalf("failed to request certificate: %v", err)
	}

	// List with PENDING_VALIDATION filter.
	listOutput, err := client.ListCertificates(ctx, &acm.ListCertificatesInput{
		CertificateStatuses: []types.CertificateStatus{
			types.CertificateStatusPendingValidation,
		},
	})
	if err != nil {
		t.Fatalf("failed to list certificates: %v", err)
	}

	// All returned certificates should be PENDING_VALIDATION.
	for _, cert := range listOutput.CertificateSummaryList {
		if cert.Status != types.CertificateStatusPendingValidation {
			t.Errorf("unexpected status: got %s, want PENDING_VALIDATION", cert.Status)
		}
	}

	t.Logf("Listed %d certificates with PENDING_VALIDATION status", len(listOutput.CertificateSummaryList))
}

func TestACM_DeleteCertificate(t *testing.T) {
	client := newACMClient(t)
	ctx := t.Context()

	// Request a certificate.
	requestOutput, err := client.RequestCertificate(ctx, &acm.RequestCertificateInput{
		DomainName: aws.String("delete-test.example.com"),
	})
	if err != nil {
		t.Fatalf("failed to request certificate: %v", err)
	}

	arn := *requestOutput.CertificateArn

	// Delete the certificate.
	_, err = client.DeleteCertificate(ctx, &acm.DeleteCertificateInput{
		CertificateArn: aws.String(arn),
	})
	if err != nil {
		t.Fatalf("failed to delete certificate: %v", err)
	}

	t.Logf("Deleted certificate: %s", arn)

	// Verify it's deleted by trying to describe it.
	_, err = client.DescribeCertificate(ctx, &acm.DescribeCertificateInput{
		CertificateArn: aws.String(arn),
	})
	if err == nil {
		t.Fatal("expected error when describing deleted certificate")
	}

	t.Logf("Verified certificate is deleted")
}

func TestACM_ImportCertificate(t *testing.T) {
	client := newACMClient(t)
	ctx := t.Context()

	// Import a certificate (using dummy data - the emulator doesn't validate).
	importOutput, err := client.ImportCertificate(ctx, &acm.ImportCertificateInput{
		Certificate: []byte("-----BEGIN CERTIFICATE-----\nMIICert...\n-----END CERTIFICATE-----"),
		PrivateKey:  []byte("-----BEGIN PRIVATE KEY-----\nMIIPrivateKey...\n-----END PRIVATE KEY-----"),
	})
	if err != nil {
		t.Fatalf("failed to import certificate: %v", err)
	}

	if importOutput.CertificateArn == nil {
		t.Fatal("certificate ARN is nil")
	}

	arn := *importOutput.CertificateArn
	t.Logf("Imported certificate: %s", arn)

	// Describe the imported certificate.
	describeOutput, err := client.DescribeCertificate(ctx, &acm.DescribeCertificateInput{
		CertificateArn: aws.String(arn),
	})
	if err != nil {
		t.Fatalf("failed to describe certificate: %v", err)
	}

	if describeOutput.Certificate.Type != types.CertificateTypeImported {
		t.Errorf("type mismatch: got %s, want IMPORTED", describeOutput.Certificate.Type)
	}

	if describeOutput.Certificate.Status != types.CertificateStatusIssued {
		t.Errorf("status mismatch: got %s, want ISSUED", describeOutput.Certificate.Status)
	}

	t.Logf("Verified imported certificate")
}

func TestACM_RequestCertificateWithOptions(t *testing.T) {
	client := newACMClient(t)
	ctx := t.Context()

	// Request certificate with key algorithm option.
	requestOutput, err := client.RequestCertificate(ctx, &acm.RequestCertificateInput{
		DomainName:   aws.String("options-test.example.com"),
		KeyAlgorithm: types.KeyAlgorithmEcPrime256v1,
		Options: &types.CertificateOptions{
			CertificateTransparencyLoggingPreference: types.CertificateTransparencyLoggingPreferenceEnabled,
		},
	})
	if err != nil {
		t.Fatalf("failed to request certificate: %v", err)
	}

	arn := *requestOutput.CertificateArn

	// Describe to verify options.
	describeOutput, err := client.DescribeCertificate(ctx, &acm.DescribeCertificateInput{
		CertificateArn: aws.String(arn),
	})
	if err != nil {
		t.Fatalf("failed to describe certificate: %v", err)
	}

	if describeOutput.Certificate.KeyAlgorithm != types.KeyAlgorithmEcPrime256v1 {
		t.Errorf("key algorithm mismatch: got %s, want EC_prime256v1", describeOutput.Certificate.KeyAlgorithm)
	}

	t.Logf("Requested certificate with options: %s", arn)
}

func TestACM_DescribeNonExistentCertificate(t *testing.T) {
	client := newACMClient(t)
	ctx := t.Context()

	// Try to describe a non-existent certificate.
	_, err := client.DescribeCertificate(ctx, &acm.DescribeCertificateInput{
		CertificateArn: aws.String("arn:aws:acm:us-east-1:000000000000:certificate/non-existent"),
	})
	if err == nil {
		t.Fatal("expected error when describing non-existent certificate")
	}

	t.Logf("Got expected error: %v", err)
}

func TestACM_DeleteNonExistentCertificate(t *testing.T) {
	client := newACMClient(t)
	ctx := t.Context()

	// Try to delete a non-existent certificate.
	_, err := client.DeleteCertificate(ctx, &acm.DeleteCertificateInput{
		CertificateArn: aws.String("arn:aws:acm:us-east-1:000000000000:certificate/non-existent"),
	})
	if err == nil {
		t.Fatal("expected error when deleting non-existent certificate")
	}

	t.Logf("Got expected error: %v", err)
}

func TestACM_GetCertificate(t *testing.T) {
	client := newACMClient(t)
	ctx := t.Context()

	// Import a certificate first (imported certificates are in ISSUED state).
	importOutput, err := client.ImportCertificate(ctx, &acm.ImportCertificateInput{
		Certificate:      []byte("-----BEGIN CERTIFICATE-----\nMIICert...\n-----END CERTIFICATE-----"),
		PrivateKey:       []byte("-----BEGIN PRIVATE KEY-----\nMIIPrivateKey...\n-----END PRIVATE KEY-----"),
		CertificateChain: []byte("-----BEGIN CERTIFICATE-----\nMIIChain...\n-----END CERTIFICATE-----"),
	})
	if err != nil {
		t.Fatalf("failed to import certificate: %v", err)
	}

	arn := *importOutput.CertificateArn

	// Get the certificate.
	getOutput, err := client.GetCertificate(ctx, &acm.GetCertificateInput{
		CertificateArn: aws.String(arn),
	})
	if err != nil {
		t.Fatalf("failed to get certificate: %v", err)
	}

	if getOutput.Certificate == nil {
		t.Fatal("certificate body is nil")
	}

	t.Logf("Got certificate: %s", arn)
}

func TestACM_GetCertificatePendingValidation(t *testing.T) {
	client := newACMClient(t)
	ctx := t.Context()

	// Request a certificate (will be in PENDING_VALIDATION state).
	requestOutput, err := client.RequestCertificate(ctx, &acm.RequestCertificateInput{
		DomainName: aws.String("pending.example.com"),
	})
	if err != nil {
		t.Fatalf("failed to request certificate: %v", err)
	}

	arn := *requestOutput.CertificateArn

	// Try to get the certificate (should fail because it's not issued).
	_, err = client.GetCertificate(ctx, &acm.GetCertificateInput{
		CertificateArn: aws.String(arn),
	})
	if err == nil {
		t.Fatal("expected error when getting pending certificate")
	}

	t.Logf("Got expected error for pending certificate: %v", err)
}
