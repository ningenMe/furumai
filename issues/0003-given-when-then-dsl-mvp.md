---
status: open
created: 2026-08-11
related: docs/core-design-direction.md, issues/0002-go-project-scaffold.md
---

# 0003: given/when/then DSLの最小実装

`docs/core-design-direction.md`（3章: テスト記述方式、5章: Core
architecture の Test Definition Layer）に基づき、`given`/`when`/`then`
を表現する内部DSLの最小実装を行う。

## やること

- Goのbuilder API/関数リテラルで `given`/`when`/`then` ステップを
  記述できるDSLを実装する
- テストは1つずつ逐次実行できればよい（並列実行スケジューラは対象外）
- ステップの中身はスタブ（no-op、または固定の成功/失敗を返す関数）で
  よい。実際のHTTP等のadapterは0004で扱う
- Goのtable-driven testパターンでparameterized testが書けることを
  サンプルで示す
- 実行結果（成功/失敗）をconsoleに表示する

## 制約

- Stimulus/Observation adapterの実装はしない（0004以降）
- 並列/直列実行スケジューラの実装はしない（別タスク）
- Environment Managerは対象外

## Definition of Done

- サンプルテストが `given`/`when`/`then` ブロックで記述できる
- 該当コマンド（例: `furumai test`）でテストが実行され、成功/失敗が
  consoleに表示される
- parameterized testのサンプルが動作する
