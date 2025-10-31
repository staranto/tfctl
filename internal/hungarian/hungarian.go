// Copyright (c) 2025 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package hungarian

import (
	"regexp"
	"strings"
)

// IsHungarian returns true if any component of the Terraform type (split by
// '_') appears in the name. Matching is case-insensitive and checks both
// substring containment and token equality when the name is split by
// non-alphanumeric chars.
func IsHungarian(typ string, name string) bool {
	if typ == "" || name == "" {
		return false
	}

	typeLower := strings.ToLower(typ)
	nameLower := strings.ToLower(name)

	// tokens from the type, e.g. "aws_s3_bucket" -> ["aws","s3","bucket"]
	typeTokens := strings.Split(typeLower, "_")

	// nameParts are tokens from the name split by non-alphanumeric separators
	// e.g. "my-thing_aws.widget" -> ["my","thing","aws","widget"]
	splitRe := regexp.MustCompile(`[^a-z0-9]+`)
	nameParts := splitRe.Split(nameLower, -1)

	for _, tok := range typeTokens {
		if tok == "" {
			continue
		}

		// 1) If the token appears as a whole name part, it's a match.
		for _, p := range nameParts {
			if p == tok {
				return true
			}
		}

		// 2) Also treat any substring occurrence as a match (covers
		//    cases like "tmp_aws" or "my-thing-aws-widget").
		if strings.Contains(nameLower, tok) {
			return true
		}
	}

	return false
}
