package globalaccelerator

import (
	"testing"
)

func createTestAccelerators(t *testing.T, storage *MemoryStorage, count int) {
	t.Helper()

	ctx := t.Context()

	for range count {
		_, err := storage.CreateAccelerator(ctx, &CreateAcceleratorRequest{
			Name:             "test-accelerator",
			IdempotencyToken: "token",
		})
		if err != nil {
			t.Fatalf("failed to create accelerator: %v", err)
		}
	}
}

func TestListAccelerators_DefaultMaxResults(t *testing.T) {
	storage := NewMemoryStorage()
	createTestAccelerators(t, storage, 5)

	accelerators, nextToken, err := storage.ListAccelerators(t.Context(), 0, "")
	if err != nil {
		t.Fatalf("failed to list accelerators: %v", err)
	}

	if len(accelerators) != 5 {
		t.Errorf("expected 5 accelerators, got %d", len(accelerators))
	}

	if nextToken != "" {
		t.Errorf("expected empty nextToken when all results fit, got %s", nextToken)
	}
}

func TestListAccelerators_Pagination_Page1(t *testing.T) {
	storage := NewMemoryStorage()
	createTestAccelerators(t, storage, 5)

	accelerators, nextToken, err := storage.ListAccelerators(t.Context(), 2, "")
	if err != nil {
		t.Fatalf("failed to list accelerators page 1: %v", err)
	}

	if len(accelerators) != 2 {
		t.Errorf("page 1: expected 2 accelerators, got %d", len(accelerators))
	}

	if nextToken == "" {
		t.Fatal("page 1: expected non-empty nextToken")
	}
}

func TestListAccelerators_Pagination_AllPages(t *testing.T) {
	storage := NewMemoryStorage()
	createTestAccelerators(t, storage, 5)
	ctx := t.Context()

	// Collect all accelerators through pagination.
	var allAccelerators []*Accelerator

	nextToken := ""

	for {
		accelerators, token, err := storage.ListAccelerators(ctx, 2, nextToken)
		if err != nil {
			t.Fatalf("failed to list accelerators: %v", err)
		}

		allAccelerators = append(allAccelerators, accelerators...)

		if token == "" {
			break
		}

		nextToken = token
	}

	// Verify total count.
	if len(allAccelerators) != 5 {
		t.Errorf("expected 5 total accelerators, got %d", len(allAccelerators))
	}

	// Verify uniqueness.
	seen := make(map[string]bool)

	for _, acc := range allAccelerators {
		if seen[acc.AcceleratorArn] {
			t.Errorf("duplicate accelerator found: %s", acc.AcceleratorArn)
		}

		seen[acc.AcceleratorArn] = true
	}
}

func TestListAccelerators_EmptyStorage(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := t.Context()

	accelerators, nextToken, err := storage.ListAccelerators(ctx, 10, "")
	if err != nil {
		t.Fatalf("failed to list accelerators: %v", err)
	}

	if len(accelerators) != 0 {
		t.Errorf("expected 0 accelerators, got %d", len(accelerators))
	}

	if nextToken != "" {
		t.Errorf("expected empty nextToken, got %s", nextToken)
	}
}

func TestListAccelerators_InvalidNextToken(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := t.Context()

	// Create an accelerator.
	_, err := storage.CreateAccelerator(ctx, &CreateAcceleratorRequest{
		Name:             "test-accelerator",
		IdempotencyToken: "token",
	})
	if err != nil {
		t.Fatalf("failed to create accelerator: %v", err)
	}

	// Use an invalid nextToken.
	accelerators, nextToken, err := storage.ListAccelerators(ctx, 10, "invalid-token")
	if err != nil {
		t.Fatalf("failed to list accelerators: %v", err)
	}

	// With an invalid token, we should start from the beginning.
	if len(accelerators) != 1 {
		t.Errorf("expected 1 accelerator, got %d", len(accelerators))
	}

	if nextToken != "" {
		t.Errorf("expected empty nextToken, got %s", nextToken)
	}
}
