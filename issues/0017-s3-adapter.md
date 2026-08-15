---
status: done
created: 2026-08-15
related: docs/adapter-capability-catalog.md
---

# 0017: S3 (Object Storage) adapter実装

`docs/adapter-capability-catalog.md`のObject Storageセクションに
基づき、S3 adapterを実装する。

## やること

- Stimulus adapter: `PutObject(key, content, opts)`
- Observation: `Object{Key, Content, Metadata}`（または非存在）を
  取得する`GetObject(key)`

## 制約

- SigV4署名等、AWSのAPI認証を自前実装するのは非現実的なため、
  `aws-sdk-go-v2`（`config`+`service/s3`）をDB driver/Kafka clientと
  同格の必須例外として許容する
- 非存在の表現は`furumai.ThenEqual`の構造比較には乗せず、
  `GetObject`が`(nil, nil)`を返す形にして呼び出し側にプレーンなnil
  チェックをしてもらう（`*Object`のnil比較を`Diff`の型付き/無型nilの
  差異に巻き込みたくないため）
- ローカル環境にはS3互換ストレージ/Dockerが無いため、ローカルでの
  実接続確認は行わない。確認はCIのLocalStackサービスコンテナに寄せる

## Definition of Done

- `PutObject`でオブジェクトに刺激を与え、`GetObject`で取得した
  `Object`を`furumai.ThenEqual`で検証するサンプルテストがCI上で
  実際に実行でき、期待通りにpass/failが判定される

## 実装メモ

- `adapter/s3`に`Stimulus`（`PutObject`、`WithMetadata`オプション、
  `GetObject`）を実装。
- `aws-sdk-go-v2`はモジュールが細かく分かれているため、`go mod
  tidy`後で直接+indirect合わせて16パッケージになった。他adapterより
  かなり重いが、AWS API認証を自前実装するのは非現実的なため許容する
  （DynamoDB adapterでも同じ依存を共有する想定）。
- `examples/s3_test.go`は`S3_ENDPOINT`/`S3_BUCKET`が無ければ
  `t.Skip`する統合テスト。CIに`localstack/localstack`
  （`SERVICES: s3`）のサービスコンテナを追加し、`go test`実行前に
  `aws s3 mb`でバケットを作成するステップを挟んでいる
  （GitHub Actions実行環境にaws-cliが入っている前提。未検証）。
- 注意: LocalStack/aws-cliのCI設定は未検証（このrunnerにDocker/
  aws-cliが無いため）。
- 注意: 他adapterのPR（MySQL #13、Kafka #15、Redis #16、PostgreSQL
  #17、Cassandra #20）も同じ`.github/workflows/ci.yml`の`services:`
  ブロックを編集しているため、複数マージする際は`services:`エントリの
  手動統合が必要。
