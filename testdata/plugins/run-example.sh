#!/usr/bin/env bash
#
# End-to-end example: build the sample plugin and load it at runtime.
#
# Go plugins require the plugin and the host binary to be built with the same
# Go version, architecture, and identical dependency build IDs. Because of
# this, the runtime load is demonstrated here with a small standalone host
# program instead of inside `go test` (see plugin_load_test.go).
#
# Usage: bash testdata/plugins/run-example.sh
set -euo pipefail

repo_root="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$repo_root"

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

# 1. Build the plugin as a shared object.
go build -buildmode=plugin -o "$tmp_dir/testplugin.so" ./testdata/plugins/testplugin

# 2. Build a small host program that registers the standard extension points
#    and loads every .so from a directory.
mkdir -p "$tmp_dir/host"
cat > "$tmp_dir/host/main.go" <<'EOF'
package main

import (
	"fmt"
	"os"

	"IACForge/src/extension"
	"IACForge/src/schema"
	"IACForge/src/validation"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: host <plugin-dir>")
		os.Exit(2)
	}
	s := schema.CoreSchema()
	vEngine := validation.NewEngine(s)
	m := extension.NewManager()
	m.RegisterExtensionPoint(extension.NewEntityKindsExtensionPoint(s))
	m.RegisterExtensionPoint(extension.NewRelationTypesExtensionPoint(s))
	m.RegisterExtensionPoint(extension.NewValidationRulesExtensionPoint(vEngine))
	m.RegisterExtensionPoint(extension.NewRendererExtensionPoint(extension.NewRendererRegistry()))
	m.RegisterExtensionPoint(extension.NewRootKindsExtensionPoint(vEngine))

	if err := m.LoadFromDir(os.Args[1]); err != nil {
		fmt.Fprintf(os.Stderr, "LoadFromDir failed: %v\n", err)
		os.Exit(1)
	}

	ext, ok := m.GetExtension("testplugin.sample")
	if !ok {
		fmt.Fprintln(os.Stderr, "testplugin.sample was not registered")
		os.Exit(1)
	}
	fmt.Printf("loaded extension %s (namespace=%s, kinds=%d)\n",
		ext.Manifest.ID, ext.Manifest.Namespace, len(ext.EntityKinds))
	fmt.Printf("schema recognizes %s: %v\n", ext.EntityKinds[0].Kind, s.HasEntityKind(ext.EntityKinds[0].Kind))
}
EOF

# 3. Run the host against the directory containing the plugin.
go run "$tmp_dir/host/main.go" "$tmp_dir"
