package dynamodb

import (
	"context"
	"testing"
)

//nolint:gocognit,cyclop,funlen // Test function exercises multiple SET arithmetic scenarios sequentially.
func TestSetArithmetic(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage("http://localhost:4566")
	ctx := context.Background()

	_, err := s.CreateTable(ctx, &CreateTableRequest{
		TableName: "test-arithmetic",
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

	t.Run("SET path = path + :val", func(t *testing.T) {
		t.Parallel()

		key := Item{"pk": {S: ptr("arith-add")}}

		// Insert initial item with Counter=5.
		_, err := s.PutItem(ctx, "test-arithmetic", Item{
			"pk":      {S: ptr("arith-add")},
			"Counter": {N: ptr("5")},
		}, false, ConditionInput{})
		if err != nil {
			t.Fatal(err)
		}

		// UpdateItem: SET Counter = Counter + :incr
		result, err := s.UpdateItem(ctx, "test-arithmetic", key,
			"SET Counter = Counter + :incr",
			nil,
			map[string]AttributeValue{":incr": {N: ptr("3")}},
			ReturnValuesAllNew,
			ConditionInput{},
		)
		if err != nil {
			t.Fatal(err)
		}

		if result["Counter"].N == nil || *result["Counter"].N != "8" {
			t.Errorf("Counter = %v, want 8", result["Counter"])
		}
	})

	t.Run("SET path = path - :val", func(t *testing.T) {
		t.Parallel()

		key := Item{"pk": {S: ptr("arith-sub")}}

		_, err := s.PutItem(ctx, "test-arithmetic", Item{
			"pk":      {S: ptr("arith-sub")},
			"Counter": {N: ptr("10")},
		}, false, ConditionInput{})
		if err != nil {
			t.Fatal(err)
		}

		result, err := s.UpdateItem(ctx, "test-arithmetic", key,
			"SET Counter = Counter - :decr",
			nil,
			map[string]AttributeValue{":decr": {N: ptr("4")}},
			ReturnValuesAllNew,
			ConditionInput{},
		)
		if err != nil {
			t.Fatal(err)
		}

		if result["Counter"].N == nil || *result["Counter"].N != "6" {
			t.Errorf("Counter = %v, want 6", result["Counter"])
		}
	})

	t.Run("SET path = if_not_exists(path, :default) + :incr (new item)", func(t *testing.T) {
		t.Parallel()

		key := Item{"pk": {S: ptr("arith-ifne-new")}}

		result, err := s.UpdateItem(ctx, "test-arithmetic", key,
			"SET #count = if_not_exists(#count, :zero) + :incr",
			map[string]string{"#count": "Counter"},
			map[string]AttributeValue{
				":zero": {N: ptr("0")},
				":incr": {N: ptr("1")},
			},
			ReturnValuesAllNew,
			ConditionInput{},
		)
		if err != nil {
			t.Fatal(err)
		}

		if result["Counter"].N == nil || *result["Counter"].N != "1" {
			t.Errorf("Counter = %v, want 1", result["Counter"])
		}

		// Second call should increment to 2.
		result, err = s.UpdateItem(ctx, "test-arithmetic", key,
			"SET #count = if_not_exists(#count, :zero) + :incr",
			map[string]string{"#count": "Counter"},
			map[string]AttributeValue{
				":zero": {N: ptr("0")},
				":incr": {N: ptr("1")},
			},
			ReturnValuesAllNew,
			ConditionInput{},
		)
		if err != nil {
			t.Fatal(err)
		}

		if result["Counter"].N == nil || *result["Counter"].N != "2" {
			t.Errorf("Counter = %v, want 2", result["Counter"])
		}
	})

	t.Run("SET with multiple assignments including if_not_exists + arithmetic", func(t *testing.T) {
		t.Parallel()

		key := Item{"pk": {S: ptr("arith-multi")}}

		result, err := s.UpdateItem(ctx, "test-arithmetic", key,
			"SET #count = if_not_exists(#count, :zero) + :incr, ExpiresAt = :exp",
			map[string]string{"#count": "Counter"},
			map[string]AttributeValue{
				":zero": {N: ptr("0")},
				":incr": {N: ptr("1")},
				":exp":  {N: ptr("9999")},
			},
			ReturnValuesAllNew,
			ConditionInput{},
		)
		if err != nil {
			t.Fatal(err)
		}

		if result["Counter"].N == nil || *result["Counter"].N != "1" {
			t.Errorf("Counter = %v, want 1", result["Counter"])
		}

		if result["ExpiresAt"].N == nil || *result["ExpiresAt"].N != "9999" {
			t.Errorf("ExpiresAt = %v, want 9999", result["ExpiresAt"])
		}
	})
}
