package cloudwatchlogs

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	defaultRegion    = "us-east-1"
	defaultAccountID = "000000000000"
	defaultLimit     = 50
	maxLimit         = 10000
)

// Storage defines the CloudWatch Logs storage interface.
type Storage interface {
	CreateLogGroup(ctx context.Context, req *CreateLogGroupRequest) error
	DeleteLogGroup(ctx context.Context, name string) error
	CreateLogStream(ctx context.Context, groupName, streamName string) error
	DeleteLogStream(ctx context.Context, groupName, streamName string) error
	PutLogEvents(ctx context.Context, groupName, streamName string, events []InputLogEvent, sequenceToken string) (*PutLogEventsResponse, error)
	GetLogEvents(ctx context.Context, req *GetLogEventsRequest) (*GetLogEventsResponse, error)
	FilterLogEvents(ctx context.Context, req *FilterLogEventsRequest) (*FilterLogEventsResponse, error)
	DescribeLogGroups(ctx context.Context, req *DescribeLogGroupsRequest) (*DescribeLogGroupsResponse, error)
	DescribeLogStreams(ctx context.Context, req *DescribeLogStreamsRequest) (*DescribeLogStreamsResponse, error)
}

// logStreamData holds log stream data with events.
type logStreamData struct {
	stream *LogStream
	events []*LogEvent
}

// logGroupData holds log group data with streams.
type logGroupData struct {
	group   *LogGroup
	streams map[string]*logStreamData
}

// MemoryStorage implements Storage with in-memory data.
type MemoryStorage struct {
	mu        sync.RWMutex
	logGroups map[string]*logGroupData
	baseURL   string
}

// NewMemoryStorage creates a new in-memory CloudWatch Logs storage.
func NewMemoryStorage(baseURL string) *MemoryStorage {
	return &MemoryStorage{
		logGroups: make(map[string]*logGroupData),
		baseURL:   baseURL,
	}
}

// CreateLogGroup creates a new log group.
func (m *MemoryStorage) CreateLogGroup(_ context.Context, req *CreateLogGroupRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.logGroups[req.LogGroupName]; exists {
		return &LogsError{
			Code:    "ResourceAlreadyExistsException",
			Message: fmt.Sprintf("The specified log group already exists: %s", req.LogGroupName),
		}
	}

	now := time.Now().UnixMilli()
	logGroup := &LogGroup{
		LogGroupName:  req.LogGroupName,
		LogGroupARN:   m.buildLogGroupARN(req.LogGroupName),
		CreationTime:  now,
		KmsKeyID:      req.KmsKeyID,
		LogGroupClass: req.LogGroupClass,
	}

	m.logGroups[req.LogGroupName] = &logGroupData{
		group:   logGroup,
		streams: make(map[string]*logStreamData),
	}

	return nil
}

// DeleteLogGroup deletes a log group.
func (m *MemoryStorage) DeleteLogGroup(_ context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.logGroups[name]; !exists {
		return &LogsError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("The specified log group does not exist: %s", name),
		}
	}

	delete(m.logGroups, name)

	return nil
}

// CreateLogStream creates a new log stream in a log group.
func (m *MemoryStorage) CreateLogStream(_ context.Context, groupName, streamName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	groupData, exists := m.logGroups[groupName]
	if !exists {
		return &LogsError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("The specified log group does not exist: %s", groupName),
		}
	}

	if _, exists := groupData.streams[streamName]; exists {
		return &LogsError{
			Code:    "ResourceAlreadyExistsException",
			Message: fmt.Sprintf("The specified log stream already exists: %s", streamName),
		}
	}

	now := time.Now().UnixMilli()
	stream := &LogStream{
		LogStreamName:       streamName,
		CreationTime:        now,
		UploadSequenceToken: uuid.New().String(),
		LogStreamARN:        m.buildLogStreamARN(groupName, streamName),
	}

	groupData.streams[streamName] = &logStreamData{
		stream: stream,
		events: make([]*LogEvent, 0),
	}

	return nil
}

