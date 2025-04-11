package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-trayai/pkg/connector/client"
)

// Create a new connector resource for a tray.ai user.
func fromElementTouserResource(
	_ context.Context,
	user client.Element,
	parentResourceID *v2.ResourceId,
) (*v2.Resource, error) {
	// TODO. BB-451. We should get the email via GetUser api.
	// see https://developer.tray.ai/openapi/trayapi/tag/users/#tag/users/operation/get-user-by-id
	profile := map[string]interface{}{
		"id":       user.ID,
		"username": user.Name,
	}
	return resource.NewUserResource(
		user.Name,
		userResourceType,
		user.ID,
		[]resource.UserTraitOption{
			resource.WithStatus(v2.UserTrait_Status_STATUS_ENABLED),
			resource.WithUserProfile(profile),
		},
		resource.WithParentResourceID(parentResourceID),
	)
}

func userResource(user *client.User) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"id":          user.ID,
		"username":    user.Name,
		"email":       user.Email,
		"accountType": user.AccountType,
		"role":        user.Role.Name,
	}
	return resource.NewUserResource(
		user.Name,
		userResourceType,
		user.ID,
		[]resource.UserTraitOption{
			resource.WithStatus(v2.UserTrait_Status_STATUS_ENABLED),
			resource.WithUserProfile(profile),
		},
	)
}

type userBuilder struct {
	client *client.Client
}

func (o *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return userResourceType
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (o *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var (
		users []*v2.Resource
	)

	resp, err := o.client.ListUsers(ctx, client.ListParams{
		Cursor: pToken.Token,
		First:  pToken.Size,
	})
	if err != nil {
		return nil, "", nil, fmt.Errorf("baton-trayai: ListUsers failed: %w", err)
	}

	for _, user := range resp.Elements {
		vUser, err := fromElementTouserResource(ctx, user, parentResourceID)
		if err != nil {
			return nil, "", nil, fmt.Errorf("baton-trayai: cannot create connector resource: %w", err)
		}
		users = append(users, vUser)
	}

	if !resp.Page.HasNextPage {
		return users, "", nil, nil
	}
	return users, resp.Page.EndCursor, nil, nil
}

// Entitlements always returns an empty slice for users.
func (o *userBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *userBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func (o *userBuilder) CreateAccountCapabilityDetails(ctx context.Context) (*v2.CredentialDetailsAccountProvisioning, annotations.Annotations, error) {
	return &v2.CredentialDetailsAccountProvisioning{
		SupportedCredentialOptions: []v2.CapabilityDetailCredentialOption{
			v2.CapabilityDetailCredentialOption_CAPABILITY_DETAIL_CREDENTIAL_OPTION_NO_PASSWORD,
		},
		PreferredCredentialOption: v2.CapabilityDetailCredentialOption_CAPABILITY_DETAIL_CREDENTIAL_OPTION_NO_PASSWORD,
	}, nil, nil
}

func (o *userBuilder) CreateAccount(ctx context.Context, accountInfo *v2.AccountInfo, credentialOptions *v2.CredentialOptions) (
	connectorbuilder.CreateAccountResponse,
	[]*v2.PlaintextData,
	annotations.Annotations,
	error,
) {
	params, err := getCreateUserParams(accountInfo)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("baton-trayai: getCreateUserParams failed: %w", err)
	}

	user, err := o.client.CreateUser(ctx, params)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("baton-trayai: cannot create user: %w", err)
	}

	r, err := userResource(user)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("baton-slack: cannot create user resource: %w", err)
	}

	return &v2.CreateAccountResponse_SuccessResult{
		Resource: r,
	}, nil, nil, nil
}

func getCreateUserParams(accountInfo *v2.AccountInfo) (*client.CreateUserParams, error) {
	pMap := accountInfo.Profile.AsMap()
	email, ok := pMap["email"].(string)
	if !ok || email == "" {
		return nil, fmt.Errorf("email is required")
	}

	name, ok := pMap["name"].(string)
	if !ok || name == "" {
		return nil, fmt.Errorf("name is required")
	}

	organizationRoleId, ok := pMap["organizationRoleId"].(string)
	if !ok || organizationRoleId == "" {
		return nil, fmt.Errorf("organizationRoleId is required")
	}
	return &client.CreateUserParams{
		Name:               name,
		Email:              email,
		OrganizationRoleId: organizationRoleId,
	}, nil
}
func newUserBuilder(c *client.Client) *userBuilder {
	return &userBuilder{
		client: c,
	}
}
