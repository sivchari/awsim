package dynamodb

import (
	"context"
	"errors"
	"testing"
)

func ptr[T any](v T) *T { return &v }

//nolint:cyclop // Test function exercises multiple storage operations sequentially.
func TestConditionThroughStorage(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage("http://localhost:4566")
	ctx := context.Background()

	_, err := s.CreateTable(ctx, &CreateTableRequest{
		TableName: "test",
		KeySchema: []KeySchemaElement{
			{AttributeName: "pk", KeyType: "HASH"},
		},
		AttributeDefinitions: []AttributeDefinition{
			{AttributeName: "pk", AttributeType: "S"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// PutItem with attribute_not_exists should succeed on new item.
	_, err = s.PutItem(ctx, "test", Item{
		"pk":      {S: ptr("1")},
		"version": {N: ptr("1")},
		"status":  {S: ptr("active")},
	}, false, ConditionInput{Expression: "attribute_not_exists(pk)"})
	if err != nil {
		t.Fatalf("first put should succeed: %v", err)
	}

	// PutItem with attribute_not_exists should fail on existing item.
	_, err = s.PutItem(ctx, "test", Item{
		"pk":      {S: ptr("1")},
		"version": {N: ptr("99")},
	}, false, ConditionInput{Expression: "attribute_not_exists(pk)"})
	if err == nil {
		t.Fatal("second put should fail")
	}

	var tErr *TableError
	if !errors.As(err, &tErr) || tErr.Code != ErrCodeConditionalCheckFailed {
		t.Fatalf("expected ConditionalCheckFailedException, got: %v", err)
	}

	// Verify original item preserved.
	item, err := s.GetItem(ctx, "test", Item{"pk": {S: ptr("1")}})
	if err != nil {
		t.Fatal(err)
	}

	if item["version"].N == nil || *item["version"].N != "1" {
		t.Fatalf("version should be 1, got: %v", item["version"])
	}

	// UpdateItem with version check (optimistic locking).
	_, err = s.UpdateItem(ctx, "test", Item{"pk": {S: ptr("1")}},
		"SET version = :new", nil,
		map[string]AttributeValue{
			":cur": {N: ptr("1")},
			":new": {N: ptr("2")},
		},
		ReturnValuesAllNew,
		ConditionInput{
			Expression: "version = :cur",
			ExprValues: map[string]AttributeValue{":cur": {N: ptr("1")}},
		},
	)
	if err != nil {
		t.Fatalf("update with correct version should succeed: %v", err)
	}

	// UpdateItem with stale version should fail.
	_, err = s.UpdateItem(ctx, "test", Item{"pk": {S: ptr("1")}},
		"SET version = :new", nil,
		map[string]AttributeValue{
			":cur": {N: ptr("1")},
			":new": {N: ptr("3")},
		},
		"",
		ConditionInput{
			Expression: "version = :cur",
			ExprValues: map[string]AttributeValue{":cur": {N: ptr("1")}},
		},
	)
	if err == nil {
		t.Fatal("update with stale version should fail")
	}

	// Comparison operators: version >= 2 should pass.
	_, err = s.UpdateItem(ctx, "test", Item{"pk": {S: ptr("1")}},
		"SET version = :new", nil,
		map[string]AttributeValue{
			":new": {N: ptr("3")},
		},
		ReturnValuesAllNew,
		ConditionInput{
			Expression: "version >= :min",
			ExprValues: map[string]AttributeValue{":min": {N: ptr("2")}},
		},
	)
	if err != nil {
		t.Fatalf("update with version >= 2 should succeed: %v", err)
	}

	// Comparison: version < 2 should fail (version is now 3).
	_, err = s.UpdateItem(ctx, "test", Item{"pk": {S: ptr("1")}},
		"SET version = :new", nil,
		map[string]AttributeValue{":new": {N: ptr("4")}},
		"",
		ConditionInput{
			Expression: "version < :max",
			ExprValues: map[string]AttributeValue{":max": {N: ptr("2")}},
		},
	)
	if err == nil {
		t.Fatal("update with version < 2 should fail (version=3)")
	}

	// String comparison: status = "active" AND version > 1.
	_, err = s.UpdateItem(ctx, "test", Item{"pk": {S: ptr("1")}},
		"SET #s = :new_status", map[string]string{"#s": "status"},
		map[string]AttributeValue{":new_status": {S: ptr("done")}},
		ReturnValuesAllNew,
		ConditionInput{
			Expression: "#s = :expected AND version > :min_ver",
			ExprNames:  map[string]string{"#s": "status"},
			ExprValues: map[string]AttributeValue{
				":expected": {S: ptr("active")},
				":min_ver":  {N: ptr("1")},
			},
		},
	)
	if err != nil {
		t.Fatalf("compound condition should succeed: %v", err)
	}

	// DeleteItem with wrong condition should fail.
	_, err = s.DeleteItem(ctx, "test", Item{"pk": {S: ptr("1")}}, false, ConditionInput{
		Expression: "status = :expected",
		ExprValues: map[string]AttributeValue{":expected": {S: ptr("active")}},
	})
	if err == nil {
		t.Fatal("delete with wrong status should fail (status=done)")
	}

	// DeleteItem with correct condition should succeed.
	_, err = s.DeleteItem(ctx, "test", Item{"pk": {S: ptr("1")}}, false, ConditionInput{
		Expression: "status = :expected",
		ExprValues: map[string]AttributeValue{":expected": {S: ptr("done")}},
	})
	if err != nil {
		t.Fatalf("delete with correct status should succeed: %v", err)
	}

	// Verify deleted.
	item, err = s.GetItem(ctx, "test", Item{"pk": {S: ptr("1")}})
	if err != nil {
		t.Fatal(err)
	}

	if item != nil {
		t.Fatalf("item should be deleted, got: %v", item)
	}
}

//nolint:funlen // Table-driven test with comprehensive condition expression coverage.
func TestEvaluateCondition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		item    Item
		cond    ConditionInput
		want    bool
		wantErr bool
	}{
		{
			name: "empty expression returns true",
			item: Item{"pk": {S: ptr("1")}},
			cond: ConditionInput{},
			want: true,
		},
		{
			name: "attribute_exists succeeds when attribute present",
			item: Item{"pk": {S: ptr("1")}, "name": {S: ptr("Alice")}},
			cond: ConditionInput{
				Expression: "attribute_exists(name)",
			},
			want: true,
		},
		{
			name: "attribute_exists fails when attribute missing",
			item: Item{"pk": {S: ptr("1")}},
			cond: ConditionInput{
				Expression: "attribute_exists(name)",
			},
			want: false,
		},
		{
			name: "attribute_not_exists succeeds when attribute missing",
			item: Item{"pk": {S: ptr("1")}},
			cond: ConditionInput{
				Expression: "attribute_not_exists(name)",
			},
			want: true,
		},
		{
			name: "attribute_not_exists on nil item (new item)",
			item: nil,
			cond: ConditionInput{
				Expression: "attribute_not_exists(pk)",
			},
			want: true,
		},
		{
			name: "attribute_not_exists fails when attribute present",
			item: Item{"pk": {S: ptr("1")}, "name": {S: ptr("Alice")}},
			cond: ConditionInput{
				Expression: "attribute_not_exists(name)",
			},
			want: false,
		},
		{
			name: "string equality",
			item: Item{"pk": {S: ptr("1")}, "status": {S: ptr("active")}},
			cond: ConditionInput{
				Expression: "status = :val",
				ExprValues: map[string]AttributeValue{
					":val": {S: ptr("active")},
				},
			},
			want: true,
		},
		{
			name: "string inequality",
			item: Item{"pk": {S: ptr("1")}, "status": {S: ptr("active")}},
			cond: ConditionInput{
				Expression: "status <> :val",
				ExprValues: map[string]AttributeValue{
					":val": {S: ptr("inactive")},
				},
			},
			want: true,
		},
		{
			name: "number comparison less than",
			item: Item{"pk": {S: ptr("1")}, "age": {N: ptr("25")}},
			cond: ConditionInput{
				Expression: "age < :val",
				ExprValues: map[string]AttributeValue{
					":val": {N: ptr("30")},
				},
			},
			want: true,
		},
		{
			name: "number comparison greater equal",
			item: Item{"pk": {S: ptr("1")}, "age": {N: ptr("25")}},
			cond: ConditionInput{
				Expression: "age >= :val",
				ExprValues: map[string]AttributeValue{
					":val": {N: ptr("25")},
				},
			},
			want: true,
		},
		{
			name: "AND both true",
			item: Item{"pk": {S: ptr("1")}, "status": {S: ptr("active")}, "age": {N: ptr("25")}},
			cond: ConditionInput{
				Expression: "status = :status AND age >= :age",
				ExprValues: map[string]AttributeValue{
					":status": {S: ptr("active")},
					":age":    {N: ptr("20")},
				},
			},
			want: true,
		},
		{
			name: "AND left false",
			item: Item{"pk": {S: ptr("1")}, "status": {S: ptr("inactive")}, "age": {N: ptr("25")}},
			cond: ConditionInput{
				Expression: "status = :status AND age >= :age",
				ExprValues: map[string]AttributeValue{
					":status": {S: ptr("active")},
					":age":    {N: ptr("20")},
				},
			},
			want: false,
		},
		{
			name: "OR one true",
			item: Item{"pk": {S: ptr("1")}, "status": {S: ptr("inactive")}},
			cond: ConditionInput{
				Expression: "status = :s1 OR status = :s2",
				ExprValues: map[string]AttributeValue{
					":s1": {S: ptr("active")},
					":s2": {S: ptr("inactive")},
				},
			},
			want: true,
		},
		{
			name: "NOT expression",
			item: Item{"pk": {S: ptr("1")}, "status": {S: ptr("active")}},
			cond: ConditionInput{
				Expression: "NOT status = :val",
				ExprValues: map[string]AttributeValue{
					":val": {S: ptr("inactive")},
				},
			},
			want: true,
		},
		{
			name: "expression attribute names",
			item: Item{"pk": {S: ptr("1")}, "status": {S: ptr("active")}},
			cond: ConditionInput{
				Expression: "#s = :val",
				ExprNames:  map[string]string{"#s": "status"},
				ExprValues: map[string]AttributeValue{
					":val": {S: ptr("active")},
				},
			},
			want: true,
		},
		{
			name: "begins_with true",
			item: Item{"pk": {S: ptr("1")}, "email": {S: ptr("alice@example.com")}},
			cond: ConditionInput{
				Expression: "begins_with(email, :prefix)",
				ExprValues: map[string]AttributeValue{
					":prefix": {S: ptr("alice@")},
				},
			},
			want: true,
		},
		{
			name: "begins_with false",
			item: Item{"pk": {S: ptr("1")}, "email": {S: ptr("bob@example.com")}},
			cond: ConditionInput{
				Expression: "begins_with(email, :prefix)",
				ExprValues: map[string]AttributeValue{
					":prefix": {S: ptr("alice@")},
				},
			},
			want: false,
		},
		{
			name: "contains string",
			item: Item{"pk": {S: ptr("1")}, "name": {S: ptr("Alice Smith")}},
			cond: ConditionInput{
				Expression: "contains(name, :sub)",
				ExprValues: map[string]AttributeValue{
					":sub": {S: ptr("Smith")},
				},
			},
			want: true,
		},
		{
			name: "contains string set",
			item: Item{"pk": {S: ptr("1")}, "tags": {SS: []string{"golang", "rust", "python"}}},
			cond: ConditionInput{
				Expression: "contains(tags, :tag)",
				ExprValues: map[string]AttributeValue{
					":tag": {S: ptr("rust")},
				},
			},
			want: true,
		},
		{
			name: "contains string set missing",
			item: Item{"pk": {S: ptr("1")}, "tags": {SS: []string{"golang", "rust"}}},
			cond: ConditionInput{
				Expression: "contains(tags, :tag)",
				ExprValues: map[string]AttributeValue{
					":tag": {S: ptr("java")},
				},
			},
			want: false,
		},
		{
			name: "size comparison",
			item: Item{"pk": {S: ptr("1")}, "items": {L: []AttributeValue{{S: ptr("a")}, {S: ptr("b")}, {S: ptr("c")}}}},
			cond: ConditionInput{
				Expression: "size(items) > :min",
				ExprValues: map[string]AttributeValue{
					":min": {N: ptr("2")},
				},
			},
			want: true,
		},
		{
			name: "size comparison fails",
			item: Item{"pk": {S: ptr("1")}, "items": {L: []AttributeValue{{S: ptr("a")}}}},
			cond: ConditionInput{
				Expression: "size(items) > :min",
				ExprValues: map[string]AttributeValue{
					":min": {N: ptr("2")},
				},
			},
			want: false,
		},
		{
			name: "parenthesized expression",
			item: Item{"pk": {S: ptr("1")}, "a": {N: ptr("1")}, "b": {N: ptr("2")}},
			cond: ConditionInput{
				Expression: "(a = :v1) AND (b = :v2)",
				ExprValues: map[string]AttributeValue{
					":v1": {N: ptr("1")},
					":v2": {N: ptr("2")},
				},
			},
			want: true,
		},
		{
			name: "idempotency pattern: attribute_not_exists on pk",
			item: Item{"pk": {S: ptr("existing-id")}, "data": {S: ptr("value")}},
			cond: ConditionInput{
				Expression: "attribute_not_exists(pk)",
			},
			want: false,
		},
		{
			name: "nested path attribute_exists",
			item: Item{"pk": {S: ptr("1")}, "meta": {M: map[string]AttributeValue{"version": {N: ptr("1")}}}},
			cond: ConditionInput{
				Expression: "attribute_exists(meta.version)",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := evaluateCondition(tt.item, tt.cond)
			if (err != nil) != tt.wantErr {
				t.Errorf("evaluateCondition() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if got != tt.want {
				t.Errorf("evaluateCondition() = %v, want %v", got, tt.want)
			}
		})
	}
}
