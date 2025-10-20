// Copyright (c) 2025 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0
// no-cloc

package filters

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/staranto/tfctlgo/internal/attrs"
)

func TestBuildFilters(t *testing.T) {
	tests := []struct {
		name      string
		spec      string
		delimiter string
		want      []Filter
		wantCount int
	}{
		{
			name:      "empty spec",
			spec:      "",
			wantCount: 0,
		},
		{
			name:      "single exact match filter",
			spec:      "name=my-resource",
			wantCount: 1,
			want: []Filter{
				{Key: "name", Operand: "=", Target: "my-resource", Negate: false},
			},
		},
		{
			name:      "prefix match filter",
			spec:      "type^aws_",
			wantCount: 1,
			want: []Filter{
				{Key: "type", Operand: "^", Target: "aws_", Negate: false},
			},
		},
		{
			name:      "regex match filter",
			spec:      "tags~^env-",
			wantCount: 1,
			want: []Filter{
				{Key: "tags", Operand: "~", Target: "^env-", Negate: false},
			},
		},
		{
			name:      "negated exact match",
			spec:      "name!=test",
			wantCount: 1,
			want: []Filter{
				{Key: "name", Operand: "=", Target: "test", Negate: true},
			},
		},
		{
			name:      "negated prefix match",
			spec:      "type!^aws_",
			wantCount: 1,
			want: []Filter{
				{Key: "type", Operand: "^", Target: "aws_", Negate: true},
			},
		},
		{
			name:      "multiple filters",
			spec:      "name=test,type^aws_",
			wantCount: 2,
			want: []Filter{
				{Key: "name", Operand: "=", Target: "test", Negate: false},
				{Key: "type", Operand: "^", Target: "aws_", Negate: false},
			},
		},
		{
			name:      "greater than numeric",
			spec:      "count>5",
			wantCount: 1,
			want: []Filter{
				{Key: "count", Operand: ">", Target: "5", Negate: false},
			},
		},
		{
			name:      "less than numeric",
			spec:      "count<10",
			wantCount: 1,
			want: []Filter{
				{Key: "count", Operand: "<", Target: "10", Negate: false},
			},
		},
		{
			name:      "contains operand",
			spec:      "name@test",
			wantCount: 1,
			want: []Filter{
				{Key: "name", Operand: "@", Target: "test", Negate: false},
			},
		},
		{
			name:      "regex operand",
			spec:      "name/^test.*",
			wantCount: 1,
			want: []Filter{
				{Key: "name", Operand: "/", Target: "^test.*", Negate: false},
			},
		},
		{
			name:      "invalid filter skipped",
			spec:      "name=test,invalid-filter,type^aws_",
			wantCount: 2,
			want: []Filter{
				{Key: "name", Operand: "=", Target: "test", Negate: false},
				{Key: "type", Operand: "^", Target: "aws_", Negate: false},
			},
		},
		{
			name:      "custom delimiter",
			spec:      "name=test|type^aws_",
			delimiter: "|",
			wantCount: 2,
			want: []Filter{
				{Key: "name", Operand: "=", Target: "test", Negate: false},
				{Key: "type", Operand: "^", Target: "aws_", Negate: false},
			},
		},
		{
			name:      "key with dots",
			spec:      "backend.s3.region=us-west-2",
			wantCount: 1,
			want: []Filter{
				{Key: "backend.s3.region", Operand: "=", Target: "us-west-2", Negate: false},
			},
		},
		{
			name:      "empty target",
			spec:      "name=",
			wantCount: 1,
			want: []Filter{
				{Key: "name", Operand: "=", Target: "", Negate: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.delimiter != "" {
				t.Setenv("TFCTL_FILTER_DELIM", tt.delimiter)
			}

			got := BuildFilters(tt.spec)
			assert.Len(t, got, tt.wantCount)
			if tt.want != nil {
				for i, filter := range tt.want {
					assert.Equal(t, filter.Key, got[i].Key)
					assert.Equal(t, filter.Operand, got[i].Operand)
					assert.Equal(t, filter.Target, got[i].Target)
					assert.Equal(t, filter.Negate, got[i].Negate)
				}
			}
		})
	}
}

func TestCheckStringOperand(t *testing.T) {
	tests := []struct {
		name   string
		value  string
		filter Filter
		want   bool
	}{
		{
			name:   "exact match true",
			value:  "test",
			filter: Filter{Operand: "=", Target: "test", Negate: false},
			want:   true,
		},
		{
			name:   "exact match false",
			value:  "test",
			filter: Filter{Operand: "=", Target: "other", Negate: false},
			want:   false,
		},
		{
			name:   "negated exact match true",
			value:  "test",
			filter: Filter{Operand: "=", Target: "other", Negate: true},
			want:   true,
		},
		{
			name:   "negated exact match false",
			value:  "test",
			filter: Filter{Operand: "=", Target: "test", Negate: true},
			want:   false,
		},
		{
			name:   "prefix match true",
			value:  "aws_instance",
			filter: Filter{Operand: "^", Target: "aws_", Negate: false},
			want:   true,
		},
		{
			name:   "prefix match false",
			value:  "gcp_instance",
			filter: Filter{Operand: "^", Target: "aws_", Negate: false},
			want:   false,
		},
		{
			name:   "case insensitive match true",
			value:  "TEST",
			filter: Filter{Operand: "~", Target: "test", Negate: false},
			want:   true,
		},
		{
			name:   "case insensitive match false",
			value:  "testing",
			filter: Filter{Operand: "~", Target: "test", Negate: false},
			want:   false,
		},
		{
			name:   "contains true",
			value:  "my-test-resource",
			filter: Filter{Operand: "@", Target: "test", Negate: false},
			want:   true,
		},
		{
			name:   "contains false",
			value:  "my-resource",
			filter: Filter{Operand: "@", Target: "test", Negate: false},
			want:   false,
		},
		{
			name:   "negated contains true",
			value:  "my-resource",
			filter: Filter{Operand: "@", Target: "test", Negate: true},
			want:   true,
		},
		{
			name:   "regex match true",
			value:  "aws_instance_v1",
			filter: Filter{Operand: "/", Target: "^aws_.*_v\\d+$", Negate: false},
			want:   true,
		},
		{
			name:   "regex match false",
			value:  "instance",
			filter: Filter{Operand: "/", Target: "^aws_.*", Negate: false},
			want:   false,
		},
		{
			name:   "negated regex match",
			value:  "instance",
			filter: Filter{Operand: "/", Target: "^aws_.*", Negate: true},
			want:   true,
		},
		{
			name:   "greater than string true",
			value:  "z",
			filter: Filter{Operand: ">", Target: "a", Negate: false},
			want:   true,
		},
		{
			name:   "greater than string false",
			value:  "a",
			filter: Filter{Operand: ">", Target: "z", Negate: false},
			want:   false,
		},
		{
			name:   "less than string true",
			value:  "a",
			filter: Filter{Operand: "<", Target: "z", Negate: false},
			want:   true,
		},
		{
			name:   "invalid regex",
			value:  "test",
			filter: Filter{Operand: "/", Target: "[invalid", Negate: false},
			want:   false,
		},
		{
			name:   "unsupported operand",
			value:  "test",
			filter: Filter{Operand: "?", Target: "test", Negate: false},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkStringOperand(tt.value, tt.filter)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCheckNumericOperand(t *testing.T) {
	tests := []struct {
		name   string
		value  float64
		filter Filter
		want   bool
	}{
		{
			name:   "exact match true",
			value:  42,
			filter: Filter{Operand: "=", Target: "42", Negate: false},
			want:   true,
		},
		{
			name:   "exact match false",
			value:  42,
			filter: Filter{Operand: "=", Target: "40", Negate: false},
			want:   false,
		},
		{
			name:   "negated equal true",
			value:  42,
			filter: Filter{Operand: "=", Target: "40", Negate: true},
			want:   true,
		},
		{
			name:   "negated equal false",
			value:  42,
			filter: Filter{Operand: "=", Target: "42", Negate: true},
			want:   false,
		},
		{
			name:   "greater than true",
			value:  50,
			filter: Filter{Operand: ">", Target: "42", Negate: false},
			want:   true,
		},
		{
			name:   "greater than false",
			value:  42,
			filter: Filter{Operand: ">", Target: "50", Negate: false},
			want:   false,
		},
		{
			name:   "less than true",
			value:  42,
			filter: Filter{Operand: "<", Target: "50", Negate: false},
			want:   true,
		},
		{
			name:   "less than false",
			value:  50,
			filter: Filter{Operand: "<", Target: "42", Negate: false},
			want:   false,
		},
		{
			name:   "float value with integer target",
			value:  42.5,
			filter: Filter{Operand: ">", Target: "42", Negate: false},
			want:   true,
		},
		{
			name:   "invalid target",
			value:  42,
			filter: Filter{Operand: "=", Target: "invalid", Negate: false},
			want:   false,
		},
		{
			name:   "unsupported operand",
			value:  42,
			filter: Filter{Operand: "^", Target: "42", Negate: false},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkNumericOperand(tt.value, tt.filter)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCheckContainsOperand(t *testing.T) {
	tests := []struct {
		name   string
		value  interface{}
		filter Filter
		want   bool
	}{
		{
			name:   "slice contains true",
			value:  []any{"a", "b", "c"},
			filter: Filter{Operand: "@", Target: "b", Negate: false},
			want:   true,
		},
		{
			name:   "slice contains false",
			value:  []any{"a", "b", "c"},
			filter: Filter{Operand: "@", Target: "d", Negate: false},
			want:   false,
		},
		{
			name:   "slice not contains true",
			value:  []any{"a", "b", "c"},
			filter: Filter{Operand: "@", Target: "d", Negate: true},
			want:   true,
		},
		{
			name:   "slice not contains false",
			value:  []any{"a", "b", "c"},
			filter: Filter{Operand: "@", Target: "b", Negate: true},
			want:   false,
		},
		{
			name:   "map key exists true",
			value:  map[string]any{"key1": "value1", "key2": "value2"},
			filter: Filter{Operand: "@", Target: "key1", Negate: false},
			want:   true,
		},
		{
			name:   "map key exists false",
			value:  map[string]any{"key1": "value1", "key2": "value2"},
			filter: Filter{Operand: "@", Target: "key3", Negate: false},
			want:   false,
		},
		{
			name:   "map key not exists true",
			value:  map[string]any{"key1": "value1", "key2": "value2"},
			filter: Filter{Operand: "@", Target: "key3", Negate: true},
			want:   true,
		},
		{
			name:   "map key not exists false",
			value:  map[string]any{"key1": "value1", "key2": "value2"},
			filter: Filter{Operand: "@", Target: "key1", Negate: true},
			want:   false,
		},
		{
			name:   "unsupported type",
			value:  123,
			filter: Filter{Operand: "@", Target: "test", Negate: false},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkContainsOperand(tt.value, tt.filter)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		name   string
		value  interface{}
		want   float64
		wantOk bool
	}{
		{
			name:   "float64",
			value:  42.5,
			want:   42.5,
			wantOk: true,
		},
		{
			name:   "float32",
			value:  float32(42.5),
			want:   42.5,
			wantOk: true,
		},
		{
			name:   "int",
			value:  42,
			want:   42,
			wantOk: true,
		},
		{
			name:   "int64",
			value:  int64(42),
			want:   42,
			wantOk: true,
		},
		{
			name:   "uint32",
			value:  uint32(42),
			want:   42,
			wantOk: true,
		},
		{
			name:   "int8",
			value:  int8(10),
			want:   10,
			wantOk: true,
		},
		{
			name:   "int16",
			value:  int16(100),
			want:   100,
			wantOk: true,
		},
		{
			name:   "int32",
			value:  int32(1000),
			want:   1000,
			wantOk: true,
		},
		{
			name:   "uint",
			value:  uint(42),
			want:   42,
			wantOk: true,
		},
		{
			name:   "uint8",
			value:  uint8(50),
			want:   50,
			wantOk: true,
		},
		{
			name:   "uint16",
			value:  uint16(500),
			want:   500,
			wantOk: true,
		},
		{
			name:   "uint64",
			value:  uint64(5000),
			want:   5000,
			wantOk: true,
		},
		{
			name:   "string",
			value:  "42",
			want:   0,
			wantOk: false,
		},
		{
			name:   "nil",
			value:  nil,
			want:   0,
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := toFloat64(tt.value)
			assert.Equal(t, tt.wantOk, ok)
			if ok {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestApplyFilters(t *testing.T) {
	testData := `
	{
		"id": "res-123",
		"name": "my-resource",
		"type": "aws_instance",
		"region": "us-east-1",
		"count": 5,
		"tags": ["prod", "web"],
		"metadata": {"env": "production"},
		"description": null,
		"nested": {"inner": "value"}
	}
	`

	attrList := attrs.AttrList{
		{Key: "name", OutputKey: "name", Include: true},
		{Key: "type", OutputKey: "type", Include: true},
		{Key: "region", OutputKey: "region", Include: true},
		{Key: "count", OutputKey: "count", Include: true},
		{Key: "description", OutputKey: "description", Include: true},
		{Key: "nested", OutputKey: "nested", Include: true},
	}

	tests := []struct {
		name    string
		filters []Filter
		want    bool
	}{
		{
			name:    "no filters",
			filters: []Filter{},
			want:    true,
		},
		{
			name: "single filter match",
			filters: []Filter{
				{Key: "name", Operand: "=", Target: "my-resource", Negate: false},
			},
			want: true,
		},
		{
			name: "single filter no match",
			filters: []Filter{
				{Key: "name", Operand: "=", Target: "other", Negate: false},
			},
			want: false,
		},
		{
			name: "multiple filters all match",
			filters: []Filter{
				{Key: "name", Operand: "=", Target: "my-resource", Negate: false},
				{Key: "type", Operand: "^", Target: "aws_", Negate: false},
			},
			want: true,
		},
		{
			name: "multiple filters one fails",
			filters: []Filter{
				{Key: "name", Operand: "=", Target: "my-resource", Negate: false},
				{Key: "type", Operand: "^", Target: "gcp_", Negate: false},
			},
			want: false,
		},
		{
			name: "native filter ignored",
			filters: []Filter{
				{Key: "_native_filter", Operand: "=", Target: "value", Negate: false},
			},
			want: true,
		},
		{
			name: "missing attribute key continues",
			filters: []Filter{
				{Key: "nonexistent", Operand: "=", Target: "value", Negate: false},
			},
			want: true,
		},
		{
			name: "numeric comparison",
			filters: []Filter{
				{Key: "count", Operand: ">", Target: "3", Negate: false},
			},
			want: true,
		},
		{
			name: "missing key returns nil",
			filters: []Filter{
				{Key: "nonexistent_key", Operand: "=", Target: "value", Negate: false},
			},
			want: true,
		},
		{
			name: "null value filter fails",
			filters: []Filter{
				{Key: "description", Operand: "=", Target: "value", Negate: false},
			},
			want: false,
		},
		{
			name: "unsupported type with equals operator passes",
			filters: []Filter{
				{Key: "nested", Operand: "=", Target: "value", Negate: false},
			},
			want: true,
		},
		{
			name: "unsupported type with contains operator uses checkContainsOperand",
			filters: []Filter{
				{Key: "nested", Operand: "@", Target: "inner", Negate: false},
			},
			want: true,
		},
		{
			name: "array type with equals operator passes",
			filters: []Filter{
				{Key: "tags", Operand: "=", Target: "prod", Negate: false},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gjson.Parse(testData)
			got := applyFilters(result, attrList, tt.filters)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFilterDataset(t *testing.T) {
	testData := `
	[
		{
			"id": "res-1",
			"name": "aws-resource-1",
			"type": "aws_instance"
		},
		{
			"id": "res-2",
			"name": "gcp-resource",
			"type": "google_instance"
		},
		{
			"id": "res-3",
			"name": "aws-resource-2",
			"type": "aws_network"
		}
	]
	`

	attrList := attrs.AttrList{
		{Key: "name", OutputKey: "name", Include: true},
		{Key: "type", OutputKey: "type", Include: true},
	}

	tests := []struct {
		name      string
		spec      string
		wantCount int
		wantNames []string
	}{
		{
			name:      "no filters",
			spec:      "",
			wantCount: 3,
			wantNames: []string{"aws-resource-1", "gcp-resource", "aws-resource-2"},
		},
		{
			name:      "prefix filter",
			spec:      "type^aws_",
			wantCount: 2,
			wantNames: []string{"aws-resource-1", "aws-resource-2"},
		},
		{
			name:      "exact match filter",
			spec:      "name=gcp-resource",
			wantCount: 1,
			wantNames: []string{"gcp-resource"},
		},
		{
			name:      "no matches",
			spec:      "name=nonexistent",
			wantCount: 0,
		},
		{
			name:      "multiple filters",
			spec:      "type^aws_,name@1",
			wantCount: 1,
			wantNames: []string{"aws-resource-1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidates := gjson.Parse(testData)
			got := FilterDataset(candidates, attrList, tt.spec)
			assert.Len(t, got, tt.wantCount)
			for i, expected := range tt.wantNames {
				assert.Equal(t, expected, got[i]["name"])
			}
		})
	}
}
