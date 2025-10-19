// Copyright (c) 2025 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// setupTestConfig sets TFCTL_CFG to point to a test config file.
// Returns cleanup function that should be deferred.
func setupTestConfig(t *testing.T, testdataFile string) (cleanup func()) {
	t.Helper()

	// Get absolute path to testdata file
	configPath := filepath.Join("testdata", testdataFile)
	absPath, err := filepath.Abs(configPath)
	assert.NoError(t, err, "failed to get absolute path for test config")

	// Set TFCTL_CFG environment variable
	t.Setenv("TFCTL_CFG", absPath)

	// Reset the global Config to force reload
	Config = Type{}

	return func() {
		// Reset global Config
		Config = Type{}
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name      string
		testFile  string
		wantErr   bool
		checkFunc func(*testing.T, Type)
	}{
		{
			name:     "simple string values",
			testFile: "simple.yaml",
			wantErr:  false,
			checkFunc: func(t *testing.T, cfg Type) {
				assert.NotEmpty(t, cfg.Source)
				assert.Contains(t, cfg.Data, "region")
				assert.Equal(t, "us-east-1", cfg.Data["region"])
				assert.Equal(t, "my-bucket", cfg.Data["bucket"])
			},
		},
		{
			name:     "nested structure",
			testFile: "nested.yaml",
			wantErr:  false,
			checkFunc: func(t *testing.T, cfg Type) {
				backend, ok := cfg.Data["backend"].(map[string]interface{})
				assert.True(t, ok, "backend should be a map")
				s3, ok := backend["s3"].(map[string]interface{})
				assert.True(t, ok, "s3 should be a map")
				assert.Equal(t, "us-west-2", s3["region"])
				assert.Equal(t, "terraform-state", s3["bucket"])
			},
		},
		{
			name:     "mixed types",
			testFile: "mixed-types.yaml",
			wantErr:  false,
			checkFunc: func(t *testing.T, cfg Type) {
				assert.Equal(t, "test-project", cfg.Data["name"])
				assert.Equal(t, 1, cfg.Data["version"])
				assert.Equal(t, true, cfg.Data["enabled"])
				assert.Equal(t, 30.5, cfg.Data["timeout"])
				tags, ok := cfg.Data["tags"].([]interface{})
				assert.True(t, ok)
				assert.Len(t, tags, 2)
			},
		},
		{
			name:     "empty file",
			testFile: "empty.yaml",
			wantErr:  false,
			checkFunc: func(t *testing.T, cfg Type) {
				// Empty YAML unmarshals to nil map, which is acceptable
				assert.NotEmpty(t, cfg.Source, "should have a source path")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTestConfig(t, tt.testFile)
			defer cleanup()

			cfg, err := Load()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if tt.checkFunc != nil {
				tt.checkFunc(t, cfg)
			}
		})
	}
}

func TestLoad_NoConfigFile(t *testing.T) {
	// Set TFCTL_CFG to non-existent file
	t.Setenv("TFCTL_CFG", "/nonexistent/path/tfctl.yaml")
	Config = Type{}

	_, err := Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config file not found")
}

func TestLoad_TFCTL_CFG_IsDirectory(t *testing.T) {
	// Set TFCTL_CFG to a directory instead of a file
	t.Setenv("TFCTL_CFG", "testdata")
	Config = Type{}

	_, err := Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "points to a directory")
}

