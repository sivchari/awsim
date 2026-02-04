//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func newDynamoDBClient(t *testing.T) *dynamodb.Client {
	t.Helper()

	cfg, err := config.LoadDefaultConfig(t.Context(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test", "test", "",
		)),
	)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	return dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestDynamoDB_CreateAndDeleteTable(t *testing.T) {
	client := newDynamoDBClient(t)
	ctx := t.Context()
	tableName := "test-table-create-delete"

	// Create table.
	createOutput, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("pk"),
				KeyType:       types.KeyTypeHash,
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("pk"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	if createOutput.TableDescription == nil {
		t.Fatal("table description is nil")
	}

	if *createOutput.TableDescription.TableName != tableName {
		t.Errorf("table name mismatch: got %s, want %s", *createOutput.TableDescription.TableName, tableName)
	}

	t.Logf("Created table: %s", *createOutput.TableDescription.TableName)

	// Delete table.
	_, err = client.DeleteTable(ctx, &dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		t.Fatalf("failed to delete table: %v", err)
	}
}

func TestDynamoDB_ListTables(t *testing.T) {
	client := newDynamoDBClient(t)
	ctx := t.Context()
	tableName := "test-table-list"

	// Create table.
	_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("pk"),
				KeyType:       types.KeyTypeHash,
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("pk"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteTable(ctx, &dynamodb.DeleteTableInput{
			TableName: aws.String(tableName),
		})
	})

	// List tables.
	listOutput, err := client.ListTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		t.Fatalf("failed to list tables: %v", err)
	}

	found := false

	for _, name := range listOutput.TableNames {
		if name == tableName {
			found = true

			break
		}
	}

	if !found {
		t.Errorf("table %s not found in list", tableName)
	}
}

func TestDynamoDB_DescribeTable(t *testing.T) {
	client := newDynamoDBClient(t)
	ctx := t.Context()
	tableName := "test-table-describe"

	// Create table.
	_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("pk"),
				KeyType:       types.KeyTypeHash,
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("pk"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteTable(ctx, &dynamodb.DeleteTableInput{
			TableName: aws.String(tableName),
		})
	})

	// Describe table.
	descOutput, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		t.Fatalf("failed to describe table: %v", err)
	}

	if descOutput.Table == nil {
		t.Fatal("table is nil")
	}

	if *descOutput.Table.TableName != tableName {
		t.Errorf("table name mismatch: got %s, want %s", *descOutput.Table.TableName, tableName)
	}

	if descOutput.Table.TableStatus != types.TableStatusActive {
		t.Errorf("table status mismatch: got %s, want ACTIVE", descOutput.Table.TableStatus)
	}
}

func TestDynamoDB_PutAndGetItem(t *testing.T) {
	client := newDynamoDBClient(t)
	ctx := t.Context()
	tableName := "test-table-put-get"

	// Create table.
	_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("pk"),
				KeyType:       types.KeyTypeHash,
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("pk"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteTable(ctx, &dynamodb.DeleteTableInput{
			TableName: aws.String(tableName),
		})
	})

	// Put item.
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item: map[string]types.AttributeValue{
			"pk":   &types.AttributeValueMemberS{Value: "test-id"},
			"name": &types.AttributeValueMemberS{Value: "Test Item"},
			"age":  &types.AttributeValueMemberN{Value: "25"},
		},
	})
	if err != nil {
		t.Fatalf("failed to put item: %v", err)
	}

	// Get item.
	getOutput, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "test-id"},
		},
	})
	if err != nil {
		t.Fatalf("failed to get item: %v", err)
	}

	if getOutput.Item == nil {
		t.Fatal("item is nil")
	}

	if nameAttr, ok := getOutput.Item["name"].(*types.AttributeValueMemberS); !ok || nameAttr.Value != "Test Item" {
		t.Errorf("name attribute mismatch")
	}

	if ageAttr, ok := getOutput.Item["age"].(*types.AttributeValueMemberN); !ok || ageAttr.Value != "25" {
		t.Errorf("age attribute mismatch")
	}
}

func TestDynamoDB_DeleteItem(t *testing.T) {
	client := newDynamoDBClient(t)
	ctx := t.Context()
	tableName := "test-table-delete-item"

	// Create table.
	_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("pk"),
				KeyType:       types.KeyTypeHash,
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("pk"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteTable(ctx, &dynamodb.DeleteTableInput{
			TableName: aws.String(tableName),
		})
	})

	// Put item.
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item: map[string]types.AttributeValue{
			"pk":   &types.AttributeValueMemberS{Value: "delete-me"},
			"name": &types.AttributeValueMemberS{Value: "To Delete"},
		},
	})
	if err != nil {
		t.Fatalf("failed to put item: %v", err)
	}

	// Delete item.
	_, err = client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "delete-me"},
		},
	})
	if err != nil {
		t.Fatalf("failed to delete item: %v", err)
	}

	// Verify item is deleted.
	getOutput, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "delete-me"},
		},
	})
	if err != nil {
		t.Fatalf("failed to get item: %v", err)
	}

	if getOutput.Item != nil {
		t.Errorf("item should be nil after deletion, got %v", getOutput.Item)
	}
}

