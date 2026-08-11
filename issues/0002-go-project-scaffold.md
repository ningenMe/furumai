---
status: done
created: 2026-08-11
related: docs/core-design-direction.md
---

# 0002: Go project scaffold

`docs/core-design-direction.md` の推奨（4章: 技術スタック、5章: Core
architecture）に基づき、Furumaiの実装をGoで開始するための最小限の
プロジェクト雛形を作る。

## やること

- Go moduleを初期化する（`go.mod`）
- CLIのエントリポイント（`cmd/furumai`）を作り、`go build ./...` が
  通り、バイナリが動作することを確認する
- `.gitignore` にGoのビルド成果物を追加する

## 制約

- given/when/then DSLやadapterの実装はしない（0003, 0004で扱う）
- 実行エンジン・環境管理などのレイヤーを空パッケージとして先回りして
  作らない。実際にコードが必要になった時点（0003以降）で追加する

## Definition of Done

- `go build ./...` が成功する
- `furumai` バイナリを実行すると、動作確認できる最小限の出力
  （バージョン表示等）が得られる
