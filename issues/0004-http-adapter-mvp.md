---
status: done
created: 2026-08-11
related: docs/core-design-direction.md, issues/0003-given-when-then-dsl-mvp.md, issues/0008-structural-diff-engine.md
---

# 0004: 最初のadapter実装（HTTP）

`docs/core-design-direction.md`（5章: Core architecture の Stimulus/
Observation Adapters、9章: MVPスコープ）に基づき、最初のStimulus/
Observation adapterとしてHTTPを実装する。

## やること

- Stimulus adapter: `docs/adapter-capability-catalog.md`のHTTPセクション
  通り、`Get`/`Post`/`Put`/`Patch`/`Delete`/`Head`/`Options`でrequestを
  送信できるようにする（MVPではheader/query paramのオプションのみ。
  認証ヘルパー等は対象外）
- Observation: `Response{StatusCode, Headers, Body}`というフルステートを
  1つ返す。0008で実装した`furumai.Diff`/`furumai.ThenEqual`でそのまま
  構造比較できるようにする（Headers/Bodyはmatcher埋め込みが効くよう
  `any`型で持つ）
- given/whenはStimulus adapterを共用する前提通り、`furumai.Given`/
  `furumai.When`から同じHTTP adapterを呼び出せることを確認する

## 制約

- HTTP以外のプロトコル（DB、Kafka等）は対象外
- 外部ライブラリに依存しない（`net/http`等、標準ライブラリのみ）
- Environment Managerによる周辺基盤の起動は対象外（対象アプリ・
  テスト対象サーバーは既に起動している前提）
- Plugin architecture（adapterの動的ロード）は対象外。組み込みの
  adapterとして実装する

## Definition of Done

- HTTP requestを送信し、`Response{StatusCode, Headers, Body}`を
  `furumai.ThenEqual`で検証するサンプルテストが実際に実行でき、
  期待通りにpass/failが判定される

## 実装メモ

- `http.go`に`HTTPStimulus`（`Get`/`Post`/`Put`/`Patch`/`Delete`/`Head`/
  `Options`、`WithHeader`/`WithQuery`オプション）と`Response{StatusCode,
  Headers, Body}`を実装。`net/http`のみに依存。
- `Response`のHeaders/Bodyは`any`型で持たせ、`assert.go`のmatcher
  （`Any`/`Ignore`等）をそのまま埋め込めるようにした。
- `examples/http_test.go`に`httptest.Server`を使ったサンプルを追加
  （`net/http/httptest`は標準ライブラリなので依存最小化方針に反しない）。
