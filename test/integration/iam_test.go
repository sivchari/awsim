//go:build integration

package integration

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

func newIAMClient(t *testing.T) *iam.Client {
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

	return iam.NewFromConfig(cfg, func(o *iam.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566/iam")
	})
}

func TestIAM_CreateAndDeleteUser(t *testing.T) {
	client := newIAMClient(t)
	ctx := t.Context()
	userName := "test-user"

	// Create user
	createResult, err := client.CreateUser(ctx, &iam.CreateUserInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if createResult.User == nil {
		t.Fatal("expected user to be created")
	}

	if *createResult.User.UserName != userName {
		t.Errorf("expected user name %s, got %s", userName, *createResult.User.UserName)
	}

	// Delete user
	_, err = client.DeleteUser(ctx, &iam.DeleteUserInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		t.Fatalf("failed to delete user: %v", err)
	}
}

func TestIAM_GetUser(t *testing.T) {
	client := newIAMClient(t)
	ctx := t.Context()
	userName := "test-get-user"

	// Create user
	_, err := client.CreateUser(ctx, &iam.CreateUserInput{
		UserName: aws.String(userName),
		Path:     aws.String("/developers/"),
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteUser(ctx, &iam.DeleteUserInput{
			UserName: aws.String(userName),
		})
	})

	// Get user
	getResult, err := client.GetUser(ctx, &iam.GetUserInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}

	if getResult.User == nil {
		t.Fatal("expected user in get response")
	}

	if *getResult.User.UserName != userName {
		t.Errorf("expected user name %s, got %s", userName, *getResult.User.UserName)
	}

	if *getResult.User.Path != "/developers/" {
		t.Errorf("expected path /developers/, got %s", *getResult.User.Path)
	}
}

func TestIAM_ListUsers(t *testing.T) {
	client := newIAMClient(t)
	ctx := t.Context()
	userName := "test-list-user"

	// Create user
	_, err := client.CreateUser(ctx, &iam.CreateUserInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteUser(ctx, &iam.DeleteUserInput{
			UserName: aws.String(userName),
		})
	})

	// List users
	listResult, err := client.ListUsers(ctx, &iam.ListUsersInput{})
	if err != nil {
		t.Fatalf("failed to list users: %v", err)
	}

	found := false
	for _, user := range listResult.Users {
		if *user.UserName == userName {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected to find user %s in list", userName)
	}
}

func TestIAM_CreateAndDeleteRole(t *testing.T) {
	client := newIAMClient(t)
	ctx := t.Context()
	roleName := "test-role"

	assumeRolePolicy := `{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Principal": {"Service": "ec2.amazonaws.com"},
			"Action": "sts:AssumeRole"
		}]
	}`

	// Create role
	createResult, err := client.CreateRole(ctx, &iam.CreateRoleInput{
		RoleName:                 aws.String(roleName),
		AssumeRolePolicyDocument: aws.String(assumeRolePolicy),
	})
	if err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	if createResult.Role == nil {
		t.Fatal("expected role to be created")
	}

	if *createResult.Role.RoleName != roleName {
		t.Errorf("expected role name %s, got %s", roleName, *createResult.Role.RoleName)
	}

	// Delete role
	_, err = client.DeleteRole(ctx, &iam.DeleteRoleInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		t.Fatalf("failed to delete role: %v", err)
	}
}

func TestIAM_GetRole(t *testing.T) {
	client := newIAMClient(t)
	ctx := t.Context()
	roleName := "test-get-role"

	assumeRolePolicy := `{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Principal": {"Service": "lambda.amazonaws.com"},
			"Action": "sts:AssumeRole"
		}]
	}`

	// Create role
	_, err := client.CreateRole(ctx, &iam.CreateRoleInput{
		RoleName:                 aws.String(roleName),
		AssumeRolePolicyDocument: aws.String(assumeRolePolicy),
		Description:              aws.String("Test role"),
	})
	if err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteRole(ctx, &iam.DeleteRoleInput{
			RoleName: aws.String(roleName),
		})
	})

	// Get role
	getResult, err := client.GetRole(ctx, &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		t.Fatalf("failed to get role: %v", err)
	}

	if getResult.Role == nil {
		t.Fatal("expected role in get response")
	}

	if *getResult.Role.RoleName != roleName {
		t.Errorf("expected role name %s, got %s", roleName, *getResult.Role.RoleName)
	}
}

