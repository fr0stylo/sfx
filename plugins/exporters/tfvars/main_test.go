package main

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/fr0stylo/sfx/exporter"
)

func TestHandleTFVarsTypesAndOrder(t *testing.T) {
	opts := options{Order: []string{"json", "number"}}
	optsBytes, err := yaml.Marshal(opts)
	if err != nil {
		t.Fatalf("marshal options: %v", err)
	}

	req := exporter.Request{
		Values: map[string][]byte{
			"number": []byte("42"),
			"string": []byte("hello world"),
			"bool":   []byte("true"),
			"json":   []byte(`{"nested":"value"}`),
		},
		Options: optsBytes,
	}

	resp, err := handle(req)
	if err != nil {
		t.Fatalf("handle returned error: %v", err)
	}

	output := string(resp.Payload)

	normalized := normalizeSpaces(output)

	if !strings.Contains(normalized, "number = 42") {
		t.Fatalf("expected integer assignment, got:\n%s", output)
	}
	if !strings.Contains(normalized, "bool = true") {
		t.Fatalf("expected boolean assignment, got:\n%s", output)
	}
	if !strings.Contains(normalized, `string = "hello world"`) {
		t.Fatalf("expected quoted string, got:\n%s", output)
	}
	if !strings.Contains(normalized, `json = {`) || !strings.Contains(normalized, `nested = "value"`) {
		t.Fatalf("expected json object rendering, got:\n%s", output)
	}

	order := strings.Index(output, "json")
	order2 := strings.Index(output, "number")
	if order > order2 {
		t.Fatalf("expected json to appear before number, got:\n%s", output)
	}
}

func normalizeSpaces(s string) string {
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return s
}
