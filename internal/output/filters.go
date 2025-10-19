// Copyright (c) 2025 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package output

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/apex/log"
	"github.com/tidwall/gjson"

	"github.com/staranto/tfctlgo/internal/attrs"
	"github.com/staranto/tfctlgo/internal/driller"
)

// filterRegex is the pattern used to parse filter expressions into key, operator, and target components.
// It matches: key + operator + target, where operator can be negated with !
var filterRegex = regexp.MustCompile(`^(.*?)([!/]{1,2}|[=^~><!@]{1,2})(.*)$`)

// Filter represents a single parsed --filter expression including the key,
// operand, optional negation and target value.
type Filter struct {
	Key     string
	Negate  bool
	Operand string
	Target  string
}

// BuildFilters parses a filter specification string into a slice of Filter.
// Invalid specs (unsupported operand or malformed expression) are skipped.
func BuildFilters(spec string) []Filter {
	// Don't prealloc because we don't know what len will be and performance is
	// not critical.
	//nolint:prealloc
	var filters []Filter

	// If there are no filters specified, go home early.
	if spec == "" {
		return filters
	}

	// Default delimiter is ",", allow an override.
	delim := ","
	if d, ok := os.LookupEnv("TFCTL_FILTER_DELIM"); ok {
		delim = d
	}

	// Split the spec and iterate over each filter spec entry.
	filterSpecs := strings.Split(spec, delim)
	for _, filterSpec := range filterSpecs {
		parts := filterRegex.FindStringSubmatch(filterSpec)

		// If a supported operand was not found, log an error and throw it away.
		if parts == nil {
			log.Error("invalid filter: " + filterSpec)
			continue
		}

		// parts[2] is the operand. It may have a leading negation. If so, chop it
		// off and just use the remainder as the working operand.
		// Check if the operand begins with a negation.
		negate := strings.HasPrefix(parts[2], "!")
		if negate {
			parts[2] = strings.TrimPrefix(parts[2], "!")
		}

		// We've got a good filter, append it to the result set.
		filters = append(filters, Filter{
			Key:     parts[1],
			Negate:  negate,
			Operand: parts[2],
			Target:  parts[3],
		})
	}

	return filters
}

// FilterDataset returns a result set filtered per the provided spec. It is the
// public entry point used by SliceDiceSpit.
func FilterDataset(candidates gjson.Result, attrs attrs.AttrList, spec string) []map[string]interface{} {
	//nolint:prealloc // Don't prealloc because we don't know what len will be.
	var filteredResults []map[string]interface{}

	// Build a slice of filters from the spec once so we can discard invalid
	// entries and avoid reparsing for each candidate row.
	filters := BuildFilters(spec)

	// Iterate over the candidate dataset, checking each against the filters.
	for _, candidate := range candidates.Array() {
		if !applyFilters(candidate, attrs, filters) {
			continue
		}

		// If the filter check was successful, add each attribute from the candidate
		// to the filtered result set.
		result := make(map[string]interface{})
		for i := range attrs {
			attr := attrs[i]
			// Intentionally defer Transform to SliceDiceSpit output phase.
			// This function is responsible for filtering only; transformations
			// are applied downstream during output formatting.
			// value := attr.Transform(candidate.Get(attr.Key).Value())
			value := driller.Driller(candidate.Raw, attr.Key)
			result[attr.OutputKey] = value.Value()
		}
		filteredResults = append(filteredResults, result)
	}

	return filteredResults
}

// applyFilters returns true if the candidate row matches all of the provided
// filters. Native TF API filter keys (prefixed with _) are ignored here.
func applyFilters(candidate gjson.Result, attrs attrs.AttrList, filters []Filter) bool {
	// No filters, so go home early.
	if len(filters) == 0 {
		return true
	}

	// Iterate over the filters, checking each against the candidate.
	for _, filter := range filters {
		var key string

		// If Key starts with _, it's a native filter used by the TF API and should
		// be ignored here.
		if strings.HasPrefix(filter.Key, "_") {
			continue
		}

		// Find the attribute that matches the filter key.
		for _, attr := range attrs {
			if attr.OutputKey == filter.Key {
				key = attr.Key
				break
			}
		}

		// If an attribute matching the filter key was not found, log the condition
		// and skip this filter (continue processing other filters).
		if key == "" {
			msg := fmt.Sprintf("filter key not found: %s", filter.Key)
			log.Error(msg)
			fmt.Fprintf(os.Stderr, "warning: %s\n", msg)
			continue
		}

		// Get the value from the candidate for the key. If it's nil, fail early.
		value := driller.Driller(candidate.Raw, key).Value()
		if value == nil {
			return false
		}

		// Check the value against the filter. If it fails the check, fail early.
		result := true
		if v, ok := value.(string); ok {
			result = checkStringOperand(v, filter)
		} else if v, ok := value.(bool); ok {
			result = checkStringOperand(fmt.Sprintf("%v", v), filter)
		} else if filter.Operand == "@" {
			result = checkContainsOperand(value, filter)
		}

		if !result {
			return false
		}
	}

	return true
}

// checkContainsOperand evaluates a membership style filter (operand '@')
// against slice or map values.
func checkContainsOperand(value interface{}, filter Filter) bool {
	switch val := value.(type) {
	case []any:
		for _, item := range val {
			if item == filter.Target && !filter.Negate {
				return true
			}
		}
	case map[string]any:
		_, found := val[filter.Target]
		if filter.Negate {
			return !found
		}
		return found
	default:
		log.Error(fmt.Sprintf("unsupported type for contains filtering: %T", value))
		return false
	}
	return false
}

// checkStringOperand evaluates a string comparison style filter against the
// provided value using the operand semantics.
func checkStringOperand(value string, filter Filter) bool {
	switch filter.Operand {
	case "=":
		return value == filter.Target == !filter.Negate
	case "~":
		return strings.EqualFold(value, filter.Target) == !filter.Negate
	case "^":
		return strings.HasPrefix(value, filter.Target) == !filter.Negate
	case ">":
		return value > filter.Target == !filter.Negate
	case "<":
		return value < filter.Target == !filter.Negate
	case "@":
		return strings.Contains(value, filter.Target) == !filter.Negate
	case "/":
		matched, err := regexp.MatchString(filter.Target, value)
		if err != nil {
			log.Error("invalid regex: " + filter.Target)
			return false
		}
		return matched == !filter.Negate
	default:
		log.Error("unsupported filtering operand: " + filter.Operand)
		return false
	}
}
