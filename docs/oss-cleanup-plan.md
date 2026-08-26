# OSS公開に向けたコードクリーンアップ計画

IACForgeをOSSとして公開するための第1段階のクリーンアップ計画。
未使用コードの削除・lint修正・安全な重複統合を対象とする。

## 方針

- 対象: 未使用コードの削除 + lint修正 + 挙動を変えない重複統合
- `src/projection/` は今後独立パッケージとして維持せず、将来 Query に統合する方針とする。今回の計画では、未接続の実装とテストを削除し、Query への統合は別タスクとする
- スコープ外（別途判断）: module名リネーム、LICENSE / README / CI整備
- 全工程で以下を実行し回帰を防ぐ:
  - `go build ./...`
  - `go vet ./...`
  - `golangci-lint run ./...`
  - `go test ./...`

## 調査結果の要点

### 孤立したコード

| 対象 | 規模 | 状況 |
|------|------|------|
| `src/projection/` パッケージ | 約2,935行 | 実コードからの import がゼロ。現在は未接続。将来は Query の結果生成・変換機能として再設計する |
| `src/extension/plugin.go` | 約30行 | `Plugin`/`PluginFunc` 型が未使用。実際の読み込みは素の関数型を使用 |
| 多数のコンストラクタ・getter | — | テストのみで使用、または参照ゼロ |

### 重複コード

| 重複 | 場所 |
|------|------|
| WhereClause / Condition / Operator 評価 | query / view の2箇所（toFloat64 も2重）。Projection側は削除対象のため統合対象から除外 |
| 参照整合性チェック | parser.ResolveReferences / validation.ruleDanglingReference / Graph.ValidateIntegrity の3箇所 |
| バリデーション重複報告 | ruleValidReference + ruleValidOwner が ruleDanglingReference の部分集合 |
| ルート数チェック | validation 内で3重実装 |
| フォーマット dispatch | renderer/format.go / mcp/tools_render.go / cmd/main.go の3箇所 |
| エンティティ JSON 変換 | renderer/json.go（`properties`）と query/format.go（`spec`）でキー不一致 |
| Kind 定義メタデータ | core/kinds + schema/core_schema.go + extension/manager.go の3箇所（既に乖離） |
| attributes パース/出力 | parser / serializer で各2箇所 |
| schema builder ヘルパ | core_schema.go と aws/definitions.go |

### OSS公開のための必須対応（今回はスコープ外）

- LICENSE ファイルが無い
- README が無い
- module名が `IACForge`（大文字・非ドメイン）→ 外部から `go get` 不可
- CI (.github/) が無い
- `panic()` が本番コードに4箇所（core/graph.go:63,149 / mcp/session.go:60 / aws/definitions.go:544）
- ドキュメントと実態の不一致（docs/extension-reference.md が存在しない testdata を参照）

---

## Phase 1: 孤立パッケージ `src/projection/` の削除

- 全5ファイル削除（engine.go / operations.go / derived.go / types.go / engine_test.go）
- 実コードからの import がゼロであることを確認済み。削除後も現在の CLI・MCP の実行経路は変わらない
- これは挙動不変の重複整理ではなく、未接続機能の削除である
- Projection の機能は、将来 Query の結果生成・変換機能として再設計・統合する
- 今回のPhaseでは Query への統合実装、Query APIの変更、Projection仕様の再設計は行わない

## Phase 2: 未使用コードの削除

