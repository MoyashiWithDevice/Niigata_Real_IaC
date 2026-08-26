package extension_test

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// TestTestdataPluginBuilds verifies that the sample plugin under
// testdata/plugins compiles successfully as a Go plugin (.so). Building the
// plugin is the first part of the load path; the actual plugin.Open runtime
// load cannot run under `go test` because Go plugins require the host binary
// and the plugin to be built with identical build IDs, which `go test` does
// not guarantee. See testdata/plugins/run-example.sh for an end-to-end load.
func TestTestdataPluginBuilds(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Go plugin building is only supported on Linux")
	}
	if testing.Short() {
		t.Skip("skipping plugin build in short mode")
	}

	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}

	soPath := filepath.Join(t.TempDir(), "testplugin.so")
	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", soPath, "./testdata/plugins/testplugin")
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build test plugin: %v\n%s", err, out)
	}
}
