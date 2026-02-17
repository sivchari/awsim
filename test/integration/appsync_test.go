//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/appsync"
	"github.com/aws/aws-sdk-go-v2/service/appsync/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppSync_CreateGraphqlApi(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := createAppSyncClient(t)

	// Create a GraphQL API.
	result, err := client.CreateGraphqlApi(ctx, &appsync.CreateGraphqlApiInput{
		Name:               aws.String("test-api"),
		AuthenticationType: types.AuthenticationTypeApiKey,
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.GraphqlApi)
	assert.Equal(t, "test-api", *result.GraphqlApi.Name)
	assert.NotEmpty(t, *result.GraphqlApi.ApiId)
	assert.NotEmpty(t, *result.GraphqlApi.Arn)
	assert.NotEmpty(t, result.GraphqlApi.Uris)

	// Clean up.
	_, err = client.DeleteGraphqlApi(ctx, &appsync.DeleteGraphqlApiInput{
		ApiId: result.GraphqlApi.ApiId,
	})
	require.NoError(t, err)
}

func TestAppSync_GetGraphqlApi(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := createAppSyncClient(t)

	// Create a GraphQL API first.
	createResult, err := client.CreateGraphqlApi(ctx, &appsync.CreateGraphqlApiInput{
		Name:               aws.String("get-test-api"),
		AuthenticationType: types.AuthenticationTypeAwsIam,
	})
	require.NoError(t, err)

	apiID := createResult.GraphqlApi.ApiId

	// Get the API.
	getResult, err := client.GetGraphqlApi(ctx, &appsync.GetGraphqlApiInput{
		ApiId: apiID,
	})
	require.NoError(t, err)
	assert.NotNil(t, getResult)
	assert.NotNil(t, getResult.GraphqlApi)
	assert.Equal(t, "get-test-api", *getResult.GraphqlApi.Name)
	assert.Equal(t, *apiID, *getResult.GraphqlApi.ApiId)

	// Clean up.
	_, err = client.DeleteGraphqlApi(ctx, &appsync.DeleteGraphqlApiInput{
		ApiId: apiID,
	})
	require.NoError(t, err)
}

func TestAppSync_ListGraphqlApis(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := createAppSyncClient(t)

	// Create some APIs.
	var apiIDs []*string

	for i := 0; i < 3; i++ {
		result, err := client.CreateGraphqlApi(ctx, &appsync.CreateGraphqlApiInput{
			Name:               aws.String("list-test-api"),
			AuthenticationType: types.AuthenticationTypeApiKey,
		})
		require.NoError(t, err)
		apiIDs = append(apiIDs, result.GraphqlApi.ApiId)
	}

	// List APIs.
	listResult, err := client.ListGraphqlApis(ctx, &appsync.ListGraphqlApisInput{})
	require.NoError(t, err)
	assert.NotNil(t, listResult)
	assert.GreaterOrEqual(t, len(listResult.GraphqlApis), 3)

	// Clean up.
	for _, apiID := range apiIDs {
		_, err = client.DeleteGraphqlApi(ctx, &appsync.DeleteGraphqlApiInput{
			ApiId: apiID,
		})
		require.NoError(t, err)
	}
}

func TestAppSync_DeleteGraphqlApi(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := createAppSyncClient(t)

	// Create an API.
	createResult, err := client.CreateGraphqlApi(ctx, &appsync.CreateGraphqlApiInput{
		Name:               aws.String("delete-test-api"),
		AuthenticationType: types.AuthenticationTypeApiKey,
	})
	require.NoError(t, err)

	apiID := createResult.GraphqlApi.ApiId

	// Delete the API.
	_, err = client.DeleteGraphqlApi(ctx, &appsync.DeleteGraphqlApiInput{
		ApiId: apiID,
	})
	require.NoError(t, err)

	// Try to get the deleted API - should fail.
	_, err = client.GetGraphqlApi(ctx, &appsync.GetGraphqlApiInput{
		ApiId: apiID,
	})
	assert.Error(t, err)
}

