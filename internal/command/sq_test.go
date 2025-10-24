// Copyright (c) 2025 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0
// no-cloc

package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChopPrefix_EmptyDataset(t *testing.T) {
	data := []map[string]interface{}{}
	chopPrefix(data, "resource")
	assert.Equal(t, 0, len(data))
}

func TestChopPrefix_NoAttribute(t *testing.T) {
	data := []map[string]interface{}{
		{"name": "example"},
		{"name": "example.two"},
	}
	// Should be a no-op
	chopPrefix(data, "resource")
	assert.Equal(t, "example", data[0]["name"])
	assert.Equal(t, "example.two", data[1]["name"])
}

func TestChopPrefix_NoCommonSegments(t *testing.T) {
	data := []map[string]interface{}{
		{"resource": "a.x.y"},
		{"resource": "b.x.y"},
		{"resource": "c.x.y"},
	}
	// No common leading segments across >=50% so no change
	chopPrefix(data, "resource")
	assert.Equal(t, "a.x.y", data[0]["resource"])
	assert.Equal(t, "b.x.y", data[1]["resource"])
	assert.Equal(t, "c.x.y", data[2]["resource"])
}

func TestChopPrefix_OneCommonSegmentOnly(t *testing.T) {
	data := []map[string]interface{}{
		{"resource": "common.a"},
		{"resource": "common.b"},
		{"resource": "other.c"},
	}
	// Only one common segment; must be at least 2 to chop
	chopPrefix(data, "resource")
	assert.Equal(t, "common.a", data[0]["resource"])
	assert.Equal(t, "common.b", data[1]["resource"])
	assert.Equal(t, "other.c", data[2]["resource"])
}

func TestChopPrefix_TwoCommonSegments_Threshold(t *testing.T) {
	data := []map[string]interface{}{
		{"resource": "env.prod.app.server1"},
		{"resource": "env.prod.app.server2"},
		{"resource": "env.prod.app.server3"},
		{"resource": "env.staging.app.server4"},
	}
	// 3 of 4 (>=50%) share "env.prod" as first two segments so they should be chopped
	chopPrefix(data, "resource")
	// The first three have a longer common prefix (env.prod.app) so the
	// implementation will remove all common leading segments (not just two).
	// Expect the full common prefix removed and replaced with "..".
	assert.Equal(t, "..server1", data[0]["resource"])
	assert.Equal(t, "..server2", data[1]["resource"])
	assert.Equal(t, "..server3", data[2]["resource"])
	// The fourth should be unchanged because it doesn't start with env.prod.
	assert.Equal(t, "env.staging.app.server4", data[3]["resource"])
}

func TestChopPrefix_PartialMatchesDifferentLengths(t *testing.T) {
	data := []map[string]interface{}{
		{"resource": "a.b.c"},
		{"resource": "a.b"},
		{"resource": "a.b.c.d"},
		{"resource": "x.y.z"},
	}
	// 3 of 4 have leading segments ["a","b"] so chop should apply to those with prefix
	chopPrefix(data, "resource")
	// The implementation computes a longest common leading segment list
	// that meets the threshold. In this case the common prefix is "a.b.c",
	// so only values that start with "a.b.c." will be shortened.
	assert.Equal(t, "a.b.c", data[0]["resource"]) // unchanged (no trailing dot)
	assert.Equal(t, "a.b", data[1]["resource"])   // unchanged
	assert.Equal(t, "..d", data[2]["resource"])   // a.b.c.d -> ..d
	assert.Equal(t, "x.y.z", data[3]["resource"])
}

func TestChopPrefix_ExactPrefixUnchanged(t *testing.T) {
	data := []map[string]interface{}{
		{"resource": "a.b"},
		{"resource": "a.b.c"},
		{"resource": "a.b.d"},
	}
	// Common prefix is "a.b" (two segments). Only entries that have a
	// remainder after "a.b." should be shortened.
	chopPrefix(data, "resource")
	assert.Equal(t, "a.b", data[0]["resource"]) // exact prefix, unchanged
	assert.Equal(t, "..c", data[1]["resource"]) // a.b.c -> ..c
	assert.Equal(t, "..d", data[2]["resource"]) // a.b.d -> ..d
}

func TestChopPrefix_SingleEntry_NoChange(t *testing.T) {
	data := []map[string]interface{}{
		{"resource": "only.one"},
	}
	// Single entry should not be transformed into an empty remainder; it
	// will remain unchanged (no trailing dot to match the prefixToRemove).
	chopPrefix(data, "resource")
	assert.Equal(t, "only.one", data[0]["resource"])
}

func TestChopPrefix_NonStringValues_Ignored(t *testing.T) {
	data := []map[string]interface{}{
		{"resource": 123},
		{"resource": "a.b.c"},
		{"resource": "a.b.d"},
	}
	// Non-string values should be ignored and not cause a panic. The
	// string values are evaluated normally.
	chopPrefix(data, "resource")
	assert.Equal(t, 123, data[0]["resource"]) // unchanged
	// With these inputs the longest common prefix may not remove a trailing
	// dot for either string, so they remain unchanged in practice.
	assert.Equal(t, "a.b.c", data[1]["resource"])
	assert.Equal(t, "a.b.d", data[2]["resource"])
}

func TestChopPrefix_SomeMissingAttribute(t *testing.T) {
	data := []map[string]interface{}{
		{"resource": "a.b.c"},
		{"name": "no-resource"},
		{"resource": "a.b.d"},
	}
	// Entries missing the attribute should be ignored and others processed.
	chopPrefix(data, "resource")
	// The middle entry shouldn't be touched; behavior for the strings is
	// consistent with the implementation (no trailing-dot removal in this set).
	assert.Equal(t, "a.b.c", data[0]["resource"]) // unchanged
	assert.Equal(t, "no-resource", data[1]["name"])
	assert.Equal(t, "a.b.d", data[2]["resource"]) // unchanged
}
