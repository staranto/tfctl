// Copyright (c) 2025 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

// no-cloc
package driller

import (
	"testing"
)

func TestDriller(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		path        string
		expectedStr string
		isNil       bool
		isArray     bool
	}{
		// Simple key tests
		{
			name:        "simple string key",
			json:        `{"name": "test"}`,
			path:        "name",
			expectedStr: "test",
		},
		{
			name:        "simple number key",
			json:        `{"count": 42}`,
			path:        "count",
			expectedStr: "42",
		},
		{
			name:        "simple boolean key true",
			json:        `{"active": true}`,
			path:        "active",
			expectedStr: "true",
		},
		{
			name:        "simple boolean key false",
			json:        `{"active": false}`,
			path:        "active",
			expectedStr: "false",
		},
		{
			name:  "simple null key",
			json:  `{"value": null}`,
			path:  "value",
			isNil: true,
		},
		// Nested object tests
		{
			name:        "nested single level",
			json:        `{"user": {"name": "alice"}}`,
			path:        "user.name",
			expectedStr: "alice",
		},
		{
			name:        "nested multiple levels",
			json:        `{"root": {"sub": {"deep": "value"}}}`,
			path:        "root.sub.deep",
			expectedStr: "value",
		},
		// Array access tests - single element array
		{
			name:        "single element array returns element",
			json:        `{"items": ["only"]}`,
			path:        "items",
			expectedStr: "only",
		},
		{
			name:        "single element array of objects drills through",
			json:        `{"items": [{"id": "first"}]}`,
			path:        "items.id",
			expectedStr: "first",
		},
		// Array access tests - multi element array (returns array)
		{
			name:    "multi element array returns array",
			json:    `{"items": ["first", "second"]}`,
			path:    "items",
			isArray: true,
		},
		// Array access tests - explicit index
		{
			name:        "array with explicit index 0",
			json:        `{"items": ["first", "second", "third"]}`,
			path:        "items[0]",
			expectedStr: "first",
		},
		{
			name:        "array with explicit index 1",
			json:        `{"items": ["first", "second", "third"]}`,
			path:        "items[1]",
			expectedStr: "second",
		},
		{
			name:        "array with explicit index 2",
			json:        `{"items": ["first", "second", "third"]}`,
			path:        "items[2]",
			expectedStr: "third",
		},
		{
			name:        "array with last valid index",
			json:        `{"items": [10, 20, 30]}`,
			path:        "items[2]",
			expectedStr: "30",
		},
		// Array inside nested objects
		{
			name:        "nested object with array access explicit index",
			json:        `{"user": {"tags": ["admin", "user"]}}`,
			path:        "user.tags[0]",
			expectedStr: "admin",
		},
		{
			name:        "nested object with array access second element",
			json:        `{"user": {"tags": ["admin", "user"]}}`,
			path:        "user.tags[1]",
			expectedStr: "user",
		},
		// Array of objects
		{
			name:        "single element array of objects drills through property",
			json:        `{"users": [{"id": 1, "name": "alice"}]}`,
			path:        "users.name",
			expectedStr: "alice",
		},
		{
			name:        "array of objects with explicit index",
			json:        `{"users": [{"id": 1, "name": "alice"}, {"id": 2, "name": "bob"}]}`,
			path:        "users[1].name",
			expectedStr: "bob",
		},
		{
			name:        "array of objects with multiple levels",
			json:        `{"org": {"teams": [{"name": "backend", "lead": {"name": "alice"}}]}}`,
			path:        "org.teams[0].lead.name",
			expectedStr: "alice",
		},
		// Key names with special characters
		{
			name:        "key with hyphen",
			json:        `{"my-key": "value"}`,
			path:        "my-key",
			expectedStr: "value",
		},
		{
			name:        "key with underscore",
			json:        `{"my_key": "value"}`,
			path:        "my_key",
			expectedStr: "value",
		},
		{
			name:        "key with numbers",
			json:        `{"key123": "value"}`,
			path:        "key123",
			expectedStr: "value",
		},
		// Error cases - invalid paths
		{
			name:  "nonexistent key returns empty result",
			json:  `{"name": "test"}`,
			path:  "missing",
			isNil: true,
		},
		{
			name:  "invalid array index returns empty result",
			json:  `{"items": ["a", "b"]}`,
			path:  "items[10]",
			isNil: true,
		},
		{
			name:  "nested missing key returns empty result",
			json:  `{"user": {"name": "alice"}}`,
			path:  "user.missing",
			isNil: true,
		},
		// Empty structures
		{
			name:  "empty object returns empty result for any key",
			json:  `{}`,
			path:  "any",
			isNil: true,
		},
		{
			name:  "empty array with index returns empty result",
			json:  `{"items": []}`,
			path:  "items[0]",
			isNil: true,
		},
		// Complex real-world-like structures
		{
			name:        "terraform resource attributes",
			json:        `{"attributes": {"aws_s3_bucket": {"id": "my-bucket"}}}`,
			path:        "attributes.aws_s3_bucket.id",
			expectedStr: "my-bucket",
		},
		{
			name:        "terraform state resources array with explicit index",
			json:        `{"resources": [{"type": "aws_instance", "name": "web"}, {"type": "aws_vpc", "name": "main"}]}`,
			path:        "resources[0].type",
			expectedStr: "aws_instance",
		},
		{
			name:        "terraform state second resource",
			json:        `{"resources": [{"type": "aws_instance", "name": "web"}, {"type": "aws_vpc", "name": "main"}]}`,
			path:        "resources[1].name",
			expectedStr: "main",
		},
		{
			name:        "deeply nested terraform-like structure",
			json:        `{"root_module": {"resources": [{"type": "aws_subnet", "instances": [{"attributes": {"id": "subnet-123"}}]}]}}`,
			path:        "root_module.resources[0].instances[0].attributes.id",
			expectedStr: "subnet-123",
		},
		// Dotted paths in strings
		{
			name:        "path traversal with repeated keys",
			json:        `{"a": {"b": {"c": {"value": "found"}}}}`,
			path:        "a.b.c.value",
			expectedStr: "found",
		},
		// Array element then nested access
		{
			name:        "array element explicit index then nested access",
			json:        `{"data": [{"nested": {"value": "test"}}]}`,
			path:        "data[0].nested.value",
			expectedStr: "test",
		},
		// Multi-element array access without index
		{
			name:    "multi element array access without index returns array",
			json:    `{"data": [{"value": "first"}, {"value": "second"}]}`,
			path:    "data",
			isArray: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Driller(tt.json, tt.path)

			if tt.isNil {
				// Result should not exist or be null
				if result.Exists() && result.Type.String() != "Null" {
					t.Errorf("Expected nil/empty result but got: %v", result.Value())
				}
				return
			}

			if !result.Exists() {
				t.Errorf("Expected result but got nil/empty")
				return
			}

			if tt.isArray {
				if !result.IsArray() {
					t.Errorf("Expected array but got: %v (type: %T)", result.Value(), result.Value())
				}
				return
			}

			val := result.String()
			if val != tt.expectedStr {
				t.Errorf("Expected %q but got %q", tt.expectedStr, val)
			}
		})
	}
}

// BenchmarkDriller benchmarks the Driller function with various path depths.
func BenchmarkDriller(b *testing.B) {
	tests := []struct {
		name string
		json string
		path string
	}{
		{
			name: "simple",
			json: `{"name": "test"}`,
			path: "name",
		},
		{
			name: "nested_2_levels",
			json: `{"user": {"name": "alice"}}`,
			path: "user.name",
		},
		{
			name: "nested_4_levels",
			json: `{"root": {"sub": {"deep": {"value": "test"}}}}`,
			path: "root.sub.deep.value",
		},
		{
			name: "array_access",
			json: `{"items": [1, 2, 3]}`,
			path: "items[1]",
		},
		{
			name: "array_of_objects",
			json: `{"users": [{"id": 1, "name": "alice"}, {"id": 2, "name": "bob"}]}`,
			path: "users[0].name",
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				Driller(tt.json, tt.path)
			}
		})
	}
}
