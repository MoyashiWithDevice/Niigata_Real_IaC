#!/usr/bin/env bash
# IACForge full-feature demo script.
# Usage: ./demo/run-demo.sh [--skip-build]
set -euo pipefail

cd "$(dirname "$0")/.."

SKIP_BUILD=0
[[ "${1:-}" == "--skip-build" ]] && SKIP_BUILD=1

section() {
  echo
  echo "======================================================================"
  echo "  $*"
  echo "======================================================================"
}

run() {
  echo
  echo "\$ $*"
  "$@"
}

OUT_DIR=demo/out
mkdir -p "$OUT_DIR"

if [[ $SKIP_BUILD -eq 0 ]]; then
  section "0. Build"
  go build -o iacforge ./cmd/iacforge
fi

# ----------------------------------------------------------------------
section "1. validate: YAML infrastructure model validation"
run ./iacforge validate demo/core/model.yaml

echo
echo "--- negative case: schema violations (expect FAILED, exit != 0) ---"
./iacforge validate demo/negative/bad-model.yaml || true

echo
echo "--- negative case: dangling reference (expect Parse error) ---"
./iacforge validate demo/negative/bad-parse.yaml || true

# ----------------------------------------------------------------------
section "2. info: graph summary (directory scan: all .yaml merged)"
run ./iacforge info demo/core/

# ----------------------------------------------------------------------
section "3. render: view -> artifact (markdown / mermaid / json / svg)"
run ./iacforge render demo/core/model.yaml --format markdown
run ./iacforge render demo/core/model.yaml --format mermaid
echo "\$ ./iacforge render demo/core/model.yaml --format json   (truncated)"
./iacforge render demo/core/model.yaml --format json | head -30
echo "..."
./iacforge render demo/core/model.yaml --format svg    --output "$OUT_DIR/graph.svg"
./iacforge render demo/aws/model.yaml   --format mermaid --output "$OUT_DIR/aws.mmd"
ls -la "$OUT_DIR"

# ----------------------------------------------------------------------
section "4. query: filter by entity kind or relation type, multiple formats"
run ./iacforge query demo/core/model.yaml --kind vm --format text
run ./iacforge query demo/core/model.yaml --type depends_on --format json
run ./iacforge query demo/core/model.yaml --kind application --format mermaid

# ----------------------------------------------------------------------
section "5. AWS extension: vendor-specific kinds & relation types"
echo "The AWS extension contributes 45 entity kinds, 12 relation types,"
echo "and root authority for aws.organization."
run ./iacforge validate demo/aws/model.yaml
run ./iacforge info demo/aws/model.yaml
run ./iacforge query demo/aws/model.yaml --kind aws.ec2 --format text
run ./iacforge query demo/aws/model.yaml --type aws.registers --format text

# ----------------------------------------------------------------------
section "6. mcp: AI-agent interface over stdio (30 tools)"
python3 demo/mcp_demo.py ./iacforge

# ----------------------------------------------------------------------
section "Done"
echo "Artifacts written to $OUT_DIR/:"
ls -la "$OUT_DIR"
