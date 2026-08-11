---
status: draft
created: 2026-08-11
related: docs/core-design-direction.md, issues/0006-adapter-capability-catalog.md
---

# Adapter Capability Catalog

Furumaiが最終的にカバーすべき Stimulus（`given`/`when`）と
Observation（`then`）を一望するためのカタログ。**このドキュメントは
コードを書くためのものでも、実装状況を追うためのものでもなく、網羅
すべき構造を可視化するためのもの。** 個別のメソッドシグネチャは今後の
実装で変わりうる。何がどこまで実装済みかは `issues/` を参照する
（このドキュメントの対象外）。

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
| RPC | スキーマ付きの手続き呼び出し | gRPC, GraphQL |
| Object Storage | key付きバイナリオブジェクトの操作 | S3等 |
| Inbound Trigger | 外部からの受動的な刺激・トリガー | Webhook受信, Cron |

HTTPとProcessはそれ自体が1システムなので大分類=システムだが、RDB・KVS・
MQ・RPCは複数の具体的な製品が同じ形のcapabilityを共有する。adapter
実装は「大分類ごとの共通interface」+「製品ごとの差分吸収層」という
2層構成を想定する。

## HTTP

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

## Process（shell command）

**Stimulus（given/when）**

- `Run(cmd, args, opts)` — opts: env, cwd, stdin, timeout

**Observation（then）**

- `ExitCode`
- `Stdout` の `Equals` / `Contains` / `Matches` / `JSONEquals`
- `Stderr` の `Equals` / `Contains` / `Matches`
- `DurationWithin`

## RDB（MySQL, PostgreSQL）

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

## KVS（Redis）

**Stimulus（given/when）**

- `Set(key, value, opts: TTL)`
- `Del(key...)`
- `Expire(key, ttl)`
- データ構造別操作（`HSet`/`LPush`/`SAdd`/`ZAdd`等）
- `FlushDB`（テスト間クリーンアップ用。危険な操作なので明示的opt-inを必須にする）

**Observation（then）**

- `Exists` / `NotExists`
- `ValueEquals(key, want)`
- `TTLWithin(key, min, max)`
- `KeyPatternCount(pattern, want)`
- データ構造別assert（`ListEquals`/`SetContains`/`HashFieldEquals`等）

## MQ（Kafka）

**Stimulus（given/when）**

- `Publish(topic, key, value, headers, opts: partition)`
- `PublishJSON(topic, key, valueStruct)`

**Observation（then）**

- `ConsumeMessage(topic, within)` — 1件受信
- `MessageEquals` / `MessageJSONEquals(topic, want, within)`
- `NoMessage(topic, within)` — 一定時間メッセージが来ないことの検証（否定的検証）
- `MessageCount(topic, want, within)`
- `MessageHeaderEquals(topic, key, want)`

## MQ（Generic Queue、例: SQS/RabbitMQ）

Kafkaと同じMQ大分類のcapability（Publish/Consume）を踏襲しつつ、
queue特有の「消費したら消える」「順序保証が無い場合がある」という
性質から、ack/visibility timeoutのような概念が追加で必要になる。

## RPC（gRPC, GraphQL）

**Stimulus（given/when）**

- gRPC: Unary call、Streaming call
- GraphQL: query/mutation実行

**Observation（then）**

- gRPC: response、status code、trailer
- GraphQL: response body、errors配列

## Object Storage（S3等）

**Stimulus（given/when）**

- `PutObject`

**Observation（then）**

- `ObjectExists`
- `ObjectContentEquals`

## Inbound Trigger（Webhook受信 / Cron）

大分類として他と異なり、「Furumai側が受動的に刺激を観測する」もしくは
「時刻を進める/手動発火する」という特殊な形を取る。

- Webhook受信: 対象システムからのwebhookをFurumai側で受け止めて観測
  するための簡易HTTPサーバー（Observation側の特殊系）
- Cron/scheduled batch trigger: 時刻を進める/手動トリガーする
  （Stimulus側の特殊系）

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
- assertionの合成: `Not` / `And` / `Or`
- 汎用Health/Readiness check（TCP/HTTP。Environment Managerとの境界領域）

## 実装優先順位（案）

このカタログ自体には優先順位を持たせない。実装順の検討はここではなく
`issues/` 側で行う。参考として、design doc（`core-design-direction.md`
9章）で決まっているMVPスコープは HTTP・Process・MySQL・Kafka。RDBは
開発者が使い慣れているMySQLをPostgreSQLより先に着手する方針。

## 未決事項

- Redis / Generic Queue / RPC / Object Storage / Inbound Trigger の
  着手順序
- `Eventually`等の横断的基盤を実装上どのレイヤーに置くか（design doc
  5章のCore architectureとの対応関係）
- Generic Queueとして最初にサポートする具体的な製品（SQS/RabbitMQ等）
