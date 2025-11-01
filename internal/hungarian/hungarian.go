// Copyright (c) 2025 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package hungarian

import (
	"regexp"
	"strings"
)

// IsHungarian returns true if any component of the Terraform type (split by
// '_') appears in the name of that resource. Matching is case-insensitive and
// checks both substring containment and token equality when the name is split
// by non-alphanumeric chars.
func IsHungarian(typ string, name string) bool {
	if typ == "" || name == "" {
		return false
	}

	typeLower := strings.ToLower(typ)
	nameLower := strings.ToLower(name)

	// Split the type into a slice of tokens.
	typeTokens := strings.Split(typeLower, "_")

	// Split the name by non-alphanumeric separators into a slice of tokens.
	splitRe := regexp.MustCompile(`[^a-z0-9]+`)
	nameParts := splitRe.Split(nameLower, -1)

	// Iterate over each type token and see if it matches any name token.  If so,
	// it's Hungarian.
	for _, tok := range typeTokens {
		if tok == "" {
			continue
		}

		// If the token appears as a whole name part, it's Hungarian.
		for _, p := range nameParts {
			if p == tok {
				// Hungarian - bail out.
				return true
			}
		}

		// Also treat any substring occurrence as a match covers cases like
		// resource "aws_s3_bucket" "mybucket", where the name is jammed without
		// separators.
		if strings.Contains(nameLower, tok) {
			// Hungarian - bail out.
			return true
		}
	}

	// Not Hungarian.
	return false
}
