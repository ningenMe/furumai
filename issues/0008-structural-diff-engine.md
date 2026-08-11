---
status: done
created: 2026-08-11
related: docs/adapter-capability-catalog.md, issues/0007-assertion-model.md
---

# 0008: 構造diffエンジンとmatcher primitiveの実装

`docs/adapter-capability-catalog.md`のAssertionモデル（期待フル
ステート + オンメモリ突合）を実現する共通基盤を実装する。特定の
adapter（HTTP/DB/Kafka等）には依存しない、汎用的な構造比較の土台。

## やること

- 任意のGoの値（struct/slice/map等）同士を構造的に比較し、差分を
  一括して報告する diff関数を実装する（外部ライブラリ依存最小化の
  方針により、まず`reflect`ベースの独自実装を検討する）
- 期待値の中に埋め込めるmatcher primitiveを実装する:
  `Any()` / `Regex(pattern)` / `Within(min, max)` / `Ignore()` /
  `AnyOrder()`
- `furumai.Then`（既存の`dsl.go`）から使いやすい形にする（例:
  `furumai.ThenEqual(t, got, want)`のような形）
- 単体テストで、一致/不一致（複数箇所の差分を一括報告できること）/
  各matcherの動作を検証する

## 制約

- 特定プロトコル（HTTP/DB/Kafka等）向けの実装はしない（別タスク）
- `Eventually`（非同期待ちのポーリング）はこのタスクのスコープ外

## Definition of Done

- 構造diff関数とmatcher primitiveが実装され、`go test`で動作確認
  できる
- 差分がある場合、複数の不一致箇所がまとめて報告される
- 各matcherについてサンプル/テストがある

## 実装メモ

- `Diff(got, want any) []string` / `ThenEqual(t, got, want)` を
  `assert.go`（root packageの`furumai`）に実装。`reflect`のみで構成し
  外部ライブラリには依存していない。
- matcherを構造体フィールドに埋め込むには、そのフィールドの静的型が
  `any`である必要がある（Goの型システム上の制約）。ドキュメント・
  テストにこの制約を明記した。
- 作業中に`go build ./...`で `cmd/furumai/main.go` の初期化循環バグ
  （`commands`パッケージ変数 → `runHelp` → `printUsage` →
  `commands`）を発見した。PR #5マージ時にCIが実際に落ちていた可能性が
  高い（`gh`のトークン権限の都合でActionsの結果を確認できていなかった
  ため気づけなかった）。`commands`を関数`commands()`に変更して解消
  した。
