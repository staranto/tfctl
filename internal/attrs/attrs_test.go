// Copyright (c) 2025 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package attrs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttrList_Set(t *testing.T) {
	tests := []struct {
		name      string
		initial   AttrList
		value     string
		wantLen   int
		wantAttrs []Attr
		wantErr   bool
	}{
		{
			name:    "empty string",
			initial: AttrList{},
			value:   "",
			wantLen: 0,
		},
		{
			name:    "wildcard only",
			initial: AttrList{},
			value:   "*",
			wantLen: 0,
		},
		{
			name:    "simple key",
			initial: AttrList{},
			value:   "name",
			wantLen: 1,
			wantAttrs: []Attr{
				{Key: "attributes.name", OutputKey: "name", Include: true, TransformSpec: ""},
			},
		},
		{
			name:    "key with dot notation",
			initial: AttrList{},
			value:   ".id",
			wantLen: 1,
			wantAttrs: []Attr{
				{Key: "id", OutputKey: "id", Include: true, TransformSpec: ""},
			},
		},
		{
			name:    "excluded key with !",
			initial: AttrList{},
			value:   "!name",
			wantLen: 1,
			wantAttrs: []Attr{
				{Key: "attributes.name", OutputKey: "name", Include: false, TransformSpec: ""},
			},
		},
		{
			name:    "key with custom output name",
			initial: AttrList{},
			value:   "full-name:name",
			wantLen: 1,
			wantAttrs: []Attr{
				{Key: "attributes.full-name", OutputKey: "name", Include: true, TransformSpec: ""},
			},
		},
		{
			name:    "key with transform spec",
			initial: AttrList{},
			value:   "name::u",
			wantLen: 1,
			wantAttrs: []Attr{
				{Key: "attributes.name", OutputKey: "name", Include: true, TransformSpec: "u"},
			},
		},
		{
			name:    "full format - key:output:transform",
			initial: AttrList{},
			value:   "created-at:date:t",
			wantLen: 1,
			wantAttrs: []Attr{
				{Key: "attributes.created-at", OutputKey: "date", Include: true, TransformSpec: "t"},
			},
		},
		{
			name:    "multiple attrs comma separated",
			initial: AttrList{},
			value:   "name,email,id",
			wantLen: 3,
			wantAttrs: []Attr{
				{Key: "attributes.name", OutputKey: "name", Include: true, TransformSpec: ""},
				{Key: "attributes.email", OutputKey: "email", Include: true, TransformSpec: ""},
				{Key: "attributes.id", OutputKey: "id", Include: true, TransformSpec: ""},
			},
		},
		{
			name:    "mixed formats",
			initial: AttrList{},
			value:   ".id,name::u,!internal,created-at:date:t",
			wantLen: 4,
			wantAttrs: []Attr{
				{Key: "id", OutputKey: "id", Include: true, TransformSpec: ""},
				{Key: "attributes.name", OutputKey: "name", Include: true, TransformSpec: "u"},
				{Key: "attributes.internal", OutputKey: "internal", Include: false, TransformSpec: ""},
				{Key: "attributes.created-at", OutputKey: "date", Include: true, TransformSpec: "t"},
			},
		},
		{
			name:    "global transform with wildcard",
			initial: AttrList{},
			value:   "*::u",
			wantLen: 1,
			wantAttrs: []Attr{
				{Key: "*", OutputKey: "*", Include: false, TransformSpec: "u"},
			},
		},
		{
			name:    "nested key with dots",
			initial: AttrList{},
			value:   "user.email.address",
			wantLen: 1,
			wantAttrs: []Attr{
				{Key: "attributes.user.email.address", OutputKey: "address", Include: true, TransformSpec: ""},
			},
		},
		{
			name: "update existing attr",
			initial: AttrList{
				{Key: "attributes.name", OutputKey: "name", Include: true, TransformSpec: ""},
			},
			value:   "name::u",
			wantLen: 1,
			wantAttrs: []Attr{
				{Key: "attributes.name", OutputKey: "name", Include: true, TransformSpec: "u"},
			},
		},
		{
			name: "update existing attr by output key",
			initial: AttrList{
				{Key: "attributes.full-name", OutputKey: "name", Include: true, TransformSpec: ""},
			},
			value:   "name::l",
			wantLen: 1,
			wantAttrs: []Attr{
				{Key: "attributes.full-name", OutputKey: "name", Include: true, TransformSpec: "l"},
			},
		},
		{
			name:    "whitespace trimming",
			initial: AttrList{},
			value:   " name : display : u ",
			wantLen: 1,
			wantAttrs: []Attr{
				{Key: "attributes.name", OutputKey: "display", Include: true, TransformSpec: "u"},
			},
		},
		{
			name:    "empty output key uses json key",
			initial: AttrList{},
			value:   "name:",
			wantLen: 1,
			wantAttrs: []Attr{
				{Key: "attributes.name", OutputKey: "name", Include: true, TransformSpec: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := tt.initial
			err := a.Set(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, a, tt.wantLen)

			if tt.wantAttrs != nil {
				for i, want := range tt.wantAttrs {
					assert.Equal(t, want.Key, a[i].Key, "attr[%d].Key", i)
					assert.Equal(t, want.OutputKey, a[i].OutputKey, "attr[%d].OutputKey", i)
					assert.Equal(t, want.Include, a[i].Include, "attr[%d].Include", i)
					assert.Equal(t, want.TransformSpec, a[i].TransformSpec, "attr[%d].TransformSpec", i)
				}
			}
		})
	}
}

