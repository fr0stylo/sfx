package main

import (
	"encoding/base64"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/fr0stylo/sfx/exporter"
)

func TestHandleK8sSecretSuccess(t *testing.T) {
	opts := options{
		Name:      "app-secrets",
		Namespace: "prod",
		Type:      "Opaque",
		Labels: map[string]string{
			"app": "payments",
		},
	}
	optsBytes, err := yaml.Marshal(opts)
	if err != nil {
		t.Fatalf("marshal options: %v", err)
	}

	req := exporter.Request{
		Values: map[string][]byte{
			"api-token": []byte("secret"),
		},
		Options: optsBytes,
	}

	resp, err := handle(req)
	if err != nil {
		t.Fatalf("handle returned error: %v", err)
	}

	var manifest secretManifest
	if err := yaml.Unmarshal(resp.Payload, &manifest); err != nil {
		t.Fatalf("unmarshal manifest: %v", err)
	}

	if manifest.Metadata.Name != "app-secrets" || manifest.Metadata.Namespace != "prod" {
		t.Fatalf("unexpected metadata: %+v", manifest.Metadata)
	}
	encoded, ok := manifest.Data["api-token"]
	if !ok {
		t.Fatalf("expected api-token entry")
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if string(decoded) != "secret" {
		t.Fatalf("unexpected secret value: %q", string(decoded))
	}
}

func TestHandleK8sSecretMissingName(t *testing.T) {
	_, err := handle(exporter.Request{})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}
