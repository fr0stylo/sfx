package main

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v3"

	"github.com/fr0stylo/sfx/exporter"
)

type options struct {
	Template     string `yaml:"template"`
	TemplatePath string `yaml:"template_path"`
	Delims       struct {
		Left  string `yaml:"left"`
		Right string `yaml:"right"`
	} `yaml:"delims"`
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

	tmplSource := opts.Template
	if tmplSource == "" && opts.TemplatePath != "" {
		data, err := os.ReadFile(opts.TemplatePath)
		if err != nil {
			return exporter.Response{}, fmt.Errorf("read template file: %w", err)
		}
		tmplSource = string(data)
	}
	if tmplSource == "" {
		return exporter.Response{}, fmt.Errorf("template content not provided (set template or template_path)")
	}

	funcMap := sprig.TxtFuncMap()
	tmpl := template.New("export").Funcs(funcMap)
	if opts.Delims.Left != "" || opts.Delims.Right != "" {
		left := opts.Delims.Left
		right := opts.Delims.Right
		if left == "" || right == "" {
			return exporter.Response{}, fmt.Errorf("both delims.left and delims.right must be set")
		}
		tmpl = tmpl.Delims(left, right)
	}

	tmpl, err := tmpl.Parse(tmplSource)
	if err != nil {
		return exporter.Response{}, fmt.Errorf("parse template: %w", err)
	}

	data := map[string]any{
		"Values": stringMap(req.Values),
		"Raw":    req.Values,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return exporter.Response{}, fmt.Errorf("execute template: %w", err)
	}

	return exporter.Response{Payload: buf.Bytes()}, nil
}

func stringMap(values map[string][]byte) map[string]string {
	out := make(map[string]string, len(values))
	for k, v := range values {
		out[k] = string(v)
	}
	return out
}
