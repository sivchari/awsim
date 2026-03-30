//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/sivchari/golden"
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
		t.Fatal(err)
	}
	golden.New(t, golden.WithIgnoreFields("TableArn", "TableId", "CreationDateTime", "ResultMetadata")).Assert(t.Name()+"_create", createOutput)

	// Delete table.
	_, err = client.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{
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
		_, _ = client.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{
			TableName: aws.String(tableName),
		})
	})

	// List tables - dynamic list, skip golden test.
	_, err = client.ListTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		t.Fatal(err)
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
		_, _ = client.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{
			TableName: aws.String(tableName),
		})
	})

	// Describe table.
	descOutput, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		t.Fatal(err)
	}
	golden.New(t, golden.WithIgnoreFields("TableArn", "TableId", "CreationDateTime", "TableSizeBytes", "ItemCount", "ResultMetadata")).Assert(t.Name(), descOutput)
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
		_, _ = client.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{
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
		t.Fatal(err)
	}

	// Get item.
	getOutput, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "test-id"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	golden.New(t, golden.WithIgnoreFields("ResultMetadata")).Assert(t.Name(), getOutput)
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
		_, _ = client.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{
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
		t.Fatal(err)
	}

	// Verify item is deleted.
	getOutput, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "delete-me"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	golden.New(t, golden.WithIgnoreFields("ResultMetadata")).Assert(t.Name()+"_get_after_delete", getOutput)
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
		_, _ = client.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{
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
		t.Fatal(err)
	}
	golden.New(t, golden.WithIgnoreFields("ResultMetadata")).Assert(t.Name()+"_update", updateOutput)

	// Verify item is updated.
	getOutput, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "update-me"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	golden.New(t, golden.WithIgnoreFields("ResultMetadata")).Assert(t.Name()+"_get_after_update", getOutput)
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
		_, _ = client.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{
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
		t.Fatal(err)
	}
	golden.New(t, golden.WithIgnoreFields("ResultMetadata")).Assert(t.Name(), queryOutput)
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
		_, _ = client.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{
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
		t.Fatal(err)
	}
	golden.New(t, golden.WithIgnoreFields("ResultMetadata")).Assert(t.Name(), scanOutput)
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
		_, _ = client.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{
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
		t.Fatal(err)
	}
	golden.New(t, golden.WithIgnoreFields("ResultMetadata")).Assert(t.Name(), getOutput)
}

func TestDynamoDB_UpdateTimeToLive(t *testing.T) {
	client := newDynamoDBClient(t)
	ctx := t.Context()
	tableName := "test-table-ttl"

	// Create table.
	_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("pk"), KeyType: types.KeyTypeHash},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("pk"), AttributeType: types.ScalarAttributeTypeS},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{
			TableName: aws.String(tableName),
		})
	})

	// Enable TTL.
	updateOutput, err := client.UpdateTimeToLive(ctx, &dynamodb.UpdateTimeToLiveInput{
		TableName: aws.String(tableName),
		TimeToLiveSpecification: &types.TimeToLiveSpecification{
			AttributeName: aws.String("ttl"),
			Enabled:       aws.Bool(true),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	golden.New(t, golden.WithIgnoreFields("ResultMetadata")).Assert(t.Name()+"_update", updateOutput)

	// Describe TTL.
	describeOutput, err := client.DescribeTimeToLive(ctx, &dynamodb.DescribeTimeToLiveInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		t.Fatal(err)
	}
	golden.New(t, golden.WithIgnoreFields("ResultMetadata")).Assert(t.Name()+"_describe", describeOutput)
}
