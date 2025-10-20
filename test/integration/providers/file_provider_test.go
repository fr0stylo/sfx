package providers_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func repoRoot(tb testing.TB) string {
	tb.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		tb.Fatal("runtime.Caller returned !ok")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", ".."))
}

func requireFile(t *testing.T, path string) {
	t.Helper()

	if info, err := os.Stat(path); err != nil || info.IsDir() {
		t.Fatalf("expected file at %s; run `make build` first (stat err: %v)", path, err)
	}
}

func TestFileProviderIntegration(t *testing.T) {
	root := repoRoot(t)
	sfxBin := filepath.Join(root, "bin", "sfx")
	fileProviderBin := filepath.Join(root, "bin", "providers", "file")
	envExporterBin := filepath.Join(root, "bin", "exporters", "env")

	requireFile(t, sfxBin)
	requireFile(t, fileProviderBin)
	requireFile(t, envExporterBin)

	workdir := t.TempDir()
	secretPrefix := filepath.Join(workdir, "integration-secret-")

	configTemplate := `
providers:
  file: %[1]q
exporters:
  env: %[2]q
output:
  type: env
secrets:
  INTEGRATION_SECRET:
    ref: INTEGRATION_SECRET
    provider: file
    provider_options:
      path: %[3]q
`

	config := []byte(strings.TrimSpace(fmt.Sprintf(
		configTemplate,
		fileProviderBin,
		envExporterBin,
		secretPrefix,
	)))

	if err := os.WriteFile(filepath.Join(workdir, ".sfx.yaml"), config, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cmd := exec.Command(sfxBin, "fetch")
	cmd.Dir = workdir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		t.Fatalf("sfx fetch failed: %v (stderr: %s)", err, stderr.String())
	}

	got := strings.TrimSpace(stdout.String())
	want := "INTEGRATION_SECRET=" + secretPrefix + "secret-for-INTEGRATION_SECRET"
	if got != want {
		t.Fatalf("unexpected output\nwant: %s\ngot:  %s", want, got)
	}
}
