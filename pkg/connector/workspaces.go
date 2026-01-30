package connector

import (
	"context"
	"fmt"
	"strings"
	"sync"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	sdkResource "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-trayai/pkg/connector/client"
)

// workspaceResource is used to create a new connector resource for a tray.ai workspace.
func workspaceResource(ws client.Element) (*v2.Resource, error) {
	return sdkResource.NewGroupResource(
		ws.Name,
		workspaceResourceType,
		ws.ID,
		[]sdkResource.GroupTraitOption{
			sdkResource.WithGroupProfile(
				map[string]interface{}{
					"workspace_id":               ws.ID,
					"workspace_name":             ws.Name,
					"workspace_type":             ws.Type,
					"workspace_description":      ws.Description,
					"workspace_monthlyTaskLimit": ws.MonthlyTaskLimit,
				},
			),
		},
		sdkResource.WithAnnotation(
			&v2.ChildResourceType{ResourceTypeId: roleResourceType.Id},
		),
	)
}

type workspaceBuilder struct {
	mu             sync.Mutex
	client         *client.Client
	workspaceUsers map[string][]*client.User // map[workspaceID]users
}

var _ connectorbuilder.ResourceSyncerV2 = &workspaceBuilder{}
var _ connectorbuilder.ResourceProvisionerV2 = &workspaceBuilder{}

// ResourceType returns the workspace resource type.
func (w *workspaceBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return workspaceResourceType
}

