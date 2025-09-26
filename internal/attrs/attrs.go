// Copyright © 2025 Steve Taranto staranto@gmail.com
// SPDX-License-Identifier: MIT

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

// Attr represents each of the keys to be included in the output.  These are
// typically identified by the JSON attributes key, thus the name.
type Attr struct {
	// The JSON key to extract from the result JSON object.
	Key string
	// Should this Attr be included in output or is it just
	// intended for filtering and sorting?
	Include bool
	// The key to use in the output.  # This will also be used as the column title
	// when output=text.  # TODO Consider a separate title field.
	OutputKey string
	// Transformation spec to apply to the output value.
	TransformSpec string
}

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

		// See if there is a timezone in the config file or via a TFCTL_ or TF_ env
		// variable.  If there's not, look for a plain TZ env variable.
		tz := "" // TODO rt.GetRuntimeVar("timezone", true, "")
		if tz == "" {
			tz = os.Getenv("TZ")
		}

		// We're only going to convert if we've specifically told what TZ to use.
		// If we haven't, we'll just use the value as is.
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

	// We need to know which case transformation appears last.  This covers the
	// case where there has been a global case transformation prepended to the
	// attrs transformation and, thus, allows the attr's to carry more weight.
	// IOW...  --attrs '*::U,name::l' will be lower case.
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
		// Same logic as above re: case.  This allows a more specific length
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

type AttrList []Attr

// Return a string representation of the AttrList.  This should match the format
// of the original --attrs flag.
func (a *AttrList) String() string {
	result := make([]string, 0, len(*a))
	for _, attr := range *a {
		result = append(result, fmt.Sprintf("%s:%s:%s", attr.Key, attr.OutputKey, attr.TransformSpec))
	}
	return strings.Join(result, ",")
}

// Parse each spec from the --attrs flag and add it to the AttrList.
func (a *AttrList) Set(value string) error {
	if value == "" || value == "*" {
		return nil
	}

	const (
		jsonIdx = iota
		outputIdx
		transformIdx
	)

	// There are three : delimited fields in each spec.  The first is the key to
	// extract from the JSON object.  The second is the key to use in the output.
	// The third is the transformation spec to apply to the output value. The
	// latter two are optional.  The output key will default to the last
	// section of the JSON key.
	specs := strings.Split(value, ",")
specloop:
	for _, spec := range specs {

		// Default to including the attribute also assuming it will be a child of the
		// .attributes key of the JSON object.
		attr := Attr{
			Include: true,
		}

		fields := strings.Split(spec, ":")

		// The first field is the key to extract from the JSON payload.  If it
		// begins with a !, it is excluded from the output.
		attr.Key = strings.TrimSpace(fields[jsonIdx])
		if strings.HasPrefix(attr.Key, "!") {
			attr.Include = false
			attr.Key = attr.Key[1:]
		}

		if attr.Key == "*" {
			attr.Include = false
		}

		// Fixup the output field.  If there is only one field it is considered the
		// JSON extract key and the output key will become the last segment of the
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

		// If the attr already exists in the list (because it's one of the defaults
		// for cmd or the user double-entered it) just apply the OutputKey, Include
		// and TransformSpec to the existing Attr.

		for i := range *a {
			if (*a)[i].Key == attr.Key || (*a)[i].OutputKey == attr.Key {
				(*a)[i].Include = attr.Include
				(*a)[i].OutputKey = attr.OutputKey
				(*a)[i].TransformSpec = attr.TransformSpec
				continue specloop
			}
		}

		// Fixup the key field.  If it begins with a . that means we're working off
		// the root of the JSON objects.  If it does not, we're working off the
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

// SetGlobalTransformSpec inserts a global transform spec into the front of all
// attrs in the list.
func (alist *AttrList) SetGlobalTransformSpec() error {
	spec := ""

	// Find the global transform spec.  If there is more than one, we're not
	// dealing with it and just taking the first.
	for a := range *alist {
		if (*alist)[a].Key == "*" {
			spec = (*alist)[a].TransformSpec
			break
		}
	}

	// Return early if there is no global transform spec.
	if spec == "" {
		return nil
	}

	// Slam the global spec onto
	for a := range *alist {
		(*alist)[a].TransformSpec = spec + "," + (*alist)[a].TransformSpec
	}

	return nil
}

func (a *AttrList) Type() string {
	return "list"
}
