package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
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
		resource.WithAnnotation(
			&v2.ChildResourceType{ResourceTypeId: userResourceType.Id},
		),
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
func (w *workspaceBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	// TODO. BB-451. We should pull the workspace roles before Entitlements Implementation.
	return nil, "", nil, nil
}

// Grants returns grants for workspace.
func (w *workspaceBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	// TODO. BB-451. We should pull the workspace roles before Grants Implementation.
	return nil, "", nil, nil
}

func newWorkspaceBuild(c *client.Client) *workspaceBuilder {
	return &workspaceBuilder{
		client: c,
	}
}
