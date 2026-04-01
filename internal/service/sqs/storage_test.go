package sqs

import (
	"testing"
)

func TestMemoryStorage_ResolveQueueData_HostnameMismatch(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage("http://localhost:4566")

	ctx := t.Context()

	_, err := s.CreateQueue(ctx, "test-queue", nil)
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

	_, err := s.CreateQueue(ctx, "delete-test", nil)
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
