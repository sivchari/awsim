package kms

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	defaultRegion    = "us-east-1"
	defaultAccountID = "000000000000"
)

// Error codes.
const (
	errNotFound          = "NotFoundException"
	errInvalidKeyState   = "KMSInvalidStateException"
	errAlreadyExists     = "AlreadyExistsException"
	errInvalidAlias      = "InvalidAliasNameException"
	errDependencyTimeout = "DependencyTimeoutException"
	errInvalidCiphertext = "InvalidCiphertextException"
	errIncorrectKey      = "IncorrectKeyException"
	errDisabled          = "DisabledException"
	errInvalidKeyUsage   = "InvalidKeyUsageException"
)

// Storage defines the KMS storage interface.
type Storage interface {
	// Key operations.
	CreateKey(ctx context.Context, req *CreateKeyRequest) (*Key, error)
	GetKey(ctx context.Context, keyID string) (*Key, error)
	ListKeys(ctx context.Context, limit int32, marker string) ([]*Key, string, error)
	EnableKey(ctx context.Context, keyID string) error
	DisableKey(ctx context.Context, keyID string) error
	ScheduleKeyDeletion(ctx context.Context, keyID string, pendingWindowInDays int32) (*Key, error)

	// Cryptographic operations.
	Encrypt(ctx context.Context, keyID string, plaintext []byte, encryptionContext map[string]string) ([]byte, error)
	Decrypt(ctx context.Context, ciphertextBlob []byte, encryptionContext map[string]string, keyID string) ([]byte, string, error)
	GenerateDataKey(ctx context.Context, keyID string, keySpec string, numberOfBytes int32, encryptionContext map[string]string) ([]byte, []byte, error)

	// Alias operations.
	CreateAlias(ctx context.Context, aliasName, targetKeyID string) error
	DeleteAlias(ctx context.Context, aliasName string) error
	ListAliases(ctx context.Context, keyID string, limit int32, marker string) ([]*Alias, string, error)
	GetAlias(ctx context.Context, aliasName string) (*Alias, error)
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu      sync.RWMutex
	keys    map[string]*Key   // keyID -> Key
	aliases map[string]*Alias // aliasName -> Alias
	region  string
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		keys:    make(map[string]*Key),
		aliases: make(map[string]*Alias),
		region:  defaultRegion,
	}
}

// CreateKey creates a new KMS key.
func (s *MemoryStorage) CreateKey(_ context.Context, req *CreateKeyRequest) (*Key, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	keyID := uuid.New().String()
	arn := fmt.Sprintf("arn:aws:kms:%s:%s:key/%s", s.region, defaultAccountID, keyID)

	// Generate random key material (256-bit for AES-256).
	keyMaterial := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, keyMaterial); err != nil {
		return nil, &KMSError{Code: errDependencyTimeout, Message: "Failed to generate key material"}
	}

	keyUsage := KeyUsageEncryptDecrypt
	if req.KeyUsage != "" {
		keyUsage = KeyUsage(req.KeyUsage)
	}

	keySpec := KeySpecSymmetricDefault
	if req.KeySpec != "" {
		keySpec = KeySpec(req.KeySpec)
	}

	origin := "AWS_KMS"
	if req.Origin != "" {
		origin = req.Origin
	}

	tags := make(map[string]string)
	for _, tag := range req.Tags {
		tags[tag.TagKey] = tag.TagValue
	}

	key := &Key{
		KeyID:        keyID,
		Arn:          arn,
		Description:  req.Description,
		KeyState:     KeyStateEnabled,
		KeyUsage:     keyUsage,
		KeySpec:      keySpec,
		KeyManager:   KeyManagerCustomer,
		CreationDate: time.Now(),
		Enabled:      true,
		Origin:       origin,
		MultiRegion:  req.MultiRegion,
		Tags:         tags,
		KeyMaterial:  keyMaterial,
	}

	s.keys[keyID] = key

	return key, nil
}

// GetKey retrieves a key by ID, ARN, or alias.
func (s *MemoryStorage) GetKey(_ context.Context, keyID string) (*Key, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.getKeyLocked(keyID)
}

// getKeyLocked retrieves a key without locking (caller must hold lock).
func (s *MemoryStorage) getKeyLocked(keyID string) (*Key, error) {
	// Check if it's an alias.
	if len(keyID) > 6 && keyID[:6] == "alias/" {
		alias, ok := s.aliases[keyID]
		if !ok {
			return nil, &KMSError{Code: errNotFound, Message: "Alias " + keyID + " is not found."}
		}

		keyID = alias.TargetKeyID
	}

	// Check if it's an ARN.
	if len(keyID) > 8 && keyID[:8] == "arn:aws:" {
		// Extract key ID from ARN.
		for _, key := range s.keys {
			if key.Arn == keyID {
				return key, nil
			}
		}

		return nil, &KMSError{Code: errNotFound, Message: "Key " + keyID + " is not found."}
	}

	// Look up by key ID.
	key, ok := s.keys[keyID]
	if !ok {
		return nil, &KMSError{Code: errNotFound, Message: "Key " + keyID + " is not found."}
	}

	return key, nil
}

