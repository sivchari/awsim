package sqs

import (
	"errors"
	"testing"
	"time"
)

func TestMemoryStorage_ResolveQueueData_HostnameMismatch(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage("http://localhost:4566")

	ctx := t.Context()

	_, err := s.CreateQueue(ctx, "test-queue", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		queueURL string
		wantErr  bool
	}{
		{
			name:     "exact match",
			queueURL: "http://localhost:4566/000000000000/test-queue",
		},
		{
			name:     "different hostname",
			queueURL: "http://kumo:4566/000000000000/test-queue",
		},
		{
			name:     "different scheme and hostname",
			queueURL: "https://sqs.us-east-1.amazonaws.com/000000000000/test-queue",
		},
		{
			name:     "non-existent queue",
			queueURL: "http://localhost:4566/000000000000/non-existent",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			msg, err := s.SendMessage(ctx, tt.queueURL, "hello", 0, nil, "", "")
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("SendMessage() error = %v", err)
			}

			if msg == nil {
				t.Fatal("expected message, got nil")
			}
		})
	}
}

func TestMemoryStorage_DeleteQueue_HostnameMismatch(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage("http://localhost:4566")

	ctx := t.Context()

	_, err := s.CreateQueue(ctx, "delete-test", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Delete using a different hostname.
	err = s.DeleteQueue(ctx, "http://kumo:4566/000000000000/delete-test")
	if err != nil {
		t.Fatalf("DeleteQueue() error = %v", err)
	}

	// Verify queue is gone.
	_, err = s.GetQueueURL(ctx, "delete-test")
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}

func TestMemoryStorage_TagsLifecycle(t *testing.T) {
	t.Parallel()

	const tagValue2 = "val2"

	s := NewMemoryStorage("http://localhost:4566")
	ctx := t.Context()

	_, err := s.CreateQueue(ctx, "tagged-queue", nil, map[string]string{"key1": "val1"})
	if err != nil {
		t.Fatal(err)
	}

	tags, err := s.ListQueueTags(ctx, "http://kumo:4566/000000000000/tagged-queue")
	if err != nil {
		t.Fatalf("ListQueueTags() error = %v", err)
	}

	if len(tags) != 1 || tags["key1"] != "val1" {
		t.Fatalf("unexpected tags after create: %#v", tags)
	}

	err = s.TagQueue(ctx, "http://kumo:4566/000000000000/tagged-queue", map[string]string{"key2": tagValue2, "key1": "updated"})
	if err != nil {
		t.Fatalf("TagQueue() error = %v", err)
	}

	tags, err = s.ListQueueTags(ctx, "http://localhost:4566/000000000000/tagged-queue")
	if err != nil {
		t.Fatalf("ListQueueTags() error = %v", err)
	}

	if len(tags) != 2 || tags["key1"] != "updated" || tags["key2"] != tagValue2 {
		t.Fatalf("unexpected tags after tag: %#v", tags)
	}

	err = s.UntagQueue(ctx, "http://localhost:4566/000000000000/tagged-queue", []string{"key1"})
	if err != nil {
		t.Fatalf("UntagQueue() error = %v", err)
	}

	tags, err = s.ListQueueTags(ctx, "http://localhost:4566/000000000000/tagged-queue")
	if err != nil {
		t.Fatalf("ListQueueTags() error = %v", err)
	}

	if len(tags) != 1 || tags["key2"] != tagValue2 {
		t.Fatalf("unexpected tags after untag: %#v", tags)
	}
}

func TestMemoryStorage_ChangeMessageVisibility(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage("http://localhost:4566")
	ctx := t.Context()
	queueURL := "http://localhost:4566/000000000000/vis-queue"

	_, err := s.CreateQueue(ctx, "vis-queue", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.SendMessage(ctx, queueURL, "test body", 0, nil, "", "")
	if err != nil {
		t.Fatal(err)
	}

	msgs, err := s.ReceiveMessage(ctx, queueURL, 1, 30, 0)
	if err != nil {
		t.Fatal(err)
	}

	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	handle := msgs[0].ReceiptHandle

	// extend visibility
	err = s.ChangeMessageVisibility(ctx, queueURL, handle, 60)
	if err != nil {
		t.Fatalf("ChangeMessageVisibility() error = %v", err)
	}

	if msgs[0].VisibleAt.Before(time.Now().Add(50 * time.Second)) {
		t.Error("expected visibility extended to ~60s from now")
	}

	// invalid receipt handle
	err = s.ChangeMessageVisibility(ctx, queueURL, "bad-handle", 10)
	if err == nil {
		t.Error("expected error for invalid receipt handle")
	}

	var qErr *QueueError
	if !errors.As(err, &qErr) || qErr.Code != "MessageNotInflight" {
		t.Errorf("expected MessageNotInflight error, got %v", err)
	}
}

func TestMemoryStorage_ChangeMessageVisibility_ZeroTimeout(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage("http://localhost:4566")
	ctx := t.Context()
	queueURL := "http://localhost:4566/000000000000/vis-zero-queue"

	_, err := s.CreateQueue(ctx, "vis-zero-queue", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.SendMessage(ctx, queueURL, "hello", 0, nil, "", "")
	if err != nil {
		t.Fatal(err)
	}

	msgs, err := s.ReceiveMessage(ctx, queueURL, 1, 30, 0)
	if err != nil {
		t.Fatal(err)
	}

	handle := msgs[0].ReceiptHandle

	// set visibility to 0 — should make message visible again
	err = s.ChangeMessageVisibility(ctx, queueURL, handle, 0)
	if err != nil {
		t.Fatalf("ChangeMessageVisibility(0) error = %v", err)
	}

	// message should be receivable again
	msgs, err = s.ReceiveMessage(ctx, queueURL, 1, 30, 0)
	if err != nil {
		t.Fatal(err)
	}

	if len(msgs) != 1 {
		t.Fatalf("expected message to be visible again, got %d messages", len(msgs))
	}
}