// DeleteLogStream deletes a log stream from a log group.
func (m *MemoryStorage) DeleteLogStream(_ context.Context, groupName, streamName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	groupData, exists := m.logGroups[groupName]
	if !exists {
		return &LogsError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("The specified log group does not exist: %s", groupName),
		}
	}

	if _, exists := groupData.streams[streamName]; !exists {
		return &LogsError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("The specified log stream does not exist: %s", streamName),
		}
	}

	delete(groupData.streams, streamName)

	return nil
}

// PutLogEvents puts log events into a log stream.
func (m *MemoryStorage) PutLogEvents(_ context.Context, groupName, streamName string, events []InputLogEvent, _ string) (*PutLogEventsResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	groupData, exists := m.logGroups[groupName]
	if !exists {
		return nil, &LogsError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("The specified log group does not exist: %s", groupName),
		}
	}

	streamData, exists := groupData.streams[streamName]
	if !exists {
		return nil, &LogsError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("The specified log stream does not exist: %s", streamName),
		}
	}

	now := time.Now().UnixMilli()

	for _, event := range events {
		logEvent := &LogEvent{
			Timestamp: event.Timestamp,
			Message:   event.Message,
		}
		streamData.events = append(streamData.events, logEvent)

		// Update stream timestamps
		if streamData.stream.FirstEventTimestamp == nil {
			streamData.stream.FirstEventTimestamp = &event.Timestamp
		}

		streamData.stream.LastEventTimestamp = &event.Timestamp
		streamData.stream.LastIngestionTime = &now
		streamData.stream.StoredBytes += int64(len(event.Message))
	}

	// Update group stored bytes
	groupData.group.StoredBytes += sumEventBytes(events)

	// Generate new sequence token
	newToken := uuid.New().String()
	streamData.stream.UploadSequenceToken = newToken

	return &PutLogEventsResponse{
		NextSequenceToken: newToken,
	}, nil
}

// GetLogEvents retrieves log events from a log stream.
func (m *MemoryStorage) GetLogEvents(_ context.Context, req *GetLogEventsRequest) (*GetLogEventsResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	groupData, exists := m.logGroups[req.LogGroupName]
	if !exists {
		return nil, &LogsError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("The specified log group does not exist: %s", req.LogGroupName),
		}
	}

	streamData, exists := groupData.streams[req.LogStreamName]
	if !exists {
		return nil, &LogsError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("The specified log stream does not exist: %s", req.LogStreamName),
		}
	}

	limit := defaultLimit
	if req.Limit != nil && *req.Limit > 0 {
		limit = min(int(*req.Limit), maxLimit)
	}

	// Filter events by time range
	filteredEvents := filterEventsByTime(streamData.events, req.StartTime, req.EndTime)

	// Sort events
	startFromHead := req.StartFromHead != nil && *req.StartFromHead
	if startFromHead {
		sort.Slice(filteredEvents, func(i, j int) bool {
			return filteredEvents[i].Timestamp < filteredEvents[j].Timestamp
		})
	} else {
		sort.Slice(filteredEvents, func(i, j int) bool {
			return filteredEvents[i].Timestamp > filteredEvents[j].Timestamp
		})
	}

	// Apply limit
	if len(filteredEvents) > limit {
		filteredEvents = filteredEvents[:limit]
	}

	// Convert to output format
	outputEvents := make([]OutputLogEvent, 0, len(filteredEvents))
	now := time.Now().UnixMilli()

	for _, event := range filteredEvents {
		outputEvents = append(outputEvents, OutputLogEvent{
			Timestamp:     event.Timestamp,
			Message:       event.Message,
			IngestionTime: now,
		})
	}

	return &GetLogEventsResponse{
		Events:            outputEvents,
		NextForwardToken:  uuid.New().String(),
		NextBackwardToken: uuid.New().String(),
	}, nil
}