func TestIAM_ListRoles(t *testing.T) {
	client := newIAMClient(t)
	ctx := t.Context()
	roleName := "test-list-role"

	assumeRolePolicy := `{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Principal": {"Service": "ec2.amazonaws.com"},
			"Action": "sts:AssumeRole"
		}]
	}`

	// Create role
	_, err := client.CreateRole(ctx, &iam.CreateRoleInput{
		RoleName:                 aws.String(roleName),
		AssumeRolePolicyDocument: aws.String(assumeRolePolicy),
	})
	if err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteRole(ctx, &iam.DeleteRoleInput{
			RoleName: aws.String(roleName),
		})
	})

	// List roles
	listResult, err := client.ListRoles(ctx, &iam.ListRolesInput{})
	if err != nil {
		t.Fatalf("failed to list roles: %v", err)
	}

	found := false
	for _, role := range listResult.Roles {
		if *role.RoleName == roleName {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected to find role %s in list", roleName)
	}
}

func TestIAM_CreateAndDeletePolicy(t *testing.T) {
	client := newIAMClient(t)
	ctx := t.Context()
	policyName := "test-policy"

	policyDocument := `{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Action": "s3:GetObject",
			"Resource": "*"
		}]
	}`

	// Create policy
	createResult, err := client.CreatePolicy(ctx, &iam.CreatePolicyInput{
		PolicyName:     aws.String(policyName),
		PolicyDocument: aws.String(policyDocument),
	})
	if err != nil {
		t.Fatalf("failed to create policy: %v", err)
	}

	if createResult.Policy == nil {
		t.Fatal("expected policy to be created")
	}

	if *createResult.Policy.PolicyName != policyName {
		t.Errorf("expected policy name %s, got %s", policyName, *createResult.Policy.PolicyName)
	}

	policyArn := createResult.Policy.Arn

	// Delete policy
	_, err = client.DeletePolicy(ctx, &iam.DeletePolicyInput{
		PolicyArn: policyArn,
	})
	if err != nil {
		t.Fatalf("failed to delete policy: %v", err)
	}
}

func TestIAM_GetPolicy(t *testing.T) {
	client := newIAMClient(t)
	ctx := t.Context()
	policyName := "test-get-policy"

	policyDocument := `{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Action": "dynamodb:*",
			"Resource": "*"
		}]
	}`

	// Create policy
	createResult, err := client.CreatePolicy(ctx, &iam.CreatePolicyInput{
		PolicyName:     aws.String(policyName),
		PolicyDocument: aws.String(policyDocument),
		Description:    aws.String("Test policy"),
	})
	if err != nil {
		t.Fatalf("failed to create policy: %v", err)
	}

	policyArn := createResult.Policy.Arn

	t.Cleanup(func() {
		_, _ = client.DeletePolicy(ctx, &iam.DeletePolicyInput{
			PolicyArn: policyArn,
		})
	})

	// Get policy
	getResult, err := client.GetPolicy(ctx, &iam.GetPolicyInput{
		PolicyArn: policyArn,
	})
	if err != nil {
		t.Fatalf("failed to get policy: %v", err)
	}

	if getResult.Policy == nil {
		t.Fatal("expected policy in get response")
	}

	if *getResult.Policy.PolicyName != policyName {
		t.Errorf("expected policy name %s, got %s", policyName, *getResult.Policy.PolicyName)
	}
}

