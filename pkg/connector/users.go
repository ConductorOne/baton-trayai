package connector

import (
	"context"
	"fmt"
	"sync"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-trayai/pkg/connector/client"
)

// Create a new connector resource for a tray.ai user.
func userResource(
	_ context.Context,
	user *client.User,
	parentResourceID *v2.ResourceId,
) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"id":          user.ID,
		"email":       user.Email,
		"username":    user.Name,
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
		resource.WithParentResourceID(parentResourceID),
	)
}

type userBuilder struct {
	client *client.Client
	mu     sync.Mutex
	users  map[string]*client.User // map[userID]*User
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

	o.mu.Lock()
	defer o.mu.Unlock()

	for _, element := range resp.Elements {
		user, err := o.client.GetUser(ctx, element.ID)
		if err != nil {
			return nil, "", nil, fmt.Errorf("baton-trayai: GetUser failed: %w", err)
		}

		vUser, err := userResource(ctx, user, parentResourceID)
		if err != nil {
			return nil, "", nil, fmt.Errorf("baton-trayai: cannot create connector resource: %w", err)
		}

		o.users[user.ID] = user
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

func (o *userBuilder) getUsers() map[string]*client.User {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.users
}

func newUserBuilder(c *client.Client) *userBuilder {
	return &userBuilder{
		client: c,
		users:  make(map[string]*client.User),
	}
}
