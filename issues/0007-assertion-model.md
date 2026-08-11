---
status: done
created: 2026-08-11
related: docs/core-design-direction.md, docs/adapter-capability-catalog.md
---

# 0007: Assertionモデルの方針決定

issue #1で「まだ決めない」とされていたassertion APIの設計方針を決める。
実装はしない（コード変更なし）。方針をdocsに反映するところまで。

## やること

- `then`のassertionモデルとして、「静的な期待フルステートを定義し、
  実際の状態をオンメモリに取得して構造的に突合する」方式を採用する
- 非決定的な値（自動採番ID、timestamp、UUID等）を扱うためのmatcher
  primitive（`Any`/`Regex`/`Within`/`Ignore`等）が必要であることを
  明記する
- `docs/adapter-capability-catalog.md`の各システムのObservation
  セクションを、個別assertメソッド列挙から「フルステート取得 + 共通
  diffエンジン」の形に書き直す
- `docs/core-design-direction.md`のassertion API関連の記述（未決事項
  リスト等）を更新する

## 制約

- 実装はしない
- 「フルステート」の対象は、6章のデータ名前空間規約でフィルタされた
  テスト関係範囲であり、システム全体ではない

## Definition of Done

- assertionモデルの方針がdocsに明記されている
- matcher primitiveの必要性が明記されている
- adapter capability catalogがこの方針と整合している
