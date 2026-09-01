#!/usr/bin/env bash
# Niigata full-feature demo script (Niigata nature & tourism resources).
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

EXT_DIR="$OUT_DIR/extensions"

if [[ $SKIP_BUILD -eq 0 ]]; then
  section "0. Build"
  go build -o niigata ./cmd/niigata
  # Go plugin extensions must be built with the same toolchain as the host
  # binary, so the sample plugin is rebuilt together with it.
  mkdir -p "$EXT_DIR"
  go build -buildmode=plugin -o "$EXT_DIR/testplugin.so" ./testdata/plugins/testplugin
fi

run ./niigata version

MODEL=demo/niigata/model.yaml

# ----------------------------------------------------------------------
section "1. validate: YAML resource model validation"
run ./niigata validate "$MODEL"

echo
echo "--- negative case: schema violations (expect FAILED, exit != 0) ---"
./niigata validate demo/negative/bad-model.yaml || true

echo
echo "--- negative case: dangling reference (expect Parse error) ---"
./niigata validate demo/negative/bad-parse.yaml || true

# ----------------------------------------------------------------------
if [[ -f $EXT_DIR/testplugin.so ]]; then
  section "1b. validate: runtime extension (Go plugin) loading"
  run ./niigata validate "$MODEL" --extensions "$EXT_DIR"
fi

# ----------------------------------------------------------------------
section "2. info: graph summary (directory scan: all .yaml merged)"
run ./niigata info demo/niigata/

# ----------------------------------------------------------------------
section "3. render: view -> artifact (markdown / mermaid / json / svg)"
run ./niigata render "$MODEL" --format markdown
run ./niigata render "$MODEL" --format mermaid
echo "\$ ./niigata render $MODEL --format json   (truncated)"
./niigata render "$MODEL" --format json | head -30
echo "..."
./niigata render "$MODEL" --format svg    --output "$OUT_DIR/graph.svg"
./niigata render "$MODEL" --format mermaid --output "$OUT_DIR/niigata.mmd"
./niigata render "$MODEL" --format markdown --output "$OUT_DIR/niigata.md"
# A themed variant for dashboards and dark surfaces.
./niigata render "$MODEL" --format svg --theme dark --output "$OUT_DIR/graph-dark.svg"
# Geographic map layout: nodes placed by spec latitude/longitude.
./niigata render "$MODEL" --format svg --layout map --output "$OUT_DIR/graph-map.svg"
# Force-directed layout: physics-based node placement.
./niigata render "$MODEL" --format svg --layout force-directed --output "$OUT_DIR/graph-force.svg"
ls -la "$OUT_DIR"

# ----------------------------------------------------------------------
section "4. query: filter by entity kind or relation type, multiple formats"
run ./niigata query "$MODEL" --kind species --format text
run ./niigata query "$MODEL" --type depends_on --format json
run ./niigata query "$MODEL" --kind tourism_spot --format mermaid
run ./niigata query "$MODEL" --kind species --format markdown
run ./niigata query "$MODEL" --kind species --format json --output "$OUT_DIR/query-species.json"

# Builtin extension kinds contributed by niigata.agri-wildlife.
section "4b. query: builtin extension kinds (agri-wildlife)"
run ./niigata query demo/niigata/agri-wildlife.yaml --kind wildlife_incident --format text
run ./niigata query demo/niigata/agri-wildlife.yaml --kind crop_harvest --format text
run ./niigata validate demo/niigata/

# ----------------------------------------------------------------------
section "5. mcp: AI-agent interface over stdio"
python3 demo/mcp_demo.py ./niigata

# ----------------------------------------------------------------------
section "Done"
echo "Artifacts written to $OUT_DIR/:"
ls -la "$OUT_DIR"
