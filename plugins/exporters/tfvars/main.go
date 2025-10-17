package main

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"

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

	var buf bytes.Buffer
	for i, key := range keys {
		if i > 0 {
			buf.WriteByte('\n')
		}
		buf.WriteString(key)
		buf.WriteString(" = ")
		buf.WriteString(formatValue(req.Values[key]))
	}
	buf.WriteByte('\n')

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

func formatValue(b []byte) string {
	s := string(b)
	if strings.Contains(s, "\n") {
		return heredoc(s)
	}
	if isBool(s) || isNumeric(s) {
		return s
	}
	return strconv.Quote(s)
}

func heredoc(s string) string {
	const marker = "EOF"
	return fmt.Sprintf("<<%s\n%s\n%s", marker, s, marker)
}

func isBool(s string) bool {
	if s == "" {
		return false
	}
	_, err := strconv.ParseBool(s)
	return err == nil
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	if _, err := strconv.ParseInt(s, 10, 64); err == nil {
		return true
	}
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}
