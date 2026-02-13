package kinesis

import (
	"context"
	"crypto/md5" //nolint:gosec // MD5 used for partition key hashing, not security
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Error codes.
const (
	errResourceNotFound = "ResourceNotFoundException"
	errResourceInUse    = "ResourceInUseException"
	errInvalidArgument  = "InvalidArgumentException"
	errExpiredIterator  = "ExpiredIteratorException"
)

// Default values.
const (
	defaultShardCount       = 1
	defaultRetentionHours   = 24
	maxRecordsPerGet        = 10000
	shardIteratorExpiration = 5 * time.Minute
)

// Storage defines the Kinesis storage interface.
type Storage interface {
	// Stream operations.
	CreateStream(ctx context.Context, req *CreateStreamRequest) error
	DeleteStream(ctx context.Context, streamName string) error
	DescribeStream(ctx context.Context, streamName string, limit int32, exclusiveStartShardID string) (*Stream, []*Shard, bool, error)
	ListStreams(ctx context.Context, exclusiveStartStreamName string, limit int32) ([]*Stream, bool, error)
	ListShards(ctx context.Context, streamName string, nextToken string, maxResults int32) ([]*Shard, string, error)

	// Record operations.
	PutRecord(ctx context.Context, streamName string, data []byte, partitionKey string, explicitHashKey string) (string, string, error)
	PutRecords(ctx context.Context, streamName string, records []PutRecordsRequestEntry) ([]PutRecordsResultEntry, int32, error)
	GetShardIterator(ctx context.Context, streamName, shardID, iteratorType string, startingSeqNum string, timestamp float64) (string, error)
	GetRecords(ctx context.Context, shardIterator string, limit int32) ([]*Record, string, int64, error)

	// DispatchAction dispatches the request to the appropriate handler.
	DispatchAction(action string) bool
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu              sync.RWMutex
	streams         map[string]*streamData
	shardIterators  map[string]*shardIteratorData
	region          string
	accountID       string
	sequenceCounter uint64
}

// streamData holds stream information and its shards.
type streamData struct {
	stream *Stream
	shards map[string]*shardData
}

// shardData holds shard information and its records.
type shardData struct {
	shard   *Shard
	records []*Record
}

// shardIteratorData holds shard iterator state.
type shardIteratorData struct {
	streamName string
	shardID    string
	position   int
	expiresAt  time.Time
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		streams:        make(map[string]*streamData),
		shardIterators: make(map[string]*shardIteratorData),
		region:         "us-east-1",
		accountID:      "000000000000",
	}
}

// CreateStream creates a new stream.
func (s *MemoryStorage) CreateStream(_ context.Context, req *CreateStreamRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.streams[req.StreamName]; exists {
		return &ServiceError{Code: errResourceInUse, Message: "Stream already exists"}
	}

	shardCount := int32(defaultShardCount)
	if req.ShardCount != nil && *req.ShardCount > 0 {
		shardCount = *req.ShardCount
	}

	now := time.Now()
	stream := &Stream{
		StreamName:              req.StreamName,
		StreamARN:               fmt.Sprintf("arn:aws:kinesis:%s:%s:stream/%s", s.region, s.accountID, req.StreamName),
		StreamStatus:            StreamStatusActive,
		ShardCount:              shardCount,
		RetentionPeriodHours:    defaultRetentionHours,
		StreamCreationTimestamp: now,
		EnhancedMonitoring:      []EnhancedMetrics{{ShardLevelMetrics: []string{}}},
		StreamModeDetails:       req.StreamModeDetails,
		OpenShardCount:          shardCount,
	}

	if stream.StreamModeDetails == nil {
		stream.StreamModeDetails = &StreamModeDetails{StreamMode: "PROVISIONED"}
	}

	// Create shards.
	shards := make(map[string]*shardData)
	hashKeyRange := calculateHashKeyRanges(shardCount)

	for i := int32(0); i < shardCount; i++ {
		shardID := fmt.Sprintf("shardId-%012d", i)
		seqNum := s.nextSequenceNumber()

		shard := &Shard{
			ShardID:      shardID,
			HashKeyRange: hashKeyRange[i],
			SequenceNumberRange: SequenceNumberRange{
				StartingSequenceNumber: seqNum,
			},
		}

		shards[shardID] = &shardData{
			shard:   shard,
			records: make([]*Record, 0),
		}
	}

	s.streams[req.StreamName] = &streamData{
		stream: stream,
		shards: shards,
	}

	return nil
}

