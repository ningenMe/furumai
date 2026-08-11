---
status: draft
created: 2026-08-11
related: docs/core-design-direction.md, issues/0006-adapter-capability-catalog.md
---

# Adapter Capability Catalog

Furumaiが最終的にカバーすべき Stimulus（`given`/`when`）と
Observation（`then`）を、実装の優先順位や着手有無を問わず一望するための
カタログ。**このドキュメントはコードを書くためのものではなく、網羅する
べき構造を可視化するためのもの。** 個別のメソッドシグネチャは今後の
実装で変わりうる。

対応する design docの推奨（[`core-design-direction.md`](./core-design-direction.md)
5章 Core architecture）通り、`given`と`when`は同じStimulus adapterを
共有する。

## 大分類

具体的なシステム（MySQL、Kafkaなど）の前に、まず操作の形が共通する
大分類を定義する。同じ大分類に属するシステムは、Stimulus/Observationの
capability形状を共通化しやすい。

| 大分類 | 定義 | 該当システム例 |
|---|---|---|
| HTTP | リクエスト/レスポンス型のWeb API | — |
| Process | プロセスの起動と標準入出力/終了コード | shell command |
| RDB（Relational DB） | 行・カラムベースのデータ操作 | MySQL, PostgreSQL |
| KVS（Key-Value Store） | キーに紐づく値・TTLの操作 | Redis |
| MQ（Messaging / Message Queue） | publish/consumeによる非同期メッセージング | Kafka, SQS, RabbitMQ |
| RPC（将来） | スキーマ付きの手続き呼び出し | gRPC, GraphQL |
| Object Storage（将来） | key付きバイナリオブジェクトの操作 | S3等 |
| Inbound Trigger（将来） | 外部からの受動的な刺激・トリガー | Webhook受信, Cron |

HTTPとProcessはそれ自体が1システムなので大分類=システムだが、RDB・KVS・
MQは複数の具体的な製品が同じ形のcapabilityを共有する。adapter実装は
「大分類ごとの共通interface」+「製品ごとの差分吸収層」という2層構成を
想定する。

## Tier（優先順位）

| Tier | 対象システム | 位置づけ |
|---|---|---|
| 1 | HTTP, Process, MySQL, Kafka | design doc 9章のMVPスコープで既に決定済み |
| 2 | PostgreSQL, Redis, Generic Message Queue | issue #1で名指しされているが未着手 |
| 3 | gRPC, GraphQL, Object Storage, Webhook受信, Cron/batch trigger | issueには無いが一般的に必要になりうる将来候補 |

DB（RDB）の優先順位は **MySQL → PostgreSQL** とする（開発者が使い慣れて
いるため）。RDB同士はSQL方言差分（placeholder記法 `?` vs `$1` 等）は
あるが行・カラムベースの操作という共通の形を持つため、MySQLで
capabilityを固めてからPostgreSQLへ横展開するのは低コストな想定。
KVS（Redis）はRDBとは形が根本的に異なるため別カテゴリとし、RDBの
capabilityがある程度固まってから着手する方が、大分類共通部分と
製品固有部分の切り分けがしやすい。

## HTTP（Tier 1）

**Stimulus（given/when）**

- `Get` / `Post` / `Put` / `Patch` / `Delete` / `Head` / `Options`
- 共通オプション: header, query param, path param, timeout, redirect追従有無, cookie送信
- body: JSON, form-urlencoded, multipart/form-data, raw bytes
- 認証: Basic, Bearer token, カスタムheader

**Observation（then）**

- `StatusCode`（完全一致） / `StatusCodeClass`（2xx/4xx/5xx等）
- `Header` / `HeaderExists` / `HeaderAbsent`
- `BodyEquals` / `BodyContains` / `BodyMatches`（regex）/ `BodyEmpty`
- `BodyJSONEquals`（構造比較） / `BodyJSONPath`（部分抽出+比較）
- `ContentType`
- `ResponseTimeWithin`（duration）

## Process（Tier 1、shell command）

**Stimulus（given/when）**

- `Run(cmd, args, opts)` — opts: env, cwd, stdin, timeout

**Observation（then）**

- `ExitCode`
- `Stdout` の `Equals` / `Contains` / `Matches` / `JSONEquals`
- `Stderr` の `Equals` / `Contains` / `Matches`
- `DurationWithin`

## RDB: MySQL（Tier 1）/ PostgreSQL（Tier 2）

