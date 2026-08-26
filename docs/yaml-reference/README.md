# YAML Reference

IACForgeのYAMLファイル作成のための完全版リファレンスです。

## 目次

| ドキュメント | 内容 |
|--------------|------|
| [Document Structure](document-structure.md) | YAMLドキュメントの基本構造 |
| [Entity Syntax](entity-syntax.md) | Entity共通プロパティ、ステータス値 |
| [Entity Kinds](entity-kinds.md) | 全Entity種類の定義とプロパティ |
| [Relation Syntax](relation-syntax.md) | Relation共通構文、participantフォーマット |
| [Relation Types](relation-types.md) | 全Relation種類の定義とプロパティ |
| [References](references.md) | 参照構文（シンプル、修飾、インターフェース） |
| [Validation](validation.md) | 検証ルール、命名規則、Graph制約 |
| [Example](example.md) | 完全なインフラモデルの例 |
| [AWS Entity Kinds](aws-entity-kinds.md) | AWS拡張のEntity種類（47種）とプロパティ |
| [AWS Relation Types](aws-relation-types.md) | AWS拡張のRelation種類とAugmented Core Relations |

## 共通ルール

| ルール | 詳細 |
|--------|------|
| 必須フィールド | [Validation - 必須フィールド](validation.md#必須フィールド) |
| ネストルール | [Entity Syntax - ネスト定義](entity-syntax.md#ネスト定義nested-entity-definition) |
| ステータス値 | [Entity Syntax - ステータス値](entity-syntax.md#ステータス値) |
| 命名規則 | [Validation - 命名規則](validation.md#命名規則) |
