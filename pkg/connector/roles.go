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
	client   *client.Client
	wbuilder *workspaceBuilder
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

	ent := []*v2.Entitlement{
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
	}
	return ent, "", nil, nil
}

func (r *roleBuilder) Grants(ctx context.Context, v2Resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	if v2Resource == nil || v2Resource.ParentResourceId == nil {
		return nil, "", nil, nil
	}

	workspaceUsers, err := r.wbuilder.getWorkspaceUsers(ctx, v2Resource.ParentResourceId.Resource)
	if err != nil {
		return nil, "", nil, err
	}

	rv := make([]*v2.Grant, 0, len(workspaceUsers))
	for _, user := range workspaceUsers {
		if v2Resource.Id.Resource != user.Role.ID {
			continue
		}

		userID, err := resource.NewResourceID(userResourceType, user.ID)
		if err != nil {
			return nil, "", nil, fmt.Errorf("baton-trayai: failed to create resourceID for user: %w", err)
		}
		rv = append(rv, grant.NewGrant(v2Resource, RoleAssignmentEntitlement, userID))
	}
	return rv, "", nil, nil
}

func (r *roleBuilder) Grant(ctx context.Context, principal *v2.Resource, entitlement *v2.Entitlement) (annotations.Annotations, error) {
	if principal == nil || principal.Id == nil {
		return nil, fmt.Errorf("baton-trayai: principal is nil")
	}

	if entitlement == nil || entitlement.Resource == nil || entitlement.Resource.Id == nil {
		return nil, fmt.Errorf("baton-trayai: entitlement resource is nil")
	}

	if principal.Id.ResourceType != userResourceType.Id {
		return nil, fmt.Errorf("baton-trayai: only users can be assigned a role")
	}

	if entitlement.Resource.ParentResourceId == nil {
		return nil, fmt.Errorf("baton-trayai: entitlement resource has no parent resource id")
	}

	return nil, r.client.SetWorkspaceRole(ctx, client.SetOrDeleteWorkspaceRoleParams{
		WorkspaceID: entitlement.Resource.ParentResourceId.Resource,
		UserID:      principal.Id.Resource,
		RoleID:      entitlement.Resource.Id.Resource,
	})
}

func (r *roleBuilder) Revoke(ctx context.Context, grant *v2.Grant) (annotations.Annotations, error) {
	if grant == nil || grant.Principal == nil || grant.Principal.Id == nil {
		return nil, fmt.Errorf("baton-trayai: grant is nil")
	}

	if grant.Principal.Id.ResourceType != userResourceType.Id {
		return nil, fmt.Errorf("baton-trayai: only users can have roles revoked")
	}

	if grant.Entitlement == nil || grant.Entitlement.Resource == nil || grant.Entitlement.Resource.ParentResourceId == nil {
		return nil, fmt.Errorf("baton-trayai: entitlement resource has no parent resource id")
	}

	return nil, r.client.RemoveWorkspaceUser(ctx, client.SetOrDeleteWorkspaceRoleParams{
		WorkspaceID: grant.Entitlement.Resource.ParentResourceId.Resource,
		UserID:      grant.Principal.Id.Resource,
	})
}

func newRoleBuilder(c *client.Client, wb *workspaceBuilder) *roleBuilder {
	return &roleBuilder{
		client:   c,
		wbuilder: wb,
	}
}