// DeleteStream deletes a stream.
func (s *MemoryStorage) DeleteStream(_ context.Context, streamName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.streams[streamName]; !exists {
		return &ServiceError{Code: errResourceNotFound, Message: "Stream not found"}
	}

	delete(s.streams, streamName)

	return nil
}

// DescribeStream describes a stream.
func (s *MemoryStorage) DescribeStream(_ context.Context, streamName string, limit int32, exclusiveStartShardID string) (*Stream, []*Shard, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sd, exists := s.streams[streamName]
	if !exists {
		return nil, nil, false, &ServiceError{Code: errResourceNotFound, Message: "Stream not found"}
	}

	// Collect and sort shards.
	shards := make([]*Shard, 0, len(sd.shards))
	for _, shardData := range sd.shards {
		shards = append(shards, shardData.shard)
	}

	sort.Slice(shards, func(i, j int) bool {
		return shards[i].ShardID < shards[j].ShardID
	})

	// Apply pagination.
	startIndex := 0

	if exclusiveStartShardID != "" {
		for i, shard := range shards {
			if shard.ShardID == exclusiveStartShardID {
				startIndex = i + 1

				break
			}
		}
	}

	if limit <= 0 {
		limit = 100
	}

	endIndex := startIndex + int(limit)
	hasMoreShards := false

	if endIndex > len(shards) {
		endIndex = len(shards)
	} else {
		hasMoreShards = true
	}

	return sd.stream, shards[startIndex:endIndex], hasMoreShards, nil
}

// ListStreams lists all streams.
func (s *MemoryStorage) ListStreams(_ context.Context, exclusiveStartStreamName string, limit int32) ([]*Stream, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	// Collect and sort stream names.
	names := make([]string, 0, len(s.streams))
	for name := range s.streams {
		names = append(names, name)
	}

	sort.Strings(names)

	// Apply pagination.
	startIndex := 0

	if exclusiveStartStreamName != "" {
		for i, name := range names {
			if name == exclusiveStartStreamName {
				startIndex = i + 1

				break
			}
		}
	}

	endIndex := startIndex + int(limit)
	hasMoreStreams := false

	if endIndex > len(names) {
		endIndex = len(names)
	} else {
		hasMoreStreams = true
	}

	streams := make([]*Stream, endIndex-startIndex)
	for i, name := range names[startIndex:endIndex] {
		streams[i] = s.streams[name].stream
	}

	return streams, hasMoreStreams, nil
}

// ListShards lists shards for a stream.
func (s *MemoryStorage) ListShards(_ context.Context, streamName string, _ string, maxResults int32) ([]*Shard, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sd, exists := s.streams[streamName]
	if !exists {
		return nil, "", &ServiceError{Code: errResourceNotFound, Message: "Stream not found"}
	}

	if maxResults <= 0 {
		maxResults = 100
	}

	// Collect and sort shards.
	shards := make([]*Shard, 0, len(sd.shards))
	for _, shardData := range sd.shards {
		shards = append(shards, shardData.shard)
	}

	sort.Slice(shards, func(i, j int) bool {
		return shards[i].ShardID < shards[j].ShardID
	})

	if int32(len(shards)) > maxResults { //nolint:gosec // slice length bounded by maxResults parameter
		shards = shards[:maxResults]
	}

	return shards, "", nil
}

