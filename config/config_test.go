package config_test

import (
	"dermify-api/config"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigure(t *testing.T) {
	type input struct {
		fileInput string
		envVar    string
	}

	testCases := []struct {
		name     string
		input    input
		expected config.Configuration
	}{
		{
			name: "file input with no env var",
			input: input{
				fileInput: `
environment: test123
port: 12345`,
			},
			expected: config.Configuration{
				Environment: "test123",
				Port:        12345,
			},
		},
		{
			name: "file input with env var override",
			input: input{
				fileInput: `
environment: test123
port: 12345`,
				envVar: "overridden",
			},
			expected: config.Configuration{
				Environment: "overridden",
				Port:        12345,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "cfg.yaml")
			err := os.WriteFile(path, []byte(tc.input.fileInput), 0o600)
			assert.NoError(t, err)

			if tc.input.envVar != "" {
				t.Setenv("OVERRIDE_ENVIRONMENT", tc.input.envVar)
			}

			actual := config.Configure(path)

			assert.Equal(t, tc.expected, *actual)
		})
	}
}
