package acm

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Error codes.
const (
	errNotFound         = "ResourceNotFoundException"
	errInvalidParameter = "ValidationException"
)

// Storage defines the interface for ACM storage operations.
type Storage interface {
	RequestCertificate(ctx context.Context, req *RequestCertificateInput) (*Certificate, error)
	DescribeCertificate(ctx context.Context, arn string) (*Certificate, error)
	ListCertificates(ctx context.Context, statuses []string, maxItems int32, nextToken string) ([]*Certificate, string, error)
	DeleteCertificate(ctx context.Context, arn string) error
	GetCertificate(ctx context.Context, arn string) (*Certificate, error)
	ImportCertificate(ctx context.Context, req *ImportCertificateInput) (*Certificate, error)
}

// MemoryStorage implements Storage with in-memory data structures.
type MemoryStorage struct {
	mu           sync.RWMutex
	certificates map[string]*Certificate
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		certificates: make(map[string]*Certificate),
	}
}

// RequestCertificate requests a new certificate.
func (s *MemoryStorage) RequestCertificate(_ context.Context, req *RequestCertificateInput) (*Certificate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if req.DomainName == "" {
		return nil, &Error{
			Code:    errInvalidParameter,
			Message: "DomainName is required",
		}
	}

	// Generate certificate ARN.
	certID := uuid.New().String()
	arn := fmt.Sprintf("arn:aws:acm:us-east-1:000000000000:certificate/%s", certID)

	// Generate serial number.
	serialBytes := make([]byte, 16)
	if _, err := rand.Read(serialBytes); err != nil {
		return nil, fmt.Errorf("failed to generate serial: %w", err)
	}

	serial := hex.EncodeToString(serialBytes)

	// Determine key algorithm.
	keyAlgorithm := req.KeyAlgorithm
	if keyAlgorithm == "" {
		keyAlgorithm = "RSA_2048"
	}

	// Create domain validation options.
	domainValidations := make([]DomainValidation, 0)
	domains := []string{req.DomainName}
	domains = append(domains, req.SubjectAlternativeNames...)

	validationMethod := req.ValidationMethod
	if validationMethod == "" {
		validationMethod = "DNS"
	}

	for _, domain := range domains {
		dv := DomainValidation{
			DomainName:       domain,
			ValidationDomain: domain,
			ValidationStatus: "PENDING_VALIDATION",
			ValidationMethod: validationMethod,
		}

		if validationMethod == "DNS" {
			dv.ResourceRecord = &ResourceRecord{
				Name:  fmt.Sprintf("_acme-challenge.%s", domain),
				Type:  "CNAME",
				Value: fmt.Sprintf("_%s.acm-validations.aws.", certID[:8]),
			}
		}

		domainValidations = append(domainValidations, dv)
	}

	now := time.Now()
	cert := &Certificate{
		CertificateArn:          arn,
		DomainName:              req.DomainName,
		SubjectAlternativeNames: req.SubjectAlternativeNames,
		Status:                  "PENDING_VALIDATION",
		Type:                    "AMAZON_ISSUED",
		KeyAlgorithm:            keyAlgorithm,
		Serial:                  serial,
		Subject:                 fmt.Sprintf("CN=%s", req.DomainName),
		CreatedAt:               now,
		DomainValidationOptions: domainValidations,
		RenewalEligibility:      "INELIGIBLE",
		Options:                 req.Options,
		Tags:                    req.Tags,
	}

	s.certificates[arn] = cert

	return cert, nil
}

// DescribeCertificate retrieves certificate details.
func (s *MemoryStorage) DescribeCertificate(_ context.Context, arn string) (*Certificate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cert, exists := s.certificates[arn]
	if !exists {
		return nil, &Error{
			Code:    errNotFound,
			Message: fmt.Sprintf("Certificate with arn %s not found", arn),
		}
	}

	return cert, nil
}

// ListCertificates lists certificates with optional filtering.
func (s *MemoryStorage) ListCertificates(_ context.Context, statuses []string, maxItems int32, _ string) ([]*Certificate, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxItems <= 0 {
		maxItems = 100
	}

	certs := make([]*Certificate, 0, len(s.certificates))

	for _, cert := range s.certificates {
		// Filter by status if specified.
		if len(statuses) > 0 && !slices.Contains(statuses, cert.Status) {
			continue
		}

		certs = append(certs, cert)
	}

	// Apply pagination.
	if len(certs) > int(maxItems) {
		certs = certs[:maxItems]
	}

	return certs, "", nil
}

// DeleteCertificate deletes a certificate.
func (s *MemoryStorage) DeleteCertificate(_ context.Context, arn string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.certificates[arn]; !exists {
		return &Error{
			Code:    errNotFound,
			Message: fmt.Sprintf("Certificate with arn %s not found", arn),
		}
	}

	delete(s.certificates, arn)

	return nil
}

// GetCertificate retrieves certificate and chain.
func (s *MemoryStorage) GetCertificate(_ context.Context, arn string) (*Certificate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cert, exists := s.certificates[arn]
	if !exists {
		return nil, &Error{
			Code:    errNotFound,
			Message: fmt.Sprintf("Certificate with arn %s not found", arn),
		}
	}

	// Only issued or imported certificates can be retrieved.
	if cert.Status != "ISSUED" && cert.Type != "IMPORTED" {
		return nil, &Error{
			Code:    errNotFound,
			Message: "Certificate is not issued yet",
		}
	}

	return cert, nil
}

// ImportCertificate imports a certificate.
func (s *MemoryStorage) ImportCertificate(_ context.Context, req *ImportCertificateInput) (*Certificate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(req.Certificate) == 0 {
		return nil, &Error{
			Code:    errInvalidParameter,
			Message: "Certificate is required",
		}
	}

	if len(req.PrivateKey) == 0 {
		return nil, &Error{
			Code:    errInvalidParameter,
			Message: "PrivateKey is required",
		}
	}

	arn := req.CertificateArn
	if arn == "" {
		// Generate new ARN.
		certID := uuid.New().String()
		arn = fmt.Sprintf("arn:aws:acm:us-east-1:000000000000:certificate/%s", certID)
	}

	// Generate serial number.
	serialBytes := make([]byte, 16)
	if _, err := rand.Read(serialBytes); err != nil {
		return nil, fmt.Errorf("failed to generate serial: %w", err)
	}

	serial := hex.EncodeToString(serialBytes)
	now := time.Now()

	// Parse certificate to extract domain (simplified - just use a placeholder).
	domainName := "imported.example.com"

	cert := &Certificate{
		CertificateArn:     arn,
		DomainName:         domainName,
		Status:             "ISSUED",
		Type:               "IMPORTED",
		KeyAlgorithm:       "RSA_2048",
		Serial:             serial,
		Subject:            fmt.Sprintf("CN=%s", domainName),
		CreatedAt:          now,
		ImportedAt:         &now,
		IssuedAt:           &now,
		CertificateBody:    string(req.Certificate),
		CertificateChain:   string(req.CertificateChain),
		PrivateKey:         string(req.PrivateKey),
		RenewalEligibility: "INELIGIBLE",
		InUseBy:            []string{},
		Tags:               req.Tags,
	}

	s.certificates[arn] = cert

	return cert, nil
}
