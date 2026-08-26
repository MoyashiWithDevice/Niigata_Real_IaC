# IACForge 改善タスク(ギャップ分析)

`spec/concrete/14-entity-kinds.md` / `15-relation-types.md` / `16-core-schema.md` / `22-aws-extension.md` と、
実装(`src/schema/core_schema.go`, `src/extension/builtin/aws/`)・MCPの生スキーマを突き合わせた結果、
不足・表現しきれていない項目をタスクとして洗い出した。

- 対象: `core`=コアスキーマ, `aws`=AWS拡張, `common`=共通機構
- 優先度: 高=今すぐ対応すべき乖離, 中=よく遭遇する不足, 低=あると便利

---

## 1. 不具合・仕様乖離の修正

| # | 対象 | 優先度 | 内容 | 根拠 |
|---|------|--------|------|------|
| B1 | aws | 高 | `aws.registers` が生スキーマに存在しない。コード(`src/extension/builtin/aws/relations.go:97`)と仕様(`22-aws-extension.md:982`)には定義済みだが、MCPの `list_relation_types` に出力されない。実行中サーバが古いビルドの可能性が高いため、バイナリ再ビルド/再起動で検証し、解消しない場合は登録処理を調査する | コードと生スキーマの乖離 |
| B2 | common | 高 | `pattern` 制約が実装されていない。`src/schema/schema.go` で `_ = c.Pattern` として破棄されていたため、宣言しても検証されなかった。正規表現コンパイルによる検証を実装した | 16-core-schema.md の Constraint Types |
| B3 | core | 中 | enumを宣言しているのに未検証の文字列プロパティが多かった。`server.platform`, `network.network_type`, `storage.protocol`, `cluster.cluster_type` などにenum制約を定義した | 14-entity-kinds.md の Property 定義 |
| B4 | core | 低 | `map` 型は要素スキーマを定義できなかった。`ValueProperty` を追加し map 要素のスキーマ検証を実装した | 16-core-schema.md の Property Types |
| B5 | core | 低 | `list[object]` の要素に未知キーがあっても弾かれなかった。定義済みサブプロパティ以外のキーを弾くよう `ValidateProperty` を拡張した | 16-core-schema.md の Structured Lists |

---

## 2. 不足している Entity Kind

### 2-1. 物理インフラ

| Kind | 対象 | 優先度 | 説明 |
|------|------|--------|------|
| building / room / row | core | 中 | region→rack の間に施設階層(建物・フロア・列)を表現するKindがない |
| access_point | core | 低 | 無線AP・SSID・チャネルを表現するKindがない |
| gpu | core | 低 | GPUアクセラレータを表現するKindがない(server/vmの属性としても代替不可) |

### 2-2. ネットワーク

| Kind | 対象 | 優先度 | 説明 |
|------|------|--------|------|
| load_balancer | core | 中 | コアにLB Kindがない。オンプレLB(F5等)は `aws.load_balancer` で表現できない |
| route_table / route | core | 中 | コアにルーティングKindがない(`aws.route_table` のみ)。オンプレのスタティック/ダイナミックルートを表現できない |
| bgp_peer / as | core | 中 | BGPピア・AS番号・ルーティングプロトコルを表現するKind/属性がない |
| nat | core | 低 | コアにNAT Kindがない |
| vpn / tunnel | core | 低 | トンネル・VPN接続を表現するKind/関係がない |
| dns_server / dhcp / ntp | core | 低 | DNS/DHCP/NTP サーバーを表現するKindがない(ネットワーク属性の `dns_servers` のみ) |
| ip_pool / ipam | core | 低 | IPアドレス割当・IPAM参照を表現するKindがない |

### 2-3. ストレージ / データベース

| Kind | 対象 | 優先度 | 説明 |
|------|------|--------|------|
| database / database_cluster | core | 中 | DBが `application` か `aws.rds` にしか表現できない。コアにDB Kindがない |
| snapshot | core | 中 | コアにスナップショットKindがない(`aws.ebs_snapshot` のみ) |
| backup_schedule | core | 中 | バックアップ方針を表現するKindがない(`backs_up` 関係の属性のみ) |
| object_storage / bucket | core | 低 | コアにオブジェクトストレージKindがない(`aws.s3_bucket` のみ) |
| datastore / pool / tier | core | 低 | ストレージプール・階層を表現するKindがない |

### 2-4. アイデンティティ / セキュリティ

| Kind | 対象 | 優先度 | 説明 |
|------|------|--------|------|
| user / group / role / policy | core | 中 | IAMにしか存在しない。オンプレ・自社システムのアイデンティティを表現できない |
| service_account | core | 低 | サービスアカウントを表現するKindがない |
| secret / credential | core | 低 | 機密・資格情報(実値ではなく参照)を表現するKindがない |
| certificate / key | core | 低 | TLS証明書・鍵を表現するKindがない |

### 2-5. 運用 / 監視

| Kind | 対象 | 優先度 | 説明 |
|------|------|--------|------|
| monitor / alert / metric | core | 中 | CloudWatchにしか存在しない。オンプレ監視を表現できない |
| notification_channel | core | 低 | アラート通知先を表現するKindがない |
| log_source / log | core | 低 | ログ出力元・ロググループを表現するKindがない |

### 2-6. 論理 / 編成

| Kind | 対象 | 優先度 | 説明 |
|------|------|--------|------|
| environment | core | 中 | dev/staging/prod をKindとして表現する方法がない(labels 頼み) |
| project / team / tenant | core | 低 | 所有・編成単位を表現するKindがない |

### 2-7. 配布 / コンテナオーケストレーション

