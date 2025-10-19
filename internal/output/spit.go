// Copyright (c) 2025 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/apex/log"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/lipgloss/v2/table"
	"github.com/tidwall/gjson"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v2"

	"github.com/staranto/tfctlgo/internal/attrs"
	"github.com/staranto/tfctlgo/internal/config"
	"github.com/staranto/tfctlgo/internal/filters"
)

// Tag represents a discovered struct field tag used when emitting schema
// information (--schema flag).
type Tag struct {
	Kind     string
	Name     string
	Encoding string
}

// Print renders the tag into its display form.
func (t Tag) Print() (out string) {
	parts := []string{}
	if t.Name != "" {
		parts = append(parts, t.Name)
	}
	return strings.Join(parts, ",")
}

// DumpSchema prints a sorted list of attribute tags for the provided type.
func DumpSchema(prefix string, typ reflect.Type) {
	tags := DumpSchemaWalker(prefix, typ, 0)
	if len(tags) == 0 {
		log.Debugf("No tags found for type: %s", typ.Name())
		return
	}

	sort.Slice(tags, func(i, j int) bool {
		if tags[i].Kind == tags[j].Kind {
			return tags[i].Name < tags[j].Name
		}
		return tags[i].Kind < tags[j].Kind
	})

	for _, tag := range tags {
		fmt.Println(tag.Name)
	}

	fmt.Println("")
	fmt.Println(
		`Resource level attributes that are directly available to the --attrs flag.
For a complete schema, including relationships, use --output=raw and see the
attrs help in the documentation or man tfctl-attrs.`)
}

const maxSchemaDepth = 1

// DumpSchemaWalker recursively walks a struct type discovering jsonapi tags.
func DumpSchemaWalker(holder string, typ reflect.Type, depth int) []Tag {
	tags := make([]Tag, 0)

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		log.Debugf("field: %s, type: %s in %s", field.Name, field.Type, field.PkgPath)

		tagValue, ok := field.Tag.Lookup("jsonapi")
		if !ok {
			continue
		}

		tag := NewTag(holder, tagValue)
		if tag.Kind != "attr" {
			continue
		}

		tags = append(tags, tag)

		if depth < maxSchemaDepth {

			switch field.Type.Kind() {
			case reflect.Struct:
				tags = append(tags, DumpSchemaWalker(tag.Name, field.Type, depth+1)...)
			case reflect.Ptr:
				if field.Type.Elem().Kind() == reflect.Struct {
					holder := tag.Name
					if tag.Kind == "relation" {
						holder = fmt.Sprintf(".relationships.%s.data", tag.Name)
					}
					tags = append(tags, DumpSchemaWalker(holder, field.Type.Elem(), depth+1)...)
				}
			default:
				if strings.Contains(field.Type.String(), ".") {
					continue
				}
				log.Debugf("Presumed primitive field type: %s for %v", field.Type.Kind(), tag)
			}
		}
	}

	return tags
}

// InterfaceToString converts supported primitive or composite values to a
// string. A custom empty value may be provided.
func InterfaceToString(value interface{}, emptyValue ...string) string {
	if len(emptyValue) == 0 {
		emptyValue = []string{""}
	}

	if value == nil || reflect.ValueOf(value).IsZero() {
		return emptyValue[0]
	}

	// THINK This doesn't do what you think it does. int and bool paths are never
	// taken?
	switch value := value.(type) {
	case string:
		return value
	case int:
		return strconv.Itoa(value)
	case float64:
		// Our current use cases have no use for an actual float, so we're just
		// going to return an integer.
		return fmt.Sprintf("%.0f", value)
	case bool:
		return strconv.FormatBool(value)
	default:
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return fmt.Sprintf("%v", value)
		}
		return string(jsonBytes)
	}
}

// NewTag constructs a Tag from a raw struct tag value and an optional holder
// prefix used to build hierarchical attribute names.
func NewTag(h string, s string) Tag {
	allowed := []string{"attr"}

	tag := Tag{}

	parts := strings.Split(s, ",")
	if len(parts) > 0 {
		found := false
		for _, a := range allowed {
			if a == parts[0] {
				found = true
				break
			}
		}

		if !found {
			return tag
		}

		tag.Kind = parts[0]
	}

	if len(parts) > 1 {
		if h != "" {
			parts[1] = fmt.Sprintf("%s.%s", h, parts[1])
		}
		tag.Name = parts[1]
	}

	if len(parts) > 2 {
		tag.Encoding = parts[2]
	}

	return tag
}

