---
status: done
created: 2026-08-15
related: docs/adapter-capability-catalog.md, issues/0004-http-adapter-mvp.md
---

# 0014: GraphQL adapter実装

`docs/adapter-capability-catalog.md`のRPCセクション（GraphQL部分）に
基づき、GraphQL adapterを実装する。

## やること

- Stimulus adapter: `Execute(query, variables, opts)`でquery/mutationを
  実行する
- Observation: `Response{Data, Errors}`というフルステートを1つ返す。
  `furumai.Diff`/`furumai.ThenEqual`でそのまま構造比較できるように
  する（Data/Errorsはmatcher埋め込みが効くよう`any`型で持つ）

## 制約

- 外部ライブラリに依存しない。GraphQL over HTTPは「JSONをPOSTして
  `{data, errors}`のJSONを受け取る」だけなので、`net/http`/
  `encoding/json`のみで実現できると判断した（`adapter/rest`と同様の
  判断）

## Definition of Done

- queryを実行し、`Response{Data, Errors}`を`furumai.ThenEqual`で
  検証するサンプルテストが実際に実行でき、期待通りにpass/failが
  判定される

## 実装メモ

- `adapter/graphql`に`Stimulus`（`Execute`、`WithHeader`オプション）
  と`Response{Data, Errors}`を実装。`net/http`/`encoding/json`のみに
  依存。
- 外部サービス不要（`httptest.Server`で完結）のため、`adapter/rest`/
  `adapter/process`と同様、unit test・exampleともにgatingなしで
  ローカル・CIどちらでもそのまま実行できる。
