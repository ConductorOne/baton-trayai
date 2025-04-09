package client

// TrayType is the Tray.ai's type.
type TrayType int

const (
	Embeded TrayType = iota + 1
	Organiaztion
	Personal
	PersonalExternal
	Shared
)

// User is the Tray.ai user.
type User struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Type             TrayType `json:"type"`
	Description      string   `json:"description"`
	MonthlyTaskLimit int64    `json:"monthlyTaskLimit"`
}

type PageInfo struct {
	StartCursor     string `json:"startCursor"`
	EndCursor       string `json:"endCursor"`
	HasNextPage     bool   `json:"hasNextPage"`
	HasPreviousPage bool   `json:"hasPreviousPage"`
}
