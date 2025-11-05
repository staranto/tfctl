// Copyright (c) 2025 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package attrs

import (
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/apex/log"
)

// Attr represents each of the keys to be included in the output. These are
// typically identified by the JSON attributes key, thus the name.
type Attr struct {
	// The JSON key to extract from the result JSON object.
	Key string `yaml:"key" json:"Key"`
	// Should this Attr be included in output or is it just
	// intended for filtering and sorting?
	Include bool `yaml:"include" json:"Include"`
	// The key to use in the output. This is also used as the column title when
	// output=text. TODO Consider a separate title field.
	OutputKey string `yaml:"outputKey" json:"OutputKey"`
	// Transformation spec to apply to the output value.
	TransformSpec string `yaml:"transformSpec" json:"TransformSpec"`
}

// Transform applies the attribute's transform spec to a value and returns the
// transformed result.
func (a *Attr) Transform(value interface{}) interface{} {

	// TODO Currently only string values can be transformed.
	result, ok := value.(string)
	if !ok {
		if mapValue, ok := value.(map[string]interface{}); ok {
			return mapValue
		}
		return value
	}

	// Convert UTC time to local.
	if strings.ContainsAny(a.TransformSpec, "tT") {

		// See if there is a timezone in the config or via a TFCTL_ or TF_ env
		// variable. If there is not, look for a plain TZ env variable.
		tz := "" // TODO rt.GetRuntimeVar("timezone", true, "").
		if tz == "" {
			tz = os.Getenv("TZ")
		}

		// Convert only if a TZ was specified. Otherwise, use the value as is.
		if tz != "" {
			loc, err := time.LoadLocation(tz)
			if err == nil {
				t, err := time.Parse(time.RFC3339, result)
				if err == nil {
					local := t.In(loc)
					result = local.Format("2006-01-02T15:04:05MST")
				} else {
					log.Error("failed to parse time: " + result)
					a.TransformSpec = strings.ReplaceAll(a.TransformSpec, "t", "")
					a.TransformSpec = strings.ReplaceAll(a.TransformSpec, "T", "")
				}
			}
		}
	}

	// We need to know which case transformation appears last. This covers the
	// case where there has been a global case transformation prepended to the
	// attrs transformation and allows the attr's to carry more weight.
	// IOW... --attrs '*::U,name::l' will be lower case.
	lastL := strings.LastIndexAny(a.TransformSpec, "lL")
	lastU := strings.LastIndexAny(a.TransformSpec, "uU")

	if lastL > lastU {
		result = strings.ToLower(result)
	} else if lastU > lastL {
		result = strings.ToUpper(result)
	}

	// Is it a length-based transformation?
	if a.TransformSpec != "" {
		re := regexp.MustCompile(`-?\d+`)
		// Same logic as above re: case. This allows a more specific length
		// transformation to override a global one.
		match := re.FindAllString(a.TransformSpec, -1)
		if len(match) != 0 {
			// Take the last (overriding) match.
			l, _ := strconv.Atoi(match[len(match)-1])
			abs := int(math.Abs(float64(l)))
			if len(result) > abs {
				if l < 0 {
					lr := abs/2 - 1
					left := result[0:lr]
					right := result[len(result)-lr:]
					result = left + ".." + right
				} else {
					result = result[:l]
				}
			}
		}
	}

	return result
}

// AttrList is a collection of Attr used to shape output fields.
type AttrList []Attr

// Set parses each spec from --attrs and adds it to the AttrList.
func (a *AttrList) Set(value string) error {
	if value == "" || value == "*" {
		return nil
	}

	const (
		jsonIdx = iota
		outputIdx
		transformIdx
	)

	// There are three : delimited fields in each spec. The first is the key to
	// extract from the JSON object. The second is the key to use in the output.
	// The third is the transformation spec to apply to the output value. The
	// latter two are optional. The output key defaults to the last
	// section of the JSON key.
	specs := strings.Split(value, ",")
specloop:
	for _, spec := range specs {

		// Default to including the attribute, assuming it is a child of the
		// .attributes key of the JSON object.
		attr := Attr{
			Include: true,
		}

		fields := strings.Split(spec, ":")

		// The first field is the key to extract from the JSON payload. If it
		// begins with a !, it is excluded from the output.
		attr.Key = strings.TrimSpace(fields[jsonIdx])
		if strings.HasPrefix(attr.Key, "!") {
			attr.Include = false
			attr.Key = attr.Key[1:]
		}

		if attr.Key == "*" {
			attr.Include = false
		}

		// Fix up the output field. If there is only one field, it is the JSON
		// extract key and the output key becomes the last segment of the
		// . notation.
		if len(fields) == 1 {
			segments := strings.Split(attr.Key, ".")
			attr.OutputKey = segments[len(segments)-1]
		} else {
			if fields[outputIdx] != "" {
				attr.OutputKey = strings.TrimSpace(fields[outputIdx])
			} else {
				attr.OutputKey = attr.Key
			}
		}

		attr.TransformSpec = ""
		if len(fields) > transformIdx {
			attr.TransformSpec = strings.TrimSpace(fields[transformIdx])
		}

		// If the attr already exists in the list (because it is a default for
		// a command or the user double-entered it), apply the OutputKey, Include
		// and TransformSpec to the existing Attr.

		for i := range *a {
			if (*a)[i].Key == attr.Key || (*a)[i].OutputKey == attr.Key {
				(*a)[i].Include = attr.Include
				(*a)[i].OutputKey = attr.OutputKey
				(*a)[i].TransformSpec = attr.TransformSpec
				continue specloop
			}
		}

		// Fix up the key field. If it begins with '.', we are working off the root
		// of the JSON objects. If it does not, we are working off the
		// .attributes of the JSON objects.
		if strings.HasPrefix(attr.Key, ".") {
			attr.Key = attr.Key[1:]
		} else if attr.Key != "*" {
			attr.Key = "attributes." + attr.Key
		}

		*a = append(*a, attr)
	}

	return nil
}

// SetGlobalTransformSpec inserts a global transform spec at the front of all
// attrs in the list.
func (a *AttrList) SetGlobalTransformSpec() error {
	spec := ""

	// Find the global transform spec. If there is more than one, take the first.
	for attr := range *a {
		if (*a)[attr].Key == "*" {
			spec = (*a)[attr].TransformSpec
			break
		}
	}

	// Return early if there is no global transform spec.
	if spec == "" {
		return nil
	}

	// Prepend the global spec onto each attribute's spec.
	for attr := range *a {
		(*a)[attr].TransformSpec = spec + "," + (*a)[attr].TransformSpec
	}

	return nil
}

// String returns a string representation of the AttrList. This matches the format
// of the original --attrs flag.
func (a *AttrList) String() string {
	result := make([]string, 0, len(*a))
	for _, attr := range *a {
		result = append(result, fmt.Sprintf("%s:%s:%s", attr.Key, attr.OutputKey, attr.TransformSpec))
	}
	return strings.Join(result, ",")
}

// Type returns the flag type for use with the flag.Value interface.
func (a *AttrList) Type() string { return "list" }
