---
status: draft
created: 2026-08-11
related: docs/core-design-direction.md, issues/0006-adapter-capability-catalog.md, issues/0007-assertion-model.md
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
| NoSQL | "NoSQL"と括られる製品群。ただし操作の形は製品ごとに異なり、この大分類自体は共通capabilityを持たない（下記参照） | DynamoDB, Cassandra |
| Inbound Trigger | 外部からの受動的な刺激・トリガー | Webhook受信, Cron |

HTTPとProcessはそれ自体が1システムなので大分類=システムだが、RDB・KVS・
MQ・RPCは複数の具体的な製品が同じ形のcapabilityを共有する。adapter
実装は「大分類ごとの共通interface」+「製品ごとの差分吸収層」という
2層構成を想定する。

## Assertionモデル: 期待フルステート + オンメモリ突合

`then`は個別の値を1つずつimperativeに検証していくのではなく、次の
2ステップに集約する。

1. **取得**: テストのnamespace/filterでスコープした範囲の実際の状態を、
   まるごとオンメモリの値として取得する（Observation adapterの役割は
   ここまで）
2. **突合**: 期待する状態を`given`/`when`と同じテストコード内にstatic
   な値（struct/slice/map）として書き、取得した実際の値と構造的に比較
   （diff）する。差分があれば一括して失敗として報告する

これにより、システムごとに大量のassertメソッド（`RowExists`/
`ColumnEquals`/`HeaderEquals`/`BodyJSONEquals`...）を個別に用意する
必要がなくなり、`then`のAPIは実質「取得」+「共通diffエンジン」の2つに
収束する。以降、各システムのObservationは個別assertメソッドの列挙では
なく「フルステートの形」として記述する。

### Matcher（非決定的な値への対処）

完全一致だけだと、自動採番ID・`created_at`のようなtimestamp・UUID等の
フィールドで毎回失敗してしまう。期待値の中に埋め込めるmatcher
primitiveが必要になる。

- `Any()` — 値を問わない（存在確認のみ）
- `AnyOf(candidates...)`
- `Regex(pattern)`
- `Within(min, max)` — 数値・時刻の範囲
- `Ignore()` — 比較対象から除外
- `AnyOrder()` — スライスの順序を問わない（Kafkaのpartition跨ぎ等、順序保証が無い大分類向け）

### スコープの注意

「フルステート」は文字通りシステム全体（テーブル全体、topic全体等）
ではなく、6章のデータ名前空間規約でフィルタされた、そのテストに関係
する範囲に限定する。

### Eventually（非同期な結果を待つ）

worker処理待ちなど非同期なシナリオ（issue #1で示されている「Kafka
publish → workerが処理 → HTTP API呼び出し → DB state検証」のような
流れ）では、「取得 → 突合」を単発で行うのではなく、一定時間内に突合が
成功するまでポーリングする`Eventually(timeout, interval)`が必須になる。
HTTP/RDB/MQ/KVSいずれの`then`にも横断的に必要になるため、個別adapter
ではなくAssertionモデルの共通基盤として設計する。

## HTTP

**Stimulus（given/when）**

- `Get` / `Post` / `Put` / `Patch` / `Delete` / `Head` / `Options`
- 共通オプション: header, query param, path param, timeout, redirect追従有無, cookie送信
- body: JSON, form-urlencoded, multipart/form-data, raw bytes
- 認証: Basic, Bearer token, カスタムheader

**Observation（then） — フルステートの形**

`Response{StatusCode, Headers, Body}` を1つ取得し、期待する`Response`
値と構造比較する。部分一致・型のみ確認したい場合は、Headers/Bodyの
値としてmatcher（`Contains`/`Regex`等）を埋め込む。

## Process（shell command）

**Stimulus（given/when）**

- `Run(cmd, args, opts)` — opts: env, cwd, stdin, timeout

**Observation（then） — フルステートの形**

`Result{ExitCode, Stdout, Stderr}` を1つ取得し、期待する`Result`値と
構造比較する。

## RDB（MySQL, PostgreSQL）

**Stimulus（given/when）**

- `Exec(sql, args...)` — 任意SQLの実行
- `Seed(table, rows...)` — 構造化されたseedヘルパー
- `Truncate(table...)`

**Observation（then） — フルステートの形**

