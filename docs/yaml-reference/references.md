# References

[← README](README.md)

---

## シンプル参照

```yaml
source: srv-proxmox-01
target: vm-web-01
```

## 修飾参照（フルパス）

```yaml
source: /region-ap-northeast-1/rack-a01/srv-proxmox-01
target: vm-web-01
```

## インターフェース参照

インターフェースはパス表記で参照します（`entity/interface`）：

```yaml
participants:
  - srv-proxmox-01/eno1
  - sw-core-01/port1
```

## ネストされたエンティティへのパス参照

ネストされたエンティティは親子関係のパスで参照できます：

```yaml
# パス表記: parent/child
participants:
  source: srv-proxmox-01/net-private/eth1
  target: sw-core-01/port1
```

パスの各セグメントはエンティティIDに対応します：
- `srv-proxmox-01` - 親サーバー
- `net-private` - ネットワーク（サーバーの子）
- `eth1` - インターフェース（ネットワークの子）

## 参照ルール

- Referencesは既存のObjectsを指す必要があります
- Unknown referenceは検証エラーとなります
- Interface referenceはパス表記を使用します
- パス参照は所有権チェーンを検証します（親→子の関係が正しいこと）
