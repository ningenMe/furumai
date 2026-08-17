---
status: done
created: 2026-08-14
related: docs/adapter-capability-catalog.md, issues/0004-http-adapter-mvp.md, issues/0008-structural-diff-engine.md
---

# 0009: MySQL adapter実装

`docs/adapter-capability-catalog.md`のRDBセクション、9章のMVPスコープ
（HTTP, Process, MySQL, Kafka）に基づき、`adapter/rest`に続く2つ目の
adapterとしてMySQLを実装する。

## やること

- Stimulus adapter: `Exec(query, args...)` / `Seed(table, rows...)` /
  `Truncate(tables...)`
- Observation: `Snapshot(queries ...TableQuery) (DataSet, error)` で
  `DataSet`（テーブル名→`[]Row`のフルステート）を取得し、
  `furumai.Diff`/`furumai.ThenEqual`でそのまま構造比較できるように
  する。dbunitの`IDataSet`に準え、`SELECT *`を強制することで
  「部分的なSELECTで一部カラムだけ検証してしまう」抜け道を型/API上
  無くす（生SQLを取る`Query`のような汎用メソッドは公開しない）
- クエリ組み立て・値の正規化などDB接続不要なロジックは通常のunit test
  でカバーする
- 実際のDB接続を要するテストは、環境変数（`MYSQL_DSN`）が無ければ
  `t.Skip`する統合テストとして書く
- CI（`.github/workflows/ci.yml`）にMySQLのサービスコンテナを追加し、
  上記の統合テストを実行する

## 制約

- 外部ライブラリ依存の最小化方針により、driverは`database/sql`+
  pure Goかつ低依存な`go-sql-driver/mysql`のみ許容する（DBドライバは
  標準ライブラリでは代替できないため明示的な例外）
- PostgreSQLは対象外（優先順位通りMySQL先行）
- Environment Manager（MySQLコンテナの起動そのもの）は対象外。
  対象DBは既に起動している前提
- ローカル環境（このrunner）にはMySQL/Dockerが無いため、ローカルでの
  実DB接続確認は行わない。確認はCIのサービスコンテナに寄せる

## Definition of Done

- `Exec`/`Seed`/`Truncate`でDBに刺激を与え、`Snapshot`で取得した
  `DataSet`を`furumai.ThenEqual`で検証するサンプルテストがCI上で
  実際に実行でき、期待通りにpass/failが判定される

## 実装メモ

- `adapter/mysql`に`Stimulus`（`Exec`/`Seed`/`Truncate`/`Snapshot`）、
  `Row`（`map[string]any`）、`TableQuery{Table, Where, Args}`、
  `DataSet`（`map[string]any`、値は通常`[]Row`だが`furumai.AnyOrder`
  でラップできるよう`any`型）を実装。driverは`go-sql-driver/mysql`
  （pure Go、transitive依存は`filippo.io/edwards25519`1つのみ）。
- 当初`Query(query, args...) ([]Row, error)`という生SQLを取る
  Observationメソッドだったが、レビューで「部分的なSELECT（一部
  カラムだけ）でも構文上は通ってしまい、フルステート比較を型/API上
  強制できていない」との指摘を受け、`Snapshot`/`TableQuery`/
  `DataSet`に置き換えた。`Snapshot`は内部で常に`SELECT *`を発行する
  ため、呼び出し側がカラムを選んで一部だけ検証する抜け道が無い
  （dbunitの`IDataSet`/`QueryDataSet`に相当する設計）。
- `Row`はmap値が`any`型なので、matcherをそのまま埋め込める
  （`adapter/rest`のHeaders/Bodyと同じ設計）。`DataSet`も同様に値を
  `any`型にして、テーブル単位で`furumai.AnyOrder`をかけられるように
  した。
- driverが返す`[]byte`（VARCHAR/DECIMAL等）は`string`に正規化してから
  `Row`に詰めている（`normalize`）。これはDB接続不要なので通常の
  unit testでカバーした。
- `examples/mysql_test.go`は`MYSQL_DSN`環境変数が無ければ`t.Skip`する
  統合テスト。ローカルのrunnerにはMySQL/Dockerが無いため、動作確認は
  CIの`services: mysql`サービスコンテナに寄せた（ローカルへの
  mysql-server直接installはしない方針）。