func TestIAM_AttachAndDetachUserPolicy(t *testing.T) {
	client := newIAMClient(t)
	ctx := t.Context()
	userName := "test-attach-user"
	policyName := "test-attach-policy"

	// Create user
	_, err := client.CreateUser(ctx, &iam.CreateUserInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Create policy
	policyDocument := `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow", "Action": "*", "Resource": "*"}]}`
	createPolicyResult, err := client.CreatePolicy(ctx, &iam.CreatePolicyInput{
		PolicyName:     aws.String(policyName),
		PolicyDocument: aws.String(policyDocument),
	})
	if err != nil {
		t.Fatalf("failed to create policy: %v", err)
	}

	policyArn := createPolicyResult.Policy.Arn

	t.Cleanup(func() {
		_, _ = client.DetachUserPolicy(ctx, &iam.DetachUserPolicyInput{
			UserName:  aws.String(userName),
			PolicyArn: policyArn,
		})
		_, _ = client.DeletePolicy(ctx, &iam.DeletePolicyInput{
			PolicyArn: policyArn,
		})
		_, _ = client.DeleteUser(ctx, &iam.DeleteUserInput{
			UserName: aws.String(userName),
		})
	})

	// Attach policy to user
	_, err = client.AttachUserPolicy(ctx, &iam.AttachUserPolicyInput{
		UserName:  aws.String(userName),
		PolicyArn: policyArn,
	})
	if err != nil {
		t.Fatalf("failed to attach policy to user: %v", err)
	}

	// Detach policy from user
	_, err = client.DetachUserPolicy(ctx, &iam.DetachUserPolicyInput{
		UserName:  aws.String(userName),
		PolicyArn: policyArn,
	})
	if err != nil {
		t.Fatalf("failed to detach policy from user: %v", err)
	}
}

func TestIAM_AttachAndDetachRolePolicy(t *testing.T) {
	client := newIAMClient(t)
	ctx := t.Context()
	roleName := "test-attach-role"
	policyName := "test-attach-role-policy"

	assumeRolePolicy := `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow", "Principal": {"Service": "ec2.amazonaws.com"}, "Action": "sts:AssumeRole"}]}`

	// Create role
	_, err := client.CreateRole(ctx, &iam.CreateRoleInput{
		RoleName:                 aws.String(roleName),
		AssumeRolePolicyDocument: aws.String(assumeRolePolicy),
	})
	if err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	// Create policy
	policyDocument := `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow", "Action": "*", "Resource": "*"}]}`
	createPolicyResult, err := client.CreatePolicy(ctx, &iam.CreatePolicyInput{
		PolicyName:     aws.String(policyName),
		PolicyDocument: aws.String(policyDocument),
	})
	if err != nil {
		t.Fatalf("failed to create policy: %v", err)
	}

	policyArn := createPolicyResult.Policy.Arn

	t.Cleanup(func() {
		_, _ = client.DetachRolePolicy(ctx, &iam.DetachRolePolicyInput{
			RoleName:  aws.String(roleName),
			PolicyArn: policyArn,
		})
		_, _ = client.DeletePolicy(ctx, &iam.DeletePolicyInput{
			PolicyArn: policyArn,
		})
		_, _ = client.DeleteRole(ctx, &iam.DeleteRoleInput{
			RoleName: aws.String(roleName),
		})
	})

	// Attach policy to role
	_, err = client.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
		RoleName:  aws.String(roleName),
		PolicyArn: policyArn,
	})
	if err != nil {
		t.Fatalf("failed to attach policy to role: %v", err)
	}

	// Detach policy from role
	_, err = client.DetachRolePolicy(ctx, &iam.DetachRolePolicyInput{
		RoleName:  aws.String(roleName),
		PolicyArn: policyArn,
	})
	if err != nil {
		t.Fatalf("failed to detach policy from role: %v", err)
	}
}

