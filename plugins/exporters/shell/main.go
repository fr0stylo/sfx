package main

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"sfx/exporter"
)

type options struct {
	Shebang      string   `yaml:"shebang"`
	Header       []string `yaml:"header"`
	ExportFormat string   `yaml:"export_format"`
	Order        []string `yaml:"order"`
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

	if opts.Shebang == "" {
		opts.Shebang = "#!/usr/bin/env bash"
	}
	format := opts.ExportFormat
	if format == "" {
		format = "export"
	}

	keys := orderedKeys(req.Values, opts.Order)

	var buf bytes.Buffer
	buf.WriteString(opts.Shebang)
	buf.WriteByte('\n')
	if len(opts.Header) > 0 {
		for _, line := range opts.Header {
			buf.WriteString("# ")
			buf.WriteString(line)
			buf.WriteByte('\n')
		}
		buf.WriteByte('\n')
	}

	for _, key := range keys {
		value := shellQuote(string(req.Values[key]))
		switch format {
		case "export":
			fmt.Fprintf(&buf, "export %s=%s\n", key, value)
		case "assign":
			fmt.Fprintf(&buf, "%s=%s\n", key, value)
		default:
			return exporter.Response{}, fmt.Errorf("unknown export_format %q", format)
		}
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

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	if !strings.ContainsAny(s, " \t\r\n\"'\\$`") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
