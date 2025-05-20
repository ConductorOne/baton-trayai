package config

import (
	"testing"

	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/stretchr/testify/assert"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Trayai
		wantErr bool
	}{
		{
			name: "valid config - with auth token",
			config: &Trayai{
				AuthToken: "abc123",
			},
			wantErr: false,
		},
		{
			name: "invalid config - empty auth token",
			config: &Trayai{
				AuthToken: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := field.Validate(Config, tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), "field auth-token of type string is marked as required but it has a zero-value")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
