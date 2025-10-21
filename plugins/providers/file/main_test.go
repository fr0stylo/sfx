package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fr0stylo/sfx/provider"
)

func writeTempFile(t *testing.T, contents string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "env")
	err := os.WriteFile(path, []byte(contents), 0o600)
	require.NoError(t, err, "write temp file")
	return path
}

func optionsYAML(path string) []byte {
	return []byte(fmt.Sprintf("path: %s\n", path))
}

func TestHandleReturnsValueWhenLineMatches(t *testing.T) {
	path := writeTempFile(t, "FOO=bar\nBAR=baz\n")

	resp, err := handle(provider.Request{
		Ref:     "env://FOO",
		Options: optionsYAML(path),
	})
	require.NoError(t, err)
	assert.Equal(t, "bar", string(resp.Value))
}

func TestHandleReturnsNilWhenRefMissing(t *testing.T) {
	path := writeTempFile(t, "FOO=bar\n")

	resp, err := handle(provider.Request{
		Ref:     "env://BAZ",
		Options: optionsYAML(path),
	})
	require.NoError(t, err)
	assert.Nil(t, resp.Value)
}

func TestHandlePropagatesYAMLError(t *testing.T) {
	_, err := handle(provider.Request{
		Ref:     "env://FOO",
		Options: []byte("path: ["),
	})
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "yaml")
	}
}

func TestHandlePropagatesFileOpenError(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing")

	_, err := handle(provider.Request{
		Ref:     "env://FOO",
		Options: optionsYAML(missing),
	})
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "open file")
	}
}

func TestParseEnvFileReturnsValueWithoutSeparator(t *testing.T) {
	buf, err := parseEnvFile(bytes.NewBufferString("FOO=bar\n"), []byte("FOO"))
	require.NoError(t, err)
	assert.Equal(t, "bar", string(buf))
}
