package main

import (
	"testing"

	"gopkg.in/yaml.v3"

	"sfx/exporter"
)

func TestHandleAnsible(t *testing.T) {
	opts := options{Prefix: "secret_", Order: []string{"B", "A"}}
	optsBytes, err := yaml.Marshal(opts)
	if err != nil {
		t.Fatalf("marshal options: %v", err)
	}

	req := exporter.Request{
		Values: map[string][]byte{
			"A": []byte("valueA"),
			"B": []byte("valueB"),
		},
		Options: optsBytes,
	}

	resp, err := handle(req)
	if err != nil {
		t.Fatalf("handle returned error: %v", err)
	}

	m := map[string]string{}
	if err := yaml.Unmarshal(resp.Payload, &m); err != nil {
		t.Fatalf("unmarshal yaml: %v", err)
	}

	if m["secret_A"] != "valueA" || m["secret_B"] != "valueB" {
		t.Fatalf("unexpected map: %+v", m)
	}
}