**Stimulus（given/when）**

- `Exec(sql, args...)` — 任意SQLの実行
- `Seed(table, rows...)` — 構造化されたseedヘルパー
- `Truncate(table...)`

**Observation（then）**

- `RowExists(table, conditions)`
- `RowCount(table, conditions)` の `Equals`
- `RowEquals(table, conditions, wantRow)` — 1行の全カラム比較
- `ColumnEquals(table, conditions, column, want)`
- `QueryEquals(sql, args, want)` — 任意クエリ結果の比較
- `NoRows(table, conditions)`

MySQLとPostgreSQLはこのcapabilityを共通interfaceとして共有し、
adapter実装側でdriver・SQL方言差分（placeholder記法等）を吸収する
想定。

## KVS: Redis（Tier 2）

**Stimulus（given/when）**

- `Set(key, value, opts: TTL)`
- `Del(key...)`
- `Expire(key, ttl)`
- データ構造別操作（`HSet`/`LPush`/`SAdd`/`ZAdd`等、優先度は低め）
- `FlushDB`（テスト間クリーンアップ用。危険な操作なので明示的opt-inを必須にする）

**Observation（then）**

- `Exists` / `NotExists`
- `ValueEquals(key, want)`
- `TTLWithin(key, min, max)`
- `KeyPatternCount(pattern, want)`
- データ構造別assert（`ListEquals`/`SetContains`/`HashFieldEquals`等、優先度は低め）

## MQ: Kafka（Tier 1）

**Stimulus（given/when）**

- `Publish(topic, key, value, headers, opts: partition)`
- `PublishJSON(topic, key, valueStruct)`

**Observation（then）**

- `ConsumeMessage(topic, within)` — 1件受信
- `MessageEquals` / `MessageJSONEquals(topic, want, within)`
- `NoMessage(topic, within)` — 一定時間メッセージが来ないことの検証（否定的検証）
- `MessageCount(topic, want, within)`
- `MessageHeaderEquals(topic, key, want)`

## MQ: Generic Queue（Tier 2、例: SQS/RabbitMQ）

Kafkaと同じMQ大分類のcapability（Publish/Consume）を踏襲しつつ、
queue特有の「消費したら消える」「順序保証が無い場合がある」という
性質から、ack/visibility timeoutのような概念が追加で必要になる。
最初にどの製品をサポートするかは未決。

## 横断的な基盤（Assertion API / 非同期待ち）

個別adapter・個別大分類に属さない、共通のObservation基盤。

- 比較プリミティブ: `Equal` / `Contains` / `Matches`（regex）/ `GreaterThan` 等
- **`Eventually(fn, timeout, interval)`**: worker処理待ちなど非同期な
  シナリオ（issue #1で示されている「Kafka publish → workerが処理 →
  HTTP API呼び出し → DB state検証」のような流れ）では、単発チェック
  ではなく「一定時間内に条件が満たされることを待つ」ポーリング型の
  assertionが必須になる。HTTP/RDB/MQ/KVSいずれの `then` にも横断的に
  必要になるため、個別adapterのメソッドではなくAssertion API層の
  共通基盤として設計する。
- assertionの合成: `Not` / `And` / `Or`（優先度は低め）

## Tier 3: 将来候補

issueには明示されていないが、一般的なサーバーサイドシステムで必要に
なりうるもの。

- **RPC（gRPC, GraphQL）**: Stimulus（Unary/Streaming call、query/
  mutation実行）、Observation（response、status code/errors配列、trailer）
- **Object Storage（S3等）**: Stimulus（PutObject）、Observation
  （ObjectExists、ObjectContentEquals）
- **Inbound Trigger（Webhook受信）**: 対象システムからのwebhookを
  Furumai側で受け止めて観測するための簡易HTTPサーバー（Observation側
  の特殊系）
- **Inbound Trigger（Cron/scheduled batch trigger）**: 時刻を進める/
  手動トリガーする（Stimulus側の特殊系）
- **汎用Health/Readiness check**: TCP/HTTPヘルスチェック
  （Environment Managerとの境界領域）

## 未決事項

- Tier 2以降（PostgreSQL/Redis/Generic Queue）の着手順序
- `Eventually`等の横断的基盤を実装上どのレイヤーに置くか（design doc
  5章のCore architectureとの対応関係）
- Generic Queueとして最初にサポートする具体的な製品（SQS/RabbitMQ等）
