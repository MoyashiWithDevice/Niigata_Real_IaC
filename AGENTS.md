# IACForge 開発ガイドライン

## プロジェクト目標

IACForgeは、インフラストラクチャをモデルとして定義し、様々な表現を生成するためのフレームワークです。「インフラは知識」という哲学に基づき、モデルを唯一の信頼できる情報源とします。

## 開発方針

### 1. オブジェクトモデル優先
- オブジェクトモデルはプロジェクトのコア
- YAML、JSON、APIはあくまでモデルの異なる表現
- シリアライズフォーマットに合わせてモデルを設計しない

### 2. エンティティとリレーション
- インフラのすべてのオブジェクトはEntityとして表現
- Entity間のピア・ツー・ピアの関係はRelationとして明示的に定義
- 所有権（ownership）はツリー構造で表現
- その他の関係はすべてRelationで表現

### 3. 人間が読める形式
- インフラデータの主な著者は人間
- プレーンテキストで読みやすい状態を維持
- 生成データが手書きデータを置き換えない

### 4. 拡張性
- コアオブジェクトモデル以外はすべて拡張可能
- Entity Kinds、Relation Types、Views、Validation Rules、Providers、Renderers

### 5. ベンダーニュートラリティ
- コアオブジェクトモデルはベンダ固有の概念を理解しない
- ベンダ固有の情報はProvidersに属する
- モデルはインフラの概念を表現し、実装ではない

## コーディング規約

### 命名規則
- Entity Kind: 小文字、単数形（例: `server`, `vm`, `interface`）
- Relation Type: スネークケース（例: `connects`, `hosts`, `depends_on`）
- プロパティ: スネークケース（例: `cpu`, `memory`, `storage`）
- ID: ケバブケース推奨（例: `srv-proxmox-01`）

### ファイル構造
```
src/
├── core/              # コアオブジェクトモデル
│   ├── entity.go      # Entity基底クラス
│   ├── relation.go    # Relation基底クラス
│   ├── graph.go       # Graph管理
│   └── kinds/         # Entity Kinds定義
├── schema/            # スキーマ定義
├── validation/        # 検証エンジン
├── parser/            # YAMLパーサー
├── query/             # クエリエンジン
├── projection/        # プロジェクションエンジン
├── view/              # ビュー定義
├── renderer/          # レンダリングエンジン
└── extension/         # 拡張システム
```

### コードスタイル
- Go言語を使用
- 不変性（immutability）を重視
- 副作用のない純粋関数を推奨
- エラーハンドリングを適切に実装（エラー値の返却）
- インターフェースによる抽象化

## テスト方針

### ユニットテスト
- すべてのモジュールにユニットテストを作成
- テストフレームワーク: Go標準テスト
- カバレッジ: 80%以上を目標

### 統合テスト
- 主要なユースケースをカバー
- YAMLパーサーのRound-tripテスト
- 検証エンジンの統合テスト

### テスト実行
```bash
# テスト実行
go test ./...

# カバレッジレポート
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 用語集

| 用語 | 定義 |
|------|------|
| Entity | インフラ内のすべてのオブジェクトを表す基本的な単位 |
| Relation | Entity間のピア・ツー・ピアの関係を表す |
| Graph | EntityとRelationのコレクション |
| Ownership | Entity間の階層関係（ツリー構造） |
| Schema | モデルの構造定義 |
| Validation | Graphの整合性チェック |
| Projection | Graphを別のGraphに変換 |
| View | Projectionの出力を解釈する定義 |
| Renderer | Viewをプレゼンテーションに変換 |
| Artifact | Rendererの生成出力 |

## 注意事項

### 仕様書への準拠
- `spec/` フォルダ内の仕様書は絶対的な参照先
- 仕様書と矛盾する実装は禁止
- 仕様書の更新が必要な場合は明示的に確認

### データ整合性
- Entity IDはスコープ内でユニークであること
- 参照は常に存在するオブジェクトを指すこと
- Ownershipはサイクルを含まないこと

### パフォーマンス
- 大規模Graphでも動作すること
- 必要に応じてキャッシュを実装
- 遅延読み込みを検討

### セキュリティ
- 機密情報をコードに埋めないこと
- 外部入力を適切にバリデーション
- 拡張のロード時にサンドボックス化を検討

### シリアライゼーション
- シリアライゼーションフォーマットはYAMLのみ
- JSONタグやJSON対応は不要（仕様書 `spec/03-object-model.md` に基づきYAMLを唯一の表現形式とする）
- Go構造体のstructタグは `yaml:"..."` のみ使用する

## 開発フロー

1. 仕様書を確認
2. 実装計画を立てる（TASK.md参照）
3. コードを実装
4. テストを作成
5. テストを実行
6. レビューを受ける
7. マージ

## 参考資料

- `spec/00-philosophy.md` - プロジェクト哲学
- `TASK.md` - 実装計画
- `spec/concrete/` - 具体的な仕様書