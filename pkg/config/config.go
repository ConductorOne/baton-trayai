package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	AuthorizationTokenField = field.StringField(
		"auth-token",
		field.WithDescription("auth-token for authenticating with the service"),
		field.WithRequired(true),
	)
)

//go:generate go run ./gen
var Config = field.NewConfiguration([]field.SchemaField{AuthorizationTokenField})