func TestGetString(t *testing.T) {
	tests := []struct {
		name         string
		testFile     string
		key          string
		defaultValue []string
		want         string
		wantErr      bool
	}{
		{
			name:     "simple string value",
			testFile: "simple.yaml",
			key:      "region",
			want:     "us-east-1",
			wantErr:  false,
		},
		{
			name:     "nested string value",
			testFile: "nested.yaml",
			key:      "backend.s3.region",
			want:     "us-west-2",
			wantErr:  false,
		},
		{
			name:         "missing key with default",
			testFile:     "simple.yaml",
			key:          "missing",
			defaultValue: []string{"default-value"},
			want:         "default-value",
			wantErr:      false,
		},
		{
			name:     "missing key without default",
			testFile: "simple.yaml",
			key:      "missing",
			want:     "",
			wantErr:  true,
		},
		{
			name:     "non-string value",
			testFile: "mixed-types.yaml",
			key:      "version",
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTestConfig(t, tt.testFile)
			defer cleanup()

			// Force load
			_, _ = Load()

			got, err := GetString(tt.key, tt.defaultValue...)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetInt(t *testing.T) {
	tests := []struct {
		name         string
		testFile     string
		key          string
		defaultValue []int
		want         int
		wantErr      bool
	}{
		{
			name:     "int value",
			testFile: "mixed-types.yaml",
			key:      "version",
			want:     1,
			wantErr:  false,
		},
		{
			name:     "float value converted to int",
			testFile: "mixed-types.yaml",
			key:      "timeout",
			want:     30,
			wantErr:  false,
		},
		{
			name:     "nested int value",
			testFile: "nested.yaml",
			key:      "backend.s3.max_retries",
			want:     5,
			wantErr:  false,
		},
		{
			name:         "missing key with default",
			testFile:     "simple.yaml",
			key:          "missing",
			defaultValue: []int{60},
			want:         60,
			wantErr:      false,
		},
		{
			name:     "missing key without default",
			testFile: "simple.yaml",
			key:      "missing",
			want:     0,
			wantErr:  true,
		},
		{
			name:     "non-int value",
			testFile: "simple.yaml",
			key:      "region",
			want:     0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTestConfig(t, tt.testFile)
			defer cleanup()

			// Force load
			_, _ = Load()

			got, err := GetInt(tt.key, tt.defaultValue...)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConfig_GetWithNamespace(t *testing.T) {
	cleanup := setupTestConfig(t, "nested.yaml")
	defer cleanup()

	// Load and set namespace
	_, err := Load()
	assert.NoError(t, err)

	// Test with namespace
	Config.Namespace = "backend.s3"

	// Should find namespaced value first
	val, err := Config.get("region")
	assert.NoError(t, err)
	assert.Equal(t, "us-west-2", val)

	val, err = Config.get("bucket")
	assert.NoError(t, err)
	assert.Equal(t, "terraform-state", val)

	// Change namespace
	Config.Namespace = "backend.local"
	val, err = Config.get("region")
	assert.NoError(t, err)
	assert.Equal(t, "us-east-1", val)

	val, err = Config.get("bucket")
	assert.NoError(t, err)
	assert.Equal(t, "local-bucket", val)
}

func TestConfig_GetNestedPath(t *testing.T) {
	cleanup := setupTestConfig(t, "deep-nested.yaml")
	defer cleanup()

	_, err := Load()
	assert.NoError(t, err)

	val, err := Config.get("level1.level2.level3.value")
	assert.NoError(t, err)
	assert.Equal(t, "deep-value", val)
}

func TestConfig_LazyLoad(t *testing.T) {
	cleanup := setupTestConfig(t, "simple.yaml")
	defer cleanup()

	// Don't explicitly call Load(), just use GetString
	// This should trigger lazy loading
	val, err := GetString("region")
	assert.NoError(t, err)
	assert.Equal(t, "us-east-1", val)
	assert.NotEmpty(t, Config.Source, "Config should be loaded")
}

func TestGetString_NamespaceFallback(t *testing.T) {
	cleanup := setupTestConfig(t, "namespace.yaml")
	defer cleanup()

	_, err := Load()
	assert.NoError(t, err)

	// Set namespace
	Config.Namespace = "backend.s3"

	// Should find namespaced value
	val, err := GetString("setting")
	assert.NoError(t, err)
	assert.Equal(t, "s3-value", val)

	// Should find specific namespaced value
	val, err = GetString("specific")
	assert.NoError(t, err)
	assert.Equal(t, "s3-specific", val)

	// Non-existent key should still error
	_, err = GetString("nonexistent")
	assert.Error(t, err)
}
