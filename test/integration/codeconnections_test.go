//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/codeconnections"
	"github.com/aws/aws-sdk-go-v2/service/codeconnections/types"
)

func newCodeConnectionsClient(t *testing.T) *codeconnections.Client {
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

	return codeconnections.NewFromConfig(cfg, func(o *codeconnections.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestCodeConnections_CreateAndDeleteConnection(t *testing.T) {
	client := newCodeConnectionsClient(t)
	ctx := t.Context()

	// Create connection.
	createOutput, err := client.CreateConnection(ctx, &codeconnections.CreateConnectionInput{
		ConnectionName: aws.String("test-connection"),
		ProviderType:   types.ProviderTypeGithub,
	})
	if err != nil {
		t.Fatalf("failed to create connection: %v", err)
	}

	if createOutput.ConnectionArn == nil || *createOutput.ConnectionArn == "" {
		t.Fatal("connection ARN is empty")
	}

	connectionArn := *createOutput.ConnectionArn
	t.Logf("Created connection: %s", connectionArn)

	// Delete connection.
	_, err = client.DeleteConnection(ctx, &codeconnections.DeleteConnectionInput{
		ConnectionArn: aws.String(connectionArn),
	})
	if err != nil {
		t.Fatalf("failed to delete connection: %v", err)
	}

	t.Logf("Deleted connection: %s", connectionArn)
}

func TestCodeConnections_GetConnection(t *testing.T) {
	client := newCodeConnectionsClient(t)
	ctx := t.Context()

	// Create connection.
	createOutput, err := client.CreateConnection(ctx, &codeconnections.CreateConnectionInput{
		ConnectionName: aws.String("test-get-connection"),
		ProviderType:   types.ProviderTypeGithub,
	})
	if err != nil {
		t.Fatalf("failed to create connection: %v", err)
	}

	connectionArn := *createOutput.ConnectionArn

	t.Cleanup(func() {
		_, _ = client.DeleteConnection(ctx, &codeconnections.DeleteConnectionInput{
			ConnectionArn: aws.String(connectionArn),
		})
	})

	// Get connection.
	getOutput, err := client.GetConnection(ctx, &codeconnections.GetConnectionInput{
		ConnectionArn: aws.String(connectionArn),
	})
	if err != nil {
		t.Fatalf("failed to get connection: %v", err)
	}

	if getOutput.Connection == nil {
		t.Fatal("connection is nil")
	}

	if *getOutput.Connection.ConnectionArn != connectionArn {
		t.Errorf("connection ARN mismatch: got %s, want %s",
			*getOutput.Connection.ConnectionArn, connectionArn)
	}

	if *getOutput.Connection.ConnectionName != "test-get-connection" {
		t.Errorf("connection name mismatch: got %s, want test-get-connection",
			*getOutput.Connection.ConnectionName)
	}

	t.Logf("Got connection: %s", *getOutput.Connection.ConnectionName)
}

func TestCodeConnections_ListConnections(t *testing.T) {
	client := newCodeConnectionsClient(t)
	ctx := t.Context()

	// Create a few connections.
	var createdArns []string

	for i := 0; i < 3; i++ {
		output, err := client.CreateConnection(ctx, &codeconnections.CreateConnectionInput{
			ConnectionName: aws.String("test-list-connection-" + string(rune('a'+i))),
			ProviderType:   types.ProviderTypeGithub,
		})
		if err != nil {
			t.Fatalf("failed to create connection %d: %v", i, err)
		}

		createdArns = append(createdArns, *output.ConnectionArn)
	}

	t.Cleanup(func() {
		for _, arn := range createdArns {
			_, _ = client.DeleteConnection(ctx, &codeconnections.DeleteConnectionInput{
				ConnectionArn: aws.String(arn),
			})
		}
	})

	// List connections.
	listOutput, err := client.ListConnections(ctx, &codeconnections.ListConnectionsInput{
		MaxResults: 10,
	})
	if err != nil {
		t.Fatalf("failed to list connections: %v", err)
	}

	if len(listOutput.Connections) == 0 {
		t.Fatal("no connections returned")
	}

	t.Logf("Listed %d connections", len(listOutput.Connections))
}

func TestCodeConnections_CreateAndDeleteHost(t *testing.T) {
	client := newCodeConnectionsClient(t)
	ctx := t.Context()

	// Create host.
	createOutput, err := client.CreateHost(ctx, &codeconnections.CreateHostInput{
		Name:             aws.String("test-host"),
		ProviderType:     types.ProviderTypeGithubEnterpriseServer,
		ProviderEndpoint: aws.String("https://github.example.com"),
	})
	if err != nil {
		t.Fatalf("failed to create host: %v", err)
	}

	if createOutput.HostArn == nil || *createOutput.HostArn == "" {
		t.Fatal("host ARN is empty")
	}

	hostArn := *createOutput.HostArn
	t.Logf("Created host: %s", hostArn)

	// Delete host.
	_, err = client.DeleteHost(ctx, &codeconnections.DeleteHostInput{
		HostArn: aws.String(hostArn),
	})
	if err != nil {
		t.Fatalf("failed to delete host: %v", err)
	}

	t.Logf("Deleted host: %s", hostArn)
}

func TestCodeConnections_GetHost(t *testing.T) {
	client := newCodeConnectionsClient(t)
	ctx := t.Context()

	// Create host.
	createOutput, err := client.CreateHost(ctx, &codeconnections.CreateHostInput{
		Name:             aws.String("test-get-host"),
		ProviderType:     types.ProviderTypeGithubEnterpriseServer,
		ProviderEndpoint: aws.String("https://github.example.com"),
	})
	if err != nil {
		t.Fatalf("failed to create host: %v", err)
	}

	hostArn := *createOutput.HostArn

	t.Cleanup(func() {
		_, _ = client.DeleteHost(ctx, &codeconnections.DeleteHostInput{
			HostArn: aws.String(hostArn),
		})
	})

	// Get host.
	getOutput, err := client.GetHost(ctx, &codeconnections.GetHostInput{
		HostArn: aws.String(hostArn),
	})
	if err != nil {
		t.Fatalf("failed to get host: %v", err)
	}

	if getOutput.Name == nil || *getOutput.Name != "test-get-host" {
		t.Errorf("host name mismatch: got %v, want test-get-host", getOutput.Name)
	}

	if getOutput.ProviderEndpoint == nil || *getOutput.ProviderEndpoint != "https://github.example.com" {
		t.Errorf("provider endpoint mismatch: got %v, want https://github.example.com", getOutput.ProviderEndpoint)
	}

	t.Logf("Got host: %s", *getOutput.Name)
}

func TestCodeConnections_ListHosts(t *testing.T) {
	client := newCodeConnectionsClient(t)
	ctx := t.Context()

	// Create a host.
	createOutput, err := client.CreateHost(ctx, &codeconnections.CreateHostInput{
		Name:             aws.String("test-list-host"),
		ProviderType:     types.ProviderTypeGithubEnterpriseServer,
		ProviderEndpoint: aws.String("https://github.example.com"),
	})
	if err != nil {
		t.Fatalf("failed to create host: %v", err)
	}

	hostArn := *createOutput.HostArn

	t.Cleanup(func() {
		_, _ = client.DeleteHost(ctx, &codeconnections.DeleteHostInput{
			HostArn: aws.String(hostArn),
		})
	})

	// List hosts.
	listOutput, err := client.ListHosts(ctx, &codeconnections.ListHostsInput{
		MaxResults: 10,
	})
	if err != nil {
		t.Fatalf("failed to list hosts: %v", err)
	}

	if len(listOutput.Hosts) == 0 {
		t.Fatal("no hosts returned")
	}

	t.Logf("Listed %d hosts", len(listOutput.Hosts))
}

func TestCodeConnections_UpdateHost(t *testing.T) {
	client := newCodeConnectionsClient(t)
	ctx := t.Context()

	// Create host.
	createOutput, err := client.CreateHost(ctx, &codeconnections.CreateHostInput{
		Name:             aws.String("test-update-host"),
		ProviderType:     types.ProviderTypeGithubEnterpriseServer,
		ProviderEndpoint: aws.String("https://github.example.com"),
	})
	if err != nil {
		t.Fatalf("failed to create host: %v", err)
	}

	hostArn := *createOutput.HostArn

	t.Cleanup(func() {
		_, _ = client.DeleteHost(ctx, &codeconnections.DeleteHostInput{
			HostArn: aws.String(hostArn),
		})
	})

	// Update host.
	_, err = client.UpdateHost(ctx, &codeconnections.UpdateHostInput{
		HostArn:          aws.String(hostArn),
		ProviderEndpoint: aws.String("https://github-new.example.com"),
	})
	if err != nil {
		t.Fatalf("failed to update host: %v", err)
	}

	// Verify update.
	getOutput, err := client.GetHost(ctx, &codeconnections.GetHostInput{
		HostArn: aws.String(hostArn),
	})
	if err != nil {
		t.Fatalf("failed to get host: %v", err)
	}

	if *getOutput.ProviderEndpoint != "https://github-new.example.com" {
		t.Errorf("provider endpoint not updated: got %s, want https://github-new.example.com",
			*getOutput.ProviderEndpoint)
	}

	t.Logf("Updated host: %s", hostArn)
}

func TestCodeConnections_ConnectionNotFound(t *testing.T) {
	client := newCodeConnectionsClient(t)
	ctx := t.Context()

	// Try to get a non-existent connection.
	_, err := client.GetConnection(ctx, &codeconnections.GetConnectionInput{
		ConnectionArn: aws.String("arn:aws:codeconnections:us-east-1:000000000000:connection/non-existent"),
	})
	if err == nil {
		t.Fatal("expected error for non-existent connection")
	}
}

func TestCodeConnections_HostNotFound(t *testing.T) {
	client := newCodeConnectionsClient(t)
	ctx := t.Context()

	// Try to get a non-existent host.
	_, err := client.GetHost(ctx, &codeconnections.GetHostInput{
		HostArn: aws.String("arn:aws:codeconnections:us-east-1:000000000000:host/non-existent"),
	})
	if err == nil {
		t.Fatal("expected error for non-existent host")
	}
}

func TestCodeConnections_TagResource(t *testing.T) {
	client := newCodeConnectionsClient(t)
	ctx := t.Context()

	// Create connection.
	createOutput, err := client.CreateConnection(ctx, &codeconnections.CreateConnectionInput{
		ConnectionName: aws.String("test-tag-connection"),
		ProviderType:   types.ProviderTypeGithub,
	})
	if err != nil {
		t.Fatalf("failed to create connection: %v", err)
	}

	connectionArn := *createOutput.ConnectionArn

	t.Cleanup(func() {
		_, _ = client.DeleteConnection(ctx, &codeconnections.DeleteConnectionInput{
			ConnectionArn: aws.String(connectionArn),
		})
	})

	// Tag resource.
	_, err = client.TagResource(ctx, &codeconnections.TagResourceInput{
		ResourceArn: aws.String(connectionArn),
		Tags: []types.Tag{
			{Key: aws.String("Environment"), Value: aws.String("Test")},
			{Key: aws.String("Project"), Value: aws.String("awsim")},
		},
	})
	if err != nil {
		t.Fatalf("failed to tag resource: %v", err)
	}

	// List tags.
	listTagsOutput, err := client.ListTagsForResource(ctx, &codeconnections.ListTagsForResourceInput{
		ResourceArn: aws.String(connectionArn),
	})
	if err != nil {
		t.Fatalf("failed to list tags: %v", err)
	}

	if len(listTagsOutput.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(listTagsOutput.Tags))
	}

	t.Logf("Tagged resource with %d tags", len(listTagsOutput.Tags))
}

func TestCodeConnections_UntagResource(t *testing.T) {
	client := newCodeConnectionsClient(t)
	ctx := t.Context()

	// Create connection with tags.
	createOutput, err := client.CreateConnection(ctx, &codeconnections.CreateConnectionInput{
		ConnectionName: aws.String("test-untag-connection"),
		ProviderType:   types.ProviderTypeGithub,
		Tags: []types.Tag{
			{Key: aws.String("Environment"), Value: aws.String("Test")},
			{Key: aws.String("Project"), Value: aws.String("awsim")},
		},
	})
	if err != nil {
		t.Fatalf("failed to create connection: %v", err)
	}

	connectionArn := *createOutput.ConnectionArn

	t.Cleanup(func() {
		_, _ = client.DeleteConnection(ctx, &codeconnections.DeleteConnectionInput{
			ConnectionArn: aws.String(connectionArn),
		})
	})

	// Untag resource.
	_, err = client.UntagResource(ctx, &codeconnections.UntagResourceInput{
		ResourceArn: aws.String(connectionArn),
		TagKeys:     []string{"Environment"},
	})
	if err != nil {
		t.Fatalf("failed to untag resource: %v", err)
	}

	// List tags.
	listTagsOutput, err := client.ListTagsForResource(ctx, &codeconnections.ListTagsForResourceInput{
		ResourceArn: aws.String(connectionArn),
	})
	if err != nil {
		t.Fatalf("failed to list tags: %v", err)
	}

	if len(listTagsOutput.Tags) != 1 {
		t.Errorf("expected 1 tag, got %d", len(listTagsOutput.Tags))
	}

	t.Logf("Untagged resource, remaining tags: %d", len(listTagsOutput.Tags))
}
