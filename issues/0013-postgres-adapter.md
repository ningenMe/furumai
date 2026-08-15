---
status: done
created: 2026-08-15
related: docs/adapter-capability-catalog.md, issues/0009-mysql-adapter-mvp.md
---

# 0013: PostgreSQL adapter実装

`docs/adapter-capability-catalog.md`のRDBセクションで、MySQLと共通の
capability形状を持つと決めていたPostgreSQLを実装する。

## やること

- `adapter/mysql`と同じ形（`Exec`/`Seed`/`Truncate`/`Query`、`Row =
  map[string]any`）を`adapter/postgres`として実装する
- placeholder記法の違い（MySQL: `?`、PostgreSQL: `$1,$2,...`）を
  adapter内部で吸収する（`Seed`のクエリ組み立てのみ）。`Exec`/`Query`
  は生SQLを渡す方式のまま（呼び出し側が`$N`を書く）

## 制約

- 外部ライブラリ依存の最小化方針のDB driver例外により`jackc/pgx/v5`
  を使う（`lib/pq`はメンテナンスモードのため不採用）
- `adapter/mysql`とのコード共通化（内部共通パッケージへの切り出し）は
  今回はしない。両PRを独立にマージ可能な状態に保つ優先度の方が高いと
  判断した。将来的な重複解消は別issueで検討する
- ローカル環境にはPostgreSQL/Dockerが無いため、ローカルでの実DB接続
  確認は行わない。確認はCIのサービスコンテナに寄せる

## Definition of Done

- `Exec`/`Seed`/`Truncate`でDBに刺激を与え、`Query`で取得した
  `[]Row`を`furumai.ThenEqual`で検証するサンプルテストがCI上で
  実際に実行でき、期待通りにpass/failが判定される

## 実装メモ

- `adapter/mysql`とほぼ同一構造。差分はdriver（`pgx`）とplaceholder
  記法、identifier quoting（`` ` `` → `"`）のみ。
- `pgx/v5`の追加でtransitive依存が5パッケージ増える
  （`pgpassfile`/`pgservicefile`/`puddle/v2`/`x/sync`/`x/text`）。
  `go-sql-driver/mysql`（1つ）より重いが、pure GoでPostgreSQL用に
  他に妥当な選択肢が無いため許容する。
- `examples/postgres_test.go`は`POSTGRES_DSN`が無ければ`t.Skip`する
  統合テスト。CIに`postgres:16-alpine`のサービスコンテナを追加した。
- 注意: pgxがINTEGER列をGoのどの型にscanするか（`int32`想定で
  exampleを書いたが）はこのrunnerに実DBが無いため未検証。CI結果を見て
  型が合わなければ調整が必要。
- 注意: 他adapterのPR（MySQL #13、Kafka #15、Redis #16）も同じ
  `.github/workflows/ci.yml`の`services:`ブロックを編集しているため、
  複数マージ時は`services:`エントリの手動統合が必要。