// SliceDiceSpit orchestrates filtering, transforming, sorting and rendering
// of a dataset according to command flags and attribute specifications.
// The optional postProcess callback allows commands to apply custom transformations
// to the filtered dataset before rendering.
func SliceDiceSpit(raw bytes.Buffer,
	attrs attrs.AttrList,
	cmd *cli.Command,
	parent string,
	w io.Writer,
	postProcess func([]map[string]interface{}) error) {

	if w == nil {
		w = os.Stdout
	}

	// If raw, just dump it and go home.
	output := cmd.String("output")
	if output == "raw" {
		_, _ = os.Stdout.Write(raw.Bytes())
		return
	}

	// Note: This schema transformation is handled by the sq command via postProcess callback,
	// which allows resource hierarchies to be flattened for consistent processing.
	if resources := gjson.Parse(raw.String()).Get("resources"); resources.Exists() {
		raw = flattenState(resources, cmd.Bool("noshort"))
	}

	var fullDataset gjson.Result
	// Just keep the "data" object from the document and throw away everything
	// else, notably "included", which we don't have a use case for. We're also
	// Parsing this into JSON so that we can use the lowercase key names and not
	// the proper case names from the TFE API.
	if parent != "" {
		fullDataset = gjson.Parse(raw.String()).Get(parent)
	} else {
		fullDataset = gjson.Parse(raw.String())
	}

	filter := cmd.String("filter")

	// Note: The concrete filter is applied here to match sq command semantics.
	// Command-specific logic like --chop is handled via postProcess callback in sq.go.
	if cmd.Bool("concrete") {
		if filter != "" {
			filter += ","
		}
		filter += "mode=managed"
	}

	// Filter out the rows we don't want. Do it here so that the following
	// processes are slightly more efficient since they'll be working on a smaller
	// dataset.
	filteredDataset := filters.FilterDataset(fullDataset, attrs, filter)

	// THINK This is inefficient. We're forcing a time transformation to occur
	// for all attributes, even though many will not be a timestamp. One
	// alternative would be to look at first row of full dataset and only add the
	// time transformation to attrs that look like timestamps.
	if cmd.Bool("local") {
		for a := range attrs {
			attrs[a].TransformSpec += "t"
		}
	}

	// Transform each value in each row.
	for _, row := range filteredDataset {
		for _, attr := range attrs {
			if attr.TransformSpec != "" {
				row[attr.OutputKey] = attr.Transform(row[attr.OutputKey])
			}
		}
	}

	spec := cmd.String("sort")
	SortDataset(filteredDataset, spec)

	switch output {
	case "json":
		// Marshal the filtered dataset into a JSON document.
		// TODO Figure out how to maintain key order in the JSON document.
		jsonOutput, err := json.Marshal(filteredDataset)
		if err != nil {
			slog.Error("SliceDiceSpit()", "err", err)
		}
		os.Stdout.Write(jsonOutput)
	case "yaml":
		yamlOutput, err := yaml.Marshal(filteredDataset)
		if err != nil {
			slog.Error("SliceDiceSpit()", "err", err)
		}
		os.Stdout.Write(yamlOutput)
	default:
		// Apply command-specific post-processing
		if postProcess != nil {
			if err := postProcess(filteredDataset); err != nil {
				slog.Error("PostProcess", "err", err)
			}
		}

		TableWriter(filteredDataset, attrs, cmd, w) // TODO
	}
}

