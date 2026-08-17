---
status: done
created: 2026-08-15
related: docs/adapter-capability-catalog.md, issues/0009-mysql-adapter-mvp.md
---

# 0016: Cassandra adapter実装

`docs/adapter-capability-catalog.md`にNoSQL（DynamoDB, Cassandra）
セクションを追加した上で、そのCassandra部分を実装する。

## やること

- カタログにNoSQL大分類とCassandra/DynamoDBのcapability記述を追加
  （この2つは互いに共通capabilityを持たないことを明記する）
- Stimulus adapter: `Exec(cql, args...)`
- Observation: `Snapshot(tables ...string) (DataSet, error)`で
  テーブルごとの`[]Row`（`Row = map[string]any`）を返す。
  `adapter/mysql`/`adapter/postgres`と同じ、dbunit準拠の
  full-state-forcing interface（フィルタ無し、常に`SELECT *`）

## 制約

- CassandraのCQL binary protocolを自前実装するのは非現実的なため、
  `gocql/gocql`をDB driver例外として許容する
- Seed/Truncateのような専用ヘルパーは持たない（catalogのCassandra部分
  にも記載が無く、`Exec`で任意CQLを実行すれば足りるため）
- ローカル環境にはCassandra/Dockerが無いため、ローカルでの実クラスタ
  接続確認は行わない。確認はCIのサービスコンテナに寄せる

## Definition of Done

- CQLでrowに刺激を与え、`Snapshot`で取得した`DataSet`を
  `furumai.ThenEqual`で検証するサンプルテストがCI上で実際に実行でき、
  期待通りにpass/failが判定される

## 実装メモ

- `adapter/cassandra`に`Stimulus`（`Exec`/`Snapshot`）、`Row`、
  `DataSet`を実装。gocqlの`MapScan`が列を素のGo型（`[]byte`ではなく）
  で返すため、`adapter/mysql`/`adapter/postgres`のような`normalize`は
  不要だった。
- MySQL側(#13)のレビューで「生SQLを取る`Query`だと部分的なSELECTでも
  通ってしまい、フルステート比較を強制できていない」と指摘があり、
  `Query`を`Snapshot`に置き換えた。さらに、namespace/filter用に
  追加した`TableQuery.Where`（生CQL文字列）も「1行だけ選んで他の行を
  隠す」という同種の抜け道になると指摘があり、`Where`/`Args`ごと
  削除した。最終的に`Snapshot(tables ...string) (DataSet, error)`は
  常に無条件で`SELECT * FROM <table>`を発行し、カラム・行のどちらも
  選べない。テストごとのデータ分離はテスト独立性（`Exec`での
  `TRUNCATE`等）で担保する方針にした。
- 依存は`gocql/gocql`＋indirectで`golang/snappy`/
  `hailocab/go-hostpool`/`gopkg.in/inf.v0`の計4パッケージ。
- Cassandraのブートに時間がかかる（KeyspaceのGossip/schema
  propagation含む）ため、CIの`services: cassandra`は
  `--health-start-period=60s`と余裕を持たせた。`examples/
  cassandra_test.go`は`CASSANDRA_HOSTS`が無ければ`t.Skip`する統合
  テストで、keyspace/table自体もテスト内で`CREATE ... IF NOT EXISTS`
  している。
- 注意: このCIサービスコンテナ設定は未検証（このrunnerにCassandra/
  Dockerが無いため）。特にhealthcheckのタイミングは調整が必要になる
  可能性が高い。
- 注意: 他adapterのPR（MySQL #13、Kafka #15、Redis #16、PostgreSQL
  #17）も同じ`.github/workflows/ci.yml`の`services:`ブロックを
  編集しているため、複数マージする際は`services:`エントリの手動統合が
  必要。
