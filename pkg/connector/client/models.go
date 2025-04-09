package client

// User is the Tray.ai user.
type User struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Type             string `json:"type"`
	Description      string `json:"description"`
	MonthlyTaskLimit int64  `json:"monthlyTaskLimit"`
}

type PageInfo struct {
	StartCursor     string `json:"startCursor"`
	EndCursor       string `json:"endCursor"`
	HasNextPage     bool   `json:"hasNextPage"`
	HasPreviousPage bool   `json:"hasPreviousPage"`
}

// Workspace has the same fields as a User. It was created to avoid ambiguity.
type Workspace User