| ファイル | 削除対象 |
|---|---|
| `src/parser/yaml_parser.go` | `ParseReader` |
| `src/parser/yaml_serializer.go` | `SetIndent`, `SerializeTo` |
| `src/core/graph.go` | `MustGetEntity`, `MustGetRelation`, `ValidateIntegrity`（+対応テスト） |
| `src/view/types.go` | `LayoutHint`+`View.Layout`, `View.Audience`, `Group.Owner`, `GroupingRule.Owner`, `NewGroupingRule`, `NewAnnotationRule`, `NewLayoutHint`, `NewEntitySelector`, `NewWhereClause`, `NewCondition`, `NewAnnotation`, `WhereClause.AddCondition`, `View.AddAnnotationRule` |
| `src/view/engine.go` | `ValidateView`（+テスト参照修正） |
| `src/renderer/types.go` | `IconSet`, `Spacing`, `Theme.Icons/Spacing`, `RenderOptions.Scale`, `NewTheme` |
| `src/schema/schema.go` | `PropertyTypeEnum`, `DirectionUndirected` |
| `src/schema/core_schema.go` | `strPtr`, `boolPtr`, `intPtrInt`（テストは `intPtr` へ置換） |
| `src/schema/profile.go` | `AddRule`, `SetRequiredProperties`, `HasRequiredKind`, `HasRequiredRelation`（`HasRule`/`AddRequiredKind`/`AddRequiredRelation` は本番使用のため維持） |
| `src/extension/plugin.go` | ファイル全体 |
| `src/extension/manager.go` | `ErrExtensionNotFound` |
| `src/extension/types.go` | `Manifest.Author/SpecVersion/SchemaVersion` |
| `src/extension/` 各ファイル | test-only の getter 群（entity_kinds / relation_types / root_kinds / validation_rules / renderers） |
| `src/validation/engine.go` | `AllowedRootKinds` |
| `src/validation/types.go` | `Rule.Scope` + `Scope*` 定数（書き込みのみ・未読） |
| `src/core/kinds/kinds.go` | `AllKinds`（テスト修正）, `ValidStatuses`, `IsValidKind` |
| `src/core/types/types.go` | 22-110行のメタデータブロック + test-only 関数（core_schema.go と乖離） |
| `src/extension/builtin/aws/kinds.go` | `IsValidKind` |
| `src/extension/builtin/aws/types.go` | `AllRelationTypes` |

## Phase 3: lint 修正

- `cmd/iacforge/main.go:257` ineffassign、`main.go:322-328` の `_ = sel` 除去
- `src/query/project.go:239,245` S1025（`fmt.Sprintf("%s", str)` → 直接返却）
- `src/renderer/markdown.go` QF1012（`WriteString(fmt.Sprintf(...))` → `fmt.Fprintf`、3箇所）
- MCP の `data, _ := json.MarshalIndent(...)` エラー処理（tools_schema / tools_graph / tools_query / tools_extension）
- テストの errcheck（`AddEntity`/`AddRelation` 等 → `t.Fatalf` パターン、約20箇所）

## Phase 4: 安全な重複統合（挙動不変）

1. **フォーマット dispatch 一元化**: `renderer/format.go` の `RenderFormat` に `svg`/`json` を追加し、`cmd/main.go:251-265` と `mcp/tools_render.go:38-50` の重複スイッチを置換
2. **parser の attributes ヘルパ抽出**: `parseEntity`/`parseRelation`、`buildEntityWithChildren`/`buildRelation` の重複ブロックを共通関数化
3. **条件評価器の共通化**: query/view の `evaluateCondition` + `toFloat64` を共有パッケージへ抽出（view→query の import cycle 回避のため新規パッケージ）。削除対象のProjection実装は統合対象に含めない

### Projection の将来方針

現在の `src/projection/` は仕様上のProjectionモデルを実装しているが、CLI、MCP、その他の実行経路から利用されていない。

今後はProjectionを独立パッケージとして維持せず、Queryの結果生成・変換機能として統合する方針とする。将来のQuery統合では、以下の機能をQueryモデルの拡張候補とする。

- select
- filter
- traverse
- aggregate
- group
- transform

Queryへの統合は今回のクリーンアップ計画の対象外とし、別タスクとして設計・実装する。Query統合時には、Graphを変更しないこと、結果の決定性、ViewおよびRendererとの責務分担を改めて定義する。

## Phase 5: プラグイン(.so)機構のテスト追加 + ドキュメント修正

- `testdata/plugins/` に実プラグイン例を作成し、`LoadFromDir` の実 .so 読み込みテストを追加
- `docs/extension-reference.md` の誤記修正
  - 「起動時自動ロード」の記述（実際は `IACFORGE_EXTENSIONS` / `--extensions` のみ）
  - 存在しない `testdata/plugins/testplugin/main.go` への参照
  - 存在しないヘルパ名（`float6Ptr` 等）への言及

## Phase 6: docsと実装の整合性確認・更新

docsに書かれた機能、入力形式、出力形式、実行条件が実装と異なると、利用者が正しい手順でも動かせない。OSS公開前に、docsを実装の実際の挙動へ同期する。

### 対象ドキュメント

