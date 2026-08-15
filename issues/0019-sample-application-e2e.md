---
status: open
created: 2026-08-15
related: docs/core-design-direction.md, issues/0004-http-adapter-mvp.md
---

# 0019: サンプルアプリケーションによるE2E品質保証をCIに組み込む

furumai自体の品質保証として、「実際のアプリケーションをfurumaiで
テストしている」挙動そのものをCIに組み込む。サンプル用アプリケー
ションをまず作る。

## 背景・動機

現状の`examples/*_test.go`はadapter単体の動作確認が主目的で、
`httptest.Server`や手書きのgRPCサーバー、あるいは外部サービス
ゲート型（`MYSQL_DSN`等が無ければ`t.Skip`）の統合テストになって
いる。これらはadapterのプロトコル実装が正しいかの検証にはなって
いるが、「複数adapterをまたいだ、現実的なアプリケーションの振る舞い
をfurumaiのgiven/when/thenで検証する」というfurumai本来の使われ方
そのものは、まだCIで検証できていない。

これはfurumai自身の一番の品質保証（dogfooding）になるはずで、
core-design-direction.mdの設計判断（DSLの表現力、Stimulus/Observation
adapterの組み合わせやすさ、フルステート+matcherのassertion model）が
実際に「使える」ことを継続的に確認する意味を持つ。

## やること（このissueの範囲＝計画のみ）

- サンプルアプリケーションの技術構成を決める（下記「未決事項」参照）
- サンプルアプリケーションをリポジトリのどこに置くか決める
- サンプルアプリケーションに対するfurumaiのgiven/when/thenテスト
  （複数adapterをまたぐ現実的なシナリオ）を書く
- CIに、サンプルアプリケーションを起動した上でそのテストを実行する
  ジョブ/ステップを追加する
- `examples/`（adapter単体の動作確認）とサンプルアプリ
  （furumai全体のE2E/dogfooding）の役割分担をREADME等に明記する

このissue自体では実装はしない。方向性を固めるための計画issue。

## 未決事項

- サンプルアプリの技術構成: 最小構成（HTTPのみ）にするか、複数
  adapterをまたぐ構成（HTTP + MySQL、+Kafkaでイベント発行、等）に
  するか。後者の方がfurumaiの「複数adapterの組み合わせ」という価値を
  検証できるが、CI起動コストは上がる
- サンプルアプリの配置場所: `examples/app/`のようにリポジトリ内に
  置くか、別ディレクトリ・別Goモジュールに分離するか
- サンプルアプリの起動方法: CI上で単にGoプロセスとしてバックグラウンド
  起動するだけで足りるか、Docker Composeのような構成が必要か
- 既存の`.github/workflows/ci.yml`の`services:`ブロック（MySQL/Kafka/
  Redis/PostgreSQL/Cassandra/LocalStack、各adapterのPRで個別に追加中）
  とどう統合するか
- サンプルアプリ自体のメンテナンス（adapterのAPIが変わった際に追従が
  必要になる）をどう位置付けるか
