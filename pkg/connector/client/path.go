package client

// For API documentation, see: https://developer.tray.ai/openapi/trayapi/tag/overview/
const (
	basePath                  = "https://api.tray.io"
	listUsersPath             = "/core/v1/users"
	listWorkspacesPath        = "/core/v1/workspaces"
	listWorkspaceRolesPath    = "/core/v1/workspaces/%s/roles"
	listOrganizationRolesPath = "/core/v1/roles"
	listWorkspaceUsersPath    = "/core/v1/workspaces/%s/users"

	getUserPath             = "/core/v1/users/%s"
	getOrganizationRolePath = "/core/v1/users/%s/roles"
	getWorkspacePath        = "/core/v1/workspaces/%s"
	getWorkspaceUserPath    = "/core/v1/workspaces/%s/users/%s"
)
