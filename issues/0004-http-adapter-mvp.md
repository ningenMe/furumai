---
status: open
created: 2026-08-11
related: docs/core-design-direction.md, issues/0003-given-when-then-dsl-mvp.md
---

# 0004: 最初のadapter実装（HTTP）

`docs/core-design-direction.md`（5章: Core architecture の Stimulus/
Observation Adapters、9章: MVPスコープ）に基づき、最初のStimulus/
Observation adapterとしてHTTPを実装する。

## やること

- Stimulus adapter: HTTP requestを送信できるようにする
- Observation adapter: HTTP responseを検証（ステータスコード、body等）
  できるようにする
- 0003で実装したDSLから、これらのadapterを `when`/`then` ステップとして
  呼び出せるようにする

## 制約

- HTTP以外のプロトコル（DB、Kafka等）は対象外
- Environment Managerによる周辺基盤の起動は対象外（対象アプリ・
  テスト対象サーバーは既に起動している前提）
- Plugin architecture（adapterの動的ロード）は対象外。組み込みの
  adapterとして実装する

## Definition of Done

- HTTP requestを送信し、responseのステータスコード/bodyを検証する
  サンプルテストが実際に実行でき、期待通りにpass/failが判定される