| 対象 | 確認・更新内容 |
|---|---|
| `docs/extension-reference.md` | 外部 `.so` のロード条件、環境変数・CLIオプション、実在するサンプルパス、実在するpointer helper、Manifestの現行フィールド、拡張ポイントの一覧を実装と一致させる |
| `docs/yaml-reference/` | `objects`、`attributes`、`spec`、ネスト、ID自動生成、owner、participant、パス参照、`@`プロパティ参照の説明をparser/serializerの実際の挙動と照合する |
| `docs/yaml-reference/document-structure.md` | コメントがround-tripで保持されるという記述を検証し、未対応なら削除または「保持されない」と明記する |
| `docs/yaml-reference/validation.md` | 必須項目、status、root、ownership、参照、AWS拡張のroot権限、validation ruleのIDとseverityを実装・テストと一致させる |
| `docs/yaml-reference/entity-kinds.md` / `relation-types.md` | core schemaとAWS extensionから生成されるkind/type、プロパティ、方向、参加可能kind、件数を照合する |
| `spec/concrete/21-rendering.md` とRenderer関連docs | 実装済みの `svg`、`mermaid`、`markdown`、`json` と、未実装のPNG/PDF/Graphviz/D2/HTML/CSVを区別する。CLI/MCPの対応形式と `renderer.RenderFormat` の対応形式も分けて記載する |
| `spec/10-projection-model.md` / `spec/concrete/20-projection-model.md` | Projectionを将来Queryへ統合する方針に合わせ、現行OSSのサポート対象か、将来設計かを明記する |
| `docs/html/` | 手書きMarkdownの変更後に生成物を再生成するか、生成元と更新手順を明記する。古い仕様・件数・リンクを公開しない |

### 更新手順

1. docs内の機能名、形式名、環境変数、CLIオプション、パス、件数、severity、例で使用するAPIを抽出する
2. parser、serializer、CLI、MCP、schema、extension、renderer、validationの実装と対応付ける
3. 実装に存在しない記述は、実装を追加せずdocsから削除するか、未対応・将来予定として明記する
4. 実装に存在するがdocsにない利用者向け挙動は、最小限の説明と実行例を追加する
5. 仕様書と実装の不一致は、仕様を変更するか実装を直すかを個別に決定する。docsだけを更新して仕様との矛盾を隠さない
6. サンプルYAML、CLI例、plugin例を実際に実行する検証を追加する

### 整合性チェック項目

- docsで参照しているファイル、ディレクトリ、関数、型、helperがすべて存在する
- docsに記載されたCLI/MCPのformat、option、環境変数が実装で受け付けられる
- YAMLサンプルをparseし、必要なものはserialize後に再parseできる
- docs記載のEntity kind、Relation type、status、validation ruleがschema/extensionに存在する
- 未実装の機能を「対応済み」「自動」「保持する」と表現していない
- Markdownを正とする場合、`docs/html/`の生成物を同じ変更に含めるか、生成手順をCIで検証する

## 検証

各 Phase 終了時に以下を実行:

```bash
go build ./...
go vet ./...
golangci-lint run ./...
go test ./...
```

Phase 5およびPhase 6では、上記に加えて以下を実行する:

```bash
# docs内の存在しない参照・旧API・未対応形式を検索する
rg 'testdata/plugins|float6Ptr|PNG|PDF|Graphviz|D2|HTML|CSV' docs

# サンプルと公開手順を実環境で確認する
go run ./cmd/iacforge --help
```

## スコープ外（別途判断推奨）

- **出力形式を変える統合**:
  - renderer/json.go と query/format.go のキー差異（`properties` vs `spec`）の統一
  - validation の重複報告解消（`ruleValidReference` ⊂ `ruleDanglingReference`）
  - core kind メタデータの一元化（core/kinds + schema/core_schema.go + extension/manager.go）
- module 名リネーム（`IACForge` → `github.com/<org>/iacforge`）
- **ProjectionのQuery統合**: Phase 1では独立パッケージを削除する。select/filter/traverse/aggregate/group/transform等のQuery統合は別タスクとする
- **Projection仕様書の改訂**: ProjectionをQueryへ統合する方針に合わせ、`spec/10-projection-model.md` と `spec/concrete/20-projection-model.md` の扱いを別途決定する
- LICENSE / README / CI / CONTRIBUTING の整備
- 作業ツリー内のビルド済みバイナリ `iacforge`（21MB、gitignore 済み）の除去

## 想定効果

- 削除量: 約3,800〜4,000行（全約28,000行の約14%。内 projection が2,935行）
- 重複コードの統合による保守性向上
- lint クリーンな状態（OSS 公開時の品質ゲートとして利用可能）
