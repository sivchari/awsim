package ecr

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"sort"
	"sync"
	"time"
)

// Error codes.
const (
	errRepositoryNotFound      = "RepositoryNotFoundException"
	errRepositoryAlreadyExists = "RepositoryAlreadyExistsException"
	errImageNotFound           = "ImageNotFoundException"
	errInvalidParameter        = "InvalidParameterException"
)

// Storage defines the ECR storage interface.
type Storage interface {
	CreateRepository(ctx context.Context, req *CreateRepositoryRequest) (*Repository, error)
	DeleteRepository(ctx context.Context, repositoryName string, force bool) (*Repository, error)
	DescribeRepositories(ctx context.Context, names []string, maxResults int32, nextToken string) ([]*Repository, string, error)
	ListImages(ctx context.Context, repositoryName string, maxResults int32, nextToken string) ([]*ImageIdentifier, string, error)
	PutImage(ctx context.Context, repositoryName, imageManifest, imageTag string) (*Image, error)
	BatchGetImage(ctx context.Context, repositoryName string, imageIDs []ImageIdentifier) ([]*Image, []ImageFailure, error)
	BatchDeleteImage(ctx context.Context, repositoryName string, imageIDs []ImageIdentifier) ([]ImageIdentifier, []ImageFailure, error)
	GetAuthorizationToken(ctx context.Context) ([]AuthorizationData, error)
	DispatchAction(action string) bool
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu           sync.RWMutex
	repositories map[string]*repositoryData
	region       string
	accountID    string
}

// repositoryData holds repository information and its images.
type repositoryData struct {
	repository *Repository
	images     map[string]*Image // keyed by digest
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		repositories: make(map[string]*repositoryData),
		region:       "us-east-1",
		accountID:    "000000000000",
	}
}

// CreateRepository creates a new repository.
func (s *MemoryStorage) CreateRepository(_ context.Context, req *CreateRepositoryRequest) (*Repository, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.repositories[req.RepositoryName]; exists {
		return nil, &ServiceError{Code: errRepositoryAlreadyExists, Message: "Repository already exists"}
	}

	tagMutability := req.ImageTagMutability
	if tagMutability == "" {
		tagMutability = "MUTABLE"
	}

	now := time.Now()
	repo := &Repository{
		RepositoryArn:              fmt.Sprintf("arn:aws:ecr:%s:%s:repository/%s", s.region, s.accountID, req.RepositoryName),
		RegistryID:                 s.accountID,
		RepositoryName:             req.RepositoryName,
		RepositoryURI:              fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s", s.accountID, s.region, req.RepositoryName),
		CreatedAt:                  now,
		ImageTagMutability:         tagMutability,
		ImageScanningConfiguration: req.ImageScanningConfiguration,
		EncryptionConfiguration:    req.EncryptionConfiguration,
	}

	s.repositories[req.RepositoryName] = &repositoryData{
		repository: repo,
		images:     make(map[string]*Image),
	}

	return repo, nil
}

// DeleteRepository deletes a repository.
func (s *MemoryStorage) DeleteRepository(_ context.Context, repositoryName string, force bool) (*Repository, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rd, exists := s.repositories[repositoryName]
	if !exists {
		return nil, &ServiceError{Code: errRepositoryNotFound, Message: "Repository does not exist"}
	}

	if !force && len(rd.images) > 0 {
		return nil, &ServiceError{Code: errInvalidParameter, Message: "Repository contains images"}
	}

	delete(s.repositories, repositoryName)

	return rd.repository, nil
}

// DescribeRepositories describes repositories.
func (s *MemoryStorage) DescribeRepositories(_ context.Context, names []string, maxResults int32, _ string) ([]*Repository, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxResults <= 0 {
		maxResults = 100
	}

	var repos []*Repository

	if len(names) > 0 {
		for _, name := range names {
			if rd, exists := s.repositories[name]; exists {
				repos = append(repos, rd.repository)
			}
		}
	} else {
		for _, rd := range s.repositories {
			repos = append(repos, rd.repository)
		}
	}

	sort.Slice(repos, func(i, j int) bool {
		return repos[i].RepositoryName < repos[j].RepositoryName
	})

	if int32(len(repos)) > maxResults {
		repos = repos[:maxResults]
	}

	return repos, "", nil
}

