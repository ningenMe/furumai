---
status: done
created: 2026-08-14
related: docs/adapter-capability-catalog.md, issues/0004-http-adapter-mvp.md, issues/0009-mysql-adapter-mvp.md
---

# 0010: Process (shell command) adapter実装

`docs/adapter-capability-catalog.md`のProcessセクション、9章のMVP
スコープ（HTTP, Process, MySQL, Kafka）に基づき、shell commandの
Stimulus/Observation adapterを実装する。

## やること

- Stimulus adapter: `Run(cmd, args, opts)`。opts: env, cwd, stdin,
  timeout
- Observation: `Result{ExitCode, Stdout, Stderr}`というフルステートを
  1つ返す。`furumai.Diff`/`furumai.ThenEqual`でそのまま構造比較できる
  ようにする（Stdout/Stderrはmatcher埋め込みが効くよう`any`型で持つ）

## 制約

- 外部ライブラリに依存しない（`os/exec`等、標準ライブラリのみ）
- shellのquoting/escapingのヘルパーは対象外（argsはそのまま
  `exec.Command`に渡す）

## Definition of Done

- コマンドを実行し、`Result{ExitCode, Stdout, Stderr}`を
  `furumai.ThenEqual`で検証するサンプルテストが実際に実行でき、
  期待通りにpass/failが判定される

## 実装メモ

- `adapter/process`に`Stimulus`（`Run`、`WithEnv`/`WithDir`/
  `WithStdin`/`WithTimeout`オプション）と`Result{ExitCode, Stdout,
  Stderr}`を実装。`os/exec`のみに依存。
- 非ゼロ終了コードは`Result.ExitCode`で表現し、Goの`error`にはしない
  （コマンド自体を起動できなかった場合のみ`error`を返す）。
- 外部サービス不要のため、`adapter/rest`/`adapter/mysql`と違い
  DB接続前提の統合テストのようなgatingは無く、通常のunit test/
  exampleとしてローカル・CIどちらでもそのまま実行できる。
