package main

import (
	"testing"

	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/conductorone/baton-sdk/pkg/test"
)

func TestConfigs(t *testing.T) {
	configurationSchema := field.NewConfiguration(
		ConfigurationFields,
		FieldRelationships...,
	)

	testCases := []test.TestCase{
		{
			Configs: map[string]string{
				"auth-token": "abc123",
			},
			IsValid: true,
		},
		{
			Configs: map[string]string{
				"auth-token": "",
			},
		},
	}

	test.ExerciseTestCases(t, configurationSchema, ValidateConfig, testCases)
}
