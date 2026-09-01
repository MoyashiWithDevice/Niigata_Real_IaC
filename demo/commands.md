# demo/run-demo.sh が実行する niigata コマンド一覧

`./demo/run-demo.sh` が実行する `./niigata` コマンドのみを列挙したリファレンスです。
各コマンドの詳細な説明はスクリプト本体のセクションコメントを参照してください。

前提: リポジトリルートで実行し、`go build -o niigata ./cmd/niigata` 済みとします。

## 0. ビルド・バージョン

```bash
./niigata version
```

## 1. validate: モデル検証

```bash
# 正常系: サンプルモデルの検証
./niigata validate demo/niigata/model.yaml

# 異常系: スキーマ違反 (FAILED, 終了コード != 0 を期待)
./niigata validate demo/negative/bad-model.yaml

# 異常系: 存在しない参照 (Parse error を期待)
./niigata validate demo/negative/bad-parse.yaml
```

## 1b. validate: Go plugin 拡張のランタイムロード

プラグイン (`demo/out/extensions/testplugin.so`) がビルド済みの場合のみ実行されます。

```bash
./niigata validate demo/niigata/model.yaml --extensions demo/out/extensions
```

## 2. info: グラフサマリー (ディレクトリスキャン: 全 .yaml をマージ)

```bash
./niigata info demo/niigata/
```

## 3. render: view -> artifact

```bash
./niigata render demo/niigata/model.yaml --format markdown
./niigata render demo/niigata/model.yaml --format mermaid
./niigata render demo/niigata/model.yaml --format json          # head -30 で先頭のみ表示
./niigata render demo/niigata/model.yaml --format svg           --output demo/out/graph.svg
./niigata render demo/niigata/model.yaml --format mermaid       --output demo/out/niigata.mmd
./niigata render demo/niigata/model.yaml --format markdown      --output demo/out/niigata.md
./niigata render demo/niigata/model.yaml --format svg --theme dark --output demo/out/graph-dark.svg
./niigata render demo/niigata/model.yaml --format svg --layout map --output demo/out/graph-map.svg
./niigata render demo/niigata/model.yaml --format svg --layout force-directed --output demo/out/graph-force.svg
```

## 4. query: kind / relation type によるフィルタ

```bash
./niigata query demo/niigata/model.yaml --kind species --format text
./niigata query demo/niigata/model.yaml --type depends_on --format json
./niigata query demo/niigata/model.yaml --kind tourism_spot --format mermaid
./niigata query demo/niigata/model.yaml --kind species --format markdown
./niigata query demo/niigata/model.yaml --kind species --format json --output demo/out/query-species.json
```

## 4b. query: ビルトイン拡張 kind (niigata.agri-wildlife)

```bash
./niigata query demo/niigata/agri-wildlife.yaml --kind wildlife_incident --format text
./niigata query demo/niigata/agri-wildlife.yaml --kind crop_harvest --format text
./niigata validate demo/niigata/
```

## 5. mcp: stdio 経由の MCP サーバ

`niigata` コマンドそのものではありませんが、内部で以下を起動します。

```bash
./niigata mcp --stdio
```
