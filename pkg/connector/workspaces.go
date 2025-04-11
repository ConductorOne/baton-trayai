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

// workspaceResource is used to create a new connector resource for a tray.ai workspace.
func workspaceResource(ws client.Element) (*v2.Resource, error) {
	return resource.NewGroupResource(
		ws.Name,
		workspaceResourceType,
		ws.ID,
		[]resource.GroupTraitOption{
			resource.WithGroupProfile(
				map[string]interface{}{
					"workspace_id":               ws.ID,
					"workspace_name":             ws.Name,
					"workspace_type":             ws.Type,
					"workspace_description":      ws.Description,
					"workspace_monthlyTaskLimit": ws.MonthlyTaskLimit,
				},
			),
		},
	)
}

type workspaceBuilder struct {
	client *client.Client
}

// ResourceType returns the workspace resource type.
func (w *workspaceBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return workspaceResourceType
}

// List returns all the workspaces from the database as resource objests.
func (w *workspaceBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var (
		workspaces []*v2.Resource
	)

	resp, err := w.client.ListWorkspaces(ctx, client.ListParams{
		Cursor: pToken.Token,
		First:  pToken.Size,
	})
	if err != nil {
		return nil, "", nil, fmt.Errorf("baton-trayai: ListWorkspaces failed: %w", err)
	}

	for _, workspace := range resp.Elements {
		vWorkspace, err := workspaceResource(workspace)
		if err != nil {
			return nil, "", nil, fmt.Errorf("baton-trayai: cannot create connector resource: %w", err)
		}
		workspaces = append(workspaces, vWorkspace)
	}

	if !resp.Page.HasNextPage {
		return workspaces, "", nil, nil
	}
	return workspaces, resp.Page.EndCursor, nil, nil
}

// Entitlements returns workspace entitlements from the database as resource objects.
func (w *workspaceBuilder) Entitlements(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var ents []*v2.Entitlement
	resp, err := w.client.ListWorkspaceRoles(ctx, client.ListParams{
		Cursor:      pToken.Token,
		First:       pToken.Size,
		WorkspaceID: resource.Id.Resource,
	})
	if err != nil {
		return nil, "", nil, fmt.Errorf("baton-trayai: ListWorkspaceRoles failed: %w", err)
	}

	for _, role := range resp.Elements {
		assignmentOptions := []entitlement.EntitlementOption{
			entitlement.WithGrantableTo(userResourceType, workspaceResourceType),
			entitlement.WithDisplayName(fmt.Sprintf("%s workspace %s", resource.DisplayName, role.Name)),
			entitlement.WithDescription(fmt.Sprintf("%s access to %s in tray.ai", role.Name, resource.DisplayName)),
		}

		ents = append(ents, entitlement.NewAssignmentEntitlement(
			resource,
			role.Name,
			assignmentOptions...,
		))
	}

	if !resp.Page.HasNextPage {
		return ents, "", nil, nil
	}
	return ents, resp.Page.EndCursor, nil, nil
}

// Grants returns grants for workspace.
func (w *workspaceBuilder) Grants(ctx context.Context, r *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	resp, err := w.client.ListWorkspaceUsers(ctx, client.ListParams{
		Cursor:      pToken.Token,
		First:       pToken.Size,
		WorkspaceID: r.Id.Resource,
	})
	if err != nil {
		return nil, "", nil, fmt.Errorf("baton-trayai: ListWorkspaceUsers failed: %w", err)
	}

	grants := make([]*v2.Grant, 0, len(resp.Elements))
	for _, userID := range resp.Elements {
		user, err := w.client.GetUser(ctx, userID.ID)
		if err != nil {
			return nil, "", nil, fmt.Errorf("baton-trayai: GetUser %v failed: %w,", userID, err)
		}

		userResource, err := resource.NewResourceID(userResourceType, user.ID)
		if err != nil {
			return nil, "", nil, fmt.Errorf("baton-trayai: cannot crete connector resource id: %w", err)
		}

		grants = append(grants, grant.NewGrant(
			r,
			user.Role.Name,
			userResource,
		))
	}

	if !resp.Page.HasNextPage {
		return grants, "", nil, nil
	}
	return grants, resp.Page.EndCursor, nil, nil
}

func newWorkspaceBuild(c *client.Client) *workspaceBuilder {
	return &workspaceBuilder{
		client: c,
	}
}
