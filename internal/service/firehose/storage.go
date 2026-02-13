package firehose

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Storage defines the interface for Firehose storage operations.
type Storage interface {
	CreateDeliveryStream(ctx context.Context, input *CreateDeliveryStreamInput) (*DeliveryStream, error)
	DeleteDeliveryStream(ctx context.Context, name string, allowForceDelete bool) error
	DescribeDeliveryStream(ctx context.Context, name string, limit int32, exclusiveStartDestinationID string) (*DeliveryStream, error)
	ListDeliveryStreams(ctx context.Context, streamType, exclusiveStartName string, limit int32) ([]string, bool, error)
	PutRecord(ctx context.Context, streamName string, record Record) (string, error)
	PutRecordBatch(ctx context.Context, streamName string, records []Record) ([]PutRecordBatchResponseEntry, int32, error)
	UpdateDestination(ctx context.Context, input *UpdateDestinationInput) error
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu      sync.RWMutex
	streams map[string]*streamData
}

type streamData struct {
	stream  *DeliveryStream
	records []storedRecord
}

type storedRecord struct {
	recordID string
	data     []byte
	received time.Time
}

// NewMemoryStorage creates a new MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		streams: make(map[string]*streamData),
	}
}

// CreateDeliveryStream creates a new delivery stream.
func (s *MemoryStorage) CreateDeliveryStream(_ context.Context, input *CreateDeliveryStreamInput) (*DeliveryStream, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.streams[input.DeliveryStreamName]; exists {
		return nil, &Error{
			Code:    errResourceInUse,
			Message: fmt.Sprintf("Delivery stream %s already exists", input.DeliveryStreamName),
		}
	}

	stream := s.buildDeliveryStream(input)
	s.streams[input.DeliveryStreamName] = &streamData{
		stream:  stream,
		records: make([]storedRecord, 0),
	}

	return stream, nil
}

func (s *MemoryStorage) buildDeliveryStream(input *CreateDeliveryStreamInput) *DeliveryStream {
	now := time.Now()

	streamType := DeliveryStreamTypeDirectPut
	if input.DeliveryStreamType != "" {
		streamType = DeliveryStreamType(input.DeliveryStreamType)
	}

	arn := fmt.Sprintf("arn:aws:firehose:us-east-1:000000000000:deliverystream/%s", input.DeliveryStreamName)

	stream := &DeliveryStream{
		DeliveryStreamName:   input.DeliveryStreamName,
		DeliveryStreamARN:    arn,
		DeliveryStreamStatus: DeliveryStreamStatusActive,
		DeliveryStreamType:   streamType,
		CreateTimestamp:      now,
		LastUpdateTimestamp:  now,
		VersionID:            "1",
		HasMoreDestinations:  false,
	}

	stream.Destinations = s.buildDestinations(input)

	if input.KinesisStreamSourceConfiguration != nil {
		stream.Source = &SourceDescription{
			KinesisStreamSourceDescription: &KinesisStreamSourceDescription{
				KinesisStreamARN:       input.KinesisStreamSourceConfiguration.KinesisStreamARN,
				RoleARN:                input.KinesisStreamSourceConfiguration.RoleARN,
				DeliveryStartTimestamp: now,
			},
		}
	}

	return stream
}

