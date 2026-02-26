package globalaccelerator

import (
	"testing"
)

func TestListAccelerators_Pagination(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := t.Context()

	// Create 5 accelerators.
	for i := 0; i < 5; i++ {
		_, err := storage.CreateAccelerator(ctx, &CreateAcceleratorRequest{
			Name:             "test-accelerator",
			IdempotencyToken: "token",
		})
		if err != nil {
			t.Fatalf("failed to create accelerator: %v", err)
		}
	}

	// Test: list all with default maxResults.
	accelerators, nextToken, err := storage.ListAccelerators(ctx, 0, "")
	if err != nil {
		t.Fatalf("failed to list accelerators: %v", err)
	}

	if len(accelerators) != 5 {
		t.Errorf("expected 5 accelerators, got %d", len(accelerators))
	}

	if nextToken != "" {
		t.Errorf("expected empty nextToken when all results fit, got %s", nextToken)
	}

	// Test: paginate with maxResults=2.
	accelerators1, nextToken1, err := storage.ListAccelerators(ctx, 2, "")
	if err != nil {
		t.Fatalf("failed to list accelerators page 1: %v", err)
	}

	if len(accelerators1) != 2 {
		t.Errorf("page 1: expected 2 accelerators, got %d", len(accelerators1))
	}

	if nextToken1 == "" {
		t.Error("page 1: expected non-empty nextToken")
	}

	// Test: second page.
	accelerators2, nextToken2, err := storage.ListAccelerators(ctx, 2, nextToken1)
	if err != nil {
		t.Fatalf("failed to list accelerators page 2: %v", err)
	}

	if len(accelerators2) != 2 {
		t.Errorf("page 2: expected 2 accelerators, got %d", len(accelerators2))
	}

	if nextToken2 == "" {
		t.Error("page 2: expected non-empty nextToken")
	}

	// Verify page 1 and page 2 are different.
	for _, a1 := range accelerators1 {
		for _, a2 := range accelerators2 {
			if a1.AcceleratorArn == a2.AcceleratorArn {
				t.Errorf("duplicate accelerator found: %s", a1.AcceleratorArn)
			}
		}
	}

	// Test: third (last) page.
	accelerators3, nextToken3, err := storage.ListAccelerators(ctx, 2, nextToken2)
	if err != nil {
		t.Fatalf("failed to list accelerators page 3: %v", err)
	}

	if len(accelerators3) != 1 {
		t.Errorf("page 3: expected 1 accelerator, got %d", len(accelerators3))
	}

	if nextToken3 != "" {
		t.Errorf("page 3: expected empty nextToken, got %s", nextToken3)
	}

	// Verify we got all unique accelerators.
	seen := make(map[string]bool)
	for _, acc := range accelerators1 {
		seen[acc.AcceleratorArn] = true
	}
	for _, acc := range accelerators2 {
		seen[acc.AcceleratorArn] = true
	}
	for _, acc := range accelerators3 {
		seen[acc.AcceleratorArn] = true
	}

	if len(seen) != 5 {
		t.Errorf("expected 5 unique accelerators across all pages, got %d", len(seen))
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