// TableWriter renders the result set in a tabular form honoring color,
// titles and padding options.
func TableWriter(
	resultSet []map[string]interface{},
	attrs attrs.AttrList,
	cmd *cli.Command,
	w io.Writer) {

	if len(resultSet) == 0 {
		return
	}

	var (
		headerStyle  = lipgloss.NewStyle().Align(lipgloss.Left).Bold(true)
		cellStyle    = lipgloss.NewStyle().Padding(0, 0).Align(lipgloss.Left)
		evenRowStyle = cellStyle
		oddRowStyle  = cellStyle
	)

	if cmd.Bool("color") {
		headerColor, evenColor, oddColor := getColors("colors")

		headerStyle = headerStyle.Foreground(lipgloss.Color(headerColor))
		evenRowStyle = evenRowStyle.Foreground(lipgloss.Color(evenColor))
		oddRowStyle = oddRowStyle.Foreground(lipgloss.Color(oddColor))
	}

	var rows [][]string
	for _, result := range resultSet {
		row := make([]string, 0, len(result))
		for _, attr := range attrs {
			if !attr.Include {
				continue
			}
			row = append(row, InterfaceToString(result[attr.OutputKey], "-"))
		}
		rows = append(rows, row)
	}

	pad, _ := config.GetInt("padding", 0)
	t := table.New().
		BorderBottom(false).
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false).
		Border(lipgloss.HiddenBorder()).
		StyleFunc(func(row, col int) lipgloss.Style {
			var style lipgloss.Style
			switch {
			case row == table.HeaderRow:
				style = headerStyle
			case row%2 == 0:
				style = evenRowStyle
			default:
				style = oddRowStyle
			}

			if col > 0 {
				style = style.PaddingLeft(pad)
			}

			return style
		}).
		Headers().
		Rows(rows...)

	if cmd.Bool("titles") {
		var headers []string
		for _, attr := range attrs {
			if attr.Include {
				headers = append(headers, attr.OutputKey)
			}
		}

		// https://github.com/charmbracelet/lipgloss/issues/261
		t = t.Headers(headers...).BorderHeader(false)
	}
	fmt.Println(t)
}

// flattenState takes the state schema of each entry and flattens it into a
// schema with parent and attributes. This is done so that we can have a common
// schema for all the different types of resources.
func flattenState(resources gjson.Result, short bool) bytes.Buffer {
	var flatResources []map[string]interface{}

	for _, resource := range resources.Array() {
		common := getCommonFields(resource)

		instances := resource.Get("instances")
		for _, instance := range instances.Array() {
			flatResource := make(map[string]interface{})
			for key, value := range common {
				flatResource[key] = value
			}

			for key, value := range instance.Map() {
				flatResource[key] = value.Value()
			}

			module := ""
			if flatResource["module"] != nil {
				module = InterfaceToString(flatResource["module"]) + "."
			}

			mode := ""
			if flatResource["mode"] != "managed" {
				mode = InterfaceToString(flatResource["mode"]) + "."
			}

			indexKey := ""
			if flatResource["index_key"] != nil {
				switch v := flatResource["index_key"].(type) {
				case int, int64, float64:
					indexKey = fmt.Sprintf("[%v]", v)
				default:
					indexKey = fmt.Sprintf("[\"%v\"]", v)
				}
			}

			resourceID := fmt.Sprintf("%s%s%s.%s%s", module, mode, flatResource["type"], flatResource["name"], indexKey)
			if !short {
				re := regexp.MustCompile(`(^module.)|(.module.)`)
				resourceID = re.ReplaceAllString(resourceID, "+")
			}
			flatResource["resource"] = resourceID

			flatResources = append(flatResources, flatResource)
		}
	}

	jsonBytes, err := json.Marshal(flatResources)
	if err != nil {
		slog.Error("flattenState()", "err", err)
		return *bytes.NewBuffer([]byte{})
	}

	raw := *bytes.NewBuffer(jsonBytes)
	return raw
}

// getColors returns configured color values for table rendering.
func getColors(key string) (header string, even string, odd string) {
	header, _ = config.GetString(fmt.Sprintf("%s.title", key), "#f6be00")
	even, _ = config.GetString(fmt.Sprintf("%s.even", key), "#ffffff")
	odd, _ = config.GetString(fmt.Sprintf("%s.odd", key), "#00c8f0")
	return
}

func getCommonFields(resource gjson.Result) map[string]interface{} {
	var common = make(map[string]interface{})
	for key, value := range resource.Map() {
		if key != "instances" {
			common[key] = value.Value()
		}
	}
	return common
}
