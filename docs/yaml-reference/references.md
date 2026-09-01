# References

[← README](README.md)

---

## シンプル参照

```yaml
source: species-toki
target: forest-osado-beech
```

## 修飾参照（フルパス）

```yaml
source: /sado-island/mt-kinpoku/ground-osado-hill
target: lake-kamo
```

## パス参照

ネストされたエンティティはパス表記で参照します（`entity/child`）：

```yaml
participants:
  - sado-island/species-toki/toki-census-2025
```

## ネストされたエンティティへのパス参照

ネストされたエンティティは親子関係のパスで参照できます：

```yaml
# パス表記: parent/child
participants:
  source: sado-island/species-toki/toki-census-2025
  target: forest-osado-beech
```

パスの各セグメントはエンティティIDに対応します：
- `sado-island` - ルートエリア
- `species-toki` - 種（areaの子）
- `toki-census-2025` - 個体群調査記録（speciesの子）

## 参照ルール

- Referencesは既存のObjectsを指す必要があります
- Unknown referenceは検証エラーとなります
- パス参照は所有権チェーンを検証します（親→子の関係が正しいこと）
