package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"sfx/exporter"
)

func TestHandleTemplateInline(t *testing.T) {
	opts := options{Template: "API_TOKEN={{ index .Values \"api_token\" }}"}
	optsBytes, err := yaml.Marshal(opts)
	if err != nil {
		t.Fatalf("marshal options: %v", err)
	}

	req := exporter.Request{
		Values:  map[string][]byte{"api_token": []byte("secret")},
		Options: optsBytes,
	}

	resp, err := handle(req)
	if err != nil {
		t.Fatalf("handle returned error: %v", err)
	}

	if got := strings.TrimSpace(string(resp.Payload)); got != "API_TOKEN=secret" {
		t.Fatalf("unexpected payload: %q", got)
	}
}

func TestHandleTemplateFile(t *testing.T) {
	dir := t.TempDir()
	tplPath := filepath.Join(dir, "config.tmpl")
	if err := os.WriteFile(tplPath, []byte("count={{ len .Values }}"), 0o600); err != nil {
		t.Fatalf("write template: %v", err)
	}

	opts := options{TemplatePath: tplPath}
	optsBytes, err := yaml.Marshal(opts)
	if err != nil {
		t.Fatalf("marshal options: %v", err)
	}

	req := exporter.Request{Values: map[string][]byte{"a": []byte("1"), "b": []byte("2")}, Options: optsBytes}

	resp, err := handle(req)
	if err != nil {
		t.Fatalf("handle returned error: %v", err)
	}

	if got := strings.TrimSpace(string(resp.Payload)); got != "count=2" {
		t.Fatalf("unexpected payload: %q", got)
	}
}

func TestHandleTemplateMissingContent(t *testing.T) {
	_, err := handle(exporter.Request{})
	if err == nil || !strings.Contains(err.Error(), "template content not provided") {
		t.Fatalf("expected missing template error, got %v", err)
	}
}