func TestAttrList_SetGlobalTransformSpec(t *testing.T) {
	tests := []struct {
		name      string
		initial   AttrList
		wantSpecs []string // expected TransformSpec for each attr after applying global
		wantErr   bool
	}{
		{
			name: "no global transform",
			initial: AttrList{
				{Key: "attributes.name", TransformSpec: ""},
				{Key: "attributes.email", TransformSpec: "u"},
			},
			wantSpecs: []string{"", "u"},
		},
		{
			name: "global uppercase",
			initial: AttrList{
				{Key: "*", TransformSpec: "u"},
				{Key: "attributes.name", TransformSpec: ""},
				{Key: "attributes.email", TransformSpec: "l"},
			},
			wantSpecs: []string{"u,u", "u,", "u,l"},
		},
		{
			name: "global with length transform",
			initial: AttrList{
				{Key: "*", TransformSpec: "10"},
				{Key: "attributes.name", TransformSpec: ""},
				{Key: "attributes.title", TransformSpec: "u"},
			},
			wantSpecs: []string{"10,10", "10,", "10,u"},
		},
		{
			name: "global with multiple transforms",
			initial: AttrList{
				{Key: "*", TransformSpec: "u,20"},
				{Key: "attributes.name", TransformSpec: ""},
				{Key: "attributes.email", TransformSpec: "l"},
			},
			wantSpecs: []string{"u,20,u,20", "u,20,", "u,20,l"},
		},
		{
			name:      "empty list",
			initial:   AttrList{},
			wantSpecs: []string{},
		},
		{
			name: "only global attr",
			initial: AttrList{
				{Key: "*", TransformSpec: "u"},
			},
			wantSpecs: []string{"u,u"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := tt.initial
			err := a.SetGlobalTransformSpec()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, a, len(tt.wantSpecs))

			for i, wantSpec := range tt.wantSpecs {
				assert.Equal(t, wantSpec, a[i].TransformSpec, "attr[%d].TransformSpec", i)
			}
		})
	}
}

