---
status: done
created: 2026-08-14
related: docs/adapter-capability-catalog.md, issues/0004-http-adapter-mvp.md, issues/0009-mysql-adapter-mvp.md, issues/0010-process-adapter-mvp.md
---

# 0011: Kafka adapter実装

`docs/adapter-capability-catalog.md`のMQ（Kafka）セクション、9章の
MVPスコープ（HTTP, Process, MySQL, Kafka）に基づき、最後の1つとして
Kafka adapterを実装する。

## やること

- Stimulus adapter: `Publish(topic, key, value, headers, opts)` /
  `PublishJSON(topic, key, valueStruct)`
- Observation: 指定window内でtopicから受信した`[]Message`を取得し、
  `furumai.Diff`/`furumai.ThenEqual`でそのまま構造比較できるように
  する（Key/Value/Headersはmatcher埋め込みが効くよう`any`型で持つ）。
  Kafkaはpartition間の順序を保証しないため、`furumai.AnyOrder()`の
  利用を想定する
- given/whenはStimulus adapterを共用する前提通り、`furumai.When`から
  同じKafka adapterを呼び出せることを確認する

## 制約

- 外部ライブラリ依存の最小化方針により、driverは`database/sql`が
  無いプロトコルのため明示的な例外として`segmentio/kafka-go`
  （pure Go、cgo無し）のみ許容する
- Generic Queue（SQS/RabbitMQ等）は対象外
- Environment Manager（Kafkaクラスタそのものの起動）は対象外。
  対象クラスタは既に起動している前提
- ローカル環境（このrunner）にはKafka/Dockerが無いため、ローカルでの
  実クラスタ接続確認は行わない。確認はCIのサービスコンテナに寄せる

## Definition of Done

- topicへpublishし、受信した`[]Message`を`furumai.ThenEqual`で検証
  するサンプルテストがCI上で実際に実行でき、期待通りにpass/failが
  判定される

## 実装メモ

- `adapter/kafka`に`Stimulus`（`Publish`/`PublishJSON`）と
  `Subscription`（`Listen`で開始、`Collect(window)`でフルステートの
  `[]Message`を取得、`Close`）を実装。
- `Listen`は`when`より前（`given`直後など）に呼ぶ設計。Kafkaの
  Readerをconsumer group（呼び出しごとに一意なGroupID）・
  `StartOffset: LastOffset`で開くことで、「呼び出し時点から先に届いた
  メッセージだけを観測する」を実現している。offsetを解決するのは
  Reader作成時点なので、`when`より前に`Listen`しないと発行済み
  メッセージを取りこぼす点に注意（examples/kafka_test.goで実演）。
- 依存は`segmentio/kafka-go`＋indirectで`klauspost/compress`/
  `pierrec/lz4/v4`の3パッケージのみ。
- ヘッダ変換（`headersToMap`）などDB接続不要なロジックはunit test
  でカバー。実クラスタ込みの`examples/kafka_test.go`は
  `KAFKA_BROKERS`が無ければ`t.Skip`する統合テスト。CIに
  `bitnami/kafka`（KRaftモード、単一コンテナ、Zookeeper不要）の
  サービスコンテナを追加した。
- 注意: CI用のKafkaサービスコンテナ設定は未検証（このrunnerに
  Docker/Kafkaが無いため）。healthcheckのタイミング等、CI実行結果を
  見て調整が必要になる可能性がある。またissue #0009（MySQL）のPRも
  同じ`.github/workflows/ci.yml`の`services:`ブロックを編集している
  ため、両方をマージする際は`services:`エントリを手動で統合する必要が
  ある。