フィルタ条件（テストのnamespace）にマッチする`[]Row`をまるごと取得し、
期待する`[]Row`と構造比較する。行が存在しないことを期待する場合は
空sliceを期待値にする。auto-incrementのID・timestampはmatcherで対処
する。

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

**Observation（then） — フルステートの形**

パターン（テストのnamespace）にマッチする`map[string]Value`を取得し、
期待するmapと構造比較する。TTLは`Within(min, max)`のようなmatcherで
範囲指定する。

## MQ（Kafka）

**Stimulus（given/when）**

- `Publish(topic, key, value, headers, opts: partition)`
- `PublishJSON(topic, key, valueStruct)`

**Observation（then） — フルステートの形**

指定window内でtopicから受信した`[]Message`を取得し、期待する
`[]Message`と構造比較する。メッセージが無いことを期待する場合は
空slice。Kafkaはpartition間の順序を保証しないため、`AnyOrder()`
matcherの利用を想定する。

## MQ（Generic Queue、例: SQS/RabbitMQ）

Kafkaと同じMQ大分類のcapability（Publish/Consume、フルステート
+ matcherによるObservation）を踏襲しつつ、queue特有の「消費したら
消える」「順序保証が無い場合がある」という性質から、ack/visibility
timeoutのような概念が追加で必要になる。

## RPC（gRPC, GraphQL）

**Stimulus（given/when）**

- gRPC: Unary call、Streaming call
- GraphQL: query/mutation実行

**Observation（then） — フルステートの形**

gRPCは`Response{Message, StatusCode, Trailer}`、GraphQLは
`Response{Data, Errors}`を1つ取得し、期待値と構造比較する。

## Object Storage（S3等）

**Stimulus（given/when）**

- `PutObject`

**Observation（then） — フルステートの形**

`Object{Key, Content, Metadata}`（または非存在）を取得し、期待値と
構造比較する。

## NoSQL（DynamoDB, Cassandra）

DynamoDBとCassandraは「NoSQL」と括られることが多いが、操作の形は
共通しない（DynamoDBはKVSに近いkey-value + 属性、Cassandraは
CQLによるクエリを持ちRDBに近い）。大分類としては分類ラベルとして
まとめているだけで、capability自体は個別に定義する。

### DynamoDB

**Stimulus（given/when）**

- `PutItem(table, item)`
- `DeleteItem(table, key)`

**Observation（then） — フルステートの形**

`table`内の`[]Item`（`Item = map[string]any`）をまるごと、または
`GetItem(table, key)`で単一`Item`を取得し、期待値と構造比較する。

### Cassandra

**Stimulus（given/when）**

- `Exec(cql, args...)` — 任意CQLの実行

**Observation（then） — フルステートの形**

RDBと同様、`Query(cql, args...)`で取得した`[]Row`
（`Row = map[string]any`）を期待値と構造比較する。

## Inbound Trigger（Webhook受信 / Cron）

大分類として他と異なり、「Furumai側が受動的に刺激を観測する」もしくは
「時刻を進める/手動発火する」という特殊な形を取る。

- Webhook受信: 対象システムからのwebhookをFurumai側で受け止め、受信
  した`[]Request`をフルステートとして取得・比較する（Observation側の
  特殊系）
- Cron/scheduled batch trigger: 時刻を進める/手動トリガーする
  （Stimulus側の特殊系）

## 実装優先順位（案）

このカタログ自体には優先順位を持たせない。実装順の検討はここではなく
`issues/` 側で行う。参考として、design doc（`core-design-direction.md`
9章）で決まっているMVPスコープは HTTP・Process・MySQL・Kafka。RDBは
開発者が使い慣れているMySQLをPostgreSQLより先に着手する方針。

## 未決事項

- Generic Queue / Inbound Trigger の着手順序（着手順序自体もissues/側で
  管理する）
- 構造diffエンジンの実装方式。外部ライブラリ依存最小化の方針
  （core-design-direction.md参照）から、`go-cmp`等の外部ライブラリより
  まず`reflect.DeepEqual`ベース/標準ライブラリのみでの独自実装を優先
  検討する
- matcher primitive（`Any`/`Regex`/`Within`/`Ignore`/`AnyOrder`等）の
  具体的なAPI形状
- `Eventually`を実装上どのレイヤーに置くか（design doc 5章のCore
  architectureとの対応関係）
- Generic Queueとして最初にサポートする具体的な製品（SQS/RabbitMQ等）
