package client

// Element is the Tray.ai element.
type Element struct {
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

type User struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	AccountType string `json:"accountType"`
	Role        Role   `json:"role"`
}

type Role struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Workspace has the same fields as Element. It's created to avoid ambiguity.
type Workspace Element