func (s *MemoryStorage) buildDestinations(input *CreateDeliveryStreamInput) []DestinationDescription {
	destID := uuid.New().String()
	destinations := make([]DestinationDescription, 0)

	if input.S3DestinationConfiguration != nil {
		dest := DestinationDescription{
			DestinationID: destID,
			S3DestinationDescription: &S3DestinationDescription{
				BucketARN:         input.S3DestinationConfiguration.BucketARN,
				Prefix:            input.S3DestinationConfiguration.Prefix,
				ErrorOutputPrefix: input.S3DestinationConfiguration.ErrorOutputPrefix,
				RoleARN:           input.S3DestinationConfiguration.RoleARN,
				BufferingHints:    input.S3DestinationConfiguration.BufferingHints,
				CompressionFormat: input.S3DestinationConfiguration.CompressionFormat,
				CloudWatchLogging: input.S3DestinationConfiguration.CloudWatchLogging,
			},
		}

		destinations = append(destinations, dest)
	}

	if input.ExtendedS3DestinationConfiguration != nil {
		dest := DestinationDescription{
			DestinationID: destID,
			ExtendedS3DestinationDescription: &ExtendedS3DestinationDescription{
				BucketARN:         input.ExtendedS3DestinationConfiguration.BucketARN,
				Prefix:            input.ExtendedS3DestinationConfiguration.Prefix,
				ErrorOutputPrefix: input.ExtendedS3DestinationConfiguration.ErrorOutputPrefix,
				RoleARN:           input.ExtendedS3DestinationConfiguration.RoleARN,
				BufferingHints:    input.ExtendedS3DestinationConfiguration.BufferingHints,
				CompressionFormat: input.ExtendedS3DestinationConfiguration.CompressionFormat,
				CloudWatchLogging: input.ExtendedS3DestinationConfiguration.CloudWatchLogging,
				ProcessingConfig:  input.ExtendedS3DestinationConfiguration.ProcessingConfig,
				S3BackupMode:      input.ExtendedS3DestinationConfiguration.S3BackupMode,
			},
		}

		destinations = append(destinations, dest)
	}

	// If no destination is provided, create a default.
	if len(destinations) == 0 {
		destinations = append(destinations, DestinationDescription{
			DestinationID: destID,
		})
	}

	return destinations
}

// DeleteDeliveryStream deletes a delivery stream.
func (s *MemoryStorage) DeleteDeliveryStream(_ context.Context, name string, _ bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.streams[name]; !exists {
		return &Error{
			Code:    errResourceNotFound,
			Message: fmt.Sprintf("Delivery stream %s not found", name),
		}
	}

	delete(s.streams, name)

	return nil
}

// DescribeDeliveryStream describes a delivery stream.
func (s *MemoryStorage) DescribeDeliveryStream(_ context.Context, name string, _ int32, _ string) (*DeliveryStream, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.streams[name]
	if !exists {
		return nil, &Error{
			Code:    errResourceNotFound,
			Message: fmt.Sprintf("Delivery stream %s not found", name),
		}
	}

	return data.stream, nil
}

// ListDeliveryStreams lists delivery streams.
func (s *MemoryStorage) ListDeliveryStreams(_ context.Context, streamType, exclusiveStartName string, limit int32) ([]string, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit == 0 {
		limit = 10
	}

	if limit > 10000 {
		limit = 10000
	}

	names := s.collectStreamNames(streamType)

	sort.Strings(names)

	startIdx := s.findStartIndex(names, exclusiveStartName)
	if startIdx >= len(names) {
		return []string{}, false, nil
	}

	endIdx := startIdx + int(limit)
	hasMore := false

	if endIdx > len(names) {
		endIdx = len(names)
	} else {
		hasMore = endIdx < len(names)
	}

	return names[startIdx:endIdx], hasMore, nil
}

func (s *MemoryStorage) collectStreamNames(streamType string) []string {
	names := make([]string, 0, len(s.streams))

	for name, data := range s.streams {
		if streamType != "" && string(data.stream.DeliveryStreamType) != streamType {
			continue
		}

		names = append(names, name)
	}

	return names
}

func (s *MemoryStorage) findStartIndex(names []string, exclusiveStartName string) int {
	if exclusiveStartName == "" {
		return 0
	}

	for i, name := range names {
		if name == exclusiveStartName {
			return i + 1
		}
	}

	return 0
}

// PutRecord puts a record to a delivery stream.
func (s *MemoryStorage) PutRecord(_ context.Context, streamName string, record Record) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, exists := s.streams[streamName]
	if !exists {
		return "", &Error{
			Code:    errResourceNotFound,
			Message: fmt.Sprintf("Delivery stream %s not found", streamName),
		}
	}

	recordID := uuid.New().String()

	data.records = append(data.records, storedRecord{
		recordID: recordID,
		data:     record.Data,
		received: time.Now(),
	})

	return recordID, nil
}