func TestAttr_Transform(t *testing.T) {
	tests := []struct {
		name        string
		attr        Attr
		input       interface{}
		envVars     map[string]string
		want        interface{}
		description string
	}{
		// Non-string passthrough
		{
			name:  "non-string value unchanged",
			attr:  Attr{TransformSpec: ""},
			input: 42,
			want:  42,
		},
		{
			name:  "map value unchanged",
			attr:  Attr{TransformSpec: ""},
			input: map[string]interface{}{"key": "value"},
			want:  map[string]interface{}{"key": "value"},
		},

		// Case transformations
		{
			name:  "uppercase transform",
			attr:  Attr{TransformSpec: "u"},
			input: "hello world",
			want:  "HELLO WORLD",
		},
		{
			name:  "lowercase transform",
			attr:  Attr{TransformSpec: "l"},
			input: "HELLO WORLD",
			want:  "hello world",
		},
		{
			name:  "uppercase with U",
			attr:  Attr{TransformSpec: "U"},
			input: "hello",
			want:  "HELLO",
		},
		{
			name:  "lowercase with L",
			attr:  Attr{TransformSpec: "L"},
			input: "HELLO",
			want:  "hello",
		},
		{
			name:  "last case transform wins - lower",
			attr:  Attr{TransformSpec: "u,l"},
			input: "Hello",
			want:  "hello",
		},
		{
			name:  "last case transform wins - upper",
			attr:  Attr{TransformSpec: "l,u"},
			input: "Hello",
			want:  "HELLO",
		},

		// Length transformations
		{
			name:  "truncate to 5 chars",
			attr:  Attr{TransformSpec: "5"},
			input: "hello world",
			want:  "hello",
		},
		{
			name:  "no truncation if shorter",
			attr:  Attr{TransformSpec: "20"},
			input: "hello",
			want:  "hello",
		},
		{
			name:  "middle ellipsis with negative",
			attr:  Attr{TransformSpec: "-10"},
			input: "hello world today",
			want:  "hell..oday",
		},
		{
			name:  "middle ellipsis calculation",
			attr:  Attr{TransformSpec: "-8"},
			input: "hello world",
			want:  "hel..rld",
		},
		{
			name:  "no middle ellipsis if shorter",
			attr:  Attr{TransformSpec: "-20"},
			input: "hello",
			want:  "hello",
		},

		// Combined transformations
		{
			name:  "uppercase and truncate",
			attr:  Attr{TransformSpec: "u,10"},
			input: "hello world",
			want:  "HELLO WORL",
		},
		{
			name:  "lowercase and truncate",
			attr:  Attr{TransformSpec: "l,5"},
			input: "HELLO",
			want:  "hello",
		},
		{
			name:  "multiple length specs - last wins",
			attr:  Attr{TransformSpec: "10,5"},
			input: "hello world",
			want:  "hello",
		},
		{
			name:  "case and length with last case wins",
			attr:  Attr{TransformSpec: "u,10,l"},
			input: "Hello World",
			want:  "hello worl",
		},

		// Time transformations (when TZ is set)
		{
			name:  "time transform with TZ set",
			attr:  Attr{TransformSpec: "t"},
			input: "2024-01-15T10:30:00Z",
			envVars: map[string]string{
				"TZ": "America/New_York",
			},
			want: "2024-01-15T05:30:00EST",
		},
		{
			name:  "time transform without TZ",
			attr:  Attr{TransformSpec: "t"},
			input: "2024-01-15T10:30:00Z",
			want:  "2024-01-15T10:30:00Z",
		},
		{
			name:  "time transform with T",
			attr:  Attr{TransformSpec: "T"},
			input: "2024-01-15T10:30:00Z",
			envVars: map[string]string{
				"TZ": "UTC",
			},
			want: "2024-01-15T10:30:00UTC",
		},
		{
			name:  "invalid time format unchanged",
			attr:  Attr{TransformSpec: "t"},
			input: "not-a-time",
			envVars: map[string]string{
				"TZ": "UTC",
			},
			want: "not-a-time",
		},

		// No transform
		{
			name:  "empty spec - no transform",
			attr:  Attr{TransformSpec: ""},
			input: "hello world",
			want:  "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			got := tt.attr.Transform(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAttrList_String(t *testing.T) {
	tests := []struct {
		name     string
		attrList AttrList
		want     string
	}{
		{
			name:     "empty list",
			attrList: AttrList{},
			want:     "",
		},
		{
			name: "single attr",
			attrList: AttrList{
				{Key: "attributes.name", OutputKey: "name", TransformSpec: "u"},
			},
			want: "attributes.name:name:u",
		},
		{
			name: "multiple attrs",
			attrList: AttrList{
				{Key: "id", OutputKey: "id", TransformSpec: ""},
				{Key: "attributes.name", OutputKey: "name", TransformSpec: "u"},
				{Key: "attributes.email", OutputKey: "email", TransformSpec: "l"},
			},
			want: "id:id:,attributes.name:name:u,attributes.email:email:l",
		},
		{
			name: "attr with empty transform",
			attrList: AttrList{
				{Key: "attributes.name", OutputKey: "name", TransformSpec: ""},
			},
			want: "attributes.name:name:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.attrList.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAttrList_Type(t *testing.T) {
	a := AttrList{}
	assert.Equal(t, "list", a.Type())
}

// TestAttr_Transform_TimezonePriority tests that timezone is sourced from
// environment in the correct priority order.
func TestAttr_Transform_TimezonePriority(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		input   string
		want    string
	}{
		{
			name: "TZ env var used",
			envVars: map[string]string{
				"TZ": "America/Los_Angeles",
			},
			input: "2024-01-15T10:00:00Z",
			want:  "2024-01-15T02:00:00PST",
		},
		{
			name:    "no timezone - passthrough",
			envVars: map[string]string{},
			input:   "2024-01-15T10:00:00Z",
			want:    "2024-01-15T10:00:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all relevant env vars first
			os.Unsetenv("TZ")

			// Set test env vars
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			attr := Attr{TransformSpec: "t"}
			got := attr.Transform(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
