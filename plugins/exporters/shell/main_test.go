package main

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/fr0stylo/sfx/exporter"
)

func TestHandleShellDefault(t *testing.T) {
	req := exporter.Request{
		Values: map[string][]byte{
			"API_TOKEN": []byte("secret value"),
			"EMPTY":     []byte(""),
		},
	}

	resp, err := handle(req)
	if err != nil {
		t.Fatalf("handle returned error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(resp.Payload)), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected shebang + 2 lines, got %d: %v", len(lines), lines)
	}
	if lines[0] != "#!/usr/bin/env bash" {
		t.Fatalf("unexpected shebang: %q", lines[0])
	}
	if lines[1] != "export API_TOKEN='secret value'" {
		t.Fatalf("unexpected first export line: %q", lines[1])
	}
	if lines[2] != "export EMPTY=''" {
		t.Fatalf("unexpected second export line: %q", lines[2])
	}
}

func TestHandleShellCustomOptions(t *testing.T) {
	opts := options{
		Shebang:      "#!/bin/sh",
		Header:       []string{"Managed by sfx"},
		ExportFormat: "assign",
		Order:        []string{"B", "A"},
	}
	optsBytes, err := yaml.Marshal(opts)
	if err != nil {
		t.Fatalf("marshal options: %v", err)
	}

	req := exporter.Request{
		Values: map[string][]byte{
			"A": []byte("1"),
			"B": []byte("2"),
		},
		Options: optsBytes,
	}

	resp, err := handle(req)
	if err != nil {
		t.Fatalf("handle returned error: %v", err)
	}

	payload := string(resp.Payload)
	if !strings.HasPrefix(payload, "#!/bin/sh\n# Managed by sfx") {
		t.Fatalf("unexpected header: %q", payload)
	}
	first := strings.Index(payload, "B=")
	second := strings.Index(payload, "A=")
	if first == -1 || second == -1 || first > second {
		t.Fatalf("expected B assignment before A, got %q", payload)
	}
}