func TestIAM_CreateAndDeleteAccessKey(t *testing.T) {
	client := newIAMClient(t)
	ctx := t.Context()
	userName := "test-access-key-user"

	// Create user
	_, err := client.CreateUser(ctx, &iam.CreateUserInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	t.Cleanup(func() {
		// List and delete all access keys first
		listResult, _ := client.ListAccessKeys(ctx, &iam.ListAccessKeysInput{
			UserName: aws.String(userName),
		})
		for _, key := range listResult.AccessKeyMetadata {
			_, _ = client.DeleteAccessKey(ctx, &iam.DeleteAccessKeyInput{
				UserName:    aws.String(userName),
				AccessKeyId: key.AccessKeyId,
			})
		}
		_, _ = client.DeleteUser(ctx, &iam.DeleteUserInput{
			UserName: aws.String(userName),
		})
	})

	// Create access key
	createResult, err := client.CreateAccessKey(ctx, &iam.CreateAccessKeyInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		t.Fatalf("failed to create access key: %v", err)
	}

	if createResult.AccessKey == nil {
		t.Fatal("expected access key to be created")
	}

	if createResult.AccessKey.AccessKeyId == nil || *createResult.AccessKey.AccessKeyId == "" {
		t.Error("expected access key ID to be set")
	}

	if createResult.AccessKey.SecretAccessKey == nil || *createResult.AccessKey.SecretAccessKey == "" {
		t.Error("expected secret access key to be set")
	}

	if createResult.AccessKey.Status != types.StatusTypeActive {
		t.Errorf("expected access key status Active, got %s", createResult.AccessKey.Status)
	}

	// Delete access key
	_, err = client.DeleteAccessKey(ctx, &iam.DeleteAccessKeyInput{
		UserName:    aws.String(userName),
		AccessKeyId: createResult.AccessKey.AccessKeyId,
	})
	if err != nil {
		t.Fatalf("failed to delete access key: %v", err)
	}
}

func TestIAM_ListAccessKeys(t *testing.T) {
	client := newIAMClient(t)
	ctx := t.Context()
	userName := "test-list-access-keys-user"

	// Create user
	_, err := client.CreateUser(ctx, &iam.CreateUserInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Create access key
	createResult, err := client.CreateAccessKey(ctx, &iam.CreateAccessKeyInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		t.Fatalf("failed to create access key: %v", err)
	}

	accessKeyID := createResult.AccessKey.AccessKeyId

	t.Cleanup(func() {
		_, _ = client.DeleteAccessKey(ctx, &iam.DeleteAccessKeyInput{
			UserName:    aws.String(userName),
			AccessKeyId: accessKeyID,
		})
		_, _ = client.DeleteUser(ctx, &iam.DeleteUserInput{
			UserName: aws.String(userName),
		})
	})

	// List access keys
	listResult, err := client.ListAccessKeys(ctx, &iam.ListAccessKeysInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		t.Fatalf("failed to list access keys: %v", err)
	}

	found := false
	for _, key := range listResult.AccessKeyMetadata {
		if *key.AccessKeyId == *accessKeyID {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected to find access key %s in list", *accessKeyID)
	}
}

func TestIAM_UserNotFound(t *testing.T) {
	client := newIAMClient(t)
	ctx := t.Context()

	// Try to get non-existent user
	_, err := client.GetUser(ctx, &iam.GetUserInput{
		UserName: aws.String("non-existent-user"),
	})
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}

func TestIAM_RoleNotFound(t *testing.T) {
	client := newIAMClient(t)
	ctx := t.Context()

	// Try to get non-existent role
	_, err := client.GetRole(ctx, &iam.GetRoleInput{
		RoleName: aws.String("non-existent-role"),
	})
	if err == nil {
		t.Fatal("expected error for non-existent role")
	}
}

func TestIAM_CreateUserWithTags(t *testing.T) {
	client := newIAMClient(t)
	ctx := t.Context()
	userName := "test-user-with-tags"

	// Create user with tags
	_, err := client.CreateUser(ctx, &iam.CreateUserInput{
		UserName: aws.String(userName),
		Tags: []types.Tag{
			{Key: aws.String("Environment"), Value: aws.String("test")},
			{Key: aws.String("Project"), Value: aws.String("awsim")},
		},
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteUser(ctx, &iam.DeleteUserInput{
			UserName: aws.String(userName),
		})
	})

	// Get user and verify
	getResult, err := client.GetUser(ctx, &iam.GetUserInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}

	if getResult.User == nil {
		t.Fatal("expected user in response")
	}
}

// Helper function to pretty print JSON for debugging.
func prettyJSON(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
