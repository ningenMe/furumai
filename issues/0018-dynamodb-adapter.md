---
status: done
created: 2026-08-15
related: docs/adapter-capability-catalog.md, issues/0016-cassandra-adapter.md, issues/0017-s3-adapter.md
---

# 0018: DynamoDB adapter実装

`docs/adapter-capability-catalog.md`にNoSQL（DynamoDB, Cassandra）
セクションを追加した上で、そのDynamoDB部分を実装する（カタログ追加は
issue #16でも行っているため、どちらか先にマージされた側でこの追記は
重複／要手動統合になる）。

## やること

- Stimulus adapter: `PutItem(table, item)` / `DeleteItem(table, key)`
- Observation: `GetItem(table, key)`で単一`Item`、`Scan(table)`で
  `[]Item`（`Item = map[string]any`）を取得する

## 制約

- SigV4署名等、AWSのAPI認証を自前実装するのは非現実的なため、
  `aws-sdk-go-v2`（`config`+`service/dynamodb`+
  `feature/dynamodb/attributevalue`）をDB driver/Kafka clientと同格の
  必須例外として許容する
- 非存在の表現は`adapter/s3`の`GetObject`と同じ方針
  （`GetItem`が`(nil, nil)`を返し、呼び出し側がプレーンなnilチェック
  をする）
- ローカル環境にはDynamoDB互換ストア/Dockerが無いため、ローカルでの
  実接続確認は行わない。確認はCIのLocalStackサービスコンテナに寄せる

## Definition of Done

- `PutItem`でitemに刺激を与え、`GetItem`で取得した`Item`を
  `furumai.ThenEqual`で検証するサンプルテストがCI上で実際に実行でき、
  期待通りにpass/failが判定される

## 実装メモ

- `adapter/dynamodb`に`Stimulus`（`PutItem`/`DeleteItem`/`GetItem`/
  `Scan`）を実装。`attributevalue.MarshalMap`/`UnmarshalMap`で
  `map[string]any`とDynamoDBのAttributeValueを変換している。
- `adapter/s3`と依存の大半（`aws-sdk-go-v2`本体・`config`・認証周り）
  を共有するが、`service/dynamodb`固有の追加でindirect含め15
  パッケージ。
- `examples/dynamodb_test.go`は`DYNAMODB_ENDPOINT`が無ければ
  `t.Skip`する統合テスト。CIに`localstack/localstack`
  （`SERVICES: dynamodb`）のサービスコンテナを追加し、テーブル
  （`users`, partition key `id`: N）を事前作成するステップを挟んで
  いる。
- 注意: DynamoDBの数値属性は`UnmarshalMap`で`any`に読み込むと
  `float64`になる想定でexampleを書いたが、このrunnerに実接続環境が
  無く未検証。
- 注意: 他adapterのPR（MySQL #13、Kafka #15、Redis #16、PostgreSQL
  #17、Cassandra #20、S3 #21）も同じ`.github/workflows/ci.yml`の
  `services:`ブロック、およびCassandra PRは同じ
  `docs/adapter-capability-catalog.md`のNoSQLセクションを編集して
  いるため、複数マージする際は手動統合が必要。
