//go:build integration

package integration

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

func newCognitoClient(t *testing.T) *cognitoidentityprovider.Client {
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

	return cognitoidentityprovider.NewFromConfig(cfg, func(o *cognitoidentityprovider.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
}

func TestCognito_CreateAndDescribeUserPool(t *testing.T) {
	client := newCognitoClient(t)
	ctx := t.Context()

	// Create user pool.
	createOutput, err := client.CreateUserPool(ctx, &cognitoidentityprovider.CreateUserPoolInput{
		PoolName: aws.String("test-user-pool"),
		Policies: &types.UserPoolPolicyType{
			PasswordPolicy: &types.PasswordPolicyType{
				MinimumLength:    aws.Int32(8),
				RequireUppercase: true,
				RequireLowercase: true,
				RequireNumbers:   true,
				RequireSymbols:   false,
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create user pool: %v", err)
	}

	if createOutput.UserPool == nil || createOutput.UserPool.Id == nil {
		t.Fatal("user pool is nil")
	}

	userPoolID := *createOutput.UserPool.Id
	t.Logf("Created user pool: %s", userPoolID)

	// Describe user pool.
	describeOutput, err := client.DescribeUserPool(ctx, &cognitoidentityprovider.DescribeUserPoolInput{
		UserPoolId: aws.String(userPoolID),
	})
	if err != nil {
		t.Fatalf("failed to describe user pool: %v", err)
	}

	if *describeOutput.UserPool.Id != userPoolID {
		t.Errorf("user pool ID mismatch: got %s, want %s", *describeOutput.UserPool.Id, userPoolID)
	}

	if *describeOutput.UserPool.Name != "test-user-pool" {
		t.Errorf("name mismatch: got %s, want test-user-pool", *describeOutput.UserPool.Name)
	}

	t.Logf("Described user pool: %s", userPoolID)
}

func TestCognito_ListUserPools(t *testing.T) {
	client := newCognitoClient(t)
	ctx := t.Context()

	// Create a user pool first.
	createOutput, err := client.CreateUserPool(ctx, &cognitoidentityprovider.CreateUserPoolInput{
		PoolName: aws.String("test-list-user-pool"),
	})
	if err != nil {
		t.Fatalf("failed to create user pool: %v", err)
	}

	userPoolID := *createOutput.UserPool.Id

	// List user pools.
	listOutput, err := client.ListUserPools(ctx, &cognitoidentityprovider.ListUserPoolsInput{
		MaxResults: aws.Int32(10),
	})
	if err != nil {
		t.Fatalf("failed to list user pools: %v", err)
	}

	if len(listOutput.UserPools) == 0 {
		t.Fatal("no user pools returned")
	}

	// Find our user pool.
	found := false

	for _, pool := range listOutput.UserPools {
		if *pool.Id == userPoolID {
			found = true

			break
		}
	}

	if !found {
		t.Errorf("created user pool %s not found in list", userPoolID)
	}

	t.Logf("Listed %d user pools", len(listOutput.UserPools))
}

func TestCognito_CreateAndDescribeUserPoolClient(t *testing.T) {
	client := newCognitoClient(t)
	ctx := t.Context()

	// Create user pool first.
	poolOutput, err := client.CreateUserPool(ctx, &cognitoidentityprovider.CreateUserPoolInput{
		PoolName: aws.String("test-client-user-pool"),
	})
	if err != nil {
		t.Fatalf("failed to create user pool: %v", err)
	}

	userPoolID := *poolOutput.UserPool.Id

	// Create user pool client.
	clientOutput, err := client.CreateUserPoolClient(ctx, &cognitoidentityprovider.CreateUserPoolClientInput{
		UserPoolId: aws.String(userPoolID),
		ClientName: aws.String("test-client"),
		ExplicitAuthFlows: []types.ExplicitAuthFlowsType{
			types.ExplicitAuthFlowsTypeAllowUserPasswordAuth,
			types.ExplicitAuthFlowsTypeAllowRefreshTokenAuth,
		},
	})
	if err != nil {
		t.Fatalf("failed to create user pool client: %v", err)
	}

	clientID := *clientOutput.UserPoolClient.ClientId
	t.Logf("Created user pool client: %s", clientID)

	// Describe user pool client.
	describeOutput, err := client.DescribeUserPoolClient(ctx, &cognitoidentityprovider.DescribeUserPoolClientInput{
		UserPoolId: aws.String(userPoolID),
		ClientId:   aws.String(clientID),
	})
	if err != nil {
		t.Fatalf("failed to describe user pool client: %v", err)
	}

	if *describeOutput.UserPoolClient.ClientId != clientID {
		t.Errorf("client ID mismatch: got %s, want %s", *describeOutput.UserPoolClient.ClientId, clientID)
	}

	if *describeOutput.UserPoolClient.ClientName != "test-client" {
		t.Errorf("client name mismatch: got %s, want test-client", *describeOutput.UserPoolClient.ClientName)
	}

	t.Logf("Described user pool client: %s", clientID)
}

func TestCognito_AdminCreateAndGetUser(t *testing.T) {
	client := newCognitoClient(t)
	ctx := t.Context()

	// Create user pool.
	poolOutput, err := client.CreateUserPool(ctx, &cognitoidentityprovider.CreateUserPoolInput{
		PoolName: aws.String("test-admin-user-pool"),
	})
	if err != nil {
		t.Fatalf("failed to create user pool: %v", err)
	}

	userPoolID := *poolOutput.UserPool.Id

	// Admin create user.
	createUserOutput, err := client.AdminCreateUser(ctx, &cognitoidentityprovider.AdminCreateUserInput{
		UserPoolId:        aws.String(userPoolID),
		Username:          aws.String("testuser"),
		TemporaryPassword: aws.String("TempPass123!"),
		UserAttributes: []types.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String("testuser@example.com"),
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to admin create user: %v", err)
	}

	if *createUserOutput.User.Username != "testuser" {
		t.Errorf("username mismatch: got %s, want testuser", *createUserOutput.User.Username)
	}

	t.Logf("Created user: %s", *createUserOutput.User.Username)

	// Admin get user.
	getUserOutput, err := client.AdminGetUser(ctx, &cognitoidentityprovider.AdminGetUserInput{
		UserPoolId: aws.String(userPoolID),
		Username:   aws.String("testuser"),
	})
	if err != nil {
		t.Fatalf("failed to admin get user: %v", err)
	}

	if *getUserOutput.Username != "testuser" {
		t.Errorf("username mismatch: got %s, want testuser", *getUserOutput.Username)
	}

	t.Logf("Got user: %s", *getUserOutput.Username)
}

func TestCognito_ListUsers(t *testing.T) {
	client := newCognitoClient(t)
	ctx := t.Context()

	// Create user pool.
	poolOutput, err := client.CreateUserPool(ctx, &cognitoidentityprovider.CreateUserPoolInput{
		PoolName: aws.String("test-list-users-pool"),
	})
	if err != nil {
		t.Fatalf("failed to create user pool: %v", err)
	}

	userPoolID := *poolOutput.UserPool.Id

	// Create a user.
	_, err = client.AdminCreateUser(ctx, &cognitoidentityprovider.AdminCreateUserInput{
		UserPoolId:        aws.String(userPoolID),
		Username:          aws.String("listuser1"),
		TemporaryPassword: aws.String("TempPass123!"),
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// List users.
	listOutput, err := client.ListUsers(ctx, &cognitoidentityprovider.ListUsersInput{
		UserPoolId: aws.String(userPoolID),
	})
	if err != nil {
		t.Fatalf("failed to list users: %v", err)
	}

	if len(listOutput.Users) == 0 {
		t.Fatal("no users returned")
	}

	found := false

	for _, user := range listOutput.Users {
		if *user.Username == "listuser1" {
			found = true

			break
		}
	}

	if !found {
		t.Error("created user not found in list")
	}

	t.Logf("Listed %d users", len(listOutput.Users))
}

func TestCognito_SignUpAndConfirm(t *testing.T) {
	client := newCognitoClient(t)
	ctx := t.Context()

	// Create user pool.
	poolOutput, err := client.CreateUserPool(ctx, &cognitoidentityprovider.CreateUserPoolInput{
		PoolName: aws.String("test-signup-pool"),
	})
	if err != nil {
		t.Fatalf("failed to create user pool: %v", err)
	}

	userPoolID := *poolOutput.UserPool.Id

	// Create user pool client.
	clientOutput, err := client.CreateUserPoolClient(ctx, &cognitoidentityprovider.CreateUserPoolClientInput{
		UserPoolId: aws.String(userPoolID),
		ClientName: aws.String("signup-client"),
		ExplicitAuthFlows: []types.ExplicitAuthFlowsType{
			types.ExplicitAuthFlowsTypeAllowUserPasswordAuth,
		},
	})
	if err != nil {
		t.Fatalf("failed to create user pool client: %v", err)
	}

	clientID := *clientOutput.UserPoolClient.ClientId

	// Sign up.
	signUpOutput, err := client.SignUp(ctx, &cognitoidentityprovider.SignUpInput{
		ClientId: aws.String(clientID),
		Username: aws.String("signupuser"),
		Password: aws.String("Password123!"),
		UserAttributes: []types.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String("signupuser@example.com"),
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to sign up: %v", err)
	}

	if signUpOutput.UserConfirmed {
		t.Error("user should not be confirmed yet")
	}

	t.Logf("Signed up user: %s", *signUpOutput.UserSub)

	// Confirm sign up.
	_, err = client.ConfirmSignUp(ctx, &cognitoidentityprovider.ConfirmSignUpInput{
		ClientId:         aws.String(clientID),
		Username:         aws.String("signupuser"),
		ConfirmationCode: aws.String("123456"),
	})
	if err != nil {
		t.Fatalf("failed to confirm sign up: %v", err)
	}

	t.Log("Confirmed sign up")
}

func TestCognito_InitiateAuth(t *testing.T) {
	client := newCognitoClient(t)
	ctx := t.Context()

	// Create user pool.
	poolOutput, err := client.CreateUserPool(ctx, &cognitoidentityprovider.CreateUserPoolInput{
		PoolName: aws.String("test-auth-pool"),
	})
	if err != nil {
		t.Fatalf("failed to create user pool: %v", err)
	}

	userPoolID := *poolOutput.UserPool.Id

	// Create user pool client.
	clientOutput, err := client.CreateUserPoolClient(ctx, &cognitoidentityprovider.CreateUserPoolClientInput{
		UserPoolId: aws.String(userPoolID),
		ClientName: aws.String("auth-client"),
		ExplicitAuthFlows: []types.ExplicitAuthFlowsType{
			types.ExplicitAuthFlowsTypeAllowUserPasswordAuth,
		},
	})
	if err != nil {
		t.Fatalf("failed to create user pool client: %v", err)
	}

	clientID := *clientOutput.UserPoolClient.ClientId

	// Sign up and confirm user.
	_, err = client.SignUp(ctx, &cognitoidentityprovider.SignUpInput{
		ClientId: aws.String(clientID),
		Username: aws.String("authuser"),
		Password: aws.String("Password123!"),
	})
	if err != nil {
		t.Fatalf("failed to sign up: %v", err)
	}

	_, err = client.ConfirmSignUp(ctx, &cognitoidentityprovider.ConfirmSignUpInput{
		ClientId:         aws.String(clientID),
		Username:         aws.String("authuser"),
		ConfirmationCode: aws.String("123456"),
	})
	if err != nil {
		t.Fatalf("failed to confirm sign up: %v", err)
	}

	// Initiate auth.
	authOutput, err := client.InitiateAuth(ctx, &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: types.AuthFlowTypeUserPasswordAuth,
		ClientId: aws.String(clientID),
		AuthParameters: map[string]string{
			"USERNAME": "authuser",
			"PASSWORD": "Password123!",
		},
	})
	if err != nil {
		t.Fatalf("failed to initiate auth: %v", err)
	}

	if authOutput.AuthenticationResult == nil {
		t.Fatal("authentication result is nil")
	}

	if authOutput.AuthenticationResult.AccessToken == nil {
		t.Error("access token is nil")
	}

	if authOutput.AuthenticationResult.IdToken == nil {
		t.Error("id token is nil")
	}

	t.Log("Successfully authenticated user")
}

func TestCognito_UserPoolNotFound(t *testing.T) {
	client := newCognitoClient(t)
	ctx := t.Context()

	// Try to describe a non-existent user pool.
	_, err := client.DescribeUserPool(ctx, &cognitoidentityprovider.DescribeUserPoolInput{
		UserPoolId: aws.String("us-east-1_nonexistent"),
	})
	if err == nil {
		t.Fatal("expected error for non-existent user pool")
	}
}

func TestCognito_DeleteUserPool(t *testing.T) {
	client := newCognitoClient(t)
	ctx := t.Context()

	// Create user pool.
	createOutput, err := client.CreateUserPool(ctx, &cognitoidentityprovider.CreateUserPoolInput{
		PoolName: aws.String("test-delete-pool"),
	})
	if err != nil {
		t.Fatalf("failed to create user pool: %v", err)
	}

	userPoolID := *createOutput.UserPool.Id

	// Delete user pool.
	_, err = client.DeleteUserPool(ctx, &cognitoidentityprovider.DeleteUserPoolInput{
		UserPoolId: aws.String(userPoolID),
	})
	if err != nil {
		t.Fatalf("failed to delete user pool: %v", err)
	}

	// Verify deletion.
	_, err = client.DescribeUserPool(ctx, &cognitoidentityprovider.DescribeUserPoolInput{
		UserPoolId: aws.String(userPoolID),
	})
	if err == nil {
		t.Fatal("expected error for deleted user pool")
	}

	t.Logf("Deleted user pool: %s", userPoolID)
}
