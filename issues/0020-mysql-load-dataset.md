---
status: open
created: 2026-08-27
related: issues/0009-mysql-adapter-mvp.md
---

# 0020: MySQL用の静的fixture読み込み（`LoadDataSet`）

`adapter/mysql`に、dbunit的な「seed/期待値を静的なfixtureファイルと
して持つ」ためのIFを追加する。issue #9（MySQL adapter実装）のレビュー
過程で出てきた要望を、実装より先にIFとして固める計画issue。

## 経緯

最初に`LoadCSV(path string) ([]Row, error)`を実装したが、レビューで
以下の指摘を受けた。

- メソッド名にファイル形式（CSV）が入っているのは具体が出すぎ
- 返り値が生の`[]Row`で、既存の`DataSet`という「固有の値オブジェクト」
  になっていない

これを受けて`LoadDataSet(paths ...string) (DataSet, error)`という
シグネチャに設計し直した（ファイル名から拡張子を除いた部分をテーブル名
とみなし、複数ファイル＝複数テーブルにそのまま対応できる）。「中身の
実装より先にIFを整えよう」という方針のもと、このissueの対象PRでは
シグネチャのみ追加し、内部実装は`errors.New(...)`を返すstubのまま
にしてある。

## やること

- `LoadDataSet`の中身（実際のファイル読み込み・パース）を実装する
- ファイル形式は当面CSVを想定（ヘッダ行+データ行）だが、関数名・
  シグネチャには出さない
- セル値の型変換（DB上のINT列等と比較できるようにする）をどう
  行うかを決める（前回はint64→float64→stringの順の簡易型推定
  だったが、この方式で良いか含めて再検討）
- `examples/mysql_test.go`を`LoadDataSet`を使う形に戻す
- DB接続不要なロジック（パース・型変換）はunit testでカバーする

## 制約

- `DataSet`/`Row`は`adapter/mysql`固有の型のまま（`adapter/postgres`
  等との共通化はしない、既存方針を踏襲）

## Definition of Done

- `LoadDataSet`が実際にファイルを読み込み、`DataSet`を返す
- `examples/mysql_test.go`が`LoadDataSet`経由のfixtureでpass/failを
  正しく判定する
