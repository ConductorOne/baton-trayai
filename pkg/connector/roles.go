package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-trayai/pkg/connector/client"
)

const RoleAssignmentEntitlement = "assigned"

func roleResource(e client.Element, parentResourceID *v2.ResourceId) (*v2.Resource, error) {
	r, err := resource.NewRoleResource(
		e.Name,
		roleResourceType,
		e.ID,
		nil,
		resource.WithParentResourceID(parentResourceID),
	)
	if err != nil {
		return nil, fmt.Errorf("baton-trayai: cannot create roleResource: %w", err)
	}
	return r, nil
}

// roleBuilder is the builder for workspace role.
type roleBuilder struct {
	client *client.Client
}

func (r *roleBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return roleResourceType
}

// List lists all the organization roles.
func (r *roleBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	if parentResourceID == nil {
		return nil, "", nil, nil
	}

	resp, err := r.client.ListWorkspaceRoles(ctx, client.ListParams{
		Cursor:      pToken.Token,
		First:       pToken.Size,
		WorkspaceID: parentResourceID.Resource,
	})
	if err != nil {
		return nil, "", nil, fmt.Errorf("baton-trayai: List workspace roles failed: %w", err)
	}

	roles := make([]*v2.Resource, 0, len(resp.Elements))
	for _, role := range resp.Elements {
		r, err := roleResource(role, parentResourceID)
		if err != nil {
			return nil, "", nil, fmt.Errorf("baton-trayai: cannot create role resource: %w", err)
		}
		roles = append(roles, r)
	}
	return roles, "", nil, nil
}

func (r *roleBuilder) Entitlements(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	if resource == nil || resource.ParentResourceId == nil {
		return nil, "", nil, nil
	}

	workspace, err := r.client.GetWorkspace(ctx, resource.ParentResourceId.Resource)
	if err != nil {
		return nil, "", nil, fmt.Errorf("baton-trayai: GetWorkspace failed: %w", err)
	}
	return []*v2.Entitlement{
			entitlement.NewAssignmentEntitlement(
				resource,
				RoleAssignmentEntitlement,
				entitlement.WithGrantableTo(userResourceType),
				entitlement.WithDescription(
					fmt.Sprintf(
						"Has the %s role in the tray.ai %s workspace",
						resource.DisplayName,
						workspace.Name,
					),
				),
				entitlement.WithDisplayName(
					fmt.Sprintf(
						"%s workspace %s role",
						workspace.Name,
						resource.DisplayName,
					),
				),
			),
		},
		"",
		nil,
		nil
}

func (r *roleBuilder) Grants(ctx context.Context, v2Resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	if v2Resource == nil || v2Resource.ParentResourceId == nil {
		return nil, "", nil, nil
	}

	users, err := r.client.ListWorkspaceUsers(ctx, client.ListParams{
		Cursor:      pToken.Token,
		First:       pToken.Size,
		WorkspaceID: v2Resource.ParentResourceId.Resource,
	})
	if err != nil {
		return nil, "", nil, fmt.Errorf("baton-trayai: ListWorkspaceUsers failed: %w", err)
	}

	rv := make([]*v2.Grant, 0, len(users.Elements))
	for _, user := range users.Elements {
		userID, err := resource.NewResourceID(userResourceType, user.ID)
		if err != nil {
			return nil, "", nil, fmt.Errorf("baton-trayai: failed to create resourceID for user: %w", err)
		}
		rv = append(rv, grant.NewGrant(v2Resource, RoleAssignmentEntitlement, userID))
	}
	return rv, users.Page.EndCursor, nil, nil
}

func newRoleBuilder(c *client.Client) *roleBuilder {
	return &roleBuilder{
		client: c,
	}
}
