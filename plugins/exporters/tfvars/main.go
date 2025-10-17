package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
	"gopkg.in/yaml.v3"

	"sfx/exporter"
)

type options struct {
	Order []string `yaml:"order"`
}

func main() {
	exporter.Run(exporter.HandlerFunc(handle))
}

func handle(req exporter.Request) (exporter.Response, error) {
	var opts options
	if len(req.Options) > 0 {
		if err := yaml.Unmarshal(req.Options, &opts); err != nil {
			return exporter.Response{}, fmt.Errorf("parse options: %w", err)
		}
	}

	keys := orderedKeys(req.Values, opts.Order)

	file := hclwrite.NewEmptyFile()
	body := file.Body()

	for _, key := range keys {
		value := req.Values[key]
		ctyVal := decodeValue(value)
		body.SetAttributeValue(key, ctyVal)
	}

	var buf bytes.Buffer
	if _, err := file.WriteTo(&buf); err != nil {
		return exporter.Response{}, fmt.Errorf("render tfvars: %w", err)
	}

	// Ensure trailing newline
	if !strings.HasSuffix(buf.String(), "\n") {
		buf.WriteByte('\n')
	}

	return exporter.Response{Payload: buf.Bytes()}, nil
}

func orderedKeys(values map[string][]byte, order []string) []string {
	seen := make(map[string]bool, len(values))
	var out []string

	for _, key := range order {
		if _, ok := values[key]; ok && !seen[key] {
			out = append(out, key)
			seen[key] = true
		}
	}

	var remaining []string
	for key := range values {
		if !seen[key] {
			remaining = append(remaining, key)
		}
	}
	sort.Strings(remaining)
	out = append(out, remaining...)

	return out
}

func decodeValue(b []byte) cty.Value {
	if len(b) == 0 {
		return cty.StringVal("")
	}
	str := strings.TrimSpace(string(b))

	if v, err := strconv.ParseBool(str); err == nil {
		return cty.BoolVal(v)
	}
	if i, err := strconv.ParseInt(str, 10, 64); err == nil {
		return cty.NumberIntVal(i)
	}
	if f, err := strconv.ParseFloat(str, 64); err == nil {
		return cty.NumberFloatVal(f)
	}

	// Try to treat the payload as JSON to support complex structures.
	if maybeJSON(str) {
		if val, err := jsonToCty([]byte(str)); err == nil && val.Type() != cty.NilType {
			return val
		}
	}

	return cty.StringVal(string(b))
}

func maybeJSON(s string) bool {
	if len(s) < 2 {
		return false
	}
	first := s[0]
	last := s[len(s)-1]
	switch first {
	case '{':
		return last == '}'
	case '[':
		return last == ']'
	case '"':
		return last == '"'
	}
	return false
}

func jsonToCty(b []byte) (cty.Value, error) {
	raw := json.RawMessage(b)
	return ctyjson.Unmarshal(raw, cty.DynamicPseudoType)
}