// PutRecordBatch puts multiple records to a delivery stream.
func (s *MemoryStorage) PutRecordBatch(_ context.Context, streamName string, records []Record) ([]PutRecordBatchResponseEntry, int32, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, exists := s.streams[streamName]
	if !exists {
		return nil, 0, &Error{
			Code:    errResourceNotFound,
			Message: fmt.Sprintf("Delivery stream %s not found", streamName),
		}
	}

	responses := make([]PutRecordBatchResponseEntry, len(records))
	now := time.Now()

	for i, record := range records {
		recordID := uuid.New().String()

		data.records = append(data.records, storedRecord{
			recordID: recordID,
			data:     record.Data,
			received: now,
		})

		responses[i] = PutRecordBatchResponseEntry{
			RecordID: recordID,
		}
	}

	return responses, 0, nil
}

// UpdateDestination updates a destination.
func (s *MemoryStorage) UpdateDestination(_ context.Context, input *UpdateDestinationInput) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, exists := s.streams[input.DeliveryStreamName]
	if !exists {
		return &Error{
			Code:    errResourceNotFound,
			Message: fmt.Sprintf("Delivery stream %s not found", input.DeliveryStreamName),
		}
	}

	if data.stream.VersionID != input.CurrentDeliveryStreamVersionID {
		return &Error{
			Code:    errInvalidArgument,
			Message: "Invalid version ID",
		}
	}

	found := false

	for i, dest := range data.stream.Destinations {
		if dest.DestinationID != input.DestinationID {
			continue
		}

		found = true

		if input.S3DestinationUpdate != nil {
			if dest.S3DestinationDescription == nil {
				dest.S3DestinationDescription = &S3DestinationDescription{}
			}

			s.applyS3Update(dest.S3DestinationDescription, input.S3DestinationUpdate)
		}

		if input.ExtendedS3DestinationUpdate != nil {
			if dest.ExtendedS3DestinationDescription == nil {
				dest.ExtendedS3DestinationDescription = &ExtendedS3DestinationDescription{}
			}

			s.applyExtendedS3Update(dest.ExtendedS3DestinationDescription, input.ExtendedS3DestinationUpdate)
		}

		data.stream.Destinations[i] = dest

		break
	}

	if !found {
		return &Error{
			Code:    errInvalidArgument,
			Message: fmt.Sprintf("Destination %s not found", input.DestinationID),
		}
	}

	versionNum, _ := strconv.Atoi(data.stream.VersionID)
	data.stream.VersionID = strconv.Itoa(versionNum + 1)
	data.stream.LastUpdateTimestamp = time.Now()

	return nil
}

func (s *MemoryStorage) applyS3Update(desc *S3DestinationDescription, update *S3DestinationUpdate) {
	if update.BucketARN != "" {
		desc.BucketARN = update.BucketARN
	}

	if update.Prefix != "" {
		desc.Prefix = update.Prefix
	}

	if update.ErrorOutputPrefix != "" {
		desc.ErrorOutputPrefix = update.ErrorOutputPrefix
	}

	if update.RoleARN != "" {
		desc.RoleARN = update.RoleARN
	}

	if update.BufferingHints != nil {
		desc.BufferingHints = update.BufferingHints
	}

	if update.CompressionFormat != "" {
		desc.CompressionFormat = update.CompressionFormat
	}

	if update.CloudWatchLogging != nil {
		desc.CloudWatchLogging = update.CloudWatchLogging
	}
}

func (s *MemoryStorage) applyExtendedS3Update(desc *ExtendedS3DestinationDescription, update *ExtendedS3DestinationUpdate) {
	if update.BucketARN != "" {
		desc.BucketARN = update.BucketARN
	}

	if update.Prefix != "" {
		desc.Prefix = update.Prefix
	}

	if update.ErrorOutputPrefix != "" {
		desc.ErrorOutputPrefix = update.ErrorOutputPrefix
	}

	if update.RoleARN != "" {
		desc.RoleARN = update.RoleARN
	}

	if update.BufferingHints != nil {
		desc.BufferingHints = update.BufferingHints
	}

	if update.CompressionFormat != "" {
		desc.CompressionFormat = update.CompressionFormat
	}

	if update.CloudWatchLogging != nil {
		desc.CloudWatchLogging = update.CloudWatchLogging
	}

	if update.ProcessingConfig != nil {
		desc.ProcessingConfig = update.ProcessingConfig
	}

	if update.S3BackupMode != "" {
		desc.S3BackupMode = update.S3BackupMode
	}
}
