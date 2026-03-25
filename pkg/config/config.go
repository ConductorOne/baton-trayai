package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	AuthorizationTokenField = field.StringField(
		"auth-token",
		field.WithDisplayName("Authentication token"),
		field.WithDescription("auth-token for authenticating with the service"),
		field.WithIsSecret(true),
		field.WithRequired(true),
	)
	BaseURLField = field.StringField(
		"base-url",
		field.WithDisplayName("Base URL"),
		field.WithDescription("Override the Tray.ai API URL (for testing)"),
		field.WithDefaultValue("https://api.tray.io"),
		field.WithHidden(true),
		field.WithExportTarget(field.ExportTargetCLIOnly),
	)
)

//go:generate go run ./gen
var Config = field.NewConfiguration(
	[]field.SchemaField{AuthorizationTokenField, BaseURLField},
	field.WithConnectorDisplayName("Tray.ai"),
	field.WithHelpUrl("/docs/baton/trayai"),
	field.WithIconUrl("/static/app-icons/trayai.svg"),
)