// List returns all the workspaces from the database as resource objests.
func (w *workspaceBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, opts sdkResource.SyncOpAttrs) ([]*v2.Resource, *sdkResource.SyncOpResults, error) {
	var (
		workspaces []*v2.Resource
	)

	pToken := opts.PageToken
	resp, err := w.client.ListWorkspaces(ctx, client.ListParams{
		Cursor: pToken.Token,
		First:  pToken.Size,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("baton-trayai: ListWorkspaces failed: %w", err)
	}

	for _, workspace := range resp.Elements {
		vWorkspace, err := workspaceResource(workspace)
		if err != nil {
			return nil, nil, fmt.Errorf("baton-trayai: cannot create connector resource: %w", err)
		}
		workspaces = append(workspaces, vWorkspace)
	}

	if !resp.Page.HasNextPage {
		return workspaces, nil, nil
	}
	return workspaces, &sdkResource.SyncOpResults{NextPageToken: resp.Page.EndCursor}, nil
}

// Entitlements returns workspace entitlements from the database as resource objects.
func (w *workspaceBuilder) Entitlements(ctx context.Context, resource *v2.Resource, opts sdkResource.SyncOpAttrs) ([]*v2.Entitlement, *sdkResource.SyncOpResults, error) {
	var ents []*v2.Entitlement
	pToken := opts.PageToken
	resp, err := w.client.ListWorkspaceRoles(ctx, client.ListParams{
		Cursor:      pToken.Token,
		First:       pToken.Size,
		WorkspaceID: resource.Id.Resource,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("baton-trayai: ListWorkspaceRoles failed: %w", err)
	}

	for _, role := range resp.Elements {
		assignmentOptions := []entitlement.EntitlementOption{
			entitlement.WithGrantableTo(userResourceType, workspaceResourceType),
			entitlement.WithDisplayName(fmt.Sprintf("%s workspace %s", resource.DisplayName, role.Name)),
			entitlement.WithDescription(fmt.Sprintf("%s access to %s in tray.ai", role.Name, resource.DisplayName)),
		}

		ent := &v2.Entitlement{
			Id:          entitlement.NewEntitlementID(resource, role.ID),
			DisplayName: role.Name,
			Slug:        role.Name,
			Purpose:     v2.Entitlement_PURPOSE_VALUE_ASSIGNMENT,
			Resource:    resource,
		}
		for _, entitlementOption := range assignmentOptions {
			entitlementOption(ent)
		}
		ents = append(ents, ent)
	}

	if !resp.Page.HasNextPage {
		return ents, nil, nil
	}
	return ents, &sdkResource.SyncOpResults{NextPageToken: resp.Page.EndCursor}, nil
}

// Grants returns grants for workspace.
func (w *workspaceBuilder) Grants(ctx context.Context, r *v2.Resource, opts sdkResource.SyncOpAttrs) ([]*v2.Grant, *sdkResource.SyncOpResults, error) {
	pToken := opts.PageToken
	params := client.ListParams{
		Cursor: pToken.Token,
		First:  pToken.Size,
	}

	isOrg, err := w.isOrganization(ctx, r.Id.Resource)
	if err != nil {
		return nil, nil, fmt.Errorf("baton-trayai: isOrganization failed: %w", err)
	}

	if !isOrg {
		params.WorkspaceID = r.Id.Resource
	}

	resp, err := w.client.ListWorkspaceUsers(ctx, params)
	if err != nil {
		return nil, nil, fmt.Errorf("baton-trayai: ListWorkspaceUsers failed: %w", err)
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	grants := make([]*v2.Grant, 0, len(resp.Elements))
	for _, userID := range resp.Elements {
		user, err := w.client.GetWorkspaceUser(ctx, userID.ID, r.Id.Resource)
		if err != nil {
			return nil, nil, fmt.Errorf("baton-trayai: GetWorkspaceUser failed: %w", err)
		}

		userResource, err := sdkResource.NewResourceID(userResourceType, user.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("baton-trayai: cannot create connector resource id: %w", err)
		}

		grants = append(grants, grant.NewGrant(
			r,
			user.Role.ID,
			userResource,
		))

		workspaceUsers := w.workspaceUsers[r.Id.Resource]
		workspaceUsers = append(workspaceUsers, user)
		w.workspaceUsers[r.Id.Resource] = workspaceUsers
	}

	if !resp.Page.HasNextPage {
		return grants, nil, nil
	}
	return grants, &sdkResource.SyncOpResults{NextPageToken: resp.Page.EndCursor}, nil
}

func (w *workspaceBuilder) getWorkspaceUsers(ctx context.Context, workspaceID string) ([]*client.User, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	workspaceUsers, ok := w.workspaceUsers[workspaceID]
	if ok {
		return workspaceUsers, nil
	}

	params := client.ListParams{
		Cursor: "",
	}
	isOrg, err := w.isOrganization(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("baton-trayai: isOrganization failed: %w", err)
	}

	if !isOrg {
		params.WorkspaceID = workspaceID
	}

	for {
		resp, err := w.client.ListWorkspaceUsers(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("baton-trayai: ListWorkspaceUsers failed: %w", err)
		}

		for _, element := range resp.Elements {
			user, err := w.client.GetWorkspaceUser(ctx, element.ID, params.WorkspaceID)
			if err != nil {
				return nil, fmt.Errorf("baton-trayai: GetWorkspaceUser failed: %w", err)
			}
			workspaceUsers = append(workspaceUsers, user)
		}

		w.workspaceUsers[workspaceID] = workspaceUsers
		if resp.Page.EndCursor == "" {
			break
		}

		if params.Cursor == resp.Page.EndCursor {
			return nil, fmt.Errorf("baton-trayai: current cursor shouldn't be equal to endCursor")
		}
		params.Cursor = resp.Page.EndCursor
	}

	w.workspaceUsers[workspaceID] = workspaceUsers
	return workspaceUsers, nil
}
func newWorkspaceBuild(c *client.Client) *workspaceBuilder {
	return &workspaceBuilder{
		client:         c,
		workspaceUsers: make(map[string][]*client.User),
	}
}

func (w *workspaceBuilder) isOrganization(ctx context.Context, workspaceID string) (bool, error) {
	resp, err := w.client.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return false, fmt.Errorf("baton-trayai: GetWorkspace failed: %w", err)
	}
	return resp.Type == "Organization", nil
}

func (w *workspaceBuilder) Grant(ctx context.Context, principal *v2.Resource, entitlement *v2.Entitlement) ([]*v2.Grant, annotations.Annotations, error) {
	if principal == nil || principal.Id == nil {
		return nil, nil, fmt.Errorf("baton-trayai: principal is nil")
	}

	if entitlement == nil || entitlement.Resource == nil || entitlement.Resource.Id == nil {
		return nil, nil, fmt.Errorf("baton-trayai: entitlement resource is nil")
	}

	if principal.Id.ResourceType != userResourceType.Id {
		return nil, nil, fmt.Errorf("baton-trayai: only users can be assigned to a workspace")
	}

	resourceIDs := strings.Split(entitlement.Id, ":")
	if len(resourceIDs) != 3 {
		return nil, nil, fmt.Errorf("baton-trayai: invalid resource ID: %s", entitlement.Id)
	}

	params := client.AddUserToWorkspaceParams{
		WorkspaceID: entitlement.Resource.Id.Resource,
		UserID:      principal.Id.Resource,
		RoleID:      resourceIDs[2],
	}
	err := w.client.AddUserToWorkspace(ctx, params)
	if err != nil {
		return nil, nil, fmt.Errorf("baton-trayai: AddUserToWorkspace failed: %w", err)
	}
	return []*v2.Grant{
		grant.NewGrant(entitlement.Resource, entitlement.Id, principal.Id),
	}, nil, nil
}

func (w *workspaceBuilder) Revoke(ctx context.Context, grant *v2.Grant) (annotations.Annotations, error) {
	if grant == nil || grant.Principal == nil || grant.Principal.Id == nil {
		return nil, fmt.Errorf("baton-trayai: grant is nil")
	}

	if grant.Principal.Id.ResourceType != userResourceType.Id {
		return nil, fmt.Errorf("baton-trayai: only users can have roles revoked")
	}

	if grant.Entitlement == nil || grant.Entitlement.Resource == nil {
		return nil, fmt.Errorf("baton-trayai: entitlement resource is empty")
	}

	params := client.SetOrDeleteWorkspaceRoleParams{
		UserID: grant.Principal.Id.Resource,
	}

	workspaceID := grant.Entitlement.Resource.Id.Resource
	isOrg, err := w.isOrganization(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("baton-trayai: isOrganization failed: %w", err)
	}

	if !isOrg {
		params.WorkspaceID = workspaceID
		return nil, w.client.RemoveWorkspaceUser(ctx, params)
	}

	return nil, w.client.RemoveOrganizationUser(ctx, params)
}