// FilterLogEvents filters log events across log streams.
func (m *MemoryStorage) FilterLogEvents(_ context.Context, req *FilterLogEventsRequest) (*FilterLogEventsResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	groupName := req.LogGroupName
	if groupName == "" {
		groupName = req.LogGroupIdentifier
	}

	groupData, exists := m.logGroups[groupName]
	if !exists {
		return nil, &LogsError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("The specified log group does not exist: %s", groupName),
		}
	}

	limit := defaultLimit
	if req.Limit != nil && *req.Limit > 0 {
		limit = min(int(*req.Limit), maxLimit)
	}

	var allEvents []FilteredLogEvent

	var searchedStreams []SearchedLogStream

	for streamName, streamData := range groupData.streams {
		// Check if we should include this stream
		if len(req.LogStreamNames) > 0 && !slices.Contains(req.LogStreamNames, streamName) {
			continue
		}

		if req.LogStreamNamePrefix != "" && !strings.HasPrefix(streamName, req.LogStreamNamePrefix) {
			continue
		}

		searchedStreams = append(searchedStreams, SearchedLogStream{
			LogStreamName:      streamName,
			SearchedCompletely: true,
		})

		// Filter events
		for i, event := range streamData.events {
			// Time filtering
			if req.StartTime != nil && event.Timestamp < *req.StartTime {
				continue
			}

			if req.EndTime != nil && event.Timestamp > *req.EndTime {
				continue
			}

			// Pattern filtering (simple substring match)
			if req.FilterPattern != "" && !strings.Contains(event.Message, req.FilterPattern) {
				continue
			}

			allEvents = append(allEvents, FilteredLogEvent{
				LogStreamName: streamName,
				Timestamp:     event.Timestamp,
				Message:       event.Message,
				IngestionTime: time.Now().UnixMilli(),
				EventID:       fmt.Sprintf("%d-%d", event.Timestamp, i),
			})
		}
	}

	// Sort by timestamp
	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i].Timestamp < allEvents[j].Timestamp
	})

	// Apply limit
	if len(allEvents) > limit {
		allEvents = allEvents[:limit]
	}

	return &FilterLogEventsResponse{
		Events:             allEvents,
		SearchedLogStreams: searchedStreams,
	}, nil
}

// DescribeLogGroups describes log groups.
func (m *MemoryStorage) DescribeLogGroups(_ context.Context, req *DescribeLogGroupsRequest) (*DescribeLogGroupsResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	limit := defaultLimit
	if req.Limit != nil && *req.Limit > 0 {
		limit = min(int(*req.Limit), maxLimit)
	}

	var groups []LogGroupResponse

	for name, groupData := range m.logGroups {
		// Apply prefix filter
		if req.LogGroupNamePrefix != "" && !strings.HasPrefix(name, req.LogGroupNamePrefix) {
			continue
		}

		// Apply pattern filter (simple substring match)
		if req.LogGroupNamePattern != "" && !strings.Contains(name, req.LogGroupNamePattern) {
			continue
		}

		// Apply class filter
		if req.LogGroupClass != "" && groupData.group.LogGroupClass != req.LogGroupClass {
			continue
		}

		groups = append(groups, LogGroupResponse{
			LogGroupName:      groupData.group.LogGroupName,
			LogGroupARN:       groupData.group.LogGroupARN,
			CreationTime:      groupData.group.CreationTime,
			RetentionInDays:   groupData.group.RetentionInDays,
			MetricFilterCount: groupData.group.MetricFilterCount,
			StoredBytes:       groupData.group.StoredBytes,
			KmsKeyID:          groupData.group.KmsKeyID,
			LogGroupClass:     groupData.group.LogGroupClass,
		})
	}

	// Sort by name
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].LogGroupName < groups[j].LogGroupName
	})

	// Handle pagination
	startIdx := 0
	if req.NextToken != "" {
		for i, g := range groups {
			if g.LogGroupName == req.NextToken {
				startIdx = i

				break
			}
		}
	}

	endIdx := min(startIdx+limit, len(groups))
	result := groups[startIdx:endIdx]

	var nextToken string
	if endIdx < len(groups) {
		nextToken = groups[endIdx].LogGroupName
	}

	return &DescribeLogGroupsResponse{
		LogGroups: result,
		NextToken: nextToken,
	}, nil
}

