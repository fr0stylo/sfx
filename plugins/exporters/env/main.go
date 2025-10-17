package main

import (
	"fmt"
	"html/template"
	"sort"
	"strconv"
	"strings"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v3"

	"github.com/fr0stylo/sfx/exporter"
)

func main() {
	exporter.Run(exporter.HandlerFunc(handle))
}

type Options struct {
	KeyTemplate string `yaml:"key_template"`
}

type TemplateValue struct {
	Value any
}

const defaultKeyTemplate = "{{ .Value | upper }}"

func handle(req exporter.Request) (exporter.Response, error) {
	var opts Options
	if err := yaml.Unmarshal(req.Options, &opts); err != nil {
		return exporter.Response{}, err
	}
	if opts.KeyTemplate == "" {
		opts.KeyTemplate = defaultKeyTemplate
	}

	keyTmpl, err := template.New("key").Funcs(sprig.FuncMap()).Parse(opts.KeyTemplate)
	if err != nil {
		return exporter.Response{}, fmt.Errorf("parse key template: %w", err)
	}

	var keys []string
	for k := range req.Values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		err := keyTmpl.Execute(&b, TemplateValue{Value: k})
		if err != nil {
			return exporter.Response{}, err
		}
		b.WriteByte('=')
		b.WriteString(formatValue(req.Values[k]))
	}
	b.WriteByte('\n')

	return exporter.Response{Payload: []byte(b.String())}, nil
}

func formatValue(value []byte) string {
	s := string(value)
	if s == "" {
		return `""`
	}

	if strings.ContainsAny(s, " \t\r\n\"'\\") {
		return strconv.Quote(s)
	}
	return s
}