func TestAppSync_CreateDataSource(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := createAppSyncClient(t)

	// Create an API first.
	apiResult, err := client.CreateGraphqlApi(ctx, &appsync.CreateGraphqlApiInput{
		Name:               aws.String("datasource-test-api"),
		AuthenticationType: types.AuthenticationTypeApiKey,
	})
	require.NoError(t, err)

	apiID := apiResult.GraphqlApi.ApiId

	// Create a data source.
	dsResult, err := client.CreateDataSource(ctx, &appsync.CreateDataSourceInput{
		ApiId:       apiID,
		Name:        aws.String("test-datasource"),
		Type:        types.DataSourceTypeNone,
		Description: aws.String("Test data source"),
	})
	require.NoError(t, err)
	assert.NotNil(t, dsResult)
	assert.NotNil(t, dsResult.DataSource)
	assert.Equal(t, "test-datasource", *dsResult.DataSource.Name)
	assert.NotEmpty(t, *dsResult.DataSource.DataSourceArn)

	// Clean up.
	_, err = client.DeleteGraphqlApi(ctx, &appsync.DeleteGraphqlApiInput{
		ApiId: apiID,
	})
	require.NoError(t, err)
}

func TestAppSync_CreateResolver(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := createAppSyncClient(t)

	// Create an API first.
	apiResult, err := client.CreateGraphqlApi(ctx, &appsync.CreateGraphqlApiInput{
		Name:               aws.String("resolver-test-api"),
		AuthenticationType: types.AuthenticationTypeApiKey,
	})
	require.NoError(t, err)

	apiID := apiResult.GraphqlApi.ApiId

	// Create a data source first.
	_, err = client.CreateDataSource(ctx, &appsync.CreateDataSourceInput{
		ApiId: apiID,
		Name:  aws.String("resolver-datasource"),
		Type:  types.DataSourceTypeNone,
	})
	require.NoError(t, err)

	// Create a resolver.
	resolverResult, err := client.CreateResolver(ctx, &appsync.CreateResolverInput{
		ApiId:          apiID,
		TypeName:       aws.String("Query"),
		FieldName:      aws.String("getItem"),
		DataSourceName: aws.String("resolver-datasource"),
	})
	require.NoError(t, err)
	assert.NotNil(t, resolverResult)
	assert.NotNil(t, resolverResult.Resolver)
	assert.Equal(t, "Query", *resolverResult.Resolver.TypeName)
	assert.Equal(t, "getItem", *resolverResult.Resolver.FieldName)
	assert.NotEmpty(t, *resolverResult.Resolver.ResolverArn)

	// Clean up.
	_, err = client.DeleteGraphqlApi(ctx, &appsync.DeleteGraphqlApiInput{
		ApiId: apiID,
	})
	require.NoError(t, err)
}

func TestAppSync_StartSchemaCreation(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := createAppSyncClient(t)

	// Create an API first.
	apiResult, err := client.CreateGraphqlApi(ctx, &appsync.CreateGraphqlApiInput{
		Name:               aws.String("schema-test-api"),
		AuthenticationType: types.AuthenticationTypeApiKey,
	})
	require.NoError(t, err)

	apiID := apiResult.GraphqlApi.ApiId

	// Start schema creation.
	schema := []byte(`
		type Query {
			getItem(id: ID!): Item
		}
		type Item {
			id: ID!
			name: String
		}
	`)

	schemaResult, err := client.StartSchemaCreation(ctx, &appsync.StartSchemaCreationInput{
		ApiId:      apiID,
		Definition: schema,
	})
	require.NoError(t, err)
	assert.NotNil(t, schemaResult)
	// Status should be SUCCESS or PROCESSING.
	assert.NotEmpty(t, schemaResult.Status)

	// Clean up.
	_, err = client.DeleteGraphqlApi(ctx, &appsync.DeleteGraphqlApiInput{
		ApiId: apiID,
	})
	require.NoError(t, err)
}

func createAppSyncClient(t *testing.T) *appsync.Client {
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

	return appsync.NewFromConfig(cfg, func(o *appsync.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566/appsync")
	})
}
