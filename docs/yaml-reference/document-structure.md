# Document Structure

[← README](README.md)

---

YAML documentは単一のGraphを表現します。

## 基本構造

```yaml
objects:
  # すべてのEntityとRelationをここに記述
```

## コメント

YAMLのコメントはパース時に破棄され、round-trip変換（parse → serialize）では
保持されません。コメントに意味を持たせないでください。

```yaml
# エリア情報
objects:
  - id: sado-island
    kind: area
    name: Sado Island
    attributes:
      # プライマリロケーション
      status: active
```

上記をparseして再serializeすると、コメント行は出力に含まれません。

## 記述順序

- Objectの順序には意味はありません
- Implementationは可能な限り順序を保持すべきです