// ListKeys lists all keys.
func (s *MemoryStorage) ListKeys(_ context.Context, limit int32, _ string) ([]*Key, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	keys := make([]*Key, 0, len(s.keys))
	for _, key := range s.keys {
		keys = append(keys, key)
		if int32(len(keys)) >= limit {
			break
		}
	}

	return keys, "", nil
}

// EnableKey enables a key.
func (s *MemoryStorage) EnableKey(_ context.Context, keyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key, err := s.getKeyLocked(keyID)
	if err != nil {
		return err
	}

	if key.KeyState == KeyStatePendingDeletion {
		return &KMSError{
			Code:    errInvalidKeyState,
			Message: "Key " + keyID + " is pending deletion.",
		}
	}

	key.KeyState = KeyStateEnabled
	key.Enabled = true

	return nil
}

// DisableKey disables a key.
func (s *MemoryStorage) DisableKey(_ context.Context, keyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key, err := s.getKeyLocked(keyID)
	if err != nil {
		return err
	}

	if key.KeyState == KeyStatePendingDeletion {
		return &KMSError{
			Code:    errInvalidKeyState,
			Message: "Key " + keyID + " is pending deletion.",
		}
	}

	key.KeyState = KeyStateDisabled
	key.Enabled = false

	return nil
}

// ScheduleKeyDeletion schedules a key for deletion.
func (s *MemoryStorage) ScheduleKeyDeletion(_ context.Context, keyID string, pendingWindowInDays int32) (*Key, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key, err := s.getKeyLocked(keyID)
	if err != nil {
		return nil, err
	}

	if key.KeyState == KeyStatePendingDeletion {
		return nil, &KMSError{
			Code:    errInvalidKeyState,
			Message: "Key " + keyID + " is pending deletion.",
		}
	}

	if pendingWindowInDays < 7 || pendingWindowInDays > 30 {
		pendingWindowInDays = 30
	}

	deletionDate := time.Now().AddDate(0, 0, int(pendingWindowInDays))
	key.KeyState = KeyStatePendingDeletion
	key.Enabled = false
	key.DeletionDate = &deletionDate
	key.PendingDeletionWindow = pendingWindowInDays

	return key, nil
}

// Encrypt encrypts plaintext using a key.
func (s *MemoryStorage) Encrypt(_ context.Context, keyID string, plaintext []byte, _ map[string]string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key, err := s.getKeyLocked(keyID)
	if err != nil {
		return nil, err
	}

	if key.KeyState != KeyStateEnabled {
		return nil, &KMSError{
			Code:    errDisabled,
			Message: "Key " + keyID + " is disabled.",
		}
	}

	if key.KeyUsage != KeyUsageEncryptDecrypt {
		return nil, &KMSError{
			Code:    errInvalidKeyUsage,
			Message: "Key " + keyID + " is not configured for encryption.",
		}
	}

	// Use AES-GCM for encryption.
	block, err := aes.NewCipher(key.KeyMaterial)
	if err != nil {
		return nil, &KMSError{Code: errDependencyTimeout, Message: "Encryption failed"}
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, &KMSError{Code: errDependencyTimeout, Message: "Encryption failed"}
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, &KMSError{Code: errDependencyTimeout, Message: "Encryption failed"}
	}

	// Prepend key ID (36 bytes UUID) + nonce to ciphertext for decryption lookup.
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	result := make([]byte, 0, 36+len(nonce)+len(ciphertext))
	result = append(result, []byte(key.KeyID)...)
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	return result, nil
}

// Decrypt decrypts ciphertext.
func (s *MemoryStorage) Decrypt(_ context.Context, ciphertextBlob []byte, _ map[string]string, requestKeyID string) ([]byte, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Extract key ID from ciphertext (first 36 bytes).
	if len(ciphertextBlob) < 36 {
		return nil, "", &KMSError{Code: errInvalidCiphertext, Message: "Invalid ciphertext"}
	}

	embeddedKeyID := string(ciphertextBlob[:36])

	// If a key ID was specified, verify it matches.
	if requestKeyID != "" {
		key, err := s.getKeyLocked(requestKeyID)
		if err != nil {
			return nil, "", err
		}

		if key.KeyID != embeddedKeyID {
			return nil, "", &KMSError{Code: errIncorrectKey, Message: "The key ID in the ciphertext does not match the specified key."}
		}
	}

	key, err := s.getKeyLocked(embeddedKeyID)
	if err != nil {
		return nil, "", err
	}

	if key.KeyState != KeyStateEnabled {
		return nil, "", &KMSError{
			Code:    errDisabled,
			Message: "Key " + embeddedKeyID + " is disabled.",
		}
	}

	// Use AES-GCM for decryption.
	block, err := aes.NewCipher(key.KeyMaterial)
	if err != nil {
		return nil, "", &KMSError{Code: errDependencyTimeout, Message: "Decryption failed"}
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, "", &KMSError{Code: errDependencyTimeout, Message: "Decryption failed"}
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertextBlob) < 36+nonceSize {
		return nil, "", &KMSError{Code: errInvalidCiphertext, Message: "Invalid ciphertext"}
	}

	nonce := ciphertextBlob[36 : 36+nonceSize]
	ciphertext := ciphertextBlob[36+nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, "", &KMSError{Code: errInvalidCiphertext, Message: "Invalid ciphertext"}
	}

	return plaintext, key.KeyID, nil
}

