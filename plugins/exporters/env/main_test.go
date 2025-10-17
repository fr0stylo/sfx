package main

import (
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/fr0stylo/sfx/exporter"
)

func TestHandleEnvDefault(t *testing.T) {
	req := exporter.Request{
		Values: map[string][]byte{
			"db-password": []byte("hunter2"),
			"API_TOKEN":   []byte("value with spaces"),
			"EMPTY":       []byte(""),
		},
	}

	resp, err := handle(req)
	if err != nil {
		t.Fatalf("handle returned error: %v", err)
	}

	got := string(resp.Payload)
	want := "API_TOKEN=\"value with spaces\"\nEMPTY=\"\"\nDB-PASSWORD=hunter2\n"

	if got != want {
		t.Fatalf("unexpected payload:\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestHandleEnvCustomTemplate(t *testing.T) {
	optsBytes, err := yaml.Marshal(Options{KeyTemplate: "{{ .Value | lower }}"})
	if err != nil {
		t.Fatalf("marshal options: %v", err)
	}

	req := exporter.Request{
		Values: map[string][]byte{
			"API_TOKEN": []byte("secret"),
		},
		Options: optsBytes,
	}

	resp, err := handle(req)
	if err != nil {
		t.Fatalf("handle returned error: %v", err)
	}

	got := string(resp.Payload)
	want := "api_token=secret\n"

	if got != want {
		t.Fatalf("unexpected payload:\nwant %q\ngot  %q", want, got)
	}
}