// ListImages lists images in a repository.
func (s *MemoryStorage) ListImages(_ context.Context, repositoryName string, maxResults int32, _ string) ([]*ImageIdentifier, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rd, exists := s.repositories[repositoryName]
	if !exists {
		return nil, "", &ServiceError{Code: errRepositoryNotFound, Message: "Repository does not exist"}
	}

	if maxResults <= 0 {
		maxResults = 100
	}

	var imageIDs []*ImageIdentifier

	for _, img := range rd.images {
		imageIDs = append(imageIDs, &ImageIdentifier{
			ImageDigest: img.ImageDigest,
			ImageTag:    img.ImageID.ImageTag,
		})
	}

	if int32(len(imageIDs)) > maxResults {
		imageIDs = imageIDs[:maxResults]
	}

	return imageIDs, "", nil
}

// PutImage puts an image into a repository.
func (s *MemoryStorage) PutImage(_ context.Context, repositoryName, imageManifest, imageTag string) (*Image, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rd, exists := s.repositories[repositoryName]
	if !exists {
		return nil, &ServiceError{Code: errRepositoryNotFound, Message: "Repository does not exist"}
	}

	digest := calculateDigest(imageManifest)

	img := &Image{
		RegistryID:     s.accountID,
		RepositoryName: repositoryName,
		ImageManifest:  imageManifest,
		ImageDigest:    digest,
		ImageID: &ImageIdentifier{
			ImageDigest: digest,
			ImageTag:    imageTag,
		},
	}

	rd.images[digest] = img

	return img, nil
}

// BatchGetImage gets multiple images.
func (s *MemoryStorage) BatchGetImage(_ context.Context, repositoryName string, imageIDs []ImageIdentifier) ([]*Image, []ImageFailure, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rd, exists := s.repositories[repositoryName]
	if !exists {
		return nil, nil, &ServiceError{Code: errRepositoryNotFound, Message: "Repository does not exist"}
	}

	var images []*Image
	var failures []ImageFailure

	for _, id := range imageIDs {
		found := false

		for _, img := range rd.images {
			if (id.ImageDigest != "" && img.ImageDigest == id.ImageDigest) ||
				(id.ImageTag != "" && img.ImageID.ImageTag == id.ImageTag) {
				images = append(images, img)
				found = true

				break
			}
		}

		if !found {
			failures = append(failures, ImageFailure{
				ImageID:       &id,
				FailureCode:   "ImageNotFound",
				FailureReason: "Image not found",
			})
		}
	}

	return images, failures, nil
}

// BatchDeleteImage deletes multiple images.
func (s *MemoryStorage) BatchDeleteImage(_ context.Context, repositoryName string, imageIDs []ImageIdentifier) ([]ImageIdentifier, []ImageFailure, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rd, exists := s.repositories[repositoryName]
	if !exists {
		return nil, nil, &ServiceError{Code: errRepositoryNotFound, Message: "Repository does not exist"}
	}

	var deleted []ImageIdentifier
	var failures []ImageFailure

	for _, id := range imageIDs {
		found := false

		for digest, img := range rd.images {
			if (id.ImageDigest != "" && img.ImageDigest == id.ImageDigest) ||
				(id.ImageTag != "" && img.ImageID.ImageTag == id.ImageTag) {
				delete(rd.images, digest)
				deleted = append(deleted, ImageIdentifier{
					ImageDigest: img.ImageDigest,
					ImageTag:    img.ImageID.ImageTag,
				})
				found = true

				break
			}
		}

		if !found {
			failures = append(failures, ImageFailure{
				ImageID:       &id,
				FailureCode:   "ImageNotFound",
				FailureReason: "Image not found",
			})
		}
	}

	return deleted, failures, nil
}

// GetAuthorizationToken returns authorization tokens.
func (s *MemoryStorage) GetAuthorizationToken(_ context.Context) ([]AuthorizationData, error) {
	token := base64.StdEncoding.EncodeToString([]byte("AWS:password"))
	expiresAt := time.Now().Add(12 * time.Hour)

	return []AuthorizationData{
		{
			AuthorizationToken: token,
			ExpiresAt:          float64(expiresAt.Unix()),
			ProxyEndpoint:      fmt.Sprintf("https://%s.dkr.ecr.%s.amazonaws.com", s.accountID, s.region),
		},
	}, nil
}

// DispatchAction checks if the action is valid.
func (s *MemoryStorage) DispatchAction(_ string) bool {
	return true
}

// calculateDigest calculates SHA256 digest of the manifest.
func calculateDigest(manifest string) string {
	hash := sha256.Sum256([]byte(manifest))

	return fmt.Sprintf("sha256:%x", hash)
}
