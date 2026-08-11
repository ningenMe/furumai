---
status: done
created: 2026-08-11
related: docs/core-design-direction.md, issues/0004-http-adapter-mvp.md
---

# 0006: Adapter capability catalogの作成

個別adapterを実装する前に、Furumaiが最終的にどこまでのStimulus/
Observationをカバーすべきかを一望できる資料を作る。実装の優先順位や
着手有無は問わず、まず網羅性のある構造を可視化することが目的（実装は
このタスクのスコープ外）。

## やること

- 対象システム（HTTP、DB、Kafka、queue、shellなど）を分類し、優先順位
  （Tier）を付ける
- DBは関係DB（Postgres/MySQL）とKVS（Redis）で操作の形が異なるため、
  カテゴリを分けて整理する
- 各システムについて、Stimulus（given/when）とObservation（then）の
  メソッドをメソッド粒度で網羅的に洗い出す
- HTTP/DB/Kafka等を横断する共通基盤（Assertion API、非同期処理を待つ
  ポーリング型assertion等）を別立てで整理する
- `docs/` 配下に一覧性のあるカタログ資料として作成する

## 制約

- 実装はしない（コード変更なし）。あくまで構造を可視化する資料作成
- Tier分け・優先順位は提案ベースとし、確定はユーザーとの合意による

## Definition of Done

- `docs/` 配下に、対象システム × Stimulus/Observationメソッドの
  網羅的なカタログ資料が存在する
- 優先順位（Tier）が明記されている
