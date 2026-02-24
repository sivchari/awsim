//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newOrganizationsClient(t *testing.T) *organizations.Client {
	t.Helper()

	cfg, err := config.LoadDefaultConfig(t.Context(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test", "test", "",
		)),
	)
	require.NoError(t, err)

	return organizations.NewFromConfig(cfg, func(o *organizations.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

// ensureNoOrganization deletes any existing organization to ensure clean test state.
func ensureNoOrganization(t *testing.T, client *organizations.Client) {
	t.Helper()

	ctx := t.Context()

	// Check if organization exists
	_, err := client.DescribeOrganization(ctx, &organizations.DescribeOrganizationInput{})
	if err != nil {
		// No organization exists, we're good
		return
	}

	// Organization exists, try to delete it
	// First, list and handle any member accounts (except management account)
	listOutput, _ := client.ListAccounts(ctx, &organizations.ListAccountsInput{})
	if listOutput != nil && len(listOutput.Accounts) > 1 {
		// There are member accounts, which would prevent deletion
		// For test purposes, we skip deletion if there are member accounts
		// The test will need to handle this case
		return
	}

	// Delete the organization
	_, _ = client.DeleteOrganization(ctx, &organizations.DeleteOrganizationInput{})
}

func TestOrganizations_CreateOrganization(t *testing.T) {
	client := newOrganizationsClient(t)
	ctx := t.Context()

	// Ensure clean state
	ensureNoOrganization(t, client)

	// Create organization
	createOutput, err := client.CreateOrganization(ctx, &organizations.CreateOrganizationInput{
		FeatureSet: types.OrganizationFeatureSetAll,
	})
	require.NoError(t, err)
	require.NotNil(t, createOutput.Organization)
	assert.NotEmpty(t, *createOutput.Organization.Id)
	assert.NotEmpty(t, *createOutput.Organization.Arn)
	assert.Equal(t, types.OrganizationFeatureSetAll, createOutput.Organization.FeatureSet)

	t.Cleanup(func() {
		// Clean up: Delete all accounts except management account, then delete organization
		// List accounts and delete non-management accounts
		listOutput, _ := client.ListAccounts(ctx, &organizations.ListAccountsInput{})
		if listOutput != nil {
			for _, acc := range listOutput.Accounts {
				if *acc.Id != *createOutput.Organization.MasterAccountId {
					// In real implementation, we would close accounts
					// For now, just skip as awsim doesn't support CloseAccount
				}
			}
		}
		_, _ = client.DeleteOrganization(ctx, &organizations.DeleteOrganizationInput{})
	})
}

func TestOrganizations_DescribeOrganization(t *testing.T) {
	client := newOrganizationsClient(t)
	ctx := t.Context()

	// Ensure clean state
	ensureNoOrganization(t, client)

	// Create organization first
	createOutput, err := client.CreateOrganization(ctx, &organizations.CreateOrganizationInput{
		FeatureSet: types.OrganizationFeatureSetAll,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteOrganization(ctx, &organizations.DeleteOrganizationInput{})
	})

	// Describe organization
	descOutput, err := client.DescribeOrganization(ctx, &organizations.DescribeOrganizationInput{})
	require.NoError(t, err)
	require.NotNil(t, descOutput.Organization)
	assert.Equal(t, *createOutput.Organization.Id, *descOutput.Organization.Id)
}

func TestOrganizations_DescribeOrganization_NotInOrganization(t *testing.T) {
	client := newOrganizationsClient(t)
	ctx := t.Context()

	// Ensure clean state
	ensureNoOrganization(t, client)

	// Try to describe without being in an organization
	_, err := client.DescribeOrganization(ctx, &organizations.DescribeOrganizationInput{})
	require.Error(t, err)
}

func TestOrganizations_CreateAccount(t *testing.T) {
	client := newOrganizationsClient(t)
	ctx := t.Context()

	// Ensure clean state
	ensureNoOrganization(t, client)

	// Create organization first
	_, err := client.CreateOrganization(ctx, &organizations.CreateOrganizationInput{
		FeatureSet: types.OrganizationFeatureSetAll,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteOrganization(ctx, &organizations.DeleteOrganizationInput{})
	})

	// Create account
	createOutput, err := client.CreateAccount(ctx, &organizations.CreateAccountInput{
		AccountName: aws.String("Test Account"),
		Email:       aws.String("test@example.com"),
	})
	require.NoError(t, err)
	require.NotNil(t, createOutput.CreateAccountStatus)
	assert.NotEmpty(t, *createOutput.CreateAccountStatus.Id)
	assert.Equal(t, types.CreateAccountStateSucceeded, createOutput.CreateAccountStatus.State)
}

func TestOrganizations_ListAccounts(t *testing.T) {
	client := newOrganizationsClient(t)
	ctx := t.Context()

	// Ensure clean state
	ensureNoOrganization(t, client)

	// Create organization first
	_, err := client.CreateOrganization(ctx, &organizations.CreateOrganizationInput{
		FeatureSet: types.OrganizationFeatureSetAll,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteOrganization(ctx, &organizations.DeleteOrganizationInput{})
	})

	// Create a test account
	_, err = client.CreateAccount(ctx, &organizations.CreateAccountInput{
		AccountName: aws.String("Test Account"),
		Email:       aws.String("test@example.com"),
	})
	require.NoError(t, err)

	// List accounts
	listOutput, err := client.ListAccounts(ctx, &organizations.ListAccountsInput{})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(listOutput.Accounts), 2) // At least management + test account
}

func TestOrganizations_DescribeAccount(t *testing.T) {
	client := newOrganizationsClient(t)
	ctx := t.Context()

	// Ensure clean state
	ensureNoOrganization(t, client)

	// Create organization first
	orgOutput, err := client.CreateOrganization(ctx, &organizations.CreateOrganizationInput{
		FeatureSet: types.OrganizationFeatureSetAll,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteOrganization(ctx, &organizations.DeleteOrganizationInput{})
	})

	// Describe the management account
	descOutput, err := client.DescribeAccount(ctx, &organizations.DescribeAccountInput{
		AccountId: orgOutput.Organization.MasterAccountId,
	})
	require.NoError(t, err)
	require.NotNil(t, descOutput.Account)
	assert.Equal(t, *orgOutput.Organization.MasterAccountId, *descOutput.Account.Id)
}

func TestOrganizations_DescribeAccount_NotFound(t *testing.T) {
	client := newOrganizationsClient(t)
	ctx := t.Context()

	// Ensure clean state
	ensureNoOrganization(t, client)

	// Create organization first
	_, err := client.CreateOrganization(ctx, &organizations.CreateOrganizationInput{
		FeatureSet: types.OrganizationFeatureSetAll,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteOrganization(ctx, &organizations.DeleteOrganizationInput{})
	})

	// Try to describe non-existent account
	_, err = client.DescribeAccount(ctx, &organizations.DescribeAccountInput{
		AccountId: aws.String("000000000000"),
	})
	require.Error(t, err)
}

func TestOrganizations_CreateOrganizationalUnit(t *testing.T) {
	client := newOrganizationsClient(t)
	ctx := t.Context()

	// Ensure clean state
	ensureNoOrganization(t, client)

	// Create organization first
	_, err := client.CreateOrganization(ctx, &organizations.CreateOrganizationInput{
		FeatureSet: types.OrganizationFeatureSetAll,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteOrganization(ctx, &organizations.DeleteOrganizationInput{})
	})

	// Get root ID
	rootsOutput, err := client.ListRoots(ctx, &organizations.ListRootsInput{})
	require.NoError(t, err)
	require.NotEmpty(t, rootsOutput.Roots)
	rootID := rootsOutput.Roots[0].Id

	// Create organizational unit
	ouOutput, err := client.CreateOrganizationalUnit(ctx, &organizations.CreateOrganizationalUnitInput{
		Name:     aws.String("Test OU"),
		ParentId: rootID,
	})
	require.NoError(t, err)
	require.NotNil(t, ouOutput.OrganizationalUnit)
	assert.Equal(t, "Test OU", *ouOutput.OrganizationalUnit.Name)
	assert.NotEmpty(t, *ouOutput.OrganizationalUnit.Id)
}

func TestOrganizations_ListOrganizationalUnitsForParent(t *testing.T) {
	client := newOrganizationsClient(t)
	ctx := t.Context()

	// Ensure clean state
	ensureNoOrganization(t, client)

	// Create organization first
	_, err := client.CreateOrganization(ctx, &organizations.CreateOrganizationInput{
		FeatureSet: types.OrganizationFeatureSetAll,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteOrganization(ctx, &organizations.DeleteOrganizationInput{})
	})

	// Get root ID
	rootsOutput, err := client.ListRoots(ctx, &organizations.ListRootsInput{})
	require.NoError(t, err)
	require.NotEmpty(t, rootsOutput.Roots)
	rootID := rootsOutput.Roots[0].Id

	// Create organizational unit
	_, err = client.CreateOrganizationalUnit(ctx, &organizations.CreateOrganizationalUnitInput{
		Name:     aws.String("Test OU 1"),
		ParentId: rootID,
	})
	require.NoError(t, err)

	_, err = client.CreateOrganizationalUnit(ctx, &organizations.CreateOrganizationalUnitInput{
		Name:     aws.String("Test OU 2"),
		ParentId: rootID,
	})
	require.NoError(t, err)

	// List OUs
	listOutput, err := client.ListOrganizationalUnitsForParent(ctx, &organizations.ListOrganizationalUnitsForParentInput{
		ParentId: rootID,
	})
	require.NoError(t, err)
	assert.Len(t, listOutput.OrganizationalUnits, 2)
}

func TestOrganizations_ListRoots(t *testing.T) {
	client := newOrganizationsClient(t)
	ctx := t.Context()

	// Ensure clean state
	ensureNoOrganization(t, client)

	// Create organization first
	_, err := client.CreateOrganization(ctx, &organizations.CreateOrganizationInput{
		FeatureSet: types.OrganizationFeatureSetAll,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = client.DeleteOrganization(ctx, &organizations.DeleteOrganizationInput{})
	})

	// List roots
	rootsOutput, err := client.ListRoots(ctx, &organizations.ListRootsInput{})
	require.NoError(t, err)
	assert.Len(t, rootsOutput.Roots, 1)
	assert.NotEmpty(t, *rootsOutput.Roots[0].Id)
	assert.Equal(t, "Root", *rootsOutput.Roots[0].Name)
}

func TestOrganizations_DeleteOrganization(t *testing.T) {
	client := newOrganizationsClient(t)
	ctx := t.Context()

	// Ensure clean state
	ensureNoOrganization(t, client)

	// Create organization first
	_, err := client.CreateOrganization(ctx, &organizations.CreateOrganizationInput{
		FeatureSet: types.OrganizationFeatureSetAll,
	})
	require.NoError(t, err)

	// Delete organization
	_, err = client.DeleteOrganization(ctx, &organizations.DeleteOrganizationInput{})
	require.NoError(t, err)

	// Verify it's deleted
	_, err = client.DescribeOrganization(ctx, &organizations.DescribeOrganizationInput{})
	require.Error(t, err)
}

func TestOrganizations_DeleteOrganization_NotEmpty(t *testing.T) {
	client := newOrganizationsClient(t)
	ctx := t.Context()

	// Ensure clean state
	ensureNoOrganization(t, client)

	// Create organization first
	_, err := client.CreateOrganization(ctx, &organizations.CreateOrganizationInput{
		FeatureSet: types.OrganizationFeatureSetAll,
	})
	require.NoError(t, err)

	// Create a test account
	_, err = client.CreateAccount(ctx, &organizations.CreateAccountInput{
		AccountName: aws.String("Test Account"),
		Email:       aws.String("test@example.com"),
	})
	require.NoError(t, err)

	// Try to delete organization (should fail because it's not empty)
	_, err = client.DeleteOrganization(ctx, &organizations.DeleteOrganizationInput{})
	require.Error(t, err)
}

func TestOrganizations_EndToEnd(t *testing.T) {
	client := newOrganizationsClient(t)
	ctx := t.Context()

	// Ensure clean state
	ensureNoOrganization(t, client)

	// 1. Create organization
	orgOutput, err := client.CreateOrganization(ctx, &organizations.CreateOrganizationInput{
		FeatureSet: types.OrganizationFeatureSetAll,
	})
	require.NoError(t, err)
	require.NotNil(t, orgOutput.Organization)

	t.Cleanup(func() {
		_, _ = client.DeleteOrganization(ctx, &organizations.DeleteOrganizationInput{})
	})

	// 2. Describe organization
	descOutput, err := client.DescribeOrganization(ctx, &organizations.DescribeOrganizationInput{})
	require.NoError(t, err)
	assert.Equal(t, *orgOutput.Organization.Id, *descOutput.Organization.Id)

	// 3. List roots
	rootsOutput, err := client.ListRoots(ctx, &organizations.ListRootsInput{})
	require.NoError(t, err)
	require.Len(t, rootsOutput.Roots, 1)
	rootID := rootsOutput.Roots[0].Id

	// 4. Create organizational unit
	ouOutput, err := client.CreateOrganizationalUnit(ctx, &organizations.CreateOrganizationalUnitInput{
		Name:     aws.String("Production"),
		ParentId: rootID,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, *ouOutput.OrganizationalUnit.Id)

	// 5. List OUs for root
	ousOutput, err := client.ListOrganizationalUnitsForParent(ctx, &organizations.ListOrganizationalUnitsForParentInput{
		ParentId: rootID,
	})
	require.NoError(t, err)
	assert.Len(t, ousOutput.OrganizationalUnits, 1)

	// 6. List accounts
	accountsOutput, err := client.ListAccounts(ctx, &organizations.ListAccountsInput{})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(accountsOutput.Accounts), 1) // At least the management account

	// 7. Describe management account
	descAccountOutput, err := client.DescribeAccount(ctx, &organizations.DescribeAccountInput{
		AccountId: orgOutput.Organization.MasterAccountId,
	})
	require.NoError(t, err)
	assert.Equal(t, *orgOutput.Organization.MasterAccountId, *descAccountOutput.Account.Id)
}
