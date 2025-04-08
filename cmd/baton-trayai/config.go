package main

import (
	"fmt"

	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/spf13/viper"
)

var (
	AuthorizationTokenField = field.StringField(
		"auth-token",
		field.WithDescription("auth-token for authenticating with the service"),
		field.WithRequired(true),
	)

	// ConfigurationFields defines the external configuration required for the
	// connector to run. Note: these fields can be marked as optional or
	// required.
	ConfigurationFields = []field.SchemaField{AuthorizationTokenField}

	// FieldRelationships defines relationships between the fields listed in
	// ConfigurationFields that can be automatically validated. For example, a
	// username and password can be required together, or an access token can be
	// marked as mutually exclusive from the username password pair.
	FieldRelationships = []field.SchemaFieldRelationship{}
)

// ValidateConfig is run after the configuration is loaded, and should return an
// error if it isn't valid. Implementing this function is optional, it only
// needs to perform extra validations that cannot be encoded with configuration
// parameters.
func ValidateConfig(v *viper.Viper) error {
	authToken := v.GetString(AuthorizationTokenField.FieldName)
	if len(authToken) == 0 {
		return fmt.Errorf("required field 'auth-token' is missing")
	}
	return nil
}