// DescribeLogStreams describes log streams in a log group.
func (m *MemoryStorage) DescribeLogStreams(_ context.Context, req *DescribeLogStreamsRequest) (*DescribeLogStreamsResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	groupName := req.LogGroupName
	if groupName == "" {
		groupName = req.LogGroupIdentifier
	}

	groupData, exists := m.logGroups[groupName]
	if !exists {
		return nil, &LogsError{
			Code:    "ResourceNotFoundException",
			Message: fmt.Sprintf("The specified log group does not exist: %s", groupName),
		}
	}

	limit := defaultLimit
	if req.Limit != nil && *req.Limit > 0 {
		limit = min(int(*req.Limit), maxLimit)
	}

	var streams []LogStreamResponse

	for name, streamData := range groupData.streams {
		// Apply prefix filter
		if req.LogStreamNamePrefix != "" && !strings.HasPrefix(name, req.LogStreamNamePrefix) {
			continue
		}

		streams = append(streams, LogStreamResponse{
			LogStreamName:       streamData.stream.LogStreamName,
			CreationTime:        streamData.stream.CreationTime,
			FirstEventTimestamp: streamData.stream.FirstEventTimestamp,
			LastEventTimestamp:  streamData.stream.LastEventTimestamp,
			LastIngestionTime:   streamData.stream.LastIngestionTime,
			UploadSequenceToken: streamData.stream.UploadSequenceToken,
			Arn:                 streamData.stream.LogStreamARN,
			StoredBytes:         streamData.stream.StoredBytes,
		})
	}

	// Sort by name or last event time
	orderBy := req.OrderBy
	if orderBy == "" {
		orderBy = "LogStreamName"
	}

	descending := req.Descending != nil && *req.Descending

	switch orderBy {
	case "LastEventTime":
		sort.Slice(streams, func(i, j int) bool {
			iTime := int64(0)
			jTime := int64(0)

			if streams[i].LastEventTimestamp != nil {
				iTime = *streams[i].LastEventTimestamp
			}

			if streams[j].LastEventTimestamp != nil {
				jTime = *streams[j].LastEventTimestamp
			}

			if descending {
				return iTime > jTime
			}

			return iTime < jTime
		})
	default: // LogStreamName
		sort.Slice(streams, func(i, j int) bool {
			if descending {
				return streams[i].LogStreamName > streams[j].LogStreamName
			}

			return streams[i].LogStreamName < streams[j].LogStreamName
		})
	}

	// Handle pagination
	startIdx := 0
	if req.NextToken != "" {
		for i, s := range streams {
			if s.LogStreamName == req.NextToken {
				startIdx = i

				break
			}
		}
	}

	endIdx := min(startIdx+limit, len(streams))
	result := streams[startIdx:endIdx]

	var nextToken string
	if endIdx < len(streams) {
		nextToken = streams[endIdx].LogStreamName
	}

	return &DescribeLogStreamsResponse{
		LogStreams: result,
		NextToken:  nextToken,
	}, nil
}

// buildLogGroupARN builds an ARN for a log group.
func (m *MemoryStorage) buildLogGroupARN(name string) string {
	return fmt.Sprintf("arn:aws:logs:%s:%s:log-group:%s",
		defaultRegion, defaultAccountID, name)
}

// buildLogStreamARN builds an ARN for a log stream.
func (m *MemoryStorage) buildLogStreamARN(groupName, streamName string) string {
	return fmt.Sprintf("arn:aws:logs:%s:%s:log-group:%s:log-stream:%s",
		defaultRegion, defaultAccountID, groupName, streamName)
}

// filterEventsByTime filters events by time range.
func filterEventsByTime(events []*LogEvent, startTime, endTime *int64) []*LogEvent {
	var result []*LogEvent

	for _, event := range events {
		if startTime != nil && event.Timestamp < *startTime {
			continue
		}

		if endTime != nil && event.Timestamp > *endTime {
			continue
		}

		result = append(result, event)
	}

	return result
}

// sumEventBytes sums the bytes of all events.
func sumEventBytes(events []InputLogEvent) int64 {
	var total int64

	for _, event := range events {
		total += int64(len(event.Message))
	}

	return total
}