// GenerateDataKey generates a data key.
func (s *MemoryStorage) GenerateDataKey(_ context.Context, keyID string, keySpec string, numberOfBytes int32, _ map[string]string) ([]byte, []byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key, err := s.getKeyLocked(keyID)
	if err != nil {
		return nil, nil, err
	}

	if key.KeyState != KeyStateEnabled {
		return nil, nil, &KMSError{
			Code:    errDisabled,
			Message: "Key " + keyID + " is disabled.",
		}
	}

	if key.KeyUsage != KeyUsageEncryptDecrypt {
		return nil, nil, &KMSError{
			Code:    errInvalidKeyUsage,
			Message: "Key " + keyID + " is not configured for encryption.",
		}
	}

	// Determine key size.
	var keySize int32
	switch keySpec {
	case "AES_256":
		keySize = 32
	case "AES_128":
		keySize = 16
	default:
		if numberOfBytes > 0 {
			keySize = numberOfBytes
		} else {
			keySize = 32 // Default to AES-256
		}
	}

	// Generate plaintext data key.
	plaintext := make([]byte, keySize)
	if _, err := io.ReadFull(rand.Reader, plaintext); err != nil {
		return nil, nil, &KMSError{Code: errDependencyTimeout, Message: "Failed to generate data key"}
	}

	// Encrypt the data key using the KMS key.
	block, err := aes.NewCipher(key.KeyMaterial)
	if err != nil {
		return nil, nil, &KMSError{Code: errDependencyTimeout, Message: "Encryption failed"}
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, &KMSError{Code: errDependencyTimeout, Message: "Encryption failed"}
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, &KMSError{Code: errDependencyTimeout, Message: "Encryption failed"}
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	encryptedKey := make([]byte, 0, 36+len(nonce)+len(ciphertext))
	encryptedKey = append(encryptedKey, []byte(key.KeyID)...)
	encryptedKey = append(encryptedKey, nonce...)
	encryptedKey = append(encryptedKey, ciphertext...)

	return plaintext, encryptedKey, nil
}

// CreateAlias creates an alias for a key.
func (s *MemoryStorage) CreateAlias(_ context.Context, aliasName, targetKeyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate alias name.
	if len(aliasName) < 7 || aliasName[:6] != "alias/" {
		return &KMSError{Code: errInvalidAlias, Message: "Alias must begin with 'alias/'"}
	}

	// Check if alias already exists.
	if _, ok := s.aliases[aliasName]; ok {
		return &KMSError{Code: errAlreadyExists, Message: "Alias " + aliasName + " already exists."}
	}

	// Verify target key exists.
	key, err := s.getKeyLocked(targetKeyID)
	if err != nil {
		return err
	}

	aliasArn := fmt.Sprintf("arn:aws:kms:%s:%s:%s", s.region, defaultAccountID, aliasName)
	now := time.Now()

	s.aliases[aliasName] = &Alias{
		AliasName:       aliasName,
		AliasArn:        aliasArn,
		TargetKeyID:     key.KeyID,
		CreationDate:    now,
		LastUpdatedDate: now,
	}

	return nil
}

// DeleteAlias deletes an alias.
func (s *MemoryStorage) DeleteAlias(_ context.Context, aliasName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.aliases[aliasName]; !ok {
		return &KMSError{Code: errNotFound, Message: "Alias " + aliasName + " is not found."}
	}

	delete(s.aliases, aliasName)

	return nil
}

// ListAliases lists aliases.
func (s *MemoryStorage) ListAliases(_ context.Context, keyID string, limit int32, _ string) ([]*Alias, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	aliases := make([]*Alias, 0)

	for _, alias := range s.aliases {
		if keyID != "" && alias.TargetKeyID != keyID {
			// If keyID is specified, filter by it.
			// Also need to resolve keyID if it's an alias or ARN.
			resolvedKey, err := s.getKeyLocked(keyID)
			if err != nil {
				continue
			}

			if alias.TargetKeyID != resolvedKey.KeyID {
				continue
			}
		}

		aliases = append(aliases, alias)
		if int32(len(aliases)) >= limit {
			break
		}
	}

	return aliases, "", nil
}

// GetAlias retrieves an alias by name.
func (s *MemoryStorage) GetAlias(_ context.Context, aliasName string) (*Alias, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	alias, ok := s.aliases[aliasName]
	if !ok {
		return nil, &KMSError{Code: errNotFound, Message: "Alias " + aliasName + " is not found."}
	}

	return alias, nil
}
