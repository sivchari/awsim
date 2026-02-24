//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRoute53Client(t *testing.T) *route53.Client {
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

	return route53.NewFromConfig(cfg, func(o *route53.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestRoute53_CreateHostedZone(t *testing.T) {
	t.Parallel()

	client := newRoute53Client(t)
	ctx := t.Context()

	result, err := client.CreateHostedZone(ctx, &route53.CreateHostedZoneInput{
		Name:            aws.String("example.com"),
		CallerReference: aws.String("test-create-hosted-zone"),
		HostedZoneConfig: &types.HostedZoneConfig{
			Comment:     aws.String("Test hosted zone"),
			PrivateZone: false,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.HostedZone.Id)
	assert.Equal(t, "example.com.", *result.HostedZone.Name)
	assert.Equal(t, types.ChangeStatusInsync, result.ChangeInfo.Status)
	assert.NotEmpty(t, result.DelegationSet.NameServers)

	// Clean up.
	_, err = client.DeleteHostedZone(ctx, &route53.DeleteHostedZoneInput{
		Id: result.HostedZone.Id,
	})
	require.NoError(t, err)
}

func TestRoute53_GetHostedZone(t *testing.T) {
	t.Parallel()

	client := newRoute53Client(t)
	ctx := t.Context()

	// Create hosted zone first.
	createResult, err := client.CreateHostedZone(ctx, &route53.CreateHostedZoneInput{
		Name:            aws.String("get-test.example.com"),
		CallerReference: aws.String("test-get-hosted-zone"),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteHostedZone(ctx, &route53.DeleteHostedZoneInput{
			Id: createResult.HostedZone.Id,
		})
	})

	// Get hosted zone.
	getResult, err := client.GetHostedZone(ctx, &route53.GetHostedZoneInput{
		Id: createResult.HostedZone.Id,
	})
	require.NoError(t, err)
	assert.Equal(t, *createResult.HostedZone.Id, *getResult.HostedZone.Id)
	assert.Equal(t, *createResult.HostedZone.Name, *getResult.HostedZone.Name)
	assert.NotEmpty(t, getResult.DelegationSet.NameServers)
}

func TestRoute53_ListHostedZones(t *testing.T) {
	t.Parallel()

	client := newRoute53Client(t)
	ctx := t.Context()

	// Create hosted zone first.
	createResult, err := client.CreateHostedZone(ctx, &route53.CreateHostedZoneInput{
		Name:            aws.String("list-test.example.com"),
		CallerReference: aws.String("test-list-hosted-zones"),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteHostedZone(ctx, &route53.DeleteHostedZoneInput{
			Id: createResult.HostedZone.Id,
		})
	})

	// List hosted zones.
	listResult, err := client.ListHostedZones(ctx, &route53.ListHostedZonesInput{})
	require.NoError(t, err)
	require.NotNil(t, listResult)

	// Find our hosted zone.
	found := false
	for _, zone := range listResult.HostedZones {
		if *zone.Id == *createResult.HostedZone.Id {
			found = true
			break
		}
	}
	assert.True(t, found, "Hosted zone should be in list")
}

func TestRoute53_DeleteHostedZone(t *testing.T) {
	t.Parallel()

	client := newRoute53Client(t)
	ctx := t.Context()

	// Create hosted zone first.
	createResult, err := client.CreateHostedZone(ctx, &route53.CreateHostedZoneInput{
		Name:            aws.String("delete-test.example.com"),
		CallerReference: aws.String("test-delete-hosted-zone"),
	})
	require.NoError(t, err)

	// Delete hosted zone.
	deleteResult, err := client.DeleteHostedZone(ctx, &route53.DeleteHostedZoneInput{
		Id: createResult.HostedZone.Id,
	})
	require.NoError(t, err)
	assert.Equal(t, types.ChangeStatusInsync, deleteResult.ChangeInfo.Status)

	// Verify it's deleted.
	_, err = client.GetHostedZone(ctx, &route53.GetHostedZoneInput{
		Id: createResult.HostedZone.Id,
	})
	require.Error(t, err)
}

func TestRoute53_ChangeResourceRecordSets(t *testing.T) {
	t.Parallel()

	client := newRoute53Client(t)
	ctx := t.Context()

	// Create hosted zone first.
	createResult, err := client.CreateHostedZone(ctx, &route53.CreateHostedZoneInput{
		Name:            aws.String("records-test.example.com"),
		CallerReference: aws.String("test-change-record-sets"),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		// Delete records first.
		_, _ = client.ChangeResourceRecordSets(ctx, &route53.ChangeResourceRecordSetsInput{
			HostedZoneId: createResult.HostedZone.Id,
			ChangeBatch: &types.ChangeBatch{
				Changes: []types.Change{
					{
						Action: types.ChangeActionDelete,
						ResourceRecordSet: &types.ResourceRecordSet{
							Name: aws.String("www.records-test.example.com."),
							Type: types.RRTypeA,
							TTL:  aws.Int64(300),
							ResourceRecords: []types.ResourceRecord{
								{Value: aws.String("192.0.2.1")},
							},
						},
					},
				},
			},
		})
		_, _ = client.DeleteHostedZone(ctx, &route53.DeleteHostedZoneInput{
			Id: createResult.HostedZone.Id,
		})
	})

	// Create record set.
	changeResult, err := client.ChangeResourceRecordSets(ctx, &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: createResult.HostedZone.Id,
		ChangeBatch: &types.ChangeBatch{
			Comment: aws.String("Adding A record"),
			Changes: []types.Change{
				{
					Action: types.ChangeActionCreate,
					ResourceRecordSet: &types.ResourceRecordSet{
						Name: aws.String("www.records-test.example.com."),
						Type: types.RRTypeA,
						TTL:  aws.Int64(300),
						ResourceRecords: []types.ResourceRecord{
							{Value: aws.String("192.0.2.1")},
						},
					},
				},
			},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, types.ChangeStatusInsync, changeResult.ChangeInfo.Status)
}

func TestRoute53_ListResourceRecordSets(t *testing.T) {
	t.Parallel()

	client := newRoute53Client(t)
	ctx := t.Context()

	// Create hosted zone first.
	createResult, err := client.CreateHostedZone(ctx, &route53.CreateHostedZoneInput{
		Name:            aws.String("list-records-test.example.com"),
		CallerReference: aws.String("test-list-record-sets"),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		// Delete records first.
		_, _ = client.ChangeResourceRecordSets(ctx, &route53.ChangeResourceRecordSetsInput{
			HostedZoneId: createResult.HostedZone.Id,
			ChangeBatch: &types.ChangeBatch{
				Changes: []types.Change{
					{
						Action: types.ChangeActionDelete,
						ResourceRecordSet: &types.ResourceRecordSet{
							Name: aws.String("api.list-records-test.example.com."),
							Type: types.RRTypeCname,
							TTL:  aws.Int64(300),
							ResourceRecords: []types.ResourceRecord{
								{Value: aws.String("app.example.com")},
							},
						},
					},
				},
			},
		})
		_, _ = client.DeleteHostedZone(ctx, &route53.DeleteHostedZoneInput{
			Id: createResult.HostedZone.Id,
		})
	})

	// Create a record set.
	_, err = client.ChangeResourceRecordSets(ctx, &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: createResult.HostedZone.Id,
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{
				{
					Action: types.ChangeActionCreate,
					ResourceRecordSet: &types.ResourceRecordSet{
						Name: aws.String("api.list-records-test.example.com."),
						Type: types.RRTypeCname,
						TTL:  aws.Int64(300),
						ResourceRecords: []types.ResourceRecord{
							{Value: aws.String("app.example.com")},
						},
					},
				},
			},
		},
	})
	require.NoError(t, err)

	// List record sets.
	listResult, err := client.ListResourceRecordSets(ctx, &route53.ListResourceRecordSetsInput{
		HostedZoneId: createResult.HostedZone.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, listResult)

	// Find our record.
	found := false
	for _, record := range listResult.ResourceRecordSets {
		if *record.Name == "api.list-records-test.example.com." && record.Type == types.RRTypeCname {
			found = true
			break
		}
	}
	assert.True(t, found, "Record set should be in list")
}

func TestRoute53_UpsertResourceRecordSet(t *testing.T) {
	t.Parallel()

	client := newRoute53Client(t)
	ctx := t.Context()

	// Create hosted zone first.
	createResult, err := client.CreateHostedZone(ctx, &route53.CreateHostedZoneInput{
		Name:            aws.String("upsert-test.example.com"),
		CallerReference: aws.String("test-upsert-record-set"),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		// Delete records first.
		_, _ = client.ChangeResourceRecordSets(ctx, &route53.ChangeResourceRecordSetsInput{
			HostedZoneId: createResult.HostedZone.Id,
			ChangeBatch: &types.ChangeBatch{
				Changes: []types.Change{
					{
						Action: types.ChangeActionDelete,
						ResourceRecordSet: &types.ResourceRecordSet{
							Name: aws.String("test.upsert-test.example.com."),
							Type: types.RRTypeA,
							TTL:  aws.Int64(600),
							ResourceRecords: []types.ResourceRecord{
								{Value: aws.String("192.0.2.2")},
							},
						},
					},
				},
			},
		})
		_, _ = client.DeleteHostedZone(ctx, &route53.DeleteHostedZoneInput{
			Id: createResult.HostedZone.Id,
		})
	})

	// Upsert (create) record set.
	_, err = client.ChangeResourceRecordSets(ctx, &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: createResult.HostedZone.Id,
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{
				{
					Action: types.ChangeActionUpsert,
					ResourceRecordSet: &types.ResourceRecordSet{
						Name: aws.String("test.upsert-test.example.com."),
						Type: types.RRTypeA,
						TTL:  aws.Int64(300),
						ResourceRecords: []types.ResourceRecord{
							{Value: aws.String("192.0.2.1")},
						},
					},
				},
			},
		},
	})
	require.NoError(t, err)

	// Upsert (update) record set.
	_, err = client.ChangeResourceRecordSets(ctx, &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: createResult.HostedZone.Id,
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{
				{
					Action: types.ChangeActionUpsert,
					ResourceRecordSet: &types.ResourceRecordSet{
						Name: aws.String("test.upsert-test.example.com."),
						Type: types.RRTypeA,
						TTL:  aws.Int64(600),
						ResourceRecords: []types.ResourceRecord{
							{Value: aws.String("192.0.2.2")},
						},
					},
				},
			},
		},
	})
	require.NoError(t, err)

	// Verify update.
	listResult, err := client.ListResourceRecordSets(ctx, &route53.ListResourceRecordSetsInput{
		HostedZoneId: createResult.HostedZone.Id,
	})
	require.NoError(t, err)

	found := false
	for _, record := range listResult.ResourceRecordSets {
		if *record.Name == "test.upsert-test.example.com." && record.Type == types.RRTypeA {
			found = true
			assert.Equal(t, int64(600), *record.TTL)
			assert.Equal(t, "192.0.2.2", *record.ResourceRecords[0].Value)
			break
		}
	}
	assert.True(t, found, "Updated record set should be in list")
}