// PutRecord puts a single record to a stream.
func (s *MemoryStorage) PutRecord(_ context.Context, streamName string, data []byte, partitionKey string, explicitHashKey string) (string, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sd, exists := s.streams[streamName]
	if !exists {
		return "", "", &ServiceError{Code: errResourceNotFound, Message: "Stream not found"}
	}

	// Determine shard based on hash key.
	hashKey := explicitHashKey
	if hashKey == "" {
		hashKey = computeHashKey(partitionKey)
	}

	shardID := s.findShardForHashKey(sd, hashKey)
	if shardID == "" {
		return "", "", &ServiceError{Code: errInvalidArgument, Message: "Could not determine shard for partition key"}
	}

	seqNum := s.nextSequenceNumber()
	record := &Record{
		Data:                        data,
		PartitionKey:                partitionKey,
		SequenceNumber:              seqNum,
		ApproximateArrivalTimestamp: time.Now(),
	}

	sd.shards[shardID].records = append(sd.shards[shardID].records, record)

	return shardID, seqNum, nil
}

// PutRecords puts multiple records to a stream.
func (s *MemoryStorage) PutRecords(_ context.Context, streamName string, records []PutRecordsRequestEntry) ([]PutRecordsResultEntry, int32, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sd, exists := s.streams[streamName]
	if !exists {
		return nil, 0, &ServiceError{Code: errResourceNotFound, Message: "Stream not found"}
	}

	results := make([]PutRecordsResultEntry, len(records))

	for i, entry := range records {
		hashKey := entry.ExplicitHashKey
		if hashKey == "" {
			hashKey = computeHashKey(entry.PartitionKey)
		}

		shardID := s.findShardForHashKey(sd, hashKey)
		if shardID == "" {
			results[i] = PutRecordsResultEntry{
				ErrorCode:    errInvalidArgument,
				ErrorMessage: "Could not determine shard for partition key",
			}

			continue
		}

		seqNum := s.nextSequenceNumber()
		record := &Record{
			Data:                        entry.Data,
			PartitionKey:                entry.PartitionKey,
			SequenceNumber:              seqNum,
			ApproximateArrivalTimestamp: time.Now(),
		}

		sd.shards[shardID].records = append(sd.shards[shardID].records, record)

		results[i] = PutRecordsResultEntry{
			ShardID:        shardID,
			SequenceNumber: seqNum,
		}
	}

	var failedCount int32

	for _, r := range results {
		if r.ErrorCode != "" {
			failedCount++
		}
	}

	return results, failedCount, nil
}

// GetShardIterator gets a shard iterator.
func (s *MemoryStorage) GetShardIterator(_ context.Context, streamName, shardID, iteratorType string, startingSeqNum string, _ float64) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sd, exists := s.streams[streamName]
	if !exists {
		return "", &ServiceError{Code: errResourceNotFound, Message: "Stream not found"}
	}

	shardData, exists := sd.shards[shardID]
	if !exists {
		return "", &ServiceError{Code: errResourceNotFound, Message: "Shard not found"}
	}

	position := 0

	switch iteratorType {
	case string(ShardIteratorTypeTrimHorizon):
		position = 0
	case string(ShardIteratorTypeLatest):
		position = len(shardData.records)
	case string(ShardIteratorTypeAtSequenceNumber):
		position = s.findPositionAtSequenceNumber(shardData.records, startingSeqNum)
	case string(ShardIteratorTypeAfterSequenceNumber):
		position = s.findPositionAtSequenceNumber(shardData.records, startingSeqNum) + 1
	default:
		return "", &ServiceError{Code: errInvalidArgument, Message: "Invalid ShardIteratorType"}
	}

	iteratorID := fmt.Sprintf("%s:%s:%d:%d", streamName, shardID, position, time.Now().UnixNano())
	encodedIterator := base64.StdEncoding.EncodeToString([]byte(iteratorID))

	s.shardIterators[encodedIterator] = &shardIteratorData{
		streamName: streamName,
		shardID:    shardID,
		position:   position,
		expiresAt:  time.Now().Add(shardIteratorExpiration),
	}

	return encodedIterator, nil
}