| Kind | 対象 | 優先度 | 説明 |
|------|------|--------|------|
| deployment / release / pipeline / job / cron | core | 低 | 配布・ジョブスケジュールを表現するKindがない |
| kubernetes.* (pod / deployment / service / ingress / namespace / configmap / pvc) | ext | 低 | K8sワークロードを表現する拡張がない(`cluster` で大雑把に代替している) |

---

## 3. 既存 Kind の不足属性

| Kind | 不足属性 | 優先度 | 説明 |
|------|----------|--------|------|
| server | hostname / fqdn | 中 | ホスト名をラベル頼みでしか持てない |
| server | gpu | 低 | GPU構成を表現できない |
| server | psu(冗長電源) / power_watts | 低 | 消費電力・PSU冗長性を表現できない |
| server | form_factor / rack_u | 低 | 1U/2U・ラック内位置を表現できない |
| server | boot_mode (uefi/bios) / firmware_version | 低 | ファームウェア情報を表現できない |
| vm | hostname | 中 | 同上 |
| vm | snapshot | 中 | スナップショット参照ができない |
| vm | ha_policy / migration_policy | 低 | HA・移行ポリシーを表現できない |
| vm | boot_order / firmware | 低 | ブート順を表現できない |
| interface | ipv6 | 中 | IPv6アドレスを表現できない(`ip_address` は文字列listのみ) |
| interface | admin_state / link_state | 中 | 運用状態(up/down)を表現できない(statusとは別) |
| interface | duplex / poe | 低 | 二重モード・PoE供給を表現できない |
| interface | parent_interface | 低 | bond/VRRP等の親子関係を明示的に張る属性が無い(ネストはある) |
| network | ipv6_cidr | 中 | IPv6ネットワークを表現できない |
| network | domain_name / search_domains | 低 | DNSドメイン情報がない |
| network | ntp_servers / dhcp 設定 | 低 | NTP/DHCP設定がない |
| network | ipam 参照 | 低 | IPAM連携がない |
| network | mtu | 低 | MTUが interface にしかない |
| container | env / volume_mounts | 中 | 環境変数・マウントを表現できない |
| container | replicas / restart_policy / health_check | 中 | スケール・再起動・ヘルスチェックを表現できない |
| container | cpu_limit / memory_limit の構造化 | 低 | 現在 string で非構造(cpu/memoryリソースを数値で検証不可) |
| volume | iops / throughput | 中 | 性能特性を表現できない |
| volume | encrypted / wwn / lun / pool | 低 | 暗号化・識別子・所属プールを表現できない |
| storage | pool / tier / datastore | 中 | プール・階層を表現できない |
| storage | encryption / replication | 低 | 暗号化・レプリケーション設定を表現できない |
| application | environment / runtime | 中 | 実行環境を表現できない |
| application | endpoint / health_check | 低 | エンドポイント・ヘルスチェックを表現できない |
| open_port | service 名の enum | 低 | `process` が自由文字列で、サービス名を正規化できない |
| power_distribution | 消費電力 / 負荷 / 出力口ごとの接続 | 低 | PDU負荷を表現できない |
| cluster | kubernetes_version / platform | 低 | クラウドK8sを単一Kindで表現するのが弱い |
| region | クラウドプロバイダ識別 | 中 | オンプレ/クラウドを横断するときに provider を表現できない |
| rack | row / aisle / 位置 | 低 | 施設内位置を表現できない |

---

## 4. 不足している Relation Type

| # | 対象 | 優先度 | 内容 |
|---|------|--------|------|
| R1 | core | 中 | `routes`(ルーティング)・`peers`(BGP) が無い。ルーティング関係を張れない |
| R2 | core | 中 | `deploys_to` / `exposes` / `publishes` が無い。配布・公開関係を張れない |
| R3 | core | 中 | `logs_to` が無い。ログ出力関係を張れない |
| R4 | core | 低 | `authenticates` / `authorizes` が無い。認証・認可関係を張れない |
| R5 | core | 低 | `schedules` が無い。ジョブスケジュール関係を張れない |
| R6 | core | 低 | `connects` が interface-interface 限定(`core_schema.go:326`)。cable↔interface 等の直接接続を張れない(仕様 `15-relation-types.md:476` と実装は一致) |
| R7 | common | 低 | 関係に時間属性(有効期間 / SLA)を載せる型が無い |

---

## 5. 横断的な表現力の限界

| # | 対象 | 優先度 | 内容 |
|---|------|--------|------|
| C1 | core | 中 | `status` enum に `provisioning` / `error` / `decommissioned` が無い。計画・運用中・削除済みの区別をラベルでしか表現できない |
| C2 | common | 中 | `labels` / `extensions` が自由帳になっていて、タイポ・誤ったキーが検証されない |
| C3 | common | 中 | 日時 / 期間 / TTL / cron を表す型が無い(すべて文字列)。`backs_up.schedule` 等も文字列 |
| C4 | common | 低 | 単位(speed, capacity)を型として表現できない。数値は無単位 |
| C5 | core | 低 | 所有・ガバナンス系のメタデータ(owner, team, cost_center, compliance, sla)が型として無い |

---

## 対応の優先順位(提案)

1. **B1** — 仕様と実装の乖離なので最優先で修正する
2. **2-2(load_balancer / route_table), 2-3(database / snapshot / backup_schedule), 3-1(hostname/fqdn), 4-R1** — モデリング上よく遭遇する不足
3. その他は「コアに追加するか、拡張(例: onprem, kubernetes)として追加するか」を判断してから着手する

> 注: 拡張で追加する場合は `spec/concrete/` に仕様書を先に作り、実装はネームスペース付きKind(`namespace.kind`)として追加すること(AGENTS.md の開発フロー参照)。