func TestDynamoDB_UpdateItem(t *testing.T) {
	client := newDynamoDBClient(t)
	ctx := t.Context()
	tableName := "test-table-update-item"

	// Create table.
	_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("pk"),
				KeyType:       types.KeyTypeHash,
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("pk"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteTable(ctx, &dynamodb.DeleteTableInput{
			TableName: aws.String(tableName),
		})
	})

	// Put initial item.
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item: map[string]types.AttributeValue{
			"pk":   &types.AttributeValueMemberS{Value: "update-me"},
			"name": &types.AttributeValueMemberS{Value: "Original"},
		},
	})
	if err != nil {
		t.Fatalf("failed to put item: %v", err)
	}

	// Update item.
	updateOutput, err := client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "update-me"},
		},
		UpdateExpression: aws.String("SET #n = :name"),
		ExpressionAttributeNames: map[string]string{
			"#n": "name",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":name": &types.AttributeValueMemberS{Value: "Updated"},
		},
		ReturnValues: types.ReturnValueAllNew,
	})
	if err != nil {
		t.Fatalf("failed to update item: %v", err)
	}

	if nameAttr, ok := updateOutput.Attributes["name"].(*types.AttributeValueMemberS); !ok || nameAttr.Value != "Updated" {
		t.Errorf("name attribute mismatch after update")
	}

	// Verify item is updated.
	getOutput, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "update-me"},
		},
	})
	if err != nil {
		t.Fatalf("failed to get item: %v", err)
	}

	if nameAttr, ok := getOutput.Item["name"].(*types.AttributeValueMemberS); !ok || nameAttr.Value != "Updated" {
		t.Errorf("name attribute mismatch: expected Updated")
	}
}

func TestDynamoDB_Query(t *testing.T) {
	client := newDynamoDBClient(t)
	ctx := t.Context()
	tableName := "test-table-query"

	// Create table with sort key.
	_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("pk"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("sk"),
				KeyType:       types.KeyTypeRange,
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("pk"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("sk"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteTable(ctx, &dynamodb.DeleteTableInput{
			TableName: aws.String(tableName),
		})
	})

	// Put multiple items.
	items := []struct {
		pk   string
		sk   string
		data string
	}{
		{"user-1", "item-1", "data1"},
		{"user-1", "item-2", "data2"},
		{"user-1", "item-3", "data3"},
		{"user-2", "item-1", "data4"},
	}

	for _, item := range items {
		_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(tableName),
			Item: map[string]types.AttributeValue{
				"pk":   &types.AttributeValueMemberS{Value: item.pk},
				"sk":   &types.AttributeValueMemberS{Value: item.sk},
				"data": &types.AttributeValueMemberS{Value: item.data},
			},
		})
		if err != nil {
			t.Fatalf("failed to put item: %v", err)
		}
	}

	// Query items for user-1.
	queryOutput, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		KeyConditionExpression: aws.String("pk = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "user-1"},
		},
	})
	if err != nil {
		t.Fatalf("failed to query items: %v", err)
	}

	if queryOutput.Count != 3 {
		t.Errorf("expected 3 items, got %d", queryOutput.Count)
	}
}

func TestDynamoDB_Scan(t *testing.T) {
	client := newDynamoDBClient(t)
	ctx := t.Context()
	tableName := "test-table-scan"

	// Create table.
	_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("pk"),
				KeyType:       types.KeyTypeHash,
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("pk"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteTable(ctx, &dynamodb.DeleteTableInput{
			TableName: aws.String(tableName),
		})
	})

	// Put multiple items.
	for i := 0; i < 5; i++ {
		_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(tableName),
			Item: map[string]types.AttributeValue{
				"pk":   &types.AttributeValueMemberS{Value: "item-" + string(rune('a'+i))},
				"data": &types.AttributeValueMemberS{Value: "data"},
			},
		})
		if err != nil {
			t.Fatalf("failed to put item: %v", err)
		}
	}

	// Scan all items.
	scanOutput, err := client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		t.Fatalf("failed to scan items: %v", err)
	}

	if scanOutput.Count != 5 {
		t.Errorf("expected 5 items, got %d", scanOutput.Count)
	}
}

func TestDynamoDB_CompositeKey(t *testing.T) {
	client := newDynamoDBClient(t)
	ctx := t.Context()
	tableName := "test-table-composite-key"

	// Create table with composite key.
	_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("pk"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("sk"),
				KeyType:       types.KeyTypeRange,
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("pk"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("sk"),
				AttributeType: types.ScalarAttributeTypeN,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteTable(ctx, &dynamodb.DeleteTableInput{
			TableName: aws.String(tableName),
		})
	})

	// Put item with composite key.
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item: map[string]types.AttributeValue{
			"pk":   &types.AttributeValueMemberS{Value: "user-1"},
			"sk":   &types.AttributeValueMemberN{Value: "100"},
			"name": &types.AttributeValueMemberS{Value: "Test User"},
		},
	})
	if err != nil {
		t.Fatalf("failed to put item: %v", err)
	}

	// Get item with composite key.
	getOutput, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "user-1"},
			"sk": &types.AttributeValueMemberN{Value: "100"},
		},
	})
	if err != nil {
		t.Fatalf("failed to get item: %v", err)
	}

	if getOutput.Item == nil {
		t.Fatal("item is nil")
	}

	if nameAttr, ok := getOutput.Item["name"].(*types.AttributeValueMemberS); !ok || nameAttr.Value != "Test User" {
		t.Errorf("name attribute mismatch")
	}
}