// GetRecords gets records from a shard.
func (s *MemoryStorage) GetRecords(_ context.Context, shardIterator string, limit int32) ([]*Record, string, int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	iterData, exists := s.shardIterators[shardIterator]
	if !exists {
		return nil, "", 0, &ServiceError{Code: errInvalidArgument, Message: "Invalid shard iterator"}
	}

	if time.Now().After(iterData.expiresAt) {
		delete(s.shardIterators, shardIterator)

		return nil, "", 0, &ServiceError{Code: errExpiredIterator, Message: "Shard iterator has expired"}
	}

	sd, exists := s.streams[iterData.streamName]
	if !exists {
		return nil, "", 0, &ServiceError{Code: errResourceNotFound, Message: "Stream not found"}
	}

	shardData, exists := sd.shards[iterData.shardID]
	if !exists {
		return nil, "", 0, &ServiceError{Code: errResourceNotFound, Message: "Shard not found"}
	}

	if limit <= 0 || limit > maxRecordsPerGet {
		limit = maxRecordsPerGet
	}

	startPos := iterData.position
	endPos := min(startPos+int(limit), len(shardData.records))

	records := make([]*Record, endPos-startPos)
	copy(records, shardData.records[startPos:endPos])

	// Create next iterator.
	delete(s.shardIterators, shardIterator)

	nextIteratorID := fmt.Sprintf("%s:%s:%d:%d", iterData.streamName, iterData.shardID, endPos, time.Now().UnixNano())
	nextIterator := base64.StdEncoding.EncodeToString([]byte(nextIteratorID))

	s.shardIterators[nextIterator] = &shardIteratorData{
		streamName: iterData.streamName,
		shardID:    iterData.shardID,
		position:   endPos,
		expiresAt:  time.Now().Add(shardIteratorExpiration),
	}

	return records, nextIterator, 0, nil
}

// DispatchAction checks if the action is valid.
func (s *MemoryStorage) DispatchAction(_ string) bool {
	return true
}

// Helper functions.

func (s *MemoryStorage) nextSequenceNumber() string {
	seq := atomic.AddUint64(&s.sequenceCounter, 1)

	return fmt.Sprintf("%021d", seq)
}

func (s *MemoryStorage) findShardForHashKey(sd *streamData, hashKey string) string {
	hashKeyBig := new(big.Int)
	hashKeyBig.SetString(hashKey, 10)

	for shardID, shardData := range sd.shards {
		startKey := new(big.Int)
		endKey := new(big.Int)

		startKey.SetString(shardData.shard.HashKeyRange.StartingHashKey, 10)
		endKey.SetString(shardData.shard.HashKeyRange.EndingHashKey, 10)

		if hashKeyBig.Cmp(startKey) >= 0 && hashKeyBig.Cmp(endKey) <= 0 {
			return shardID
		}
	}

	return ""
}

func (s *MemoryStorage) findPositionAtSequenceNumber(records []*Record, seqNum string) int {
	for i, r := range records {
		if r.SequenceNumber == seqNum {
			return i
		}
	}

	return 0
}

func computeHashKey(partitionKey string) string {
	hash := md5.Sum([]byte(partitionKey)) //nolint:gosec // MD5 used for partition key hashing, not security
	hashHex := hex.EncodeToString(hash[:])

	hashInt := new(big.Int)
	hashInt.SetString(hashHex, 16)

	return hashInt.String()
}

func calculateHashKeyRanges(shardCount int32) []HashKeyRange {
	maxHashKey := new(big.Int)
	maxHashKey.SetString("340282366920938463463374607431768211455", 10) // 2^128 - 1

	ranges := make([]HashKeyRange, shardCount)
	shardSize := new(big.Int).Div(maxHashKey, big.NewInt(int64(shardCount)))

	for i := range shardCount {
		start := new(big.Int).Mul(shardSize, big.NewInt(int64(i)))
		end := new(big.Int).Mul(shardSize, big.NewInt(int64(i+1)))
		end.Sub(end, big.NewInt(1))

		if i == shardCount-1 {
			end = maxHashKey
		}

		ranges[i] = HashKeyRange{
			StartingHashKey: start.String(),
			EndingHashKey:   end.String(),
		}
	}

	return ranges
}
